package logger

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

type Field struct {
	Key  string
	Data any
}

var NopLogger Logger = &nopLogger{}

type nopLogger struct {
}

func (n *nopLogger) Debug(msg string, fields ...Field) {
}

func (n *nopLogger) Info(msg string, fields ...Field) {
}

func (n *nopLogger) Warn(msg string, fields ...Field) {
}

func (n *nopLogger) Error(msg string, fields ...Field) {
}

func (n *nopLogger) Fatal(msg string, fields ...Field) {
}
