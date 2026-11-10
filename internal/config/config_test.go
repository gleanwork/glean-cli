package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// mockKeyring implements a file-based mock of the keyring for testing
type mockKeyring struct {
	dir         string
	err         error  // Used to simulate keyring errors
	serviceName string // Test-specific service name
}

func (m *mockKeyring) Get(service, key string) (string, error) {
	if service != m.serviceName {
		return "", keyring.ErrNotFound
	}
	if m.err != nil {
		return "", m.err
	}
	data, err := os.ReadFile(filepath.Join(m.dir, service+"_"+key))
	if err != nil {
		if os.IsNotExist(err) {
			return "", keyring.ErrNotFound
		}
		return "", err
	}
	return string(data), nil
}

func (m *mockKeyring) Set(service, key, value string) error {
	if service != m.serviceName {
		return fmt.Errorf("attempted to write to non-test service: %s", service)
	}
	if m.err != nil {
		return m.err
	}
	err := os.MkdirAll(m.dir, 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.dir, service+"_"+key), []byte(value), 0600)
}

func (m *mockKeyring) Delete(service, key string) error {
	if service != m.serviceName {
		return keyring.ErrNotFound
	}
	if m.err != nil {
		return m.err
	}
	err := os.Remove(filepath.Join(m.dir, service+"_"+key))
	if err != nil {
		if os.IsNotExist(err) {
			return keyring.ErrNotFound
		}
		return err
	}
	return nil
}

func setupTestKeyring(t *testing.T) (*mockKeyring, func()) {
	t.Helper()

	// Create a temporary directory for the mock keyring
	tmpDir := t.TempDir()

	// Store the original keyring implementation and service name
	originalKeyring := keyringImpl
	originalService := ServiceName

	// Use a test-specific service name
	testService := "glean-cli-test"
	ServiceName = testService

	// Create and set up the mock keyring
	mock := &mockKeyring{
		dir:         tmpDir,
		serviceName: testService,
	}
	keyringImpl = mock

	// Return the mock and cleanup function
	return mock, func() {
		keyringImpl = originalKeyring
		ServiceName = originalService
	}
}

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory for the config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Store the original config path
	originalPath := ConfigPath
	ConfigPath = configPath

	// Return the config path and cleanup function
	return configPath, func() {
		ConfigPath = originalPath
	}
}

// isolateAuthState redirects config path, service name, and keyring to
// temporary locations so tests never touch real credentials.
// This mirrors authtest.IsolateAuthState but lives in the config package
// to avoid a circular import (authtest imports config).
func isolateAuthState(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)

	oldConfigPath := ConfigPath
	ConfigPath = filepath.Join(home, ".glean", "config.json")
	t.Cleanup(func() { ConfigPath = oldConfigPath })

	oldServiceName := ServiceName
	ServiceName = "glean-cli-test-isolated"
	t.Cleanup(func() { ServiceName = oldServiceName })

	keyring.MockInit()
}

func TestNormalizeServerURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace returns empty",
			input: "   ",
			want:  "",
		},
		{
			name:  "no scheme prepends https",
			input: "acme-be.glean.com",
			want:  "https://acme-be.glean.com",
		},
		{
			name:  "https preserved",
			input: "https://acme-be.glean.com",
			want:  "https://acme-be.glean.com",
		},
		{
			name:  "http preserved for localhost",
			input: "http://localhost:8080",
			want:  "http://localhost:8080",
		},
		{
			name:  "http preserved for non-localhost",
			input: "http://acme-be.glean.com",
			want:  "http://acme-be.glean.com",
		},
		{
			name:  "trailing slash stripped",
			input: "https://acme-be.glean.com/",
			want:  "https://acme-be.glean.com",
		},
		{
			name:  "multiple trailing slashes stripped",
			input: "https://acme-be.glean.com///",
			want:  "https://acme-be.glean.com",
		},
		{
			name:  "vanity URL preserved",
			input: "acmecorp-pl.glean.com",
			want:  "https://acmecorp-pl.glean.com",
		},
		{
			name:  "obfuscated URL preserved",
			input: "a7c3d91b-be.glean.com",
			want:  "https://a7c3d91b-be.glean.com",
		},
		{
			name:  "surrounding whitespace trimmed",
			input: "  https://acme-be.glean.com  ",
			want:  "https://acme-be.glean.com",
		},
		{
			name:  "localhost without port",
			input: "localhost",
			want:  "https://localhost",
		},
		{
			name:  "localhost with port no scheme",
			input: "localhost:8080",
			want:  "https://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeServerURL(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeServerURL_Idempotent(t *testing.T) {
	inputs := []string{
		"acme-be.glean.com",
		"https://acme-be.glean.com",
		"https://acme-be.glean.com/",
		"http://localhost:8080",
		"acmecorp-pl.glean.com",
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			once := NormalizeServerURL(in)
			twice := NormalizeServerURL(once)
			assert.Equal(t, once, twice, "normalizer must be idempotent")
		})
	}
}

func TestConfigPath(t *testing.T) {
	t.Run("default config path", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)
		expected := filepath.Join(homeDir, ".glean", "config.json")
		assert.Equal(t, expected, ConfigPath)
	})
}

func TestConfigOperations(t *testing.T) {
	// Set up both keyring and config file
	mock, cleanupKeyring := setupTestKeyring(t)
	configPath, cleanupConfig := setupTestConfig(t)
	defer cleanupKeyring()
	defer cleanupConfig()

	t.Run("save and load config with working keyring", func(t *testing.T) {
		// Save config
		err := SaveConfig("https://linkedin-be.glean.com", "test-token")
		require.NoError(t, err)

		// Load config
		cfg, err := LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, "https://linkedin-be.glean.com", cfg.GleanServerURL)
		assert.Equal(t, "test-token", cfg.GleanToken)

		// Verify config file was also created
		assert.FileExists(t, configPath)
	})

	t.Run("fallback to config file when keyring fails", func(t *testing.T) {
		// First save config successfully
		err := SaveConfig("https://linkedin-be.glean.com", "test-token")
		require.NoError(t, err)

		// Now simulate keyring failure
		mock.err = assert.AnError

		// Load config should still work using file
		cfg, err := LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, "https://linkedin-be.glean.com", cfg.GleanServerURL)
		assert.Equal(t, "test-token", cfg.GleanToken)
	})

	t.Run("clear config removes from both storages", func(t *testing.T) {
		// First save some config
		err := SaveConfig("https://linkedin-be.glean.com", "test-token")
		require.NoError(t, err)

		// Reset mock error
		mock.err = nil

		// Clear config
		err = ClearConfig()
		require.NoError(t, err)

		// Verify keyring is cleared
		_, err = keyringImpl.Get(ServiceName, serverURLKey)
		assert.Equal(t, keyring.ErrNotFound, err)

		// Verify config file is removed
		_, err = os.Stat(configPath)
		assert.True(t, os.IsNotExist(err))

		// Load config should return empty values
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Empty(t, cfg.GleanServerURL)
		assert.Empty(t, cfg.GleanToken)
	})

	t.Run("save with both storages failing", func(t *testing.T) {
		// Simulate keyring failure
		mock.err = assert.AnError

		// Create config directory
		configDir := filepath.Dir(configPath)
		err := os.MkdirAll(configDir, 0700)
		require.NoError(t, err)

		// Remove any existing config file
		os.Remove(configPath)

		// Create a file instead of the config directory to make writes fail
		err = os.Remove(configDir)
		require.NoError(t, err)
		err = os.WriteFile(configDir, []byte("not a directory"), 0600)
		require.NoError(t, err)
		defer func() {
			// Clean up for other tests
			os.Remove(configDir)
			os.MkdirAll(configDir, 0700)
		}()

		err = SaveConfig("https://linkedin-be.glean.com", "test-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save config")
		assert.Contains(t, err.Error(), "keyring:")
		assert.Contains(t, err.Error(), "file:")

		// Reset mock error for other tests
		mock.err = nil
	})
}

