package utils

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
)

// Logger wraps PocketBase logger with tag support
type Logger struct {
	app *pocketbase.PocketBase
	tag string
}

// NewLogger creates a new logger with a base tag
func NewLogger(app *pocketbase.PocketBase, tag string) *Logger {
	return &Logger{
		app: app,
		tag: tag,
	}
}

// WithTag creates a child logger with additional tag
func (l *Logger) WithTag(subtag string) *Logger {
	return &Logger{
		app: l.app,
		tag: fmt.Sprintf("%s:%s", l.tag, subtag),
	}
}

// Info logs info level message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.app.Logger().Info(l.formatMsg(msg), args...)
}

// Debug logs debug level message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.app.Logger().Debug(l.formatMsg(msg), args...)
}

// Warn logs warning level message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.app.Logger().Warn(l.formatMsg(msg), args...)
}

// Error logs error level message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.app.Logger().Error(l.formatMsg(msg), args...)
}

// Infof logs formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.app.Logger().Info(l.formatMsg(fmt.Sprintf(format, args...)))
}

// Debugf logs formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.app.Logger().Debug(l.formatMsg(fmt.Sprintf(format, args...)))
}

// Warnf logs formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.app.Logger().Warn(l.formatMsg(fmt.Sprintf(format, args...)))
}

// Errorf logs formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.app.Logger().Error(l.formatMsg(fmt.Sprintf(format, args...)))
}

// formatMsg adds tag prefix to message
func (l *Logger) formatMsg(msg string) string {
	return fmt.Sprintf("[%s] %s", l.tag, msg)
}
