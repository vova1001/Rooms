package config

import (
	"fmt"
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

type ConfigRDB struct {
	RedisAdr  string
	RedisPass string
}

type ConfigSMTP struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

type ConfigOTP struct {
	Secret string
}

// Локально загрузит .env.
// В Docker отсутствие файла не является ошибкой,
// потому что переменные передаёт Docker Compose.
func LoadEnv() {
	_ = godotenv.Load(".env")
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

func LoadCfgSMTP() (*ConfigSMTP, error) {
	cfg := &ConfigSMTP{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),
	}

	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST is empty")
	}
	if cfg.SMTPPort == "" {
		return nil, fmt.Errorf("SMTP_PORT is empty")
	}
	if cfg.SMTPUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME is empty")
	}
	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD is empty")
	}
	if cfg.SMTPFrom == "" {
		return nil, fmt.Errorf("SMTP_FROM is empty")
	}

	return cfg, nil
}

func LoadCfgOTP() (*ConfigOTP, error) {
	secret := os.Getenv("OTP_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("OTP_SECRET is empty")
	}

	return &ConfigOTP{
		Secret: secret,
	}, nil
}
