package data

import (
	"testing"
	"time"
)

func TestToken_New(t *testing.T) {
	userSeed.clean()
	userSeed.seed()
	_, err := daWrappers.Tokens.New(userSeed.v.ID, 24*time.Hour, ScopeAuthentication)
	if err != nil {
		t.Fatalf("Failed to insert token: %v", err)
	}
}

//TODO: add more test for tokens
