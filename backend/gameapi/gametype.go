package gameapi

import (
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

func (h *Handler) listGameTypes(w http.ResponseWriter, r *http.Request) {
	gameTypes, err := h.gameTypeRepository.FindAll(r.Context())
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error reading games types")
	}

	response := utils.Map(gameTypes, func(gt *models.GameType) gameType {
		return dbToUi(gt)
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func dbToUi(gt *models.GameType) gameType {
	return gameType{
		Code:        utils.IdToCode(gt.ID),
		Name:        gt.Name,
		ScoringType: gt.ScoringType,
		Version:     gt.Version,
	}
}

func (h *Handler) createGameType(w http.ResponseWriter, r *http.Request) {
	var gameType gameType
	if err := json.NewDecoder(r.Body).Decode(&gameType); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error parsing incoming request")
		return
	}

	gameTypeDb := models.GameType{
		Version:     1,
		ScoringType: gameType.ScoringType,
		Name:        gameType.Name,
		Labels:      nil,
		Teams:       nil,
		MinPlayers:  1,
		MaxPlayers:  6,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := h.gameTypeRepository.Create(r.Context(), &gameTypeDb); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error storing gametype %v", gameTypeDb)
		return
	}

	gameType.Code = utils.IdToCode(gameTypeDb.ID)
	gameType.Version = gameTypeDb.Version

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(gameType); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error encoding response %v", gameType)
	}
}

func (h *Handler) getGameType(w http.ResponseWriter, r *http.Request) {
	id, err := utils.CodeToID(chi.URLParam(r, "code"))
	if err != nil {
		http.Error(w, "Invalid game type ID/code", http.StatusBadRequest)
		return
	}

	gameType, err := h.gameTypeRepository.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if gameType == nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "error navigating gametype id %s", id)
		return
	}

	res := dbToUi(gameType)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error encoding response %v", res)
	}
}

func (h *Handler) updateGameType(w http.ResponseWriter, r *http.Request) {
	id, err := utils.CodeToID(chi.URLParam(r, "code"))
	if err != nil {
		http.Error(w, "Invalid game type ID/code", http.StatusBadRequest)
		return
	}

	var gameType models.GameType
	if err := json.NewDecoder(r.Body).Decode(&gameType); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	gameType.ID = id
	if err := h.gameTypeRepository.Update(r.Context(), &gameType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) deleteGameType(w http.ResponseWriter, r *http.Request) {
	id, err := utils.CodeToID(chi.URLParam(r, "code"))
	if err != nil {
		http.Error(w, "Invalid game type ID/code", http.StatusBadRequest)
		return
	}

	// Change from gameRoundRepository to gameTypeRepository
	if err := h.gameTypeRepository.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type gameType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ScoringType string `json:"scoring_type"`
	Version     int64  `json:"version"`
}