func TestClearTokenFromStorage(t *testing.T) {
	isolateAuthState(t)

	t.Run("clears token but preserves server URL", func(t *testing.T) {
		require.NoError(t, SaveConfig("https://linkedin-be.glean.com", "stale-api-token"))

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "stale-api-token", cfg.GleanToken)

		require.NoError(t, ClearTokenFromStorage())

		cfg, err = LoadConfig()
		require.NoError(t, err)
		assert.Empty(t, cfg.GleanToken, "token should be cleared")
		assert.Equal(t, "https://linkedin-be.glean.com", cfg.GleanServerURL, "server URL should be preserved")
	})

	t.Run("no-op when no token exists", func(t *testing.T) {
		// Clear state from previous subtest.
		_ = ClearConfig()

		require.NoError(t, SaveServerURLToFile("https://acme-be.glean.com"))

		require.NoError(t, ClearTokenFromStorage())

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Empty(t, cfg.GleanToken)
		assert.Equal(t, "https://acme-be.glean.com", cfg.GleanServerURL)
	})

	t.Run("clears the file token when no keyring is available", func(t *testing.T) {
		_ = ClearConfig()

		// e.g. a devcontainer with no Secret Service daemon running
		keyring.MockInitWithError(errors.New("The name org.freedesktop.secrets was not provided by any .service files"))
		t.Cleanup(keyring.MockInit)

		// SaveConfig falls back to file-only storage when the keyring is unavailable.
		require.NoError(t, SaveConfig("https://acme-be.glean.com", "stale-api-token"))

		require.NoError(t, ClearTokenFromStorage())

		cfg, err := loadFromFile()
		require.NoError(t, err)
		assert.Empty(t, cfg.GleanToken, "token should be cleared from the config file")
		assert.Equal(t, "https://acme-be.glean.com", cfg.GleanServerURL)
	})

	t.Run("still fails when a keyring token can be read but not deleted", func(t *testing.T) {
		_, cleanup := setupTestKeyring(t)
		defer cleanup()
		keyringImpl = &undeletableKeyring{token: "stale-api-token"}

		err := ClearTokenFromStorage()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error clearing token from keyring")
	})
}

// undeletableKeyring holds a readable token that cannot be deleted, like a
// keychain that allows reads but denies the delete.
type undeletableKeyring struct{ token string }

func (k *undeletableKeyring) Get(_, key string) (string, error) {
	if key == tokenKey {
		return k.token, nil
	}
	return "", keyring.ErrNotFound
}

func (k *undeletableKeyring) Set(_, _, _ string) error { return nil }

func (k *undeletableKeyring) Delete(_, _ string) error { return errors.New("delete denied") }

func TestLoadConfigEnvPriority(t *testing.T) {
	isolateAuthState(t)

	t.Run("GLEAN_API_TOKEN overrides keyring", func(t *testing.T) {
		t.Setenv("GLEAN_API_TOKEN", "env-token")
		t.Setenv("GLEAN_SERVER_URL", "https://env-be.glean.com")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "env-token", cfg.GleanToken)
		assert.Equal(t, "https://env-be.glean.com", cfg.GleanServerURL)
	})

	t.Run("falls through to keyring when env vars absent", func(t *testing.T) {
		require.NoError(t, SaveConfig("https://linkedin-be.glean.com", "keyring-token"))

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "keyring-token", cfg.GleanToken)
	})
}

func TestLoadConfig_EnvTokenWithKeyringHost(t *testing.T) {
	isolateAuthState(t)

	err := SaveConfig("https://myhost.glean.com", "")
	require.NoError(t, err)
	t.Setenv("GLEAN_API_TOKEN", "env-token")

	result, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "env-token", result.GleanToken)
	assert.Equal(t, "https://myhost.glean.com", result.GleanServerURL, "server URL from keyring must be used even when token comes from env")
}

