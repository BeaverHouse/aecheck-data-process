package logic

import (
	"os"
	"strconv"
)

// Retrieves an environment variable with a default string value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Retrieves an environment variable with a default integer value
func GetIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return valueInt
}

// Environment check functions for GO_ENV
func IsLocalEnv() bool {
	return GetEnv("GO_ENV", "local") == "local"
}

func IsDevelopmentEnv() bool {
	return GetEnv("GO_ENV", "local") == "development"
}

func IsProductionEnv() bool {
	return GetEnv("GO_ENV", "local") == "production"
}
