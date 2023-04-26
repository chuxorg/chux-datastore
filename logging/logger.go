package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

type Logger struct {
	*log.Logger
	level LogLevel
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", 0),
		level:  level,
	}
}

func (l *Logger) iso8601Formatter(prefix, format string, v ...interface{}) string {
	now := time.Now().UTC().Format(time.RFC3339)
	msg := fmt.Sprintf(format, v...)
	return fmt.Sprintf("%s%s %s\n", prefix, now, msg)
}

func (l *Logger) SetOutput(w io.Writer) {
	l.Logger.SetOutput(w)
}

func (l *Logger) SetLogLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l == nil {
		log.Printf("[INFO] "+format+"\n", v...)
		return
	}
	if l.level <= LogLevelInfo {
		l.Output(2, l.iso8601Formatter("[INFO] chux-datastore ", format, v...))
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l == nil {
		log.Printf("[DEBUG] "+format+"\n", v...)
		return
	}
	if l.level <= LogLevelDebug {
		l.Output(2, l.iso8601Formatter("[DEBUG] chux-datastore ", format, v...))
	}
}

func (l *Logger) Warning(format string, v ...interface{}) {
	if l == nil {
		log.Printf("[WARNING] "+format+"\n", v...)
		return
	}
	if l.level <= LogLevelWarning {
		l.Output(2, l.iso8601Formatter("[WARNING] chux-datastore ", format, v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l == nil {
		log.Printf("[ERROR] "+format+"\n", v...)
		return
	}
	if l.level <= LogLevelError {
		l.Output(2, l.iso8601Formatter("[ERROR] chux-datastore ", format, v...))
	}
}
