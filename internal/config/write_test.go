package config

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestWriteStoresOnlyNonSecretFieldsAndRoundTrips(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/icuvisor/config.json"
	plainValue := strings.Join([]string{"must", "not", "write"}, "-")
	err := Write(context.Background(), path, Config{APIKey: plainValue, AthleteID: "i12345", Timezone: "Europe/Madrid", APIBaseURL: DefaultAPIBaseURL}, WriteOptions{})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written config: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "api_key") || strings.Contains(content, plainValue) {
		t.Fatalf("written config leaked API key: %s", content)
	}
	if !strings.Contains(content, `"athlete_id": "i12345"`) || !strings.Contains(content, `"timezone": "Europe/Madrid"`) {
		t.Fatalf("written config missing normalized fields: %s", content)
	}
	for _, want := range []string{`"credential_ref"`, `"type": "keychain"`, `"service": "icuvisor"`, `"account": "intervals-icu-api-key"`} {
		if !strings.Contains(content, want) {
			t.Fatalf("written config missing credential reference %q: %s", want, content)
		}
	}
	if strings.Contains(content, "api_base_url") {
		t.Fatalf("default api_base_url should be omitted: %s", content)
	}
	loaded, err := Load(context.Background(), Options{Path: path, Env: map[string]string{}, CredentialStore: &fakeCredentialStore{value: "keychain-key"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.APIKey != "keychain-key" || loaded.AthleteID != "i12345" || loaded.Timezone != "Europe/Madrid" || loaded.APIBaseURL != DefaultAPIBaseURL {
		t.Fatalf("loaded config = %+v", loaded)
	}
}

func TestLoadWriteReloadRoundTripBytes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fixturePath := dir + "/fixture.json"
	writtenPath := dir + "/written.json"
	fixture := []byte("{\n  \"athlete_id\": \"I12345\",\n  \"timezone\": \"Europe/Madrid\",\n  \"api_base_url\": \"https://example.test/api\"\n}\n")
	if err := os.WriteFile(fixturePath, fixture, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	loaded, err := Load(context.Background(), Options{Path: fixturePath, Env: map[string]string{EnvAPIKey: "env-key"}})
	if err != nil {
		t.Fatalf("Load() fixture error = %v", err)
	}
	if err := Write(context.Background(), writtenPath, loaded, WriteOptions{}); err != nil {
		t.Fatalf("Write() round-trip error = %v", err)
	}
	written, err := os.ReadFile(writtenPath)
	if err != nil {
		t.Fatalf("read written config: %v", err)
	}
	expected := "{\n  \"credential_ref\": {\n    \"type\": \"keychain\",\n    \"service\": \"icuvisor\",\n    \"account\": \"intervals-icu-api-key\"\n  },\n  \"athlete_id\": \"i12345\",\n  \"timezone\": \"Europe/Madrid\",\n  \"api_base_url\": \"https://example.test/api\"\n}\n"
	if string(written) != expected {
		t.Fatalf("written config bytes mismatch:\n got %q\nwant %q", string(written), expected)
	}

	reloaded, err := Load(context.Background(), Options{Path: writtenPath, Env: map[string]string{EnvAPIKey: "env-key"}})
	if err != nil {
		t.Fatalf("Load() written error = %v", err)
	}
	if reloaded.AthleteID != loaded.AthleteID || reloaded.Timezone != loaded.Timezone || reloaded.APIBaseURL != loaded.APIBaseURL {
		t.Fatalf("reloaded config = %+v, want athlete/timezone/base URL from %+v", reloaded, loaded)
	}
}

func TestWriteRefusesClobberWithoutAllowOverwrite(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/config.json"
	if err := Write(context.Background(), path, Config{AthleteID: "i12345", Timezone: "UTC"}, WriteOptions{}); err != nil {
		t.Fatalf("initial Write() error = %v", err)
	}
	if err := Write(context.Background(), path, Config{AthleteID: "i67890", Timezone: "UTC"}, WriteOptions{}); err == nil {
		t.Fatal("second Write() error = nil, want clobber refusal")
	}
	if err := Write(context.Background(), path, Config{AthleteID: "i67890", Timezone: "UTC"}, WriteOptions{AllowOverwrite: true}); err != nil {
		t.Fatalf("overwrite Write() error = %v", err)
	}
	loaded, err := Load(context.Background(), Options{Path: path, Env: map[string]string{EnvAPIKey: "env-key"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.AthleteID != "i67890" {
		t.Fatalf("AthleteID = %q, want i67890", loaded.AthleteID)
	}
}
