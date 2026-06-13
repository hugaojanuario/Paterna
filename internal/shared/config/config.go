package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHOST    string
	DBNAME    string
	DBUSER    string
	DBPASS    string
	DBPORT    string
	DBSSLMODE string
}

func LoadEnv() *Config {

	_ = godotenv.Load()

	host, ok := os.LookupEnv("DBHOST")
	if !ok {
		log.Fatal("DBHOST not set")
	}
	name, ok := os.LookupEnv("DBNAME")
	if !ok {
		log.Fatal("DBNAME not set")
	}
	user, ok := os.LookupEnv("DBUSER")
	if !ok {
		log.Fatal("DBUSER not set")
	}
	pass, ok := os.LookupEnv("DBPASS")
	if !ok {
		log.Fatal("DBPASS not set")
	}
	port, ok := os.LookupEnv("DBPORT")
	if !ok {
		log.Fatal("DBPORT not set")
	}
	sslmode, ok := os.LookupEnv("DBSSLMODE")
	if !ok {
		log.Fatal("DBSSLMODE not set")
	}

	return &Config{
		DBHOST:    host,
		DBNAME:    name,
		DBUSER:    user,
		DBPASS:    pass,
		DBPORT:    port,
		DBSSLMODE: sslmode,
	}
}
