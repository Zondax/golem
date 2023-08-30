package zdbconfig

import (
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func getDBLogger(config LogConfig) logger.Interface {
	logLevel := logger.Error
	switch config.LogLevel {
	case "info":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	case "silent":
		logLevel = logger.Silent
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(config.SlowThreshold) * time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
			Colorful:                  config.Colorful,
		},
	)
}
