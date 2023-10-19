package zdbconfig

import (
	"github.com/zondax/golem/pkg/constants"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

const (
	defaultPrefix = "\r\n"
)

var stringToLevel = map[string]logger.LogLevel{
	constants.InfoLevel:  logger.Info,
	constants.WarnLevel:  logger.Warn,
	constants.ErrorLevel: logger.Error,
	constants.FatalLevel: logger.Silent,
}

func getDBLogger(config LogConfig) logger.Interface {
	logLevel := logger.Error

	gormLevel, ok := stringToLevel[strings.ToLower(config.LogLevel)]
	if ok {
		logLevel = gormLevel
	}

	prefix := defaultPrefix
	if config.Prefix != "" {
		prefix = config.Prefix
	}

	return logger.New(
		log.New(os.Stdout, prefix, log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(config.SlowThreshold) * time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
			ParameterizedQueries:      config.ParameterizedQuery,
			Colorful:                  config.Colorful,
		},
	)
}
