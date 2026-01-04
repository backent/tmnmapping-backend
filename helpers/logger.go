package helpers

import (
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger initializes the global logger with JSON formatter
func InitLogger() *logrus.Logger {
	logger := logrus.New()

	// Set JSON formatter for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Set log level from environment or default to INFO
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	Logger = logger
	return logger
}

// GetLogger returns the global logger instance with default fields
func GetLogger() *logrus.Entry {
	if Logger == nil {
		InitLogger()
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "tmn-backend-api"
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	return Logger.WithFields(logrus.Fields{
		"service":     serviceName,
		"environment": environment,
	})
}

// getCallerInfo returns file, line, and function name of the caller
func getCallerInfo(skip int) (file string, line int, function string) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown", 0, "unknown"
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		function = fn.Name()
		// Remove package path, keep only function name
		parts := strings.Split(function, ".")
		if len(parts) > 0 {
			function = parts[len(parts)-1]
		}
	}

	// Get only filename, not full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}

	return file, line, function
}

// getStackTrace returns a formatted stack trace
func getStackTrace(skip int) string {
	buf := make([]byte, 1024*64)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	// Split by newlines and skip the first few lines (runtime info)
	lines := strings.Split(stack, "\n")
	if len(lines) > skip*2+2 {
		return strings.Join(lines[skip*2+2:], "\n")
	}
	return stack
}

// LogInfo logs an info message with optional fields
func LogInfo(message string, fields map[string]interface{}) {
	entry := GetLogger()
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Info(message)
}

// LogWarn logs a warning message with optional fields
func LogWarn(message string, fields map[string]interface{}) {
	entry := GetLogger()
	file, line, function := getCallerInfo(1)
	entry = entry.WithFields(logrus.Fields{
		"file":     file,
		"line":     line,
		"function": function,
	})
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Warn(message)
}

// LogError logs an error message with optional fields and automatically includes caller info
func LogError(message string, err error, fields map[string]interface{}) {
	entry := GetLogger()
	file, line, function := getCallerInfo(1)

	entry = entry.WithFields(logrus.Fields{
		"file":     file,
		"line":     line,
		"function": function,
	})

	if err != nil {
		entry = entry.WithField("error", err.Error())
	}
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Error(message)
}

// LogErrorWithStackTrace logs an error with full stack trace (useful for panics)
func LogErrorWithStackTrace(message string, err error, fields map[string]interface{}) {
	entry := GetLogger()
	file, line, function := getCallerInfo(1)
	stackTrace := getStackTrace(2)

	entry = entry.WithFields(logrus.Fields{
		"file":        file,
		"line":        line,
		"function":    function,
		"stack_trace": stackTrace,
	})

	if err != nil {
		entry = entry.WithField("error", err.Error())
	}
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Error(message)
}

