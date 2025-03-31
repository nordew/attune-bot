package logger

import (
	"context"
	"os"

	"log/slog"
)

type SLogger struct {
	logger *slog.Logger
}

func NewSLogger() *SLogger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	return &SLogger{
		logger: logger,
	}
}

func (l *SLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, flattenArgs(args)...)
}

func (l *SLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, flattenArgs(args)...)
}

func (l *SLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, flattenArgs(args)...)
}

func (l *SLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, flattenArgs(args)...)
}

func (l *SLogger) With(args ...interface{}) Logger {
	newLogger := l.logger.With(flattenArgs(args)...)
	return &SLogger{
		logger: newLogger,
	}
}

func flattenArgs(args []interface{}) []any {
	flat := make([]any, 0, len(args))
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "unknown"
		}
		flat = append(flat, key, args[i+1])
	}
	return flat
}
