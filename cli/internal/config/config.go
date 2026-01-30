package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// ConfigDirName is the name of the config directory in user's home
	ConfigDirName = ".envie"

	// CredentialsFileName is the name of the credentials file
	CredentialsFileName = "credentials.json"
)

// Credentials stores CLI authentication information
type Credentials struct {
	Token string `json:"token"`
}

// GetConfigDir returns the path to the Envie config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ConfigDirName), nil
}

// GetCredentialsPath returns the path to the credentials file
func GetCredentialsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, CredentialsFileName), nil
}

// StoreCredentials saves the token to the credentials file
func StoreCredentials(creds *Credentials) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory with restricted permissions
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	credsPath, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(credsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// LoadCredentials loads the token from the credentials file
func LoadCredentials() (*Credentials, error) {
	credsPath, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(credsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated: run 'envie auth' first")
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	if creds.Token == "" {
		return nil, fmt.Errorf("credentials file is empty or invalid")
	}

	return &creds, nil
}

// ClearCredentials removes the credentials file
func ClearCredentials() error {
	credsPath, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	if err := os.Remove(credsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	return nil
}

// GetToken retrieves the token from environment variable or credentials file
func GetToken() (string, error) {
	// 1. Check environment variable first (highest priority)
	if token := os.Getenv("ENVIE_TOKEN"); token != "" {
		return token, nil
	}

	// 2. Check credentials file
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}

	return creds.Token, nil
}
