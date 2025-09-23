package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/chris-a-kaiser-7/go-rest-template/internal/data"
)

var apiKeyValue string

func TestApiKeyCreate(t *testing.T) {
	rBody := `{
		"keyName": "testKey"
	}`
	code, _, body := ts.post(t, "/v1/keys", strings.NewReader(rBody))

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	pattern := `"apiKey":\s*{\s*"id":\s*\d+,\s*"user_id":\s*\d+,\s*"key_name":\s*"[^"]*",\s*"key_value":\s*"\w*"\s*}`
	re := regexp.MustCompile(pattern)

	if re.Match(body) {
		t.Errorf("want body to match regex but got %q", string(body))
	}
}

func TestApiKeyValidate(t *testing.T) {
	k := createKey(t, "testKey2")

	rBody := fmt.Sprintf(`{
		"keyName": "%s"
	}`, k)
	code, _, body := ts.post(t, "/v1/keys", strings.NewReader(rBody))
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}
	t.Logf("%s", body)

}

func createKey(t *testing.T, name string) []byte {
	rBody := fmt.Sprintf(`{ "keyName": "%s" }`, name)
	code, _, body := ts.post(t, "/v1/keys", strings.NewReader(rBody))
	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	type resp struct {
		ApiKey data.ApiKeyData `json:"apiKey"`
	}
	r := resp{}

	err := json.Unmarshal(body, &r)
	if err != nil {
		t.Errorf("error unmarshalling keys: %v", err)
	}
	return r.ApiKey.Key
}
