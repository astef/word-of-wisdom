package config

import (
	"errors"
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

func ReadBigIntWithDefault(key string, defaultValue string, onFailure func(err error)) *big.Int {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	result := &big.Int{}
	result, success := result.SetString(value, 0)
	if !success {
		onFailure(errors.New("failed parsing env var value " + value))
	}
	return result
}

func ReadIntWithDefault(key string, defaultValue int, onFailure func(err error)) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		onFailure(err)
	}

	return value
}
