package main

import (
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/frontendfs"
	"github.com/andriyg76/bgl/gameapi"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/userapi"
	log "github.com/andriyg76/glog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	log.Info("Starting...")

	mongourl := os.Getenv("MONGODB_URI")
	mongodb, err := db.NewMongoDB(mongourl)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:, connection: %s %v", mongourl, err)
	}

	userRepository, err := repositories.NewUserRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise usersRepository")
	}

	gameRoundRepository, err := repositories.NewGameRoundRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise gameRoundRepository")
	}

	gameTypeRepository, err := repositories.NewGameTypeRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise gameRoundRepository")
	}
	log.Info("Database connector initialised")

	userService := services.NewUserService(userRepository)

	log.Info("Services initialised...")

	gameApiHandler := gameapi.NewHandler(userService, gameRoundRepository, gameTypeRepository)
	authHandler := auth.NewDefaultHandler(userRepository)
	userProfileHandler := userapi.NewHandler(userRepository)

	log.Info("Handlers instances connector initialised")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/google", authHandler.HandleBeginLoginFlow)
		r.Post("/auth/google/callback", authHandler.GoogleCallbackHandler)
		r.Post("/auth/logout", authHandler.LogoutHandler)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authHandler.Middleware)
			// Add your protected endpoints here
			r.Get("/user", userProfileHandler.GetUserHandler)

			r.Post("/user/alias/exist", userProfileHandler.CheckAliasUniquenessHandler)
			r.Put("/user/update", userProfileHandler.GetUserHandler)

			r.Put("/admin/user/create", userProfileHandler.AdminCreateUserHandler)

			gameApiHandler.RegisterRoutes(r)
		})
		r.Handle("/*", http.NotFoundHandler())
	})

	// Reverse proxy for other requests
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" {
		log.Info("Running revese proxy to url: %s", frontendURL)
		proxyURL, err := url.Parse(frontendURL)
		if err != nil {
			_ = log.Error("Can't parse frontend url: %v", err)
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
	log.Info("Http routers configured...")

	listenAddress := ":8080"

	log.Info("Listening on %s... Ctrl+C to break server processing", listenAddress)

	if err := http.ListenAndServe(listenAddress, r); err != nil {
		_ = log.Error("Error attaching to listen socket %v", err)
		os.Exit(1)
	}
	log.Info("Exiting...")
}
