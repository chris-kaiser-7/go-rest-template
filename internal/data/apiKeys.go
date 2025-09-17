package data

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chris-a-kaiser-7/go-rest-template/internal/validator"
)

type ApiKey struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	UserId    int64     `json:"user_id"`
	KeyName   string    `json:"key_name"`
	KeyValue  []byte    `json:"key_value"`
}

// Wraper for sql.DB connection pool and logger
type ApiKeyDataAccess struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

const keyLength = 64 //TODO: update this to be in config

// Create will generate the apiKey value and hash update the apiKeyData with the value.
// Param apiKeyData: apiKey data to create. Should contain userId and keyName.
// Return Value is unhandled errors.
func (m ApiKeyDataAccess) Create(apiKeyData ApiKey) error {
	query := `
		INSERT INTO api_keys (user_id, key_name, key_hash) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at
		`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Generate API key from crypto/rand then hash agianst SHA512.
	unencodedKey := make([]byte, keyLength)
	_, err := rand.Read(unencodedKey)
	if err != nil {
		return err //TODO: this should handle the error / retry
	}
	hashedKey := sha512.Sum512(unencodedKey)

	// Do db query to insert the key
	args := []interface{}{apiKeyData.UserId, apiKeyData.KeyName, hashedKey}
	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&apiKeyData.ID, &apiKeyData.CreatedAt)
	if err != nil {
		return err
	}

	// Encode key in base64
	key := make([]byte, keyLength)
	base64.StdEncoding.Encode(key, unencodedKey)
	apiKeyData.KeyValue = key
	return nil
}

// Get will retrieve an api key with the provided keyValue as key by comparing key hashes
// Return Value 1, apiKey data retrieved
// Return Value is unhandled errors.
func (m ApiKeyDataAccess) Get(key []byte) (ApiKey, error) {
	query := `
		SELECT id, created_at, user_id, key_name
        	FROM api_keys
 		WHERE key_hash = $1, activated = TRUE
 		`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Take byte64 encoded key and decode it.
	unencodedKey := make([]byte, keyLength)
	base64.StdEncoding.Decode(unencodedKey, key)
	hashedKey := sha512.Sum512(unencodedKey)

	// Do db query to check if the hashedKey exists
	var apiKeyData ApiKey
	err := m.DB.QueryRowContext(ctx, query, hashedKey).Scan(
		&apiKeyData.ID,
		&apiKeyData.CreatedAt,
		&apiKeyData.UserId,
		&apiKeyData.KeyName)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows): // Scan() will return a sql.ErrNoRows if there is no match
			return ApiKey{}, ErrRecordNotFound
		default:
			return ApiKey{}, err
		}
	}

	return apiKeyData, nil
}

// Delete will set activated to be false. for supplied apikey id.
// Param id: id of the apiKey to be deactivated
// Return Value is unhandled errors.
func (m ApiKeyDataAccess) Deactivate(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		UPDATE api_keys
		SET activated = false
		WHERE id = $1 
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result
	// object to get the number of rows affected by the query.
	// If no rows were affected, then the record was not found.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetALl will return a slice of activated apiKeys and Metadata with provided optional user_id filter.
// Param user_id: optional param for filter apikeys by specific user id. -1 will should all users.
// Param filters: filter struct that provides sort and pagination information.
// Return Value 1, Slice of ApiKeys: This return value is the set of keys found with applied filters and pagination.
// Return Value 2, Metadata: This is the metadata information that includes pagination information.
// Return Value 3 is unhandled errors.
func (m ApiKeyDataAccess) GetAll(user_id int, filters Filters) ([]ApiKey, Metadata, error) {
	// Note count(*) OVER() is used for getting the total count of records returned.
	// The (user_id = $1 OR $1 = '') clause allows for an optional filter by user_id.
	// ORDER BY %s %s, id ASC interpolates the sort column and direction from
	// the filter with id being secondary sort for matches.
	// LIMIT $3 OFFSET $4 apply limit and offset from the filter.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, user_id, key_name
		FROM api_keys
		WHERE (user_id = $1 OR $1 = -1) AND (activated = TRUE)
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`,
		filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{user_id, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Defer a call to rows.Close() to ensure that the result set is closed before GetAll returns.
	defer func() {
		if err := rows.Close(); err != nil {
			m.ErrorLog.Println(err)
		}
	}()

	// Parse the data from rows
	totalRecords := 0
	apiKeys := []ApiKey{}
	for rows.Next() {
		var apiKey ApiKey
		err := rows.Scan(
			&totalRecords,
			&apiKey.ID,
			&apiKey.CreatedAt,
			&apiKey.UserId,
			&apiKey.KeyName,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		apiKeys = append(apiKeys, apiKey)
	}

	// rows.Err() retrieves any error that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and pagination parameters
	// from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return apiKeys, metadata, nil
}

// ValidateApiKey runs validation checks on the ApiKey type.
func ValidateApiKey(v *validator.Validator, apiKey *ApiKey) {
	v.Check(apiKey.KeyName != "", "name", "must be provided")
	v.Check(len(apiKey.KeyName) <= 500, "name", "must not be more than 500 bytes long")
}
