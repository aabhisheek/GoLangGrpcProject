package observability

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

// InitLogger initializes the logger
func InitLogger(cfg LoggerConfig) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// LoggerFromContext extracts logger from context or returns global logger
func LoggerFromContext(logger *zap.Logger) *zap.Logger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return logger
}

