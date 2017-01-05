package log

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
