package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

const usageResolution = time.Minute * 60

type ApiKeyUsage struct {
	Id          int64     `json:"id"`
	KeyId       int64     `json:"key_id"`
	BucketStart time.Time `json:"bucket_start"`
	UsageCount  int       `json:"usage_count"`
}

type ApiKeyUsageDataAccess struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (da ApiKeyUsageDataAccess) logUsage(keyId int64) error {
	updateQuery := `
		UPDATE api_key_usage
		SET usage_count = usage_count + 1
		WHERE key_id = $1 
		`
	insertQuery := `
		INSERT INTO api_key_usage (key_id) 
		VALUES ($1) 
		`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	keyUsage := ApiKeyUsage{KeyId: keyId}
	err := da.GetLatestUsageOfKey(&keyUsage)
	if err != nil && err != ErrRecordNotFound {
		return err
	}

	// Insert a new bucket if a recent bucket is not found
	if err == ErrRecordNotFound || keyUsage.BucketStart.Add(usageResolution).Before(time.Now()) {
		result, err := da.DB.ExecContext(ctx, insertQuery, keyUsage.KeyId)
		if err != nil {
			return err
		}
		rowCount, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowCount == 0 {
			return ErrRecordNotFound
		}
	} else {
		result, err := da.DB.ExecContext(ctx, updateQuery, keyUsage.KeyId)
		if err != nil {
			return err
		}
		rowCount, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowCount == 0 {
			return ErrRecordNotFound
		}
	}

	return nil
}

func (da ApiKeyUsageDataAccess) GetLatestUsageOfKey(keyUsage *ApiKeyUsage) error {
	query := `
		SELECT id, bucket_start, usage_count
        	FROM api_key_usage
 		WHERE key_id = $1
		ORDER BY bucket_start DESC LIMIT 1
 		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := da.DB.QueryRowContext(ctx, query, keyUsage.KeyId).Scan(
		&keyUsage.Id,
		&keyUsage.BucketStart,
		&keyUsage.UsageCount)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}

func (da ApiKeyUsageDataAccess) GetAllUsageData() (int, error) {
	return da.GetUsageDataByUser(-1)
}

func (da ApiKeyUsageDataAccess) GetUsageDataByUser(user_id int64) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT COALESCE(SUM(a.usage_count), 0)
		FROM api_key_usage a
		INNER JOIN api_keys b ON a.key_id = b.id
		WHERE (user_id = $1 OR $1 = -1)
		`

	usageCount := 0
	err := da.DB.QueryRowContext(ctx, query, user_id).Scan(&usageCount)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, nil
		default:
			return -1, err
		}
	}
	return usageCount, nil
}

func (da ApiKeyUsageDataAccess) GetUsageDataByKey(key_id int64) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT COALESCE(SUM(usage_count), 0)
		FROM api_key_usage
		WHERE (key_id = $1)
		`

	usageCount := 0
	err := da.DB.QueryRowContext(ctx, query, key_id).Scan(&usageCount)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, nil
		default:
			return -1, err
		}
	}
	return usageCount, nil
}
