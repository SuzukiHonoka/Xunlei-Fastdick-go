package main

import (
	"encoding/json"
	"os"
)

const ConfigPath = "config.json"

type Account struct {
	PhoneNumber uint64
	Password    string
	Session     *Session
}

func CheckConfig() {
	if _, err := os.Stat(ConfigPath); err != nil {
		panic("no config file found")
	}
}

func LoadConfig() *Account {
	var account Account
	b, err := os.ReadFile(ConfigPath)
	CheckError(err)
	err = json.Unmarshal(b, &account)
	CheckError(err)
	return &account
}
