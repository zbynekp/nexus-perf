package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	*zap.SugaredLogger
	logger *zap.Logger
}

// New creates a logger that writes to stdout/stderr.
func New(verbosity int) (*Logger, error) {
	return NewWithFile(verbosity, "")
}

// NewWithFile creates a logger with optional file output.
func NewWithFile(verbosity int, logFile string) (*Logger, error) {
	var level zapcore.Level
	switch verbosity {
	case 0:
		level = zapcore.ErrorLevel
	case 1:
		level = zapcore.InfoLevel
	case 2:
		level = zapcore.DebugLevel
	case 3: // trace — zap has no finer level than debug
		level = zapcore.DebugLevel
	default:
		level = zapcore.DebugLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)

	if logFile != "" {
		cfg.OutputPaths = []string{logFile}
		cfg.ErrorOutputPaths = []string{logFile}
	} else {
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true

	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{
		SugaredLogger: l.Sugar(),
		logger:        l,
	}, nil
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	if l.logger != nil {
		_ = l.logger.Sync()
	}
	return l.SugaredLogger.Sync()
}
