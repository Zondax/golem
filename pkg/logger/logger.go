package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"sync"
)

const (
	ConsoleEncode        = "console"
	initializingLogError = "initializing logger error: "
)

var (
	defaultConfig = Config{
		Level: "info",
	}
	baseLogger *zap.Logger
	lock       sync.RWMutex
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

var stringToLevel = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
	"panic": zapcore.PanicLevel,
}

func InitLogger(config Config) {
	lock.Lock()
	defer lock.Unlock()

	baseLogger = configureAndBuildLogger(config)
	zap.ReplaceGlobals(baseLogger)
}
func NewLogger(opts ...interface{}) *Logger {
	lock.Lock()
	defer lock.Unlock()

	var config Config
	var fields []Field

	for _, opt := range opts {
		switch opt := opt.(type) {
		case Config:
			config = opt
		case Field:
			fields = append(fields, opt)
		}
	}

	if config == (Config{}) {
		config = DefaultConfig()
	}

	logger := configureAndBuildLogger(config).With(toZapFields(fields)...)
	return &Logger{logger: logger}
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

	level := zapcore.InfoLevel
	if l, ok := stringToLevel[strings.ToLower(config.Level)]; ok {
		level = l
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(initializingLogError + err.Error())
	}

	return logger
}

func Sync() error {
	lock.Lock()
	defer lock.Unlock()

	return baseLogger.Sync()
}

func DefaultConfig() Config {
	return defaultConfig
}

func toZapFields(fields []Field) []zap.Field {
	var zapFields []zap.Field
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}
