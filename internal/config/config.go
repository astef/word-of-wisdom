package config

import (
	"math/big"
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

func MustReadBigIntWithDefault(key string, defaultValue string) *big.Int {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	result := &big.Int{}
	result, success := result.SetString(value, 0)
	if !success {
		panic("failed parsing env var value " + value)
	}
	return result
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
