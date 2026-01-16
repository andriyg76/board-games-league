package main

import (
	"context"
	stdlog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/andriyg76/bgl/api"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/frontendfs"
	"github.com/andriyg76/bgl/gameapi"
	bglmiddleware "github.com/andriyg76/bgl/middleware"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/userapi"
	"github.com/andriyg76/bgl/wizardapi"
	log "github.com/andriyg76/glog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gopkg.in/natefinch/lumberjack.v2"
)

// cacheAdapter adapts cache implementations to CleanableCache interface
type cacheAdapter struct {
	cleanExpired func() int
	size         func() int
}

func (a *cacheAdapter) CleanExpired() int {
	return a.cleanExpired()
}

func (a *cacheAdapter) Size() int {
	return a.size()
}

const (
	logMaxSizeMB   = 100
	logMaxBackups  = 7
	logMaxAgeDays  = 30
	logCompress    = true
	logPermissions = 0o755
)

func setupLogging() (*stdlog.Logger, func(), error) {
	logDir := strings.TrimSpace(os.Getenv("LOG_DIR"))
	if logDir == "" {
		return nil, func() {}, nil
	}

	if err := os.MkdirAll(logDir, logPermissions); err != nil {
		return nil, nil, err
	}

	serverWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "server.log"),
		MaxSize:    logMaxSizeMB,
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAgeDays,
		Compress:   logCompress,
	}
	accessWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "access.log"),
		MaxSize:    logMaxSizeMB,
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAgeDays,
		Compress:   logCompress,
	}

	log.ToWriters(serverWriter, serverWriter, log.INFO)
	accessLogger := stdlog.New(accessWriter, "", stdlog.LstdFlags)

	cleanup := func() {
		_ = serverWriter.Close()
		_ = accessWriter.Close()
	}

	return accessLogger, cleanup, nil
}

func main() {
	accessLogger, cleanup, err := setupLogging()
	if err != nil {
		stdlog.Fatalf("Failed to configure logging: %v", err)
	}
	defer cleanup()

	log.Info("Starting...")
	auth.LogSuperAdmins()

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

	wizardGameRepository, err := repositories.NewWizardGameRepository(mongodb)
	if err != nil {
		log.Fatal("Failed to initialise wizardGameRepository %v", err)
	}

	log.Info("Database connector initialised")

	// Initialize caches first (needed for services)
	idCodeCache := services.NewIdAndCodeCache()
	leagueCache := services.NewLeagueCache()
	membershipCache := services.NewMembershipCache()
	userCache := services.NewUserCache(idCodeCache)

	userService := services.NewUserService(userRepository, userCache)
	sessionService := services.NewSessionService(sessionRepository, userRepository)
	requestService := services.NewRequestService()
	geoIPService := services.NewGeoIPService()
	leagueService := services.NewLeagueService(
		leagueRepository,
		leagueMembershipRepository,
		leagueInvitationRepository,
		userRepository,
		gameRoundRepository,
	)

	gameTypeService := services.NewGameTypeService(gameTypeRepository)
	gameEventHub := services.NewGameEventHub()

	log.Info("Services initialised...")

	// Завантаження вбудованих типів ігор
	if err := gameTypeService.LoadBuiltInGames(context.Background()); err != nil {
		log.Warn("Failed to load built-in game types: %v", err)
	}

	// Initialize cache cleanup service
	cacheCleanupService := services.NewCacheCleanupService()

	// Register caches for cleanup using adapter functions
	// Since CleanExpired and Size are not part of business interfaces,
	// we create adapters that wrap the cache implementations
	// We also register the original caches for stats
	cacheCleanupService.RegisterCacheWithStats("IdAndCode", &cacheAdapter{
		cleanExpired: func() int {
			if c, ok := idCodeCache.(interface{ CleanExpired() int }); ok {
				return c.CleanExpired()
			}
			return 0
		},
		size: func() int {
			if c, ok := idCodeCache.(interface{ Size() int }); ok {
				return c.Size()
			}
			return 0
		},
	}, idCodeCache)
	cacheCleanupService.RegisterCacheWithStats("League", &cacheAdapter{
		cleanExpired: func() int {
			if c, ok := leagueCache.(interface{ CleanExpired() int }); ok {
				return c.CleanExpired()
			}
			return 0
		},
		size: func() int {
			if c, ok := leagueCache.(interface{ Size() int }); ok {
				return c.Size()
			}
			return 0
		},
	}, leagueCache)
	cacheCleanupService.RegisterCacheWithStats("Membership", &cacheAdapter{
		cleanExpired: func() int {
			if c, ok := membershipCache.(interface{ CleanExpired() int }); ok {
				return c.CleanExpired()
			}
			return 0
		},
		size: func() int {
			if c, ok := membershipCache.(interface{ Size() int }); ok {
				return c.Size()
			}
			return 0
		},
	}, membershipCache)
	cacheCleanupService.RegisterCacheWithStats("User", &cacheAdapter{
		cleanExpired: func() int {
			if c, ok := userCache.(interface{ CleanExpired() int }); ok {
				return c.CleanExpired()
			}
			return 0
		},
		size: func() int {
			if c, ok := userCache.(interface{ Size() int }); ok {
				return c.Size()
			}
			return 0
		},
	}, userCache)

	log.Info("Caches initialised and registered")

	// Create league middleware
	leagueMiddleware := bglmiddleware.NewLeagueMiddleware(leagueService, idCodeCache)

	gameApiHandler := gameapi.NewHandler(userService, gameRoundRepository, gameTypeRepository, leagueService, leagueMiddleware, idCodeCache)
	wizardApiHandler := wizardapi.NewHandler(wizardGameRepository, gameRoundRepository, gameTypeRepository, leagueService, userService, idCodeCache, gameEventHub)
	authHandler := auth.NewDefaultHandler(userRepository, sessionService, requestService)
	userProfileHandler := userapi.NewHandlerWithServices(userRepository, sessionRepository, geoIPService)
	diagnosticsHandler := api.NewDiagnosticsHandler(requestService, geoIPService, cacheCleanupService)

	log.Info("Handlers instances connector initialised")

	r := chi.NewRouter()
	if accessLogger != nil {
		r.Use(bglmiddleware.AccessLog(accessLogger))
	} else {
		r.Use(middleware.Logger)
	}

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

			gameApiHandler.RegisterRoutes(r, wizardApiHandler)
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

	// Start cache cleanup service
	cacheCleanupService.Start(ctx, 15*time.Minute)
	log.Info("Cache cleanup service started")

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
