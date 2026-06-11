package logger

import (
	"log"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	level LogLevel
}

func NewLogger(level string) *Logger {
	var l LogLevel
	switch level {
	case "debug":
		l = DEBUG
	case "warn":
		l = WARN
	case "error":
		l = ERROR
	default:
		l = INFO
	}
	return &Logger{level: l}
}

func (l *Logger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		log.Println("[DEBUG] ", v)
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.level <= INFO {
		log.Println("[INFO] ", v)
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WARN {
		log.Println("[WARN] ", v)
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.level <= ERROR {
		log.Println("[ERROR] ", v)
	}
}
