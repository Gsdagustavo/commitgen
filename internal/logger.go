package internal

import (
	"context"
	"log/slog"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() *zap.Logger {
	zapLogger := initZap()

	handler := NewZapHandler(zapLogger)
	slog.SetDefault(slog.New(handler))
	return zapLogger
}

func initZap() *zap.Logger {
	const timeKey = "timestamp"
	var infoLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

	// Create lumberjack for file logger
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	// Configs for file logger
	fileSyncer := zapcore.AddSync(lumberJackLogger)
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoderConfig.TimeKey = timeKey
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	fileCore := zapcore.NewCore(
		fileEncoder,
		fileSyncer,
		infoLevel,
	)

	// Configs for console logger
	consoleSyncer := zapcore.AddSync(os.Stdout)
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.ConsoleSeparator = " | "
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoderConfig.TimeKey = timeKey
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		consoleSyncer,
		infoLevel,
	)

	// Combine both file and console logger
	core := zapcore.NewTee(fileCore, consoleCore)
	logger := zap.New(core)

	return logger
}

type ZapHandler struct {
	logger *zap.Logger
}

func NewZapHandler(l *zap.Logger) slog.Handler {
	return &ZapHandler{logger: l}
}

func (h *ZapHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		return h.logger.Core().Enabled(zap.DebugLevel)
	case slog.LevelInfo:
		return h.logger.Core().Enabled(zap.InfoLevel)
	case slog.LevelWarn:
		return h.logger.Core().Enabled(zap.WarnLevel)
	case slog.LevelError:
		return h.logger.Core().Enabled(zap.ErrorLevel)
	default:
		return h.logger.Core().Enabled(zap.InfoLevel)
	}
}

func (h *ZapHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make([]zap.Field, 0, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields = append(fields, zap.Any(a.Key, a.Value.Any()))
		return true
	})

	switch r.Level {
	case slog.LevelDebug:
		h.logger.Debug(r.Message, fields...)
	case slog.LevelInfo:
		h.logger.Info(r.Message, fields...)
	case slog.LevelWarn:
		h.logger.Warn(r.Message, fields...)
	case slog.LevelError:
		h.logger.Error(r.Message, fields...)
	default:
		h.logger.Info(r.Message, fields...)
	}

	return nil
}

func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zap.Field, 0, len(attrs))
	for _, a := range attrs {
		fields = append(fields, zap.Any(a.Key, a.Value.Any()))
	}
	return &ZapHandler{
		logger: h.logger.With(fields...),
	}
}

func (h *ZapHandler) WithGroup(name string) slog.Handler {
	return &ZapHandler{
		logger: h.logger.Named(name),
	}
}
