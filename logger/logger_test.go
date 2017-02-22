package logger_test

import (
	"testing"

	"github.com/gkarlik/quark-go/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogrusLogger(t *testing.T) {
	l := logger.NewLogger()
	l.SetLevel(logger.DebugLevel)

	assert.Panics(t, func() {
		l.Panic("Test panic")
	})

	assert.Panics(t, func() {
		l.PanicWithFields(logger.Fields{"panic": true}, "Test panic")
	})

	// it's hard to test fatal behavior - skipping for now
	// l.Fatal("Test fatal")
	// l.FatalWithFields(logger.Fields{"fatal": true}, "Test fatal")

	l.Debug("Test debug")
	l.DebugWithFields(logger.Fields{"debug": true}, "Test debug")

	l.Error("Test error")
	l.ErrorWithFields(logger.Fields{"error": true}, "Test error")

	l.Warning("Test warning")
	l.WarningWithFields(logger.Fields{"warning": true}, "Test warning")

	l.Info("Test info")
	l.InfoWithFields(logger.Fields{"info": true}, "Test info")

	assert.Equal(t, true, l.IsLevel(logger.DebugLevel))
	assert.Equal(t, false, l.IsLevel(logger.ErrorLevel))
}

func TestLackOfInternalLogger(t *testing.T) {
	assert.Panics(t, func() {
		logger.SetInternalLogger(nil)
	})
}

func TestSetInternalLogger(t *testing.T) {
	l := logger.NewLogger()

	logger.SetInternalLogger(l)

	assert.Equal(t, l, logger.Log())
}
