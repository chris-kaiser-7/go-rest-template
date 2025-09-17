package data

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"testing"
	"time"
)

var testDB *sql.DB

var daWrappers DataAccessWrapers

var dsn string

func TestMain(m *testing.M) {
	var err error

	flag.StringVar(&dsn, "dsn", "", "Test DB connection string")
	flag.Parse()
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}
	defer testDB.Close()

	//TODO: conisder support for migrations.
	if err := resetSchema(testDB); err != nil {
		log.Fatalf("failed to reset schema: %v", err)
	}

	//TODO: figure out testing race conditions
	daWrappers = InitDataAccess(testDB)

	os.Exit(m.Run())
}

func resetSchema(db *sql.DB) error {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS users CASCADE;
		DROP TABLE IF EXISTS tokens CASCADE;
		DROP TABLE IF EXISTS permissions CASCADE;
		DROP TABLE IF EXISTS users_permissions CASCADE;
		DROP TABLE IF EXISTS api_key_usage CASCADE;
		DROP TABLE IF EXISTS api_keys CASCADE;

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
			user_id BIGINT                      NOT NULL REFERENCES users ON DELETE CASCADE,
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
			user_id       BIGINT NOT NULL REFERENCES users ON DELETE CASCADE,
			permission_id BIGINT NOT NULL REFERENCES permissions ON DELETE CASCADE,
			PRIMARY KEY (user_id, permission_id)
		);

		INSERT INTO permissions (code)
		VALUES ('keys:read'), ('keys:delete'), ('keys:add');

		CREATE TABLE IF NOT EXISTS api_keys
		(
		  id         BIGSERIAL PRIMARY KEY,
		  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
		  user_id    BIGINT NOT NULL,
		  key_name   TEXT NOT NULL,
		  key_hash   INT UNIQUE NOT NULL,
		  activated  BOOL DEFAULT TRUE,
		  CONSTRAINT user_id FOREIGN KEY (id)
		  REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS api_key_usage
		(
		  id            BIGSERIAL PRIMARY KEY,
		  key_id        BIGINT NOT NULL,
		  bucket_start  TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(), --TODO: double check this
		  usage_count   INT DEFAULT 0
		  CONSTRAINT key_id FOREIGN KEY (id)
		  REFERENCES api_keys(id)
		);
	`)
	return err
}

func seedUserData() {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
		`

	userPass := password{}
	userPass.Set("asdf1235")
	args := []interface{}{"TestUser1", "validEmail@gmail.com", userPass.hash, true}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := daWrappers.Users.DB.QueryRowContext(ctx, query, args...)
	if err != nil {
		log.Fatalf("failed to insert seed user: %v", err)
	}
}
