package data

import (
	"fmt"
	"testing"
)

// type ApiKey struct {
// 	ID        int64     `json:"id"`
// 	CreatedAt time.Time `json:"-"`
// 	UserId    string    `json:"user_id"`
// 	KeyName   string    `json:"key_name"`
// 	Key  []byte    `json:"key_value"`
// }

func TestApiKey_CreateValid(t *testing.T) {
	testKey := ApiKeyData{UserId: 1, KeyName: "testKey1"}
	err := daWrappers.ApiKeys.Create(&testKey)
	if err != nil {
		t.Fatalf("failed to create testKey1: %v", err)
	}
	if len(testKey.Key) == 0 {
		t.Fatalf("expected Key to contain apikey but got %s", testKey.Key)
	}
}

func TestApiKey_CreateGetValid(t *testing.T) {
	testKey := ApiKeyData{UserId: 1, KeyName: "testKey2"}
	err := daWrappers.ApiKeys.Create(&testKey)
	if err != nil {
		t.Fatalf("failed to create testKey2: %v", err)
	}
	fetchedKey, err := daWrappers.ApiKeys.Get(testKey.Key)
	if err != nil {
		t.Fatalf("failed to fetch key with value %s: %v", testKey.Key, err)
	}
	if fetchedKey.KeyName != "testKey2" {
		t.Fatalf("expected feteched key to have keyName = testKey2, but got %s", fetchedKey.KeyName)
	}
	if fetchedKey.UserId != 1 {
		t.Fatalf("expected feteched key to have userId = 1 testKey2, but got %s", fetchedKey.KeyName)
	}
}

func TestApiKey_GetInvalid(t *testing.T) {
	badKey := make(Secret, keyLen64.GetDecodedLen())
	badKey.PopulateRand()
	badKey64 := make([]byte, keyLen64.GetEncodedLen())
	badKey.GetBase64encoded(badKey64)

	_, err := daWrappers.ApiKeys.Get(badKey64)
	if err != ErrRecordNotFound {
		t.Fatalf("expected err ErrRecordNotFound but got: %v", err)
	}
}

func TestApiKey_CreateGetAll(t *testing.T) {
	for i := 0; i < 10; i++ {
		testKey := ApiKeyData{UserId: 1, KeyName: fmt.Sprintf("testKeyBulk-%d", i)}
		err := daWrappers.ApiKeys.Create(&testKey)
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
	for _, key := range fetchedKeys {
		t.Log(key)
	}
	if err != nil {
		t.Fatalf("failed GetAll keys: %v", err)
	}
	if len(fetchedKeys) != metadata.TotalRecords {
		t.Fatalf("expected len(fetchedKeys) == metadata.TotalRecords, got %d != %d", len(fetchedKeys), metadata.TotalRecords)
	}
}
