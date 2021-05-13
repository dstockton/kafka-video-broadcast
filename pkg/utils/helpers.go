package utils

import (
	"os"
)

// GetEnv is a helper function to get environment variables
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
