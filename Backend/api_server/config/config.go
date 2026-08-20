package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

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

func LoadEnv() {
	_ = godotenv.Load(".env")
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
