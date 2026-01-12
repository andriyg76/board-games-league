package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andriyg76/bgl/api"
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
		log.Fatal("Failed to initialise usersRepository %v", err)
	}

	gameRoundRepository, err := repositories.NewGameRoundRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise gameRoundRepository %v", err)
	}

	gameTypeRepository, err := repositories.NewGameTypeRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise gameTypeRepository %v", err)
	}

	sessionRepository, err := repositories.NewSessionRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise sessionRepository %v", err)
	}

	leagueRepository, err := repositories.NewLeagueRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise leagueRepository %v", err)
	}

	leagueMembershipRepository, err := repositories.NewLeagueMembershipRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise leagueMembershipRepository %v", err)
	}

	leagueInvitationRepository, err := repositories.NewLeagueInvitationRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise leagueInvitationRepository %v", err)
	}

	log.Info("Database connector initialised")

	userService := services.NewUserService(userRepository)
	sessionService := services.NewSessionService(sessionRepository, userRepository)
	requestService := services.NewRequestService()
	geoIPService := services.NewGeoIPService()
	leagueService := services.NewLeagueService(
		leagueRepository,
		leagueMembershipRepository,
		leagueInvitationRepository,
		gameRoundRepository,
		userRepository,
	)

	log.Info("Services initialised...")

	gameApiHandler := gameapi.NewHandler(userService, gameRoundRepository, gameTypeRepository, leagueService)
	authHandler := auth.NewDefaultHandler(userRepository, sessionService)
	userProfileHandler := userapi.NewHandlerWithServices(userRepository, sessionRepository, geoIPService)
	diagnosticsHandler := api.NewDiagnosticsHandler(requestService, geoIPService)

	log.Info("Handlers instances connector initialised")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/google", authHandler.HandleBeginLoginFlow)
		r.Post("/auth/google/callback", authHandler.GoogleCallbackHandler)
		r.Post("/auth/logout", authHandler.LogoutHandler)
		r.Post("/auth/refresh", authHandler.RefreshTokenHandler)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authHandler.Middleware)
			// Add your protected endpoints here
			r.Get("/user", userProfileHandler.GetUserHandler)
			r.Get("/user/sessions", userProfileHandler.GetUserSessionsHandler)

			r.Post("/user/alias/exist", userProfileHandler.CheckAliasUniquenessHandler)
			r.Put("/user/update", userProfileHandler.UpdateUser)

			r.Put("/admin/user/create", userProfileHandler.AdminCreateUserHandler)
			r.Get("/admin/diagnostics", diagnosticsHandler.GetDiagnosticsHandler)

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

	// Start background cleanup task for expired sessions
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sessionService.CleanupExpiredSessions(ctx); err != nil {
					_ = log.Error("Failed to cleanup expired sessions: %v", err)
				} else {
					log.Info("Cleaned up expired sessions")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Info("Listening on :8080... Ctrl+C to break server processing")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			_ = log.Error("Error attaching to listen socket %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Info("Shutting down server...")

	// Cancel background tasks
	cancel()

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = log.Error("Server shutdown error: %v", err)
	}

	log.Info("Exiting...")
}
