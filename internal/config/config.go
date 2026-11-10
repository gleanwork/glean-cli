// Package config manages the Glean CLI's configuration, providing secure storage
// of credentials using the system keyring with fallback to file-based storage.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gleanwork/glean-cli/internal/debug"
	"github.com/gleanwork/glean-cli/internal/fileutil"
	"github.com/zalando/go-keyring"
)

var (
	cfgLog     = debug.New("config:load")
	keyringLog = debug.New("config:keyring")
)

// keyringProvider defines operations for secure credential storage.
type keyringProvider interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

// ServiceName is the service identifier used for keyring operations.
var ServiceName = "glean-cli"

// ConfigPath is the path to the fallback config file.
var ConfigPath string

const (
	serverURLKey = "server_url"
	tokenKey     = "token"
)

// Config holds the Glean API credentials and connection settings.
type Config struct {
	GleanServerURL    string `json:"server_url"`
	GleanToken        string `json:"token"`
	OAuthClientID     string `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string `json:"oauth_client_secret,omitempty"`
}

// MaskToken masks a token by showing only the first and last 4 characters
// and replacing the rest with asterisks.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

// NormalizeServerURL canonicalizes a Glean server URL value.
// It trims surrounding whitespace, strips trailing slashes, and ensures a
// scheme is present (defaulting to https). Existing schemes are preserved,
// so "http://localhost:8080" stays on http.
//
// The function is idempotent — applying it twice yields the same result as
// applying it once.
//
// Examples:
//
//	NormalizeServerURL("acme-be.glean.com")          → "https://acme-be.glean.com"
//	NormalizeServerURL("https://acme-be.glean.com")  → "https://acme-be.glean.com"
//	NormalizeServerURL("https://acme-be.glean.com/") → "https://acme-be.glean.com"
//	NormalizeServerURL("http://localhost:8080")      → "http://localhost:8080"
//	NormalizeServerURL("")                           → ""
func NormalizeServerURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.TrimRight(s, "/")
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	return s
}

// legacyHostEnvError is returned when the caller has the retired GLEAN_HOST
// environment variable set without a matching GLEAN_SERVER_URL. It surfaces
// the rename in a single message rather than letting the user hit a generic
// "not configured" error downstream.
const legacyHostEnvError = `the GLEAN_HOST environment variable is no longer supported. Use GLEAN_SERVER_URL instead.

  export GLEAN_SERVER_URL=<your Glean server URL>

See https://developers.glean.com/get-started/authentication for how to find your server URL.`

// LoadConfig retrieves configuration using the following priority order:
//  1. Environment variables (GLEAN_API_TOKEN, GLEAN_SERVER_URL)
//  2. System keyring
//  3. ~/.glean/config.json
//
// If GLEAN_HOST is set without GLEAN_SERVER_URL, LoadConfig returns an error
// describing the rename rather than falling through to a "not configured"
// message from a downstream caller.
func LoadConfig() (*Config, error) {
	if os.Getenv("GLEAN_SERVER_URL") == "" && os.Getenv("GLEAN_HOST") != "" {
		return nil, fmt.Errorf("%s", legacyHostEnvError)
	}

	cfg := loadFromEnv()
	cfgLog.Log("env: server_url=%t token=%t", cfg.GleanServerURL != "", cfg.GleanToken != "")

	if cfg.GleanServerURL == "" || cfg.GleanToken == "" {
		keyringCfg := loadFromKeyring()
		if cfg.GleanServerURL == "" {
			cfg.GleanServerURL = keyringCfg.GleanServerURL
		}
		if cfg.GleanToken == "" {
			cfg.GleanToken = keyringCfg.GleanToken
		}
		cfgLog.Log("after keyring: server_url=%t token=%t", cfg.GleanServerURL != "", cfg.GleanToken != "")
	}

	if cfg.GleanServerURL == "" || cfg.GleanToken == "" {
		fileCfg, err := loadFromFile()
		if err != nil {
			cfgLog.Log("config file error: %v", err)
			return nil, err
		}
		if cfg.GleanServerURL == "" {
			cfg.GleanServerURL = fileCfg.GleanServerURL
		}
		if cfg.GleanToken == "" {
			cfg.GleanToken = fileCfg.GleanToken
		}
		if cfg.OAuthClientID == "" {
			cfg.OAuthClientID = fileCfg.OAuthClientID
		}
		if cfg.OAuthClientSecret == "" {
			cfg.OAuthClientSecret = fileCfg.OAuthClientSecret
		}
		cfgLog.Log("after file: server_url=%t token=%t", cfg.GleanServerURL != "", cfg.GleanToken != "")
	}

	cfgLog.Log("resolved server_url=%s token=%t", cfg.GleanServerURL, cfg.GleanToken != "")
	return cfg, nil
}

// loadFromEnv reads config values from environment variables.
// GLEAN_API_TOKEN takes precedence over all other credential sources.
// GLEAN_SERVER_URL is normalized on read so downstream callers always see a
// canonical form (scheme included, no trailing slash).
func loadFromEnv() *Config {
	cfg := &Config{}
	if v := os.Getenv("GLEAN_API_TOKEN"); v != "" {
		cfg.GleanToken = v
	}
	if v := os.Getenv("GLEAN_SERVER_URL"); v != "" {
		cfg.GleanServerURL = NormalizeServerURL(v)
	}
	return cfg
}

// SaveConfig stores the server URL and token in both the system keyring and file storage.
func SaveConfig(serverURL, token string) error {
	if serverURL != "" {
		serverURL = NormalizeServerURL(serverURL)
	}

	var keyringErr error
	if serverURL != "" {
		if err := keyringImpl.Set(ServiceName, serverURLKey, serverURL); err != nil {
			keyringErr = err
		}
	}
	if token != "" {
		if err := keyringImpl.Set(ServiceName, tokenKey, token); err != nil {
			keyringErr = err
		}
	}

	cfg := &Config{}
	existingCfg, err := loadFromFile()
	if err == nil {
		cfg = existingCfg
	}

	if serverURL != "" {
		cfg.GleanServerURL = serverURL
	}
	if token != "" {
		cfg.GleanToken = token
	}

	fileErr := saveToFile(cfg)

	switch {
	case keyringErr != nil && fileErr != nil:
		return fmt.Errorf("failed to save config: keyring: %v, file: %v", keyringErr, fileErr)
	case fileErr != nil:
		cfgLog.Log("warning: config file write failed (keyring OK): %v", fileErr)
		return nil
	case keyringErr != nil:
		cfgLog.Log("keyring unavailable, config saved to file only: %v", keyringErr)
		return nil
	default:
		return nil
	}
}

// SaveServerURLToFile persists only the server URL in ~/.glean/config.json
// without touching the system keyring. This is intended for OAuth flows where
// the URL is not secret and persisting it should not trigger OS keychain prompts.
func SaveServerURLToFile(serverURL string) error {
	if serverURL != "" {
		serverURL = NormalizeServerURL(serverURL)
	}

	cfg := &Config{}
	existingCfg, err := loadFromFile()
	if err == nil {
		cfg = existingCfg
	}
	cfg.GleanServerURL = serverURL

	return saveToFile(cfg)
}

// ClearTokenFromStorage removes only the API token from keyring and config file,
// leaving the server URL and other settings intact. This is used during OAuth login to
// prevent a stale API token from shadowing newly obtained OAuth credentials.
func ClearTokenFromStorage() error {
	cfgLog.Log("clearing stale API token from storage")
	// Remove token from keyring (ignore not-found).
	if err := keyringImpl.Delete(ServiceName, tokenKey); err != nil && err != keyring.ErrNotFound {
		// If the keyring can't be read either (for example, no Secret Service
		// daemon in a container), no token stored there can shadow the new
		// credentials, so fall through to the config file like SaveConfig does.
		// Only fail when a token is still readable from the keyring.
		if _, getErr := keyringImpl.Get(ServiceName, tokenKey); getErr == nil {
			return fmt.Errorf("error clearing token from keyring: %w", err)
		}
		keyringLog.Log("keyring unavailable, clearing token from config file only: %v", err)
	}

	// Remove token from config file while preserving other fields.
	cfg, err := loadFromFile()
	if err != nil {
		return nil // no file to update
	}
	if cfg.GleanToken != "" {
		cfg.GleanToken = ""
		if err := saveToFile(cfg); err != nil {
			return fmt.Errorf("error clearing token from config file: %w", err)
		}
	}
	return nil
}

// ClearConfig removes all stored configuration from both keyring and file storage.
func ClearConfig() error {
	var keyringErr error

	if err := keyringImpl.Delete(ServiceName, serverURLKey); err != nil && err != keyring.ErrNotFound {
		keyringErr = err
	}
	if err := keyringImpl.Delete(ServiceName, tokenKey); err != nil && err != keyring.ErrNotFound {
		keyringErr = err
	}

	if ConfigPath != "" {
		if err := os.Remove(ConfigPath); err != nil && !os.IsNotExist(err) {
			if keyringErr != nil {
				return fmt.Errorf("failed to clear config: keyring error: %v, file error: %v", keyringErr, err)
			}
			return fmt.Errorf("error removing config file: %w", err)
		}
	}

	if keyringErr != nil {
		return fmt.Errorf("error clearing keyring: %w", keyringErr)
	}

	return nil
}

// systemKeyring provides the default keyring implementation.
type systemKeyring struct{}

func (s *systemKeyring) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}

