package rewrite_to_root

import (
	"net/http"
	"os"
	"path/filepath"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RewriteToRoot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		if !fileExists(filepath.Join("frontend/build", r.URL.Path)) {
			// Rewrite the request to the root resource
			r.URL.Path = "/"
		}
		next.ServeHTTP(w, r)
	})
}
