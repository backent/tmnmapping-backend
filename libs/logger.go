package libs

import (
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/sirupsen/logrus"
)

// NewLogger provides the logger instance for dependency injection
func NewLogger() *logrus.Logger {
	return helpers.Logger
}
