package data

import (
	"testing"
)

func TestApiKey_GetSanity(t *testing.T) {
	setupSeedApiKey()
	testData, err := apiKeySeed.da.Get(apiKeySeed.v.Key)
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}
	if len(testData.Key) == 0 {
		t.Fatalf("expected Key to contain apikey but got %s", testData.Key)
	}
}

func TestApiKey_CreateSanity(t *testing.T) {
	setupSeedUserForApiKey()
	err := apiKeySeed.da.Create(&apiKeySeed.v)
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	if len(apiKeySeed.v.Key) == 0 {
		t.Fatalf("expected Key to contain apikey but got %s", apiKeySeed.v.Key)
	}
}

func TestApiKey_CreateGetValid(t *testing.T) {
	setupSeedUserForApiKey()
	err := apiKeySeed.da.Create(&apiKeySeed.v)
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	fetchedKey, err := apiKeySeed.da.Get(apiKeySeed.v.Key)
	if err != nil {
		t.Fatalf("failed to fetch key with value %s: %v", apiKeySeed.v.Key, err)
	}
	if fetchedKey.KeyName != apiKeySeed.mockDefault.KeyName {
		t.Fatalf("expected feteched key to have keyName == %s, but got %s", apiKeySeed.mockDefault.KeyName, fetchedKey.KeyName)
	}
	if fetchedKey.UserId != userSeed.v.ID {
		t.Fatalf("expected feteched key to have userId == %d, but got %d", userSeed.v.ID, fetchedKey.UserId)
	}
}

func TestApiKey_GetInvalid(t *testing.T) {
	setupSeedApiKey()
	badKey := make(Secret, key64.DecodedLen)
	badKey.PopulateRand()
	badKey64 := make([]byte, key64.EncodedLen)
	badKey.GetBase64encoded(badKey64)

	_, err := daWrappers.ApiKeys.Get(badKey64)
	if err != ErrRecordNotFound {
		t.Fatalf("expected err ErrRecordNotFound but got: %v", err)
	}
}

func TestApiKey_CreateGetAll(t *testing.T) {
	setupSeedUserForApiKey()
	apiKeySeed.seedCount(userSeed.v.ID, 10)

	filter := Filters{
		Page:         1,
		PageSize:     20,
		Sort:         "id",
		SortSafeList: []string{"id", "-id"},
	}

	fetchedKeys, metadata, err := daWrappers.ApiKeys.GetAll(-1, filter)
	// for _, key := range fetchedKeys {
	// 	t.Log(key)
	// }
	if err != nil {
		t.Fatalf("failed GetAll keys: %v", err)
	}
	if len(fetchedKeys) != metadata.TotalRecords {
		t.Fatalf("expected len(fetchedKeys) == metadata.TotalRecords, got %d != %d", len(fetchedKeys), metadata.TotalRecords)
	}
}

// seed.clean()
// seed.seedCount(10)
// for t := range seed.tGroup {
//
// }
//
// if len(seed.t.Key) == 0 {
// 	t.Fatalf("expected Key to contain apikey but got %s", testData.Key)
// }
