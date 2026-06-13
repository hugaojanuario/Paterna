package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHOST     string
	DBNAME     string
	DBUSER     string
	DBPASSWORD string
	DBPORT     string
	DBSSLMODE  string
	JWTSECRET  string
}

func LoadEnv() *Config {

	_ = godotenv.Load()

	host, ok := os.LookupEnv("DB_HOST")
	if !ok {
		log.Fatal("DB_HOST not set")
	}
	name, ok := os.LookupEnv("DB_NAME")
	if !ok {
		log.Fatal("DB_NAME not set")
	}
	user, ok := os.LookupEnv("DB_USER")
	if !ok {
		log.Fatal("DB_USER not set")
	}
	pass, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		log.Fatal("DB_PASSWORD not set")
	}
	port, ok := os.LookupEnv("DB_PORT")
	if !ok {
		log.Fatal("DB_PORT not set")
	}
	sslmode, ok := os.LookupEnv("DB_SSLMODE")
	if !ok {
		log.Fatal("DB_SSLMODE not set")
	}
	jwtsecret, ok := os.LookupEnv("JWT_SECRET")
	if !ok {
		log.Fatal("JWT_SECRET not set")
	}

	return &Config{
		DBHOST:     host,
		DBNAME:     name,
		DBUSER:     user,
		DBPASSWORD: pass,
		DBPORT:     port,
		DBSSLMODE:  sslmode,
		JWTSECRET:  jwtsecret,
	}
}