func (s *systemKeyring) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}

func (s *systemKeyring) Delete(service, key string) error {
	return keyring.Delete(service, key)
}

var keyringImpl keyringProvider = &systemKeyring{}

func init() {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		ConfigPath = filepath.Join(homeDir, ".glean", "config.json")
	}
}

func loadFromKeyring() *Config {
	cfg := &Config{}

	v, err := keyringImpl.Get(ServiceName, serverURLKey)
	if err == nil {
		cfg.GleanServerURL = v
	} else {
		keyringLog.Log("get %s: %v", serverURLKey, err)
		// Legacy migration: fall back to the pre-rename "host" key and
		// migrate its value forward on successful read.
		if legacy, legacyErr := keyringImpl.Get(ServiceName, "host"); legacyErr == nil && legacy != "" {
			cfg.GleanServerURL = NormalizeServerURL(legacy)
			keyringLog.Log("migrating legacy keyring key 'host' -> '%s' (%s)", serverURLKey, cfg.GleanServerURL)
			if setErr := keyringImpl.Set(ServiceName, serverURLKey, cfg.GleanServerURL); setErr == nil {
				if delErr := keyringImpl.Delete(ServiceName, "host"); delErr != nil && delErr != keyring.ErrNotFound {
					keyringLog.Log("best-effort delete of legacy 'host' key failed: %v", delErr)
				}
			} else {
				keyringLog.Log("best-effort rewrite of migrated keyring value failed: %v", setErr)
			}
		}
	}

	if token, err := keyringImpl.Get(ServiceName, tokenKey); err == nil {
		cfg.GleanToken = token
	} else {
		keyringLog.Log("get %s: %v", tokenKey, err)
	}

	return cfg
}

