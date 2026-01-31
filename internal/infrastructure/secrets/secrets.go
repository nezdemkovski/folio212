package secrets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nezdemkovski/folio212/internal/shared/constants"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

// Service identifies this app in the OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service).
// Replace this when bootstrapping a new CLI.
const Service = constants.AppName

// Common secret keys used by this CLI.
// Add more as needed (e.g., "github-token", "slack-webhook", etc.).
const (
	KeyAPIToken            = "api-token"
	KeyTrading212APISecret = "t212-api-secret"
)

// Source indicates where a secret was retrieved from.
type Source string

const (
	SourceEnv     Source = "environment"
	SourceKeyring Source = "keyring"
	SourceFile    Source = "config_file"
	SourceNone    Source = "none"
)

// TimeoutError indicates a keyring operation timed out.
type TimeoutError struct {
	message string
}

func (e *TimeoutError) Error() string {
	return e.message
}

// Get retrieves a secret using the following priority:
// 1. Environment variable (FOLIO212_<KEY> format, e.g., FOLIO212_API_TOKEN)
// 2. OS keyring (with 3-second timeout)
// 3. Config file (insecure fallback for headless environments)
//
// Returns empty string and SourceNone if the secret doesn't exist anywhere.
func Get(key string) (value string, source Source, err error) {
	// 1. Check environment variable first (works everywhere, explicit override)
	envKey := toEnvVar(key)
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue, SourceEnv, nil
	}

	// 2. Try OS keyring (with timeout to prevent hanging)
	value, err = getFromKeyringWithTimeout(key)
	if err == nil && value != "" {
		return value, SourceKeyring, nil
	}
	// If keyring fails (unavailable/timeout), continue to file fallback
	keyringErr := err

	// 3. Fall back to config file (insecure but works on headless servers)
	value, err = getFromFile(key)
	if err == nil && value != "" {
		return value, SourceFile, nil
	}

	// Nothing found anywhere
	if keyringErr != nil && !errors.Is(keyringErr, keyring.ErrNotFound) {
		// Return the keyring error if it wasn't just "not found"
		return "", SourceNone, fmt.Errorf("failed to get secret %q (keyring error: %w)", key, keyringErr)
	}
	return "", SourceNone, nil
}

// Set stores a secret with the following priority:
// 1. OS keyring (secure, desktop environments)
// 2. Config file fallback (insecure, but necessary for headless/Docker)
//
// Returns the source where the secret was stored and whether it used insecure storage.
func Set(key, value string) (source Source, insecure bool, err error) {
	// Try to store in OS keyring first
	err = setInKeyringWithTimeout(key, value)
	if err == nil {
		// Successfully stored in keyring, clean up any file-based secret
		_ = deleteFromFile(key)
		return SourceKeyring, false, nil
	}

	// Keyring failed (timeout, unavailable, etc.), fall back to file
	// This is necessary for Docker/CI/headless servers
	if fileErr := setInFile(key, value); fileErr != nil {
		return SourceNone, true, fmt.Errorf("failed to store secret in keyring (%w) and file (%w)", err, fileErr)
	}

	return SourceFile, true, nil
}

// Delete removes a secret from all storage locations (keyring + file).
func Delete(key string) error {
	var errs []error

	// Delete from keyring
	if err := deleteFromKeyringWithTimeout(key); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("keyring: %w", err))
	}

	// Delete from file
	if err := deleteFromFile(key); err != nil {
		errs = append(errs, fmt.Errorf("file: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to delete secret %q: %v", key, errs)
	}
	return nil
}

// Keyring operations with timeouts (learned from github.com/cli/cli)

func getFromKeyringWithTimeout(key string) (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)
	go func() {
		defer close(ch)
		val, err := keyring.Get(Service, key)
		ch <- struct {
			val string
			err error
		}{val, err}
	}()
	select {
	case res := <-ch:
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", &TimeoutError{"timeout while trying to get secret from keyring"}
	}
}

func setInKeyringWithTimeout(key, value string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Set(Service, key, value)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &TimeoutError{"timeout while trying to set secret in keyring"}
	}
}

func deleteFromKeyringWithTimeout(key string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Delete(Service, key)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &TimeoutError{"timeout while trying to delete secret from keyring"}
	}
}

// File-based secret storage (insecure fallback for headless environments)

type secretsFile struct {
	Secrets map[string]string `yaml:"secrets"`
}

func getSecretsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, constants.ConfigDirName, "secrets.yml"), nil
}

func loadSecretsFile() (*secretsFile, error) {
	path, err := getSecretsFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &secretsFile{Secrets: make(map[string]string)}, nil
		}
		return nil, err
	}

	var sf secretsFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, err
	}
	if sf.Secrets == nil {
		sf.Secrets = make(map[string]string)
	}
	return &sf, nil
}

func saveSecretsFile(sf *secretsFile) error {
	path, err := getSecretsFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(sf)
	if err != nil {
		return err
	}

	// Use restrictive permissions for secrets file
	return os.WriteFile(path, data, 0o600)
}

func getFromFile(key string) (string, error) {
	sf, err := loadSecretsFile()
	if err != nil {
		return "", err
	}
	return sf.Secrets[key], nil
}

func setInFile(key, value string) error {
	sf, err := loadSecretsFile()
	if err != nil {
		return err
	}
	sf.Secrets[key] = value
	return saveSecretsFile(sf)
}

func deleteFromFile(key string) error {
	sf, err := loadSecretsFile()
	if err != nil {
		return err
	}
	delete(sf.Secrets, key)
	return saveSecretsFile(sf)
}

// toEnvVar converts a secret key to environment variable format.
// e.g., "api-token" -> "FOLIO212_API_TOKEN"
func toEnvVar(key string) string {
	// Convert to uppercase and replace hyphens with underscores
	envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	return strings.ToUpper(constants.AppName) + "_" + envKey
}
