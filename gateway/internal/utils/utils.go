// Package utils utils.go contains helper-func
package utils

import (
	"os"
	"strconv"
)

// GetEnvInt returns an environment variable as an integer
func GetEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
