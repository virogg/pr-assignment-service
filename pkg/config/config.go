package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort string
	DB         DatabaseConfig
	LogLevel   string
}

func Load() (*Config, error) {
	port, err := getEnv("PORT")
	if err != nil {
		return nil, err
	}

	dbHost, err := getEnv("POSTGRES_HOST")
	if err != nil {
		return nil, err
	}

	dbUser, err := getEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := getEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := getEnv("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	dbPort, err := getEnv("POSTGRES_PORT")
	if err != nil {
		return nil, err
	}

	logLevel := getEnvWithDefault("LOG_LEVEL", "dev")

	return &Config{
		ServerPort: port,
		DB: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
			User:     dbUser,
			Password: dbPassword,
		},
		LogLevel: logLevel,
	}, nil
}

func MustLoadConfig() *Config {
	config, err := Load()
	if err != nil {
		panic(err)
	}

	return config
}

func getEnv(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("%s is not set in environment or .env file", key)
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
