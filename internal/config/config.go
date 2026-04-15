package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Port           int
	DataDir        string
	DBPath         string
	ReposDir       string
	LogLevel       string
	HealthTimeout  int // seconds
	HealthInterval int // seconds
	DockerHost     string
}

func Load() *Config {
	dataDir := getEnv("COMPOSARR_DATA_DIR", "./data")

	return &Config{
		Port:           getEnvInt("COMPOSARR_PORT", 8080),
		DataDir:        dataDir,
		DBPath:         filepath.Join(dataDir, "composarr.db"),
		ReposDir:       filepath.Join(dataDir, "repos"),
		LogLevel:       getEnv("COMPOSARR_LOG_LEVEL", "info"),
		HealthTimeout:  getEnvInt("COMPOSARR_HEALTH_TIMEOUT", 120),
		HealthInterval: getEnvInt("COMPOSARR_HEALTH_INTERVAL", 5),
		DockerHost:     getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
