package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ConfigDB struct {
	DBHost    string
	DBName    string
	DBPass    string
	DBPort    string
	DBUser    string
	DBSSLMode string
}

func LoadCfgDB() (*ConfigDB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env not found")
	}
	cfgBD := &ConfigDB{
		DBHost:    os.Getenv("DB_HOST"),
		DBName:    os.Getenv("DB_NAME"),
		DBPass:    os.Getenv("DB_PASS"),
		DBPort:    os.Getenv("DB_PORT"),
		DBUser:    os.Getenv("DB_USER"),
		DBSSLMode: os.Getenv("DB_SSLMODE"),
	}
	return cfgBD, nil
}
