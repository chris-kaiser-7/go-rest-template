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
	code, _, body := ts.request(t, methodPOST, "/v1/keys", strings.NewReader(rBody))

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	pattern := `"apiKey":\s*{\s*"id":\s*\d+,\s*"user_id":\s*\d+,\s*"key_name":\s*"[^"]*",\s*"key_value":\s*"\w*"\s*}`
	re := regexp.MustCompile(pattern)

	if re.Match(body) {
		t.Errorf("want body to match regex but got %q", string(body))
	}
}

func TestApiKeyDelete(t *testing.T) {
	k := createKey(t, "testKey2")
	code, _, body := ts.request(t, methodDELETE, fmt.Sprintf("/v1/keys/%d", k.Id), nil)

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	expected := []byte("API key succesfully deactivated.")
	if !bytes.Contains(body, expected) {
		t.Errorf("expected body to contain \"%s\" got %s", expected, body)
	}
}

func TestApiKeyValidate(t *testing.T) {
	k := createKey(t, "testKey3")

	rBody := fmt.Sprintf(`{
		"key": "%s"
	}`, k.Key)
	code, _, _ := ts.request(t, methodPOST, "/v1/keys/validate", strings.NewReader(rBody))
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}
}

func createKey(t *testing.T, name string) formatedApiKey {
	rBody := fmt.Sprintf(`{ "keyName": "%s" }`, name)
	code, _, body := ts.request(t, methodPOST, "/v1/keys", strings.NewReader(rBody))
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
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
