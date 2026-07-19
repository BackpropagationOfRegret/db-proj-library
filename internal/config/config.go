package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL         string
	HTTPAddr            string
	MaxLoansPerReader   int
	LoanDuration        time.Duration
	ElasticsearchURL    string
	ElasticsearchIndex  string
	MigrationsPath      string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:        env("DATABASE_URL", "postgres://library:library@localhost:5434/library?sslmode=disable"),
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		MaxLoansPerReader:  envInt("MAX_LOANS_PER_READER", 5),
		LoanDuration:       time.Duration(envInt("LOAN_DURATION_DAYS", 14)) * 24 * time.Hour,
		ElasticsearchURL:   env("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchIndex: env("ELASTICSEARCH_INDEX", "books"),
		MigrationsPath:     env("MIGRATIONS_PATH", "file://migrations"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.MaxLoansPerReader <= 0 {
		return Config{}, fmt.Errorf("MAX_LOANS_PER_READER must be > 0")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
