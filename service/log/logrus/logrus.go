package logrus

import (
	log "github.com/Sirupsen/logrus"
	logging "github.com/gkarlik/quark/service/log"
)

type logrusLogger struct {
}

// NewLogrusLogger creates logger based on logrus library
func NewLogrusLogger() *logrusLogger {
	return &logrusLogger{}
}

func (logger logrusLogger) Panic(args ...interface{}) {
	log.Panic(args...)
}

func (logger logrusLogger) PanicWithFields(fields logging.LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Panic(args...)
}

func (logger logrusLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func (logger logrusLogger) FatalWithFields(fields logging.LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Fatal(args...)
}

func (logger logrusLogger) Error(args ...interface{}) {
	log.Error(args...)
}

func (logger logrusLogger) ErrorWithFields(fields logging.LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Error(args...)
}

func (logger logrusLogger) Warning(args ...interface{}) {
	log.Warning(args...)
}

func (logger logrusLogger) WarningWithFields(fields logging.LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Warning(args...)
}

func (logger logrusLogger) Info(args ...interface{}) {
	log.Info(args...)
}

func (logger logrusLogger) InfoWithFields(fields logging.LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Info(args...)
}

func (logger logrusLogger) Debug(args ...interface{}) {
	log.Debug(args...)
}

func (logger logrusLogger) DebugWithFields(fields logging.LogFields, args ...interface{}) {
	log.WithFields(log.Fields(fields)).Debug(args...)
}

func (logger logrusLogger) SetLogLevel(level logging.LogLevel) {
	log.SetLevel(log.DebugLevel)
}
