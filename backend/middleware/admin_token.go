package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/andriyg76/bgl/utils"
)

// AdminTokenMiddleware validates the ADMIN_API_TOKEN from Authorization header
func AdminTokenMiddleware(next http.Handler) http.Handler {
	expectedToken := os.Getenv("ADMIN_API_TOKEN")
	if expectedToken == "" {
		// If token is not configured, reject all requests
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, nil, "ADMIN_API_TOKEN is not configured")
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != expectedToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
