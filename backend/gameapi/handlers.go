package gameapi

import (
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	gameRoundRepository repositories.GameRoundRepository
	gameTypeRepository  repositories.GameTypeRepository
	userService         services.UserService
	leagueService       services.LeagueService
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/game_rounds", func(r chi.Router) {
		r.Get("/", h.listGameRounds)
		r.Post("/", h.startGame)
		r.Get("/{code}", h.getGameRound)
		r.Put("/{code}", h.updateGameRound)
		r.Put("/{code}/finalize", h.finalizeGame)
		r.Put("/{code}/players/{userId}/score", h.updatePlayerScore)
	})

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
		r.Post("/", h.createLeague)                       // Create league (superadmin)
		r.Get("/", h.listLeagues)                          // List leagues
		r.Post("/join/{token}", h.acceptInvitation)        // Accept invitation
		r.Get("/{code}", h.getLeague)                      // Get league details
		r.Get("/{code}/members", h.getLeagueMembers)       // Get league members
		r.Get("/{code}/standings", h.getLeagueStandings)   // Get league standings
		r.Post("/{code}/invitations", h.createInvitation)  // Create invitation
		r.Get("/{code}/invitations", h.listMyInvitations)  // List my active invitations
		r.Get("/{code}/invitations/expired", h.listMyExpiredInvitations)  // List my expired invitations
		r.Post("/{code}/invitations/{token}/cancel", h.cancelInvitation)  // Cancel invitation by token
		r.Post("/{code}/invitations/{token}/extend", h.extendInvitation)  // Extend invitation by 7 days
		r.Put("/{code}/members/{memberCode}/alias", h.updatePendingMemberAlias)  // Edit pending member alias
		r.Post("/{code}/ban/{userCode}", h.banUserFromLeague)     // Ban user (superadmin)
		r.Post("/{code}/archive", h.archiveLeague)         // Archive league (superadmin)
		r.Post("/{code}/unarchive", h.unarchiveLeague)     // Unarchive league (superadmin)
	})
}

func NewHandler(r services.UserService, r2 repositories.GameRoundRepository, r3 repositories.GameTypeRepository, leagueService services.LeagueService) *Handler {
	return &Handler{
		gameRoundRepository: r2,
		gameTypeRepository:  r3,
		userService:         r,
		leagueService:       leagueService,
	}
}
