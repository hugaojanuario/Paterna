package commands

import (
	"os"
	"path/filepath"
)

// paternaDir retorna ~/Library/Application Support/Paterna (macOS),
// ~/.config/Paterna (Linux) ou similar — criando se não existir.
func paternaDir() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(cfg, "Paterna")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

func initializedMarkerPath() (string, error) {
	dir, err := paternaDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".initialized"), nil
}

func isInitialized() bool {
	path, err := initializedMarkerPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

func markInitialized() error {
	path, err := initializedMarkerPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte("ok"), 0644)
}
