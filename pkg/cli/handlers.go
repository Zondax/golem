package cli

import (
	"context"
	"github.com/zondax/golem/pkg/logger"
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
		logger.Log(context.Background()).Warn("\r- Ctrl+C pressed in Terminal")

		if handler != nil {
			handler()
		}

		_ = logger.Sync() // Sync logger
		// TODO: friendly closing callback
		os.Exit(0)
	}()
}

func setupDefaultConfiguration(handler DefaultConfigHandler) {
	defaultConfigHandler = handler
}