func TestLoadConfig_EnvHostWithFileToken(t *testing.T) {
	isolateAuthState(t)

	err := saveToFile(&Config{GleanToken: "file-token"})
	require.NoError(t, err)
	t.Setenv("GLEAN_SERVER_URL", "https://envhost.glean.com")

	result, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "https://envhost.glean.com", result.GleanServerURL)
	assert.Equal(t, "file-token", result.GleanToken, "token from file must be used even when server URL comes from env")
}

func TestLoadConfig_LegacyHostEnv_Errors(t *testing.T) {
	isolateAuthState(t)

	t.Setenv("GLEAN_SERVER_URL", "")
	t.Setenv("GLEAN_HOST", "acme-be.glean.com")

	_, err := LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GLEAN_HOST environment variable is no longer supported")
	assert.Contains(t, err.Error(), "GLEAN_SERVER_URL")
	assert.Contains(t, err.Error(), "https://developers.glean.com/get-started/authentication")
}

func TestLoadConfig_LegacyHostEnv_IgnoredWhenNewIsSet(t *testing.T) {
	isolateAuthState(t)

	t.Setenv("GLEAN_SERVER_URL", "https://acme-be.glean.com")
	t.Setenv("GLEAN_HOST", "ignored-be.glean.com")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "https://acme-be.glean.com", cfg.GleanServerURL)
}

func TestLoadConfig_NeitherEnvSet_NoError(t *testing.T) {
	isolateAuthState(t)

	t.Setenv("GLEAN_SERVER_URL", "")
	t.Setenv("GLEAN_HOST", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.GleanServerURL)
}

func TestSaveServerURLToFile_DoesNotTouchKeyring(t *testing.T) {
	mock, cleanupKeyring := setupTestKeyring(t)
	_, cleanupConfig := setupTestConfig(t)
	defer cleanupKeyring()
	defer cleanupConfig()

	mock.err = assert.AnError
	require.NoError(t, SaveServerURLToFile("https://linkedin-be.glean.com"))

	cfg, err := loadFromFile()
	require.NoError(t, err)
	assert.Equal(t, "https://linkedin-be.glean.com", cfg.GleanServerURL)
	assert.Empty(t, cfg.GleanToken)
}

// TestLoadFromFile_MigratesLegacyHostKey verifies that a config file produced
// by the pre-rename CLI (which wrote `"host": "acme-be.glean.com"`) is read
// correctly on first load under the new code, has its value normalized to a
// full URL, and is rewritten so the legacy key does not linger.
func TestLoadFromFile_MigratesLegacyHostKey(t *testing.T) {
	isolateAuthState(t)

	legacy := []byte(`{"host":"acme-be.glean.com","token":"tok"}`)
	require.NoError(t, os.MkdirAll(filepath.Dir(ConfigPath), 0700))
	require.NoError(t, os.WriteFile(ConfigPath, legacy, 0600))

	cfg, err := loadFromFile()
	require.NoError(t, err)
	assert.Equal(t, "https://acme-be.glean.com", cfg.GleanServerURL, "legacy host value must migrate to normalized server URL")
	assert.Equal(t, "tok", cfg.GleanToken)

	// File on disk should now use the new key and drop the legacy one.
	rewritten, err := os.ReadFile(ConfigPath)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(rewritten, &raw))
	assert.Contains(t, raw, "server_url")
	assert.NotContains(t, raw, "host", "legacy 'host' key should be dropped after migration")

	// Second load must be idempotent.
	cfg2, err := loadFromFile()
	require.NoError(t, err)
	assert.Equal(t, cfg.GleanServerURL, cfg2.GleanServerURL)
}

