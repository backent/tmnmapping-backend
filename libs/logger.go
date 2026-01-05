package libs

import (
	"github.com/sirupsen/logrus"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

// NewLogger provides the logger instance for dependency injection
func NewLogger() *logrus.Logger {
	return helpers.Logger
}

