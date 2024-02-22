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

func NewLogger(config ...Config) *Logger {
	lock.Lock()
	defer lock.Unlock()

	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	zapLogger := configureAndBuildLogger(cfg)

	return &Logger{logger: zapLogger}
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

	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if level, ok := stringToLevel[strings.ToLower(config.Level)]; ok {
		cfg.Level = zap.NewAtomicLevelAt(level)
	}

	logger, err := cfg.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(initializingLogError + err.Error())
	}

	return logger
}

func Sync() error {
	lock.Lock()
	defer lock.Unlock()

	if err := baseLogger.Sync(); err != nil {
		return err
	}

	return nil
}
