package logger

import (
	"context"
	"go.uber.org/zap"
)

const (
	RequestIDKey = "request_id"
)

func (l *Logger) WithFields(extraFields ...Field) *Logger {
	zapFields := make([]zap.Field, len(extraFields))
	for i, f := range extraFields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}

	return &Logger{logger: l.logger.With(zapFields...)}
}

func getMandatoryFields(ctx context.Context) []zap.Field {
	fields := []zap.Field{}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		fields = append(fields, zap.String(RequestIDKey, requestID))
	}

	return fields
}
