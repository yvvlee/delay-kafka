package core

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const appName = "delay-kafka"

func NewLogger() (*zap.Logger, func(), error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	zapConfig.EncoderConfig.LevelKey = "_LEVEL_"
	zapConfig.EncoderConfig.TimeKey = "_TS_"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	zapConfig.EncoderConfig.NameKey = "_NAME_"
	zapConfig.EncoderConfig.MessageKey = "_MSG_"
	zapConfig.EncoderConfig.CallerKey = "_CALLER_"
	zapConfig.EncoderConfig.StacktraceKey = "_STACKTRACE_"
	logger, err := zapConfig.Build(zap.Fields(
		zap.String("_APP_", appName),
	))
	if err != nil {
		return nil, nil, err
	}
	return logger, func() { _ = logger.Sync() }, nil
}
