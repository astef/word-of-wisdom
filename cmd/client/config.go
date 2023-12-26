package main

import (
	cfgutil "github.com/astef/word-of-wisdom/internal/config"
)

type config struct {
	Address string
}

func getConfig() *config {
	return &config{
		Address: cfgutil.ReadStrWithDefault("WOW_ADDRESS", ":5000"),
	}
}
