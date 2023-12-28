package main

import (
	"embed"
	"os"
	"strings"
)

//go:embed quotes.txt
var f embed.FS

var quotes []string

func init() {
	data, err := f.ReadFile("quotes.txt")
	if err != nil {
		print("failed reading embedded file")
		os.Exit(1)
	}

	quotes = strings.Split(string(data), "\n")
}
