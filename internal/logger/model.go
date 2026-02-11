package logger

import "time"

// Event — событие логирования с поддержкой структурированных полей.
// Методы возвращают Event для цепочки вызовов; Msg завершает запись.
type Event interface {
	String(key, value string) Event
	Int(key string, value int) Event
	Error(err error) Event
	Interface(key string, value interface{}) Event
	Duration(key string, value time.Duration) Event
	Time(key string, value time.Time) Event
	Message(message string)
}
