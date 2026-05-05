package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Credentials holds the LeetCode authentication credentials.
type Credentials struct {
	Session string `json:"session"`
	CSRF    string `json:"csrf,omitempty"`
}

func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "pattern-drill"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// Load reads credentials from the config file.
// Returns an error if the file cannot be read or parsed.
func Load() (*Credentials, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// Save writes credentials to the config file with mode 0600.
func Save(c *Credentials) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "credentials.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// FromEnv reads credentials from environment variables.
// Returns nil if neither LEETCODE_SESSION nor LEETCODE_CSRF is set.
func FromEnv() *Credentials {
	session := os.Getenv("LEETCODE_SESSION")
	csrf := os.Getenv("LEETCODE_CSRF")
	if session == "" && csrf == "" {
		return nil
	}
	return &Credentials{
		Session: session,
		CSRF:    csrf,
	}
}

// Get returns credentials, with env vars taking priority over the stored file.
// Returns nil, nil if neither env vars nor stored credentials are available.
func Get() (*Credentials, error) {
	if creds := FromEnv(); creds != nil {
		return creds, nil
	}
	return Load()
}
