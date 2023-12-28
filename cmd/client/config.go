package main

import (
	"os"

	cfgutil "github.com/astef/word-of-wisdom/internal/config"
)

type config struct {
	Address   string
	QuotesNum int
}

func onConfigFailure(err error) {
	print(err.Error())
	os.Exit(1)
}

func getConfig() *config {
	return &config{
		Address:   cfgutil.ReadStrWithDefault("WOW_ADDRESS", ":5000"),
		QuotesNum: cfgutil.ReadIntWithDefault("WOW_QUOTES_NUM", 1, onConfigFailure),
	}
}
