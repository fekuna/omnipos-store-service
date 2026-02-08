package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppName      string
	AppEnv       string
	GRPCPort     string
	LoggerLvl    string
	PostgresHost string
	PostgresPort string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
}

func Load() *Config {
	return &Config{
		AppName:      getEnv("APP_NAME", "omnipos-store-service"),
		AppEnv:       getEnv("APP_ENV", "development"),
		GRPCPort:     getEnv("GRPC_PORT", ":50055"), // Store Service Port
		LoggerLvl:    getEnv("LOGGER_LEVEL", "debug"),
		PostgresHost: getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort: getEnv("POSTGRES_PORT", "5432"),
		PostgresUser: getEnv("POSTGRES_USER", "postgres"),
		PostgresPass: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:   getEnv("POSTGRES_DB", "omnipos_store_db"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid value for %s: %v. Using fallback: %d", key, err, fallback)
			return fallback
		}
		return i
	}
	return fallback
}
