package data

import (
	"testing"
	"time"
)

func TestToken_New(t *testing.T) {
	_, err := daWrappers.Tokens.New(0, 24*time.Hour, ScopeAuthentication)
	if err != nil {
		t.Fatalf("Failed to insert token: %v", err)
	}
}

//TODO: add more test for tokens