var knownConfigKeys = map[string]bool{
	"server_url":          true,
	"host":                true, // legacy; migrated to "server_url" on load
	"token":               true,
	"oauth_client_id":     true,
	"oauth_client_secret": true,
}

func validateConfigKeys(data []byte) []string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	var warnings []string
	for key := range raw {
		if !knownConfigKeys[key] {
			warnings = append(warnings, fmt.Sprintf("unknown key %q in config.json (typo?)", key))
		}
	}
	sort.Strings(warnings)
	return warnings
}

func loadFromFile() (*Config, error) {
	if ConfigPath == "" {
		return nil, fmt.Errorf("config path not set")
	}

	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfgLog.Log("config file not found: %s", ConfigPath)
			return &Config{}, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	cfgLog.Log("loaded config file: %s (%d bytes)", ConfigPath, len(data))

	for _, w := range validateConfigKeys(data) {
		cfgLog.Log("%s", w)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Legacy migration: if the file has the pre-rename "host" key but no
	// "server_url", copy the value (normalized) into GleanServerURL and
	// rewrite the file so subsequent loads see only the new shape.
	if cfg.GleanServerURL == "" {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err == nil {
			if legacyHostBytes, ok := raw["host"]; ok {
				var legacyHost string
				if err := json.Unmarshal(legacyHostBytes, &legacyHost); err == nil && legacyHost != "" {
					cfg.GleanServerURL = NormalizeServerURL(legacyHost)
					cfgLog.Log("migrating legacy config key 'host' -> 'server_url' (%s)", cfg.GleanServerURL)
					if err := saveToFile(&cfg); err != nil {
						cfgLog.Log("best-effort rewrite of migrated config failed: %v", err)
					}
				}
			}
		}
	}

	return &cfg, nil
}

func saveToFile(cfg *Config) error {
	if ConfigPath == "" {
		return fmt.Errorf("config path not set")
	}

	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0700); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := fileutil.WriteFileAtomic(ConfigPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
