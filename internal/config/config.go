package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadConfig() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("PGPORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid PGPORT: %w", err)
	}

	sslMode := strings.ToLower(getEnv("PGSSLMODE", "disable"))
	if err := validateSSLMode(sslMode); err != nil {
		return nil, err
	}

	if sslMode == "disable" {
		fmt.Println("WARNING: SSL is disabled - not recommended for production!")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("PGHOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("PGUSER", "postgres"),
			Password: getEnv("PGPASSWORD", "Daetoi30"),
			DBName:   getEnv("PGDATABASE", "chat_service"),
			SSLMode:  sslMode,
		},
	}, nil
}

func validateSSLMode(mode string) error {
	validModes := map[string]bool{
		"disable": true, "allow": true, "prefer": true,
		"require": true, "verify-ca": true, "verify-full": true,
	}
	if !validModes[mode] {
		return fmt.Errorf("invalid SSL mode: %s (allowed: disable, allow, prefer, require, verify-ca, verify-full)", mode)
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
