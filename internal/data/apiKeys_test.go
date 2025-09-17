package data

import (
	"crypto/rand"
	"fmt"
	"testing"
)

// type ApiKey struct {
// 	ID        int64     `json:"id"`
// 	CreatedAt time.Time `json:"-"`
// 	UserId    string    `json:"user_id"`
// 	KeyName   string    `json:"key_name"`
// 	KeyValue  []byte    `json:"key_value"`
// }

func TestApiKey_CreateValid(t *testing.T) {
	testKey := ApiKey{UserId: 0, KeyName: "testKey1"}
	err := daWrappers.ApiKeys.Create(testKey)
	if err != nil {
		t.Fatalf("failed to create testKey1: %v", err)
	}
	if len(testKey.KeyValue) == 0 {
		t.Fatalf("expected KeyValue to contain apikey but got %s", testKey.KeyValue)
	}
}

func TestApiKey_CreateGetValid(t *testing.T) {
	testKey := ApiKey{UserId: 0, KeyName: "testKey2"}
	err := daWrappers.ApiKeys.Create(testKey)
	if err != nil {
		t.Fatalf("failed to create testKey2: %v", err)
	}
	fetchedKey, err := daWrappers.ApiKeys.Get(testKey.KeyValue)
	if err != nil {
		t.Fatalf("failed to fetch key with value %s: %v", testKey.KeyValue, err)
	}
	if fetchedKey.KeyName != "testKey2" {
		t.Fatalf("expected feteched key to have keyName = testKey2, but got %s", fetchedKey.KeyName)
	}
	if fetchedKey.UserId != 0 {
		t.Fatalf("expected feteched key to have userId = 0 testKey2, but got %s", fetchedKey.KeyName)
	}
}

func TestApiKey_GetInvalid(t *testing.T) {
	//create random key
	badKey := make([]byte, keyLength)
	_, err := rand.Read(badKey)
	if err != nil {
		t.Fatalf("failed to create random key: %v", err)
	}

	_, err = daWrappers.ApiKeys.Get(badKey)
	if err != ErrRecordNotFound {
		t.Fatalf("expected err ErrRecordNotFound but got: %v", err)
	}
}

func TestApiKey_CreateGetAll(t *testing.T) {
	for i := 0; i < 10; i++ {
		testKey := ApiKey{UserId: 0, KeyName: fmt.Sprintf("testKeyBulk-%d", i)}
		err := daWrappers.ApiKeys.Create(testKey)
		if err != nil {
			t.Fatalf("failed to create testKeyBulk-%d: %v", i, err)
		}
	}

	filter := Filters{
		Page:         1,
		PageSize:     20,
		Sort:         "id",
		SortSafeList: []string{"id", "-id"},
	}

	fetchedKeys, metadata, err := daWrappers.ApiKeys.GetAll(-1, filter)
	if err != nil {
		t.Fatalf("failed GetAll keys: %v", err)
	}
	if len(fetchedKeys) != 10 {
		t.Fatalf("expected 10 keys, got %d", len(fetchedKeys))
	}
	if metadata.TotalRecords != 10 {
		t.Fatalf("expected metadata.TotalRecords = 10, got %d", metadata.TotalRecords)
	}
}
