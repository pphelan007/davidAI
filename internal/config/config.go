// Package config provides configuration management for the application.
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	App    AppConfig
	Log    LogConfig
	Worker WorkerConfig
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	NumWorkers int
}

// AppConfig holds application configuration
type AppConfig struct {
	Name string
	Env  string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	numWorkers, err := strconv.Atoi(getEnv("NUM_WORKERS", "1"))
	if err != nil {
		numWorkers = 1
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "gostarter"),
			Env:  getEnv("ENV", "development"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Worker: WorkerConfig{
			NumWorkers: numWorkers,
		},
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
