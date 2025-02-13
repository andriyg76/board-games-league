package main

import (
	"github.com/andriyg76/bgl/frontendfs"
	log "github.com/andriyg76/glog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/google", googleAuthHandler)
		r.Get("/auth/google/callback", googleCallbackHandler)
		r.Post("/auth/logout", logoutHandler)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware)
			// Add your protected endpoints here
			r.Get("/user", getUserHandler)
		})
		r.Handle("/*", http.NotFoundHandler())
	})

	// Reverse proxy for other requests
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" {
		log.Info("Running revese proxy to url: %s", frontendURL)
		proxyURL, err := url.Parse(frontendURL)
		if err != nil {
			log.Error("Can't parse frontend url: %v", err)
			os.Exit(1)
		} else {
			proxy := httputil.NewSingleHostReverseProxy(proxyURL)
			r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
				proxy.ServeHTTP(w, r)
			})
		}
	} else {
		log.Info("Serving frontend from embed fs")
		// Serve static files from the frontend build directory
		r.Handle("/*", http.StripPrefix("/", http.FileServer(http.FS(frontendfs.Frontend))))
	}

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Error("Error attaching to listen socket %v", err)
		os.Exit(1)
	}
	log.Info("Exiting...")
}
