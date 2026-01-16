package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func AccessLog(logger *log.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Printf(
				"%s %s %s %d %dB %s",
				r.RemoteAddr,
				r.Method,
				r.URL.RequestURI(),
				ww.Status(),
				ww.BytesWritten(),
				time.Since(start),
			)
		})
	}
}
