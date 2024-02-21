package logger

import (
	"context"
	"go.uber.org/zap/zapcore"
)

func Log() *Logger {
	return &Logger{logger: baseLogger}
}

func (l *Logger) log(ctx context.Context, level zapcore.Level, msg string, args ...interface{}) {
	fields := getMandatoryFields(ctx)

	sugaredLogger := l.logger.With(fields...).Sugar()

	if len(args) == 0 {
		switch level {
		case zapcore.DebugLevel:
			sugaredLogger.Debug(msg)
		case zapcore.InfoLevel:
			sugaredLogger.Info(msg)
		case zapcore.WarnLevel:
			sugaredLogger.Warn(msg)
		case zapcore.ErrorLevel:
			sugaredLogger.Error(msg)
		case zapcore.DPanicLevel:
			sugaredLogger.DPanic(msg)
		case zapcore.PanicLevel:
			sugaredLogger.Panic(msg)
		case zapcore.FatalLevel:
			sugaredLogger.Fatal(msg)
		}
	} else {
		switch level {
		case zapcore.DebugLevel:
			sugaredLogger.Debugf(msg, args...)
		case zapcore.InfoLevel:
			sugaredLogger.Infof(msg, args...)
		case zapcore.WarnLevel:
			sugaredLogger.Warnf(msg, args...)
		case zapcore.ErrorLevel:
			sugaredLogger.Errorf(msg, args...)
		case zapcore.DPanicLevel:
			sugaredLogger.DPanicf(msg, args...)
		case zapcore.PanicLevel:
			sugaredLogger.Panicf(msg, args...)
		case zapcore.FatalLevel:
			sugaredLogger.Fatalf(msg, args...)
		}
	}
}

func (l *Logger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.InfoLevel, msg, args...)
}

func (l *Logger) Infof(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.InfoLevel, template, args...)
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.DebugLevel, msg, args...)
}

func (l *Logger) Debugf(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.DebugLevel, template, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.WarnLevel, msg, args...)
}

func (l *Logger) Warnf(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.WarnLevel, template, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.ErrorLevel, msg, args...)
}

func (l *Logger) Errorf(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.ErrorLevel, template, args...)
}

func (l *Logger) DPanic(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.DPanicLevel, msg, args...)
}

func (l *Logger) DPanicf(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.DPanicLevel, template, args...)
}

func (l *Logger) Panic(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.PanicLevel, msg, args...)
}

func (l *Logger) Panicf(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.PanicLevel, template, args...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, zapcore.FatalLevel, msg, args...)
}

func (l *Logger) Fatalf(ctx context.Context, template string, args ...interface{}) {
	l.log(ctx, zapcore.FatalLevel, template, args...)
}
