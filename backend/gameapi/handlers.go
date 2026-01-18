package gameapi

import (
	"github.com/andriyg76/bgl/middleware"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/go-chi/chi/v5"
)

type WizardHandler interface {
	RegisterWizardLeagueRoutes(r chi.Router)
}

type Handler struct {
	gameRoundRepository repositories.GameRoundRepository
	gameTypeRepository  repositories.GameTypeRepository
	userService         services.UserService
	leagueService       services.LeagueService
	leagueMiddleware    *middleware.LeagueMiddleware
	idCodeCache         services.IdAndCodeCache
}

func (h *Handler) RegisterRoutes(r chi.Router, wizardHandler WizardHandler) {
	r.Route("/game_types", func(r chi.Router) {
		r.Get("/", h.listGameTypes)
		r.Post("/", h.createGameType)
		r.Get("/{code}", h.getGameType)
		r.Put("/{code}", h.updateGameType)
		r.Delete("/{code}", h.deleteGameType)
	})

	r.Route("/players", func(r chi.Router) {
		r.Get("/", h.listPlayers)
		r.Get("/{code}", h.getPlayer)
		r.Get("/i_am", h.iAm)
	})

	r.Route("/leagues", func(r chi.Router) {
		r.Post("/", h.createLeague)                         // Create league (superadmin)
		r.Get("/", h.listLeagues)                           // List leagues
		r.Post("/join/{token}", h.acceptInvitation)         // Accept invitation

		// Routes that require league membership - apply middleware
		r.Route("/{code}", func(r chi.Router) {
			if h.leagueMiddleware != nil {
				r.Use(h.leagueMiddleware.RequireLeagueMembership)
			}

			r.Get("/", h.getLeague)                            // Get league details
			r.Get("/members", h.getLeagueMembers)              // Get league members
			r.Get("/standings", h.getLeagueStandings)          // Get league standings
			r.Get("/suggested-players", h.getSuggestedPlayers) // Get suggested players for game

			// Game rounds routes - all under league
			r.Route("/game_rounds", func(r chi.Router) {
				r.Get("/", h.listGameRounds)                                              // List game rounds for league
				r.Post("/", h.startGame)                                                  // Create game round in league
				r.Get("/{code}", h.getGameRound)                                          // Get game round by code
				r.Put("/{code}", h.updateGameRound)                                       // Update game round
				r.Put("/{code}/roles", h.updateRoles)                                     // Update player roles
				r.Put("/{code}/scores", h.updateScores)                                   // Update player scores
				r.Put("/{code}/status", h.updateRoundStatus)                              // Update round status
				r.Put("/{code}/finalize", h.finalizeGame)                                 // Finalize game
				r.Put("/{gameRoundCode}/players/{playerCode}/score", h.updatePlayerScore) // Update player score
			})

			r.Post("/invitations", h.createInvitation)                       // Create invitation
			r.Get("/invitations", h.listMyInvitations)                       // List my active invitations
			r.Get("/invitations/expired", h.listMyExpiredInvitations)        // List my expired invitations
			r.Post("/invitations/{token}/cancel", h.cancelInvitation)        // Cancel invitation by token
			r.Post("/invitations/{token}/extend", h.extendInvitation)        // Extend invitation by 7 days
			r.Put("/members/{memberCode}/alias", h.updatePendingMemberAlias) // Edit pending member alias
			r.Post("/memberships", h.createMembershipForSuperAdmin)          // Create membership for superadmin (superadmin only)
			r.Post("/ban/{userCode}", h.banUserFromLeague)                   // Ban user (superadmin)
			r.Post("/unban/{userCode}", h.unbanUserFromLeague)               // Unban user (superadmin)
			r.Post("/archive", h.archiveLeague)                              // Archive league (superadmin)
			r.Post("/unarchive", h.unarchiveLeague)                          // Unarchive league (superadmin)

			// Wizard routes
			if wizardHandler != nil {
				r.Route("/wizard/games", func(r chi.Router) {
					wizardHandler.RegisterWizardLeagueRoutes(r)
				})
			}
		})
	})
}

// RegisterPublicRoutes registers public (no auth) endpoints.
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/leagues/join/{token}/preview", h.previewInvitation) // Preview invitation (public)
}

func NewHandler(r services.UserService, r2 repositories.GameRoundRepository, r3 repositories.GameTypeRepository, leagueService services.LeagueService, leagueMiddleware *middleware.LeagueMiddleware, idCodeCache services.IdAndCodeCache) *Handler {
	return &Handler{
		gameRoundRepository: r2,
		gameTypeRepository:  r3,
		userService:         r,
		leagueService:       leagueService,
		leagueMiddleware:    leagueMiddleware,
		idCodeCache:         idCodeCache,
	}
}
