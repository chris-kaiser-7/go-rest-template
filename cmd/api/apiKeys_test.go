package main

import (
	"net/http"
	"testing"
)

func TestApiKeyCreate(t *testing.T) {
	app := newTestApp()
	ts := newTestServer(app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/v1/keys")

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	t.Log(code)

	// expResp := `{ "apiKey": {} }`
	//
	// if string(body) != expResp {
	// 	t.Errorf("want body to equal %q,\n but got %q", expResp, string(body))
	// }
}
