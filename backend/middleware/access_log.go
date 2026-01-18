package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/utils"
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

			userCode := utils.UserCodeFromContext(r.Context())
			if userCode == "" {
				userCode = "anonymous"
			}

			requestInfo := requestService.ParseRequest(r)
			endpoint := requestInfo.BaseURL()
			if endpoint == "" {
				endpoint = r.Host
			}

			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = "<none>"
			}

			remoteIP := requestInfo.ClientIP()
			if remoteIP == "" {
				remoteIP = "<unknown>"
			}

			logger.Printf(
				"%s %s %s %d %dB %s user=%s request_id=%s remote_ip=%s",
				endpoint,
				r.Method,
				r.URL.RequestURI(),
				ww.Status(),
				ww.BytesWritten(),
				time.Since(start),
				userCode,
				requestID,
				remoteIP,
			)
		})
	}
}
