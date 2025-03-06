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
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/game_rounds", func(r chi.Router) {
		r.Post("/", h.startGame)
		r.Get("/{code}", h.getGameRound)
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
}

func NewHandler(r services.UserService, r2 repositories.GameRoundRepository, r3 repositories.GameTypeRepository) *Handler {
	return &Handler{
		gameRoundRepository: r2,
		gameTypeRepository:  r3,
		userService:         r,
	}
}
