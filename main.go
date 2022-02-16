package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// check and load Account
	CheckConfig()
	account := LoadConfig()
	// new instance
	api := NewAPI(account)
	// handle termination signals
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		api.Recover()
		fmt.Println("speedup stopped")
		os.Exit(0)
	}()
	// get portal first
	api.GetPortal()
	// check if current network supports
	api.PromptSpeedupCapability()
	// use session to login if not nil
	if account.Session != nil {
		api.LoginKey()
	} else {
		api.Login()
	}
	// check account vip info, login if session is invalid
	api.CheckAccountCapability()
	// auto renew sessions
	api.AutoRenewal()
	// upgrade the bandwidth
	api.AutoSpeedUp()
	// keep upgraded session
	api.AutoKeepAlive()
	// block main
	select {}
}
