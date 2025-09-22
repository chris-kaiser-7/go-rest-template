package data

import (
	"testing"
)

func TestApiKeyUsage_LogSanity(t *testing.T) {
	setupSeedApiKeyUsage()
	err := daWrappers.ApiKeyUsage.logUsage(apiKeySeed.v.Id)
	if err != nil {
		t.Fatalf("failed to log usage for key %d: %v", apiKeySeed.v.Id, err)
	}
}

func TestApiKeyUsage_GetLatestKeySanity(t *testing.T) {
	setupSeedApiKeyUsage()
	testKeyUsage := ApiKeyUsage{KeyId: apiKeySeed.v.Id}
	err := daWrappers.ApiKeyUsage.GetLatestUsageOfKey(&testKeyUsage)
	if err != ErrRecordNotFound {
		t.Fatalf("failed to get latest usage of key %s: %v", apiKeySeed.v.KeyName, err)
	}
}

func TestApiKeyUsage_GetUsageData(t *testing.T) {
	setupSeedApiKeyUsage()
	count, err := daWrappers.ApiKeyUsage.GetUsageData(-1)
	if err != nil {
		t.Fatalf("failed to get test usage data: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count == 0 instead got count == %d", count)
	}
}

func TestApiKeyUsage_LogAndGet(t *testing.T) {
	setupSeedApiKeyUsage()
	err := daWrappers.ApiKeyUsage.logUsage(apiKeySeed.v.Id)
	if err != nil {
		t.Fatalf("failed to log usage for key %d: %v", apiKeySeed.v.Id, err)
	}
	testKeyUsage := ApiKeyUsage{KeyId: apiKeySeed.v.Id}
	err = daWrappers.ApiKeyUsage.GetLatestUsageOfKey(&testKeyUsage)
	if err != nil {
		t.Fatalf("failed to get latest usage of key %s: %v", apiKeySeed.v.KeyName, err)
	}
	if testKeyUsage.UsageCount != 1 {
		t.Fatalf("expected useageCount == 1 instead got %d", testKeyUsage.UsageCount)
	}
}

func TestApiKeyUsage_LogAndGet10(t *testing.T) {
	setupSeedApiKeyUsage()
	for i := 0; i < 10; i += 1 {
		err := daWrappers.ApiKeyUsage.logUsage(apiKeySeed.v.Id)
		if err != nil {
			t.Fatalf("failed to log usage for key %d: %v", apiKeySeed.v.Id, err)
		}
	}
	testKeyUsage := ApiKeyUsage{KeyId: apiKeySeed.v.Id}
	err := daWrappers.ApiKeyUsage.GetLatestUsageOfKey(&testKeyUsage)
	if err != nil {
		t.Fatalf("failed to get latest usage of key %s: %v", apiKeySeed.v.KeyName, err)
	}
	if testKeyUsage.UsageCount != 10 {
		t.Fatalf("expected useageCount == 10 instead got %d", testKeyUsage.UsageCount)
	}
}
