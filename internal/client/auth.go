package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type loginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func Login(baseURL, email, password string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed (status %d): %s", resp.StatusCode, b)
	}

	var out loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Token, nil
}

func LoadCredentials() (email, password string, err error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	appDir := filepath.Join(configDir, "Paterna")

	files, err := os.ReadDir(appDir)
	if err != nil {
		return "", "", fmt.Errorf("read config dir: %w", err)
	}

	var tokenFile string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".paterna" {
			tokenFile = f.Name()
			break
		}
	}
	if tokenFile == "" {
		return "", "", fmt.Errorf("no .paterna credentials file found")
	}

	data, err := os.ReadFile(filepath.Join(appDir, tokenFile))
	if err != nil {
		return "", "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "email:"):
			email = strings.TrimSpace(strings.TrimPrefix(line, "email:"))
		case strings.HasPrefix(line, "password:"):
			password = strings.TrimSpace(strings.TrimPrefix(line, "password:"))
		}
	}

	if email == "" || password == "" {
		return "", "", fmt.Errorf("credentials missing in %s", tokenFile)
	}
	return email, password, nil
}
