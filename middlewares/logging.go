package middlewares

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

// responseWriter is a wrapper around http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware struct
type LoggingMiddleware struct{}

// NewLoggingMiddleware creates a new logging middleware instance
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// Log wraps the httprouter.Handle with logging functionality
func (m *LoggingMiddleware) Log(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := newResponseWriter(w)

		// Execute the next handler
		next(rw, r, p)

		// Calculate duration
		duration := time.Since(start)

		// Build log fields
		logFields := map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"query":       r.URL.RawQuery,
			"ip":          r.RemoteAddr,
			"user_agent":  r.UserAgent(),
			"status_code": rw.statusCode,
			"duration_ms": duration.Milliseconds(),
		}

		// Log based on status code
		logger := helpers.GetLogger()
		if rw.statusCode >= 500 {
			logger.WithFields(logFields).Error("HTTP request completed with server error")
		} else if rw.statusCode >= 400 {
			logger.WithFields(logFields).Warn("HTTP request completed with client error")
		} else {
			logger.WithFields(logFields).Info("HTTP request completed")
		}
	}
}

