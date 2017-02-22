package logger

import (
	log "github.com/Sirupsen/logrus"
)

// Fields represents map between log parameter name and value.
type Fields map[string]interface{}

// Level represents logging level.
type Level uint8

const (
	// PanicLevel represents Panic logging level.
	PanicLevel Level = iota
	// FatalLevel represents Fatal logging level.
	FatalLevel
	// ErrorLevel represents Error logging level.
	ErrorLevel
	// WarningLevel represents Warning logging level.
	WarningLevel
	// InfoLevel represents Panic Info level.
	InfoLevel
	// DebugLevel represents Debug logging level.
	DebugLevel
)

// Logger represents logging mechanism.
type Logger interface {
	Panic(args ...interface{})
	PanicWithFields(fields Fields, args ...interface{})

	Fatal(args ...interface{})
	FatalWithFields(fields Fields, args ...interface{})

	Error(args ...interface{})
	ErrorWithFields(fields Fields, args ...interface{})

	Warning(args ...interface{})
	WarningWithFields(fields Fields, args ...interface{})

	Info(args ...interface{})
	InfoWithFields(fields Fields, args ...interface{})

	Debug(args ...interface{})
	DebugWithFields(fields Fields, args ...interface{})

	SetLevel(level Level)
	IsLevel(level Level) bool
}

var logger = NewLogger()

// SetInternalLogger sets logger for quark-go framework internal logging.
// Internal logger is required and default implementation is base on "github.com/Sirupsen/logrus" library.
func SetInternalLogger(l Logger) {
	if l == nil {
		panic("Internal logger must be set!")
	}

	logger = l
}

// Log returns internally configured logger. By default it is set to "github.com/Sirupsen/logrus" library logger.
func Log() Logger {
	return logger
}

// LogrusLogger represents loging mechanism based on "github.com/Sirupsen/logrus" library.
type LogrusLogger struct {
}

// NewLogger creates logger based on logrus library. It is default logger implementation for internal logging.
func NewLogger() Logger {
	return &LogrusLogger{}
}

// Panic logs message and panics.
func (logger LogrusLogger) Panic(args ...interface{}) {
	log.Panic(args...)
}

// PanicWithFields logs message with custom fields and panic.
func (logger LogrusLogger) PanicWithFields(fields Fields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Panic(args...)
}

// Fatal logs message and calls exit(1).
func (logger LogrusLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// FatalWithFields logs message with custom fields and calls exit(1).
func (logger LogrusLogger) FatalWithFields(fields Fields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Fatal(args...)
}

// Error logs error level message.
func (logger LogrusLogger) Error(args ...interface{}) {
	log.Error(args...)
}

// ErrorWithFields logs error level message with custom fields.
func (logger LogrusLogger) ErrorWithFields(fields Fields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Error(args...)
}

// Warning logs warning level message.
func (logger LogrusLogger) Warning(args ...interface{}) {
	log.Warning(args...)
}

// WarningWithFields logs warning level message with custom fields.
func (logger LogrusLogger) WarningWithFields(fields Fields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Warning(args...)
}

// Info logs info level message.
func (logger LogrusLogger) Info(args ...interface{}) {
	log.Info(args...)
}

// InfoWithFields logs info level message with custom fields.
func (logger LogrusLogger) InfoWithFields(fields Fields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Info(args...)
}

// Debug logs debug level message.
func (logger LogrusLogger) Debug(args ...interface{}) {
	log.Debug(args...)
}

// DebugWithFields log debug level message with custom fields.
func (logger LogrusLogger) DebugWithFields(fields Fields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Debug(args...)
}

// SetLevel sets logging level.
func (logger LogrusLogger) SetLevel(level Level) {
	l := log.Level(uint8(level))

	log.SetLevel(l)
}

// IsLevel returs true if log level specified in parameter matches log level currently sets on logger.
func (logger LogrusLogger) IsLevel(level Level) bool {
	l := Level(log.GetLevel())
	if l == level {
		return true
	}
	return false
}
