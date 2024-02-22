package logger

import (
	"context"
	"go.uber.org/zap"
)

const (
	loggerKey    = "golem.logger"
	RequestIDKey = "request_id"
)

func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *Logger) DPanic(msg string) {
	l.logger.DPanic(msg)
}

func (l *Logger) Panic(msg string) {
	l.logger.Panic(msg)
}

func (l *Logger) Fatal(msg string) {
	l.logger.Fatal(msg)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.logger.Sugar().Infof(template, args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.logger.Sugar().Debugf(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.logger.Sugar().Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.logger.Sugar().Errorf(template, args...)
}

func (l *Logger) DPanicf(template string, args ...interface{}) {
	l.logger.Sugar().DPanicf(template, args...)
}

func (l *Logger) Panicf(template string, args ...interface{}) {
	l.logger.Sugar().Panicf(template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.logger.Sugar().Fatalf(template, args...)
}

func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{logger: l.logger.With(fields...)}
}

func Log(ctx context.Context) *Logger {
	logger, ok := ctx.Value(loggerKey).(*Logger)
	if !ok {
		return NewLogger()
	}
	return logger
}

func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger) //nolint
}
