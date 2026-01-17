package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/go-chi/chi/v5/middleware"
)

func AccessLog(logger *log.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	requestService := services.NewRequestService()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			userCode := "anonymous"
			if profile, ok := r.Context().Value("user").(*user_profile.UserProfile); ok && profile != nil && profile.Code != "" {
				userCode = profile.Code
			}

			requestInfo := requestService.ParseRequest(r)
			endpoint := requestInfo.BaseURL()
			if endpoint == "" {
				endpoint = r.Host
			}

			logger.Printf(
				"%s %s %s %d %dB %s user=%s",
				endpoint,
				r.Method,
				r.URL.RequestURI(),
				ww.Status(),
				ww.BytesWritten(),
				time.Since(start),
				userCode,
			)
		})
	}
}
