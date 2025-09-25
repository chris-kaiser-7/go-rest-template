package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chris-a-kaiser-7/go-rest-template/internal/data"
	"github.com/chris-a-kaiser-7/go-rest-template/internal/jsonlog"
	"github.com/chris-a-kaiser-7/go-rest-template/internal/mailer"
)

// Define a custom testServer type which anonymously embeds a httptest.Server instance.
type testServer struct {
	*httptest.Server
}

// newTestApp returns an instance of application struct
// containing mocked dependencies to be used for testing.
func newTestApp() (*application, *sql.DB) {
	cfg := config{env: "testing"}

	flag.StringVar(&cfg.db.dsn, "dsn", "", "Test DB connection string")
	flag.Parse()

	testDB, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		log.Fatalf("failed to connect to test db: %v", err)
	}

	// if err := resetSchema(testDB); err != nil {
	// 	log.Fatalf("failed to reset schema: %v", err)
	// }
	// log.Println("completed reset schema")

	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	mail := mailer.New("smtp-host", 2525, "smtp-username", "smtp-pass", "DoNotReply <3fc3f54366-09689f+1@inbox.mailtrap.io>")
	app := application{
		config: cfg,
		logger: logger,
		models: data.InitDataAccess(testDB),
		mailer: mail,
	}

	return &app, testDB
}

// Create a newTestServer helper which initializes and returns a new instance of our
// custom testServer type.
func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	// Disable redirect-following for the client. Essentially this function is called
	// after a 3xx response is received by the client, and returning the http.ErrUseLastResponse
	// error forces it to immediately return the received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

// Implement a get method on our custom testServer type. This makes a GET request to
// a given URL path on the test server, and returns the response status code, headers,
// and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	req, err := http.NewRequest("GET", ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, body
}

func (ts *testServer) request(t *testing.T, method string, urlPath string, requestBody io.Reader) (int, http.Header, []byte) {
	req, err := http.NewRequest(method, ts.URL+urlPath, requestBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, body
}

func (ts *testServer) post(t *testing.T, urlPath string, requestBody io.Reader) (int, http.Header, []byte) {
	req, err := http.NewRequest("POST", ts.URL+urlPath, requestBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, body
}
