package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

var (
	userSeed = UserTestSeed{
		mockDefault: User{
			Name:  "TestUser1",
			Email: "validEmail@gmail.com",
		},
		da: &daWrappers.Users,
	}
	apiKeySeed = ApiKeyTestSeed{
		mockDefault: ApiKeyData{
			KeyName: "seedKey",
		},
		da: &daWrappers.ApiKeys,
	}
	apiKeyUsageSeed = ApiKeyUsageTestSeed{
		da: &daWrappers.ApiKeyUsage,
	}
)

// Default boilerplate for cleaning table
func clean(tableName string, db *sql.DB) {
	query := fmt.Sprintf(`DELETE FROM %s`, tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Fatalf("failed to clean %s: %v", tableName, err)
	}
}

type UserTestSeed struct {
	mockDefault User
	v           User
	da          *UserDataAccess
}

func (s *UserTestSeed) clean() {
	copy := s.mockDefault
	s.v = copy
	clean(userTableName, s.da.DB)
}

func (s *UserTestSeed) seed() {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
		`

	userPass := password{}
	err := userPass.Set("asdf1235")
	if err != nil {
		log.Fatalf("failed to set password for seed user: %v", err)
	}
	s.v.Password = userPass
	s.v.Activated = true
	args := []any{s.mockDefault.Name, s.mockDefault.Email, s.v.Password.hash, s.v.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = s.da.DB.QueryRowContext(ctx, query, args...).Scan(&s.v.ID, &s.v.CreatedAt, &s.v.Version)
	if err != nil {
		log.Fatalf("failed to insert seed user: %v", err)
	}
}

type ApiKeyTestSeed struct {
	mockDefault ApiKeyData
	v           ApiKeyData
	vGroup      []ApiKeyData
	da          *ApiKeyDataAccess
}

func (s *ApiKeyTestSeed) clean() {
	copy := s.mockDefault
	s.v = copy
	clean(apiKeyTableName, s.da.DB)
}

func (s *ApiKeyTestSeed) seed(fkeyUserId int64) {
	query := `
		INSERT INTO api_keys (user_id, key_name, key_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	newSecret := make(Secret, key64.DecodedLen)
	newSecret.PopulateRand()
	hash := newSecret.GetHash()

	args := []any{fkeyUserId, s.mockDefault.KeyName, hash[:]}

	err := s.da.DB.QueryRowContext(ctx, query, args...).Scan(&s.v.Id, &s.v.CreatedAt)
	if err != nil {
		log.Fatalf("failed to insert seed key: %v", err)
	}
	s.v.Key = make([]byte, key64.EncodedLen)
	s.v.UserId = fkeyUserId
	s.v.KeyName = s.mockDefault.KeyName
	newSecret.GetBase64encoded(s.v.Key)
}

func (s *ApiKeyTestSeed) seedCount(fkeyUserId int64, count int) {
	query := `
		INSERT INTO api_keys (user_id, key_name, key_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := 0; i < count; i++ {
		newSecret := make(Secret, key64.DecodedLen)
		newSecret.PopulateRand()
		hash := newSecret.GetHash()
		newKey := ApiKeyData{
			UserId:  fkeyUserId,
			KeyName: fmt.Sprintf("%s-%d", s.mockDefault.KeyName, i),
		}
		args := []any{newKey.UserId, newKey.KeyName, hash[:]}

		_, err := s.da.DB.ExecContext(ctx, query, args...)
		if err != nil {
			log.Fatalf("failed to create testKeyBulk-%d: %v", i, err)
		}
		newKey.Key = make([]byte, key64.EncodedLen)
		newSecret.GetBase64encoded(newKey.Key)
		s.vGroup = append(s.vGroup, newKey)
	}
}

type ApiKeyUsageTestSeed struct {
	da *ApiKeyUsageDataAccess
}

func (s *ApiKeyUsageTestSeed) clean() {
	clean(apiKeyTableName, s.da.DB)
}

// Boilerplate for seed calls

func setupSeedUserForApiKey() {
	userSeed.clean()
	apiKeySeed.clean()
	userSeed.seed()
	apiKeySeed.v.UserId = userSeed.v.ID
}

func setupSeedApiKey() {
	userSeed.clean()
	apiKeySeed.clean()
	userSeed.seed()
	apiKeySeed.seed(userSeed.v.ID)
}

func setupSeedApiKeyUsage() {
	userSeed.clean()
	apiKeySeed.clean()
	apiKeyUsageSeed.clean()
	userSeed.seed()
	apiKeySeed.seed(userSeed.v.ID)
}
