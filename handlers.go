package golem

import (
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var defaultConfigHandler DefaultConfigHandler

func setupCloseHandler(handler CleanUpHandler) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		zap.S().Warn("\r- Ctrl+C pressed in Terminal")

		if handler != nil {
			handler()
		}

		_ = zap.S().Sync() // Sync logger
		// TODO: friendly closing callback
		os.Exit(0)
	}()
}

func setupDefaultConfiguration(handler DefaultConfigHandler) {
	defaultConfigHandler = handler
}
