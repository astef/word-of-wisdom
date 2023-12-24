package config

import (
	"os"
	"strconv"
)

func ReadStrWithDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func MustReadIntWithDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(err)
	}

	return value
}
