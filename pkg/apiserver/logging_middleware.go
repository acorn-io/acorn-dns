package apiserver

import (
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// realIP get the real IP from http request
func realIP(req *http.Request) string {
	ra := req.RemoteAddr
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// loggingMiddleware logs the incoming HTTP request & its duration.
func loggingMiddleware(logger *logrus.Entry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if remoteAddr := realIP(r); remoteAddr != "" {
				logger = logger.WithField("remoteAddr", remoteAddr)
			}

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.WithError(err.(error)).WithField("status", http.StatusInternalServerError).Error("recovered error")
					logger.Errorf("Stack %s", debug.Stack())
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)

			if !strings.Contains(r.URL.EscapedPath(), "healthz") {
				requestLogger := logger.WithFields(logrus.Fields{
					"status":   wrapped.status,
					"method":   r.Method,
					"path":     r.URL.EscapedPath(),
					"duration": time.Since(start),
				})

				msg := fmt.Sprintf("handled: %d", wrapped.status)
				if wrapped.status >= 400 {
					requestLogger.Error(msg)
				} else {
					requestLogger.Debug(msg)
				}
			}
		}

		return http.HandlerFunc(fn)
	}
}
