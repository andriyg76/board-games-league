package wizardapi

import (
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/services"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	wizardRepo    repositories.WizardGameRepository
	gameRoundRepo repositories.GameRoundRepository
	gameTypeRepo  repositories.GameTypeRepository
	leagueService services.LeagueService
	userService   services.UserService
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/wizard/games", func(r chi.Router) {
		// Game Management
		r.Post("/", h.createGame)
		r.Get("/{code}", h.getGame)
		r.Get("/by-round/{gameRoundId}", h.getGameByRoundID)
		r.Delete("/{code}", h.deleteGame)

		// Round Operations
		r.Put("/{code}/rounds/{round}/bids", h.submitBids)
		r.Put("/{code}/rounds/{round}/results", h.submitResults)
		r.Post("/{code}/rounds/{round}/complete", h.completeRound)
		r.Post("/{code}/rounds/{round}/restart", h.restartRound)
		r.Put("/{code}/rounds/{round}/edit", h.editRound)

		// Game State
		r.Get("/{code}/scoreboard", h.getScoreboard)
		r.Post("/{code}/finalize", h.finalizeGame)
		r.Post("/{code}/next-round", h.nextRound)
		r.Post("/{code}/prev-round", h.prevRound)
	})
}

func NewHandler(
	wizardRepo repositories.WizardGameRepository,
	gameRoundRepo repositories.GameRoundRepository,
	gameTypeRepo repositories.GameTypeRepository,
	leagueService services.LeagueService,
	userService services.UserService,
) *Handler {
	return &Handler{
		wizardRepo:    wizardRepo,
		gameRoundRepo: gameRoundRepo,
		gameTypeRepo:  gameTypeRepo,
		leagueService: leagueService,
		userService:   userService,
	}
}
