package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(writer http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: writer,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code and calls the underlying WriteHeader
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware logs HTTP requests with detailed information
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		// Wrap the response writer to capture status code
		wrappedWriter := newResponseWriter(writer)

		// Log incoming request
		log.Info().
			Str("method", request.Method).
			Str("path", request.URL.Path).
			Str("remote_addr", request.RemoteAddr).
			Str("user_agent", request.UserAgent()).
			Msg("Incoming request")

		// Call the next handler
		next.ServeHTTP(wrappedWriter, request)

		// Calculate request duration
		duration := time.Since(startTime)

		// Determine log level based on status code
		var logEvent *zerolog.Event
		statusCode := wrappedWriter.statusCode

		switch {
		case statusCode >= 500:
			logEvent = log.Error()
		case statusCode >= 400:
			logEvent = log.Warn()
		default:
			logEvent = log.Info()
		}

		// Log request completion with details
		logEvent.
			Str("method", request.Method).
			Str("path", request.URL.Path).
			Int("status", statusCode).
			Dur("duration", duration).
			Str("duration_ms", duration.String()).
			Msg("Request completed")
	})
}
