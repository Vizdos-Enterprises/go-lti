package lti_ports

type Logger interface {
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Debug(msg string, kv ...any)
	Error(msg string, kv ...any)
}
