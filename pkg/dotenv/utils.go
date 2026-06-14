package dotenv

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func load() {
	for i := 0; i <= 5; i++ {
		path := filepath.Join(strings.Repeat("../", i), ".env")

		if err := godotenv.Load(path); err == nil {
			return
		}
	}

	panic(".env não encontrado")
}

func get(envname string) string {
	tmp := os.Getenv(envname)

	if tmp == "" {
		panic("dotenv: " + envname + " is not set")
	}

	return tmp
}
