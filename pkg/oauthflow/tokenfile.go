package oauthflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

// ExpandPath resolves a leading "~/" to the user home directory
func ExpandPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(path, "~"), "/"))
	}
	return path
}

// LoadTokenFile reads a persisted oauth2 token; returns os.ErrNotExist if the file is absent
func LoadTokenFile(path string) (*oauth2.Token, error) {
	content, err := os.ReadFile(ExpandPath(path))
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(content, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// SaveTokenFile persists an oauth2 token with 0600 permissions via temp file + rename
func SaveTokenFile(path string, token *oauth2.Token) error {
	path = ExpandPath(path)

	content, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0o700)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".token-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	err = tmp.Chmod(0o600)
	if err != nil {
		_ = tmp.Close()
		return err
	}

	_, err = tmp.Write(content)
	if err != nil {
		_ = tmp.Close()
		return err
	}

	err = tmp.Close()
	if err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}
