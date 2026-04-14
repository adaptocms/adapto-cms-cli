package credentials

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Credentials holds persisted authentication data.
type Credentials struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TenantID     string `json:"tenant_id,omitempty"`
}

// Path returns the credentials file path (~/.config/adapto/credentials.json).
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "adapto", "credentials.json")
}

// Load reads and parses the credentials file. Returns empty Credentials if the file doesn't exist.
func Load() (*Credentials, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Credentials{}, nil
		}
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// Save writes credentials as JSON with 0600 permissions.
func Save(creds *Credentials) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// Clear deletes the credentials file.
func Clear() error {
	err := os.Remove(Path())
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
