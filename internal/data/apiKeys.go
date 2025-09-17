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

type ApiKeyData struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	UserId    int64     `json:"user_id"`
	KeyName   string    `json:"key_name"`
	Key       Secret    `json:"key_value"`
}

// length of encoded key
// const keyLength = 64 //TODO: update this to be in config
type keyLength int

func (k keyLength) GetDecodedLen() int { return base64.StdEncoding.DecodedLen(int(k)) }
func (k keyLength) GetEncodedLen() int { return int(k) }

const keyLen64 keyLength = 64

type Secret []byte

func (s Secret) PopulateRand() {
	rand.Read(s) //nolint:all rand.Read never returns error except on legacy linux systems
}

func (s Secret) PopulateFromBase64(src []byte) {
	base64.StdEncoding.Decode(s, src) //nolint:all Decode won't return an error since the size is controlled
}

func (s Secret) GetHash() [64]byte {
	return sha512.Sum512(s)
}

func (s Secret) GetBase64encoded(out []byte) {
	base64.StdEncoding.Encode(out, s)
}

// func newSecret() Secret {
// 	s := Secret{}
// 	s = make([]byte, keyLen64.GetDecodedLen())
// 	rand.Read(s) // rand.Read never returns error except on legacy linux systems
// 	s.valueBase64 = make([]byte, keyLen64.GetEncodedLen())
// 	base64.StdEncoding.Encode(s.valueBase64, s.value)
// 	s.hash = sha512.Sum512(s.value)
// 	return s
// }

// Wraper for sql.DB connection pool and logger
type ApiKeyDataAccess struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// Create will generate the apiKey value and hash update the apiKeyData with the value.
// Param apiKeyData: apiKey data to create. Should contain userId and keyName.
// Return Value is unhandled errors.
func (m ApiKeyDataAccess) Create(apiKeyData *ApiKeyData) error {
	query := `
		INSERT INTO api_keys (user_id, key_name, key_hash) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at
		`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	newSecret := make(Secret, keyLen64.GetDecodedLen())
	newSecret.PopulateRand()
	hash := newSecret.GetHash()

	// Do db query to insert the key
	args := []any{apiKeyData.UserId, apiKeyData.KeyName, hash[:]}
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&apiKeyData.ID, &apiKeyData.CreatedAt)
	if err != nil {
		return err
	}
	apiKeyData.Key = make([]byte, keyLen64.GetEncodedLen())
	newSecret.GetBase64encoded(apiKeyData.Key)

	return nil
}

// Get will retrieve an api key with the provided keyValue as key by comparing key hashes
// Return Value 1, apiKey data retrieved
// Return Value is unhandled errors.
func (m ApiKeyDataAccess) Get(key []byte) (ApiKeyData, error) {
	query := `
		SELECT id, created_at, user_id, key_name
        	FROM api_keys
 		WHERE key_hash = $1 AND activated = TRUE
 		`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	newSecret := make(Secret, keyLen64.GetDecodedLen())
	newSecret.PopulateFromBase64(key)
	hash := newSecret.GetHash()

	var apiKeyData ApiKeyData
	err := m.DB.QueryRowContext(ctx, query, hash[:]).Scan(
		&apiKeyData.ID,
		&apiKeyData.CreatedAt,
		&apiKeyData.UserId,
		&apiKeyData.KeyName)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows): // Scan() will return a sql.ErrNoRows if there is no match
			return ApiKeyData{}, ErrRecordNotFound
		default:
			return ApiKeyData{}, err
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
func (m ApiKeyDataAccess) GetAll(user_id int, filters Filters) ([]ApiKeyData, Metadata, error) {
	// Note count(*) OVER() is used for getting the total count of records returned.
	// The (user_id = $1 OR $1 = '') clause allows for an optional filter by user_id.
	// ORDER BY %s %s, id ASC interpolates the sort column and direction from
	// the filter with id being secondary sort for matches.
	// LIMIT $2 OFFSET $3 apply limit and offset from the filter.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, user_id, key_name
		FROM api_keys
		WHERE (user_id = $1 OR $1 = -1) AND (activated = TRUE)
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`,
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
	apiKeys := []ApiKeyData{}
	for rows.Next() {
		var apiKey ApiKeyData
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
func ValidateApiKey(v *validator.Validator, apiKey *ApiKeyData) {
	v.Check(apiKey.KeyName != "", "name", "must be provided")
	v.Check(len(apiKey.KeyName) <= 500, "name", "must not be more than 500 bytes long")
}
