package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	*zap.SugaredLogger
}

func NewZapLogger() *ZapLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()

	if err != nil {
		log.Fatal("Failed create logger:", err)
	}
	sugarLogger := logger.Sugar()

	return &ZapLogger{
		sugarLogger,
	}
}