// TestLoadFromFile_ServerURLTakesPrecedenceOverLegacyHost covers a config
// file that has both keys (unlikely in practice, but possible from manual
// edits during a transition). The new key wins; the legacy one is ignored.
func TestLoadFromFile_ServerURLTakesPrecedenceOverLegacyHost(t *testing.T) {
	isolateAuthState(t)

	both := []byte(`{"server_url":"https://acme-be.glean.com","host":"ignored-be.glean.com","token":"tok"}`)
	require.NoError(t, os.MkdirAll(filepath.Dir(ConfigPath), 0700))
	require.NoError(t, os.WriteFile(ConfigPath, both, 0600))

	cfg, err := loadFromFile()
	require.NoError(t, err)
	assert.Equal(t, "https://acme-be.glean.com", cfg.GleanServerURL)
}

// TestLoadFromKeyring_MigratesLegacyHostKey verifies the keyring-side
// counterpart of the file migration: a value stored under the legacy "host"
// key is surfaced as GleanServerURL and moved to the new "server_url" key,
// with the legacy one removed.
func TestLoadFromKeyring_MigratesLegacyHostKey(t *testing.T) {
	mock, cleanup := setupTestKeyring(t)
	defer cleanup()

	require.NoError(t, mock.Set(ServiceName, "host", "acme-be.glean.com"))

	cfg := loadFromKeyring()
	assert.Equal(t, "https://acme-be.glean.com", cfg.GleanServerURL)

	// New key populated, legacy key gone.
	newVal, err := mock.Get(ServiceName, "server_url")
	require.NoError(t, err)
	assert.Equal(t, "https://acme-be.glean.com", newVal)

	_, err = mock.Get(ServiceName, "host")
	assert.ErrorIs(t, err, keyring.ErrNotFound, "legacy keyring 'host' key should be deleted after migration")
}

func TestValidateConfigKeys(t *testing.T) {
	t.Run("unknown key produces warning", func(t *testing.T) {
		data := []byte(`{"server_url":"https://x.glean.com","toke":"bad"}`)
		warnings := validateConfigKeys(data)
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], `unknown key "toke"`)
	})

	t.Run("all known keys produce no warnings", func(t *testing.T) {
		data := []byte(`{"server_url":"https://x.glean.com","token":"t","oauth_client_id":"id","oauth_client_secret":"s"}`)
		warnings := validateConfigKeys(data)
		assert.Empty(t, warnings)
	})

	t.Run("unknown keys do not prevent loading", func(t *testing.T) {
		isolateAuthState(t)

		cfgData := []byte(`{"server_url":"https://test-be.glean.com","token":"tok","extra_field":"val"}`)
		err := os.MkdirAll(filepath.Dir(ConfigPath), 0700)
		require.NoError(t, err)
		err = os.WriteFile(ConfigPath, cfgData, 0600)
		require.NoError(t, err)

		cfg, err := loadFromFile()
		require.NoError(t, err)
		assert.Equal(t, "https://test-be.glean.com", cfg.GleanServerURL)
		assert.Equal(t, "tok", cfg.GleanToken)
	})

	t.Run("invalid JSON returns no warnings", func(t *testing.T) {
		warnings := validateConfigKeys([]byte(`not json`))
		assert.Nil(t, warnings)
	})
}

func TestLoadFromFile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	t.Run("load non-existent file returns empty config", func(t *testing.T) {
		cfg, err := loadFromFile()
		require.NoError(t, err)
		assert.Empty(t, cfg.GleanServerURL)
		assert.Empty(t, cfg.GleanToken)
	})

	t.Run("load invalid JSON returns error", func(t *testing.T) {
		err := os.WriteFile(ConfigPath, []byte("invalid json"), 0600)
		require.NoError(t, err)

		_, err = loadFromFile()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing config file")
	})

	t.Run("load valid config file", func(t *testing.T) {
		cfg := Config{
			GleanServerURL: "https://test-be.glean.com",
			GleanToken:     "test-token",
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(ConfigPath, data, 0600)
		require.NoError(t, err)

		loadedCfg, err := loadFromFile()
		require.NoError(t, err)
		assert.Equal(t, cfg, *loadedCfg)
	})
}
