package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/chris-a-kaiser-7/go-rest-template/internal/data"
)

var (
	ts        *testServer
	testToken string
)

func TestMain(m *testing.M) {
	app, db := newTestApp()
	defer db.Close() //nolint:errcheck
	ts = newTestServer(app.routes())
	defer ts.Close()

	requestBody := `{
		"name": "testUser",
		"email": "testEmail@gmail.com",
		"password": "testpass"
	}`

	code, _, body := post("/v1/users", strings.NewReader(requestBody))
	if code != http.StatusAccepted {
		log.Fatalf("failed to setup user code: %d body: %s", code, string(body))
	}

	type ruser struct {
		User data.User `json:"user"`
	}
	u := ruser{}
	err := json.Unmarshal(body, &u)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}
	perms, _ := app.models.Permissions.GetAllForUser(u.User.ID)
	log.Printf("permissions for test user %v", perms)

	requestBody = `{
		"email": "testEmail@gmail.com",
		"password": "testpass"
	}`
	code, _, body = post("/v1/tokens/authentication", strings.NewReader(requestBody))
	if code != http.StatusCreated {
		log.Fatalf("failed to setup token code: %d body: %s", code, string(body))
	}

	type resp struct {
		Auth data.Token `json:"authentication_token"`
	}
	r := resp{}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}
	testToken = r.Auth.Plaintext

	os.Exit(m.Run())
}

func post(urlPath string, requestBody io.Reader) (int, http.Header, []byte) {
	rs, err := ts.Client().Post(ts.URL+urlPath, "application/json", requestBody)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := rs.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		log.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
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
		VALUES ('keys:get'), ('keys:add'), ('keys:delete'), ('keys:admin');

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
		  usage_count   INT DEFAULT 1
		);
	`)
	return err
}
