package zdbconfig

import "gorm.io/gorm"

type ConnectionParams struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     uint
	Params   string
	Protocol string
}

type Config struct {
	RetryInterval    int
	MaxAttempts      int
	ConnectionParams ConnectionParams
	LogConfig        LogConfig
}

type LogConfig struct {
	LogLevel                  string
	SlowThreshold             int
	IgnoreRecordNotFoundError bool
	ParameterizedQuery        bool
	Colorful                  bool
}

func BuildGormConfig(logConfig LogConfig) *gorm.Config {
	newLogger := getDBLogger(logConfig)
	return &gorm.Config{Logger: newLogger}
}
