package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/zondax/golem/pkg/logger"
)

var defaultConfigHandler DefaultConfigHandler

func setupCloseHandler(handler CleanUpHandler) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Warn("\r- Ctrl+C pressed in Terminal")

		if handler != nil {
			handler()
		}

		logger.Sync() // Sync logger
		// TODO: friendly closing callback
		os.Exit(0)
	}()
}

func setupDefaultConfiguration(handler DefaultConfigHandler) {
	defaultConfigHandler = handler
}
