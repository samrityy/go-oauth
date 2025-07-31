package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggerMiddleware logs the start and end of requests.
func LoggerMiddleware(logger *slog.Logger, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receivedTime := time.Now()
		logger.Info("request received", "method", r.Method, "path", r.URL.Path)
		handler(w, r)
		logger.Info("request complete", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(receivedTime).Milliseconds())
	}
}
