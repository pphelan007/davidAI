// Package config provides configuration management for the application.
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Log      LogConfig
	Worker   WorkerConfig
	Temporal TemporalConfig
	Database DatabaseConfig
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
}

// TemporalConfig holds Temporal configuration
type TemporalConfig struct {
	Address   string
	Namespace string
	TaskQueue string
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

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "gostarter"),
			Env:  getEnv("ENV", "development"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Worker: WorkerConfig{},
		Temporal: TemporalConfig{
			Address:   getEnv("TEMPORAL_ADDRESS", "localhost:7233"),
			Namespace: getEnv("TEMPORAL_NAMESPACE", "default"),
			TaskQueue: getEnv("TEMPORAL_TASK_QUEUE", "davidai-task-queue"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "davidai"),
			Password: getEnv("DB_PASSWORD", "davidai"),
			DBName:   getEnv("DB_NAME", "davidai"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
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
