package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

var apiKeyValue string

const (
	methodGET    = "GET"
	methodPOST   = "POST"
	methodDELETE = "DELETE"
	methodUPDATE = "UPDATE"
)

func TestApiKeyCreate(t *testing.T) {
	rBody := `{ "keyName": "testKey" }`
	code, header, body := ts.request(t, methodPOST, "/v1/keys", strings.NewReader(rBody))

	if code != http.StatusOK {
		t.Fatalf("want %d; got %d\n body: %s", http.StatusOK, code, body)
	}

	pattern := `"apiKey":\s*{\s*"id":\s*\d+,\s*"user_id":\s*\d+,\s*"key_name":\s*"[^"]*",\s*"key_value":\s*"\w*"\s*}`
	re := regexp.MustCompile(pattern)

	if re.Match(body) {
		t.Errorf("want body to match regex but got %q", string(body))
	}

	check_owasp_headers(t, header)
}

func TestApiKeyDelete(t *testing.T) {
	k := createKey(t, "testKey2")
	code, header, body := ts.request(t, methodDELETE, fmt.Sprintf("/v1/keys/%d", k.Id), nil)

	if code != http.StatusOK {
		t.Fatalf("want %d; got %d", http.StatusOK, code)
	}

	expected := []byte("API key succesfully deactivated.")
	if !bytes.Contains(body, expected) {
		t.Errorf("expected body to contain \"%s\" got %s", expected, body)
	}

	check_owasp_headers(t, header)
}

func TestApiKeyValidate(t *testing.T) {
	k := createKey(t, "testKey3")

	rBody := fmt.Sprintf(`{
		"key": "%s"
	}`, k.Key)
	code, header, _ := ts.request(t, methodPOST, "/v1/keys/validate", strings.NewReader(rBody))
	if code != http.StatusOK {
		t.Fatalf("want %d; got %d", http.StatusOK, code)
	}
	check_owasp_headers(t, header)
}

func createKey(t *testing.T, name string) formatedApiKey {
	rBody := fmt.Sprintf(`{ "keyName": "%s" }`, name)
	code, _, body := ts.request(t, methodPOST, "/v1/keys", strings.NewReader(rBody))
	if code != http.StatusOK {
		t.Fatalf("want %d; got %d", http.StatusOK, code)
	}

	type resp struct {
		ApiKey formatedApiKey `json:"apiKey"`
	}
	r := resp{}

	err := json.Unmarshal(body, &r)
	if err != nil {
		t.Errorf("error unmarshalling keys: %v", err)
	}
	return r.ApiKey
}
