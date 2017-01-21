package logrus_test

import (
	"github.com/gkarlik/quark/logger"
	log "github.com/gkarlik/quark/logger/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogrusLogger(t *testing.T) {
	l := log.NewLogger()
	l.SetLogLevel(logger.DebugLogLevel)

	assert.Panics(t, func() {
		l.Panic("Test panic")
	})

	assert.Panics(t, func() {
		l.PanicWithFields(logger.LogFields{"panic": true}, "Test panic")
	})

	// it's hard to test fatal behavior
	// l.Fatal("Test fatal")
	// l.FatalWithFields(logger.LogFields{"fatal": true}, "Test fatal")

	l.Debug("Test debug")
	l.DebugWithFields(logger.LogFields{"debug": true}, "Test debug")

	l.Error("Test error")
	l.ErrorWithFields(logger.LogFields{"error": true}, "Test error")

	l.Warning("Test warning")
	l.WarningWithFields(logger.LogFields{"warning": true}, "Test warning")

	l.Info("Test info")
	l.InfoWithFields(logger.LogFields{"info": true}, "Test info")
}
