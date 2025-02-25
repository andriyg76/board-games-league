package gameapi

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/utils"
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

func validateLabelsAndTeams(gameType *models.GameType) error {
	// Check for duplicate label names
	labelNames := make(map[string]bool)
	for _, label := range gameType.Labels {
		if labelNames[label.Name] {
			return fmt.Errorf("duplicate label name: %s", label.Name)
		}
		labelNames[label.Name] = true
	}

	// Check for duplicate team names
	teamNames := make(map[string]bool)
	for _, team := range gameType.Teams {
		if teamNames[team.Name] {
			return fmt.Errorf("duplicate team name: %s", team.Name)
		}
		teamNames[team.Name] = true
	}

	return nil
}

func (h *Handler) createGameType(w http.ResponseWriter, r *http.Request) {
	var gameType gameType
	if err := json.NewDecoder(r.Body).Decode(&gameType); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error parsing incoming request")
		return
	}

	gameTypeDb := models.GameType{
		Version:    1,
		MinPlayers: 1,
		MaxPlayers: 6,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	uiToDb(gameType, &gameTypeDb)

	if err := validateLabelsAndTeams(&gameTypeDb); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "validation error")
		return
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
	id, err := utils.GetIDFromChiURL(r, "code")
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
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error parsing incoming request, invalid code/id")
		return
	}

	var gameType gameType
	if err := json.NewDecoder(r.Body).Decode(&gameType); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error parsing incoming request payload")
		return
	}

	if gameTypeDb, err := h.gameTypeRepository.FindByID(r.Context(), id); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error reading gametype")
		return
	} else if gameTypeDb == nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error reading gametype")
		return
	} else {
		uiToDb(gameType, gameTypeDb)
		gameTypeDb.UpdatedAt = time.Now()

		if err := validateLabelsAndTeams(gameTypeDb); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "validation error")
			return
		}

		if err := h.gameTypeRepository.Update(r.Context(), gameTypeDb); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Update ")
			return
		}
		res := dbToUi(gameTypeDb)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error encoding response %v", res)
		}
	}
}

func (h *Handler) deleteGameType(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
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

func uiToDb(gt gameType, db *models.GameType) {
	db.Name = gt.Name
	db.ScoringType = gt.ScoringType
	db.Labels = uiToDbLabels(gt.Labels)
	db.Teams = uiToDbLabels(gt.Teams)
}

func dbToUi(gt *models.GameType) gameType {
	return gameType{
		Code:        utils.IdToCode(gt.ID),
		Name:        gt.Name,
		ScoringType: gt.ScoringType,
		Version:     gt.Version,
		Labels:      dbToUiLabels(gt.Labels),
		Teams:       dbToUiLabels(gt.Teams),
	}
}

func dbToUiLabels(ls []models.Label) []label {
	return utils.Map(ls, func(l models.Label) label {
		return label{
			name:  l.Name,
			color: l.Color,
			icon:  l.Icon,
		}
	})
}

func uiToDbLabels(labels []label) []models.Label {
	return utils.Map(labels, func(label label) models.Label {
		return models.Label{
			Name:  label.name,
			Color: label.color,
			Icon:  label.icon,
		}
	})
}

type label struct {
	name  string `json:"name"`
	color string `json:"color"`
	icon  string `json:"icon"`
}

type gameType struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	ScoringType string  `json:"scoring_type"`
	Version     int64   `json:"version"`
	Labels      []label `json:"labels"`
	Teams       []label `json:"teams"`
}
