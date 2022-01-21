package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	CheckConfig()
	account := LoadConfig()
	api := NewAPI(account)
	// handle cancellation signals
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		api.Recover()
		fmt.Println("speedup stopped")
		os.Exit(1)
	}()
	// get portal first
	api.GetPortal()
	// check if current network supports
	api.PromptSpeedupCapability()
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
