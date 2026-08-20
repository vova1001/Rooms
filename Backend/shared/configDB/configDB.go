package configShared

import (
	"fmt"
	"os"
)

type ConfigDB struct {
	DBHost    string
	DBName    string
	DBPass    string
	DBPort    string
	DBUser    string
	DBSSLMode string
}

type ConfigRDB struct {
	RedisAdr  string
	RedisPass string
}

func LoadCfgDB() (*ConfigDB, error) {
	cfg := &ConfigDB{
		DBHost:    os.Getenv("DB_HOST"),
		DBName:    os.Getenv("DB_NAME"),
		DBPass:    os.Getenv("DB_PASS"),
		DBPort:    os.Getenv("DB_PORT"),
		DBUser:    os.Getenv("DB_USER"),
		DBSSLMode: os.Getenv("DB_SSLMODE"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST is empty")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME is empty")
	}
	if cfg.DBPort == "" {
		return nil, fmt.Errorf("DB_PORT is empty")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER is empty")
	}

	return cfg, nil
}

func LoadCfgRDB() (*ConfigRDB, error) {
	cfg := &ConfigRDB{
		RedisAdr:  os.Getenv("REDIS_ADDR"),
		RedisPass: os.Getenv("REDIS_PASS"),
	}

	if cfg.RedisAdr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is empty")
	}

	return cfg, nil
}
