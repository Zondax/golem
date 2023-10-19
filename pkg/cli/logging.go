package cli

import (
	"errors"
	"github.com/zondax/golem/pkg/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

var stringToLevel = map[string]zapcore.Level{
	constants.DebugLevel: zapcore.DebugLevel,
	constants.InfoLevel:  zapcore.InfoLevel,
	constants.WarnLevel:  zapcore.WarnLevel,
	constants.ErrorLevel: zapcore.ErrorLevel,
	constants.FatalLevel: zapcore.FatalLevel,
	constants.PanicLevel: zapcore.PanicLevel,
}

func InitGlobalLogger(level string) (*zap.Logger, error) {
	zapLevel, ok := stringToLevel[strings.ToLower(level)]
	if !ok {
		return nil, errors.New("log level '%s' is incorrect")
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(encoder, os.Stdout, zapLevel)
	logger := zap.New(core)

	zap.ReplaceGlobals(logger)

	return logger, nil
}
