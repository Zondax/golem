package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ConsoleEncode        = "console"
	initializingLogError = "initializing logger error: "
)

var (
	defaultConfig = Config{
		Level:    "info",
		Encoding: "json",
	}
	defaultOptions = []zap.Option{zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel)}
)

type Config struct {
	Level    string `json:"level"`
	Encoding string `json:"encoding"`
}

type Field struct {
	Key   string
	Value interface{}
}

type Logger struct {
	logger *zap.Logger
}

func init() {
	InitLogger(defaultConfig)
}

func InitLogger(config Config) {
	baseLogger := configureAndBuildLogger(config)
	zap.ReplaceGlobals(baseLogger)
}

func NewLogger(opts ...interface{}) *Logger {
	var config *Config
	var zapFields []zap.Field

	for _, opt := range opts {
		switch opt := opt.(type) {
		case Config:
			config = &opt
		case Field:
			zapFields = append(zapFields, zap.Any(opt.Key, opt.Value))
		}
	}

	logger := L().WithOptions(defaultOptions...)

	if config != nil {
		logger = configureAndBuildLogger(*config)
	}

	logger = logger.With(zapFields...)
	return &Logger{logger: logger}
}

func NewNopLogger() *Logger {
	return &Logger{logger: zap.NewNop()}
}

func NewDevelopmentLogger(fields ...Field) *Logger {
	config := defaultConfig
	config.Encoding = ConsoleEncode

	return NewLogger(config, fields)
}

func configureAndBuildLogger(config Config) *zap.Logger {
	cfg := zap.NewProductionConfig()
	if strings.EqualFold(config.Encoding, ConsoleEncode) {
		cfg = zap.NewDevelopmentConfig()
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig = encoderConfig

	level, err := zapcore.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = zapcore.InfoLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build(defaultOptions...)
	if err != nil {
		panic(initializingLogError + err.Error())
	}

	return logger
}
