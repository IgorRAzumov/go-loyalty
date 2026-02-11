package logger

// Logger — контракт логирования. Позволяет подменять реализацию (zerolog, nop, mock).
type Logger interface {
	Info() Event
	Warn() Event
	Error() Event
	Debug() Event
}
