package data

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"testing"
)

const (
	userTableName   = "users"
	apiKeyTableName = "api_keys"
)

var (
	testDB     *sql.DB
	daWrappers DataAccessWrapers
	dsn        string
)

func TestMain(m *testing.M) {
	var err error

	flag.StringVar(&dsn, "dsn", "", "Test DB connection string")
	flag.Parse()
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}
	defer testDB.Close()

	if err := resetSchema(testDB); err != nil {
		log.Fatalf("failed to reset schema: %v", err)
	}

	daWrappers = InitDataAccess(testDB)

	os.Exit(m.Run())
}

func resetSchema(db *sql.DB) error {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS users;
		DROP TABLE IF EXISTS tokens;
		DROP TABLE IF EXISTS permissions;
		DROP TABLE IF EXISTS users_permissions;
		DROP TABLE IF EXISTS api_key_usage;
		DROP TABLE IF EXISTS api_keys;

		CREATE TABLE IF NOT EXISTS users
		(
			id            BIGSERIAL PRIMARY KEY,
			created_at    TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
			name          TEXT                        NOT NULL,
			email         CITEXT UNIQUE               NOT NULL,
			password_hash BYTEA                       NOT NULL,
			activated     BOOL                        NOT NULL,
			version       INTEGER                     NOT NULL DEFAULT 1
		);
		
		CREATE TABLE IF NOT EXISTS tokens
		(
			hash    BYTEA PRIMARY KEY,
			user_id BIGINT                      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expiry  TIMESTAMP(0) WITH TIME ZONE NOT NULL,
			scope   TEXT                        NOT NULL
		);

		CREATE TABLE IF NOT EXISTS permissions
		(
			id   BIGSERIAL PRIMARY KEY,
			code TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS users_permissions
		(
			user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (user_id, permission_id)
		);

		INSERT INTO permissions (code)
		VALUES ('keys:read'), ('keys:delete'), ('keys:add');

		CREATE TABLE IF NOT EXISTS api_keys
		(
		  id         BIGSERIAL PRIMARY KEY,
		  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
		  user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		  key_name   TEXT NOT NULL,
		  key_hash   BYTEA UNIQUE NOT NULL,
		  activated  BOOL DEFAULT TRUE
		);

		CREATE TABLE IF NOT EXISTS api_key_usage
		(
		  id            BIGSERIAL PRIMARY KEY,
		  key_id        BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
		  bucket_start  TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(), --TODO: double check this
		  usage_count   INT DEFAULT 0
		);
	`)
	return err
}
