package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger interface {
	Infof(msg string, fields ...string)
	Info(msg string)
	Infow(msg string, fields ...string)
	Error(msg string)
	Errorf(msg string, fields ...string)
	Errorw(msg string, fields ...string)
}

type null struct {
}

func (n null) Infof(msg string, fields ...string) {
	return
}

func (n null) Info(msg string) {
	return
}

func (n null) Infow(msg string, fields ...string) {
	return
}

func (n null) Error(msg string) {
	return
}

func (n null) Errorf(msg string, fields ...string) {
	return
}

func (n null) Errorw(msg string, fields ...string) {
	return
}

var (
	Null Logger = &null{}
)

type zapLogger struct {
	logger *zap.Logger
}

func makeFields(fields []string) []zap.Field {
	zapField := make([]zap.Field, len(fields)/2)

	for i := 0; i < len(fields); i += 2 {
		zapField[i/2] = zap.String(fields[i], fields[i+1])
	}

	return zapField
}

func (z *zapLogger) Infof(msg string, fields ...string) {
	z.logger.Info(msg, makeFields(fields)...)
}

func (z *zapLogger) Info(msg string) {
	z.logger.Info(msg)
}

func (z *zapLogger) Infow(msg string, fields ...string) {
	z.logger.With(makeFields(fields)...).Info(msg)
}

func (z *zapLogger) Error(msg string) {
	z.logger.Info(msg)
}

func (z *zapLogger) Errorf(msg string, fields ...string) {
	z.logger.Error(msg, makeFields(fields)...)
}

func (z *zapLogger) Errorw(msg string, fields ...string) {
	z.logger.With(makeFields(fields)...).Error(msg)
}

func NewLogger() *zapLogger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	defaultLogLevel := zapcore.DebugLevel

	return &zapLogger{
		logger: zap.New(zapcore.NewTee(zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel)), zap.AddCaller()),
	}
}
