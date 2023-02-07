package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger interface {
	Infof(msg string, fields ...any)
	Info(msg string)
	Infow(msg string, fields ...string)
	Error(msg string)
	Errorf(msg string, fields ...any)
	Errorw(msg string, fields ...string)
	With(fields ...string) Logger
}

type null struct {
}

func (n null) With(fields ...string) Logger {
	return Null
}

func (n null) Infof(msg string, fields ...any) {
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

func (n null) Errorf(msg string, fields ...any) {
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

func (z *zapLogger) Infof(msg string, fields ...any) {
	z.logger.Info(fmt.Sprintf(msg, fields...))
}

func (z *zapLogger) With(fields ...string) Logger {
	return &zapLogger{
		logger: z.logger.With(makeFields(fields)...),
	}
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

func (z *zapLogger) Errorf(msg string, fields ...any) {
	z.logger.Error(fmt.Sprintf(msg, fields...))
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
