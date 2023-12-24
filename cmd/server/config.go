package main

import (
	cfgutil "github.com/astef/word-of-wisdom/internal/config"
)

type config struct {
	Address                  string
	ConnectionTimeoutMs      int
	ConnectionReadBufferSize int
}

func getConfig() *config {
	return &config{
		Address:                  cfgutil.ReadStrWithDefault("WOW_ADDRESS", ":5000"),
		ConnectionTimeoutMs:      cfgutil.MustReadIntWithDefault("WOW_CONN_TIMEOUT", 1000),
		ConnectionReadBufferSize: cfgutil.MustReadIntWithDefault("WOW_CONN_READ_BUFFER_SIZE", 64*1024), // 64KB
	}
}
