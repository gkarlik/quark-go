package logger

import (
	log "github.com/Sirupsen/logrus"
)

// LogFields represents map between parameter name and value
type LogFields map[string]interface{}

// LogLevel represents logging level
type LogLevel uint8

const (
	// PanicLogLevel represents Panic logging level
	PanicLogLevel LogLevel = iota
	// FatalLogLevel represents Fatal logging level
	FatalLogLevel
	// ErrorLogLevel represents Error logging level
	ErrorLogLevel
	// WarningLogLevel represents Warning logging level
	WarningLogLevel
	// InfoLogLevel represents Panic Info level
	InfoLogLevel
	// DebugLogLevel represents Debug logging level
	DebugLogLevel
)

// Logger represents logging mechanism
type Logger interface {
	Panic(args ...interface{})
	PanicWithFields(fields LogFields, args ...interface{})

	Fatal(args ...interface{})
	FatalWithFields(fields LogFields, args ...interface{})

	Error(args ...interface{})
	ErrorWithFields(fields LogFields, args ...interface{})

	Warning(args ...interface{})
	WarningWithFields(fields LogFields, args ...interface{})

	Info(args ...interface{})
	InfoWithFields(fields LogFields, args ...interface{})

	Debug(args ...interface{})
	DebugWithFields(fields LogFields, args ...interface{})

	SetLogLevel(level LogLevel)
}

var logger = NewLogger()

// SetInternalLogger sets logger for quark internal logging
func SetInternalLogger(l Logger) {
	if l == nil {
		panic("Internal logger must be set!")
	}

	logger = l
}

// Log returns internally configured logger. By default it is set to Logrus logger.
func Log() Logger {
	return logger
}

// LogrusLogger represents loging mechanism based on Logrus library
type LogrusLogger struct {
}

// NewLogger creates logger based on logrus library. It is default logger implementation for internal logging.
func NewLogger() Logger {
	return &LogrusLogger{}
}

// Panic logs message and panics
func (logger LogrusLogger) Panic(args ...interface{}) {
	log.Panic(args...)
}

// PanicWithFields logs message with custom fields and panic
func (logger LogrusLogger) PanicWithFields(fields LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Panic(args...)
}

// Fatal logs message and calls exit(1)
func (logger LogrusLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// FatalWithFields logs message with custom fields and calls exit(1)
func (logger LogrusLogger) FatalWithFields(fields LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Fatal(args...)
}

// Error logs error level message
func (logger LogrusLogger) Error(args ...interface{}) {
	log.Error(args...)
}

// ErrorWithFields logs error level message with custom fields
func (logger LogrusLogger) ErrorWithFields(fields LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Error(args...)
}

// Warning logs warning level message
func (logger LogrusLogger) Warning(args ...interface{}) {
	log.Warning(args...)
}

// WarningWithFields logs warning level message with custom fields
func (logger LogrusLogger) WarningWithFields(fields LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Warning(args...)
}

// Info logs info level message
func (logger LogrusLogger) Info(args ...interface{}) {
	log.Info(args...)
}

// InfoWithFields logs info level message with custom fields
func (logger LogrusLogger) InfoWithFields(fields LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Info(args...)
}

// Debug logs debug level message
func (logger LogrusLogger) Debug(args ...interface{}) {
	log.Debug(args...)
}

// DebugWithFields log debug level message with custom fields
func (logger LogrusLogger) DebugWithFields(fields LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Debug(args...)
}

// SetLogLevel sets logging level
func (logger LogrusLogger) SetLogLevel(level LogLevel) {
	log.SetLevel(log.DebugLevel)
}
