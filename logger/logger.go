package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger based on verbosity level
func New(verbosity int) *Logger {
	var level zapcore.Level

	switch verbosity {
	case 0:
		level = zapcore.ErrorLevel
	case 1:
		level = zapcore.InfoLevel
	case 2:
		level = zapcore.DebugLevel
	default:
		level = zapcore.DebugLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true

	logger, _ := config.Build()
	defer logger.Sync()

	return &Logger{logger.Sugar()}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.SugaredLogger.Sync()
}
