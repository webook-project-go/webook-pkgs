package logger

import "go.uber.org/zap"

type zapLogger struct {
	l *zap.Logger
}

func NewZapLogger(l *zap.Logger) Logger {
	return &zapLogger{
		l: l,
	}
}

func (z *zapLogger) Debug(msg string, fields ...Field) {
	z.l.Debug(msg, z.toFiled(fields)...)
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.l.Info(msg, z.toFiled(fields)...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.l.Warn(msg, z.toFiled(fields)...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.l.Error(msg, z.toFiled(fields)...)
}

func (z *zapLogger) Fatal(msg string, fields ...Field) {
	z.l.Fatal(msg, z.toFiled(fields)...)
}

func (z *zapLogger) toFiled(f []Field) []zap.Field {
	res := make([]zap.Field, 0, len(f))
	for _, field := range f {
		res = append(res, zap.Any(field.Key, field.Data))
	}
	return res
}
