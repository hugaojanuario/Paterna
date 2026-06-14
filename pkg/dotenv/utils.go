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
	// .env opcional: binário instalado via curl não tem .env no pwd
}

func get(envname string) string {
	return os.Getenv(envname)
}
