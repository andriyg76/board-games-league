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
	r.Route("/games/rounds", func(r chi.Router) {
		r.Post("/", h.startGame)
		r.Get("/{id}", h.getGameRound)
		r.Put("/{id}/finalize", h.finalizeGame)
		r.Put("/{id}/players/{userId}/score", h.updatePlayerScore)
	})

	r.Route("/games/types", func(r chi.Router) {
		r.Get("/", h.listGameTypes)
		r.Post("/", h.createGameType)
		r.Get("/{id}", h.getGameType)
		r.Put("/{id}", h.updateGameType)
		r.Delete("/{id}", h.deleteGameType)
	})
}

func NewHandler(r services.UserService, r2 repositories.GameRoundRepository, r3 repositories.GameTypeRepository) *Handler {
	return &Handler{
		gameRoundRepository: r2,
		gameTypeRepository:  r3,
		userService:         r,
	}
}
