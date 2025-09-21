package data

import (
	"testing"
)

func TestApiKeyUsage_LogSanity(t *testing.T) {
	setupSeedApiKey()
	err := daWrappers.ApiKeyUsage.logUsage(apiKeySeed.v.Id)
	if err != nil {
		t.Fatalf("failed to log usage for key %d: %v", apiKeySeed.v.Id, err)
	}
}

func TestApiKeyUsage_GetLatestKeySanity(t *testing.T) {
	setupSeedApiKey()
	testKeyUsage := ApiKeyUsage{KeyId: apiKeySeed.v.Id}
	err := daWrappers.ApiKeyUsage.GetLatestUsageOfKey(&testKeyUsage)
	if err != ErrRecordNotFound {
		t.Fatalf("failed to get latest usage of key %s: %v", apiKeySeed.v.KeyName, err)
	}
}

func TestApiKeyUsage_GetUsageData(t *testing.T) {
	setupSeedApiKey()
	count, err := daWrappers.ApiKeyUsage.GetUsageData(-1)
	if err != nil {
		t.Fatalf("failed to get test usage data: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count == 0 instead got count == %d", count)
	}
}
