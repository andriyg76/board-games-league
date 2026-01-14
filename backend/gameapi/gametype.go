package gameapi

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"net/http"
	"time"
)

// API структури

type roleAPI struct {
	Key      string            `json:"key"`
	Names    map[string]string `json:"names"`
	Color    string            `json:"color"`
	Icon     string            `json:"icon"`
	RoleType string            `json:"role_type"`
}

type gameTypeAPI struct {
	Code        string            `json:"code"`
	Key         string            `json:"key"`
	Names       map[string]string `json:"names"`
	Icon        string            `json:"icon"`
	ScoringType string            `json:"scoring_type"`
	Roles       []roleAPI         `json:"roles"`
	MinPlayers  int               `json:"min_players"`
	MaxPlayers  int               `json:"max_players"`
	BuiltIn     bool              `json:"built_in"`
	Version     int64             `json:"version"`
}

// Helper functions

func requireSuperAdmin(w http.ResponseWriter, r *http.Request) bool {
	profile, err := user_profile.GetUserProfile(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	if !auth.IsSuperAdminByExternalIDs(profile.ExternalIDs) {
		http.Error(w, "Forbidden: superadmin access required", http.StatusForbidden)
		return false
	}
	return true
}

// Handlers

func (h *Handler) listGameTypes(w http.ResponseWriter, r *http.Request) {
	gameTypes, err := h.gameTypeRepository.FindAll(r.Context())
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error reading games types")
		return
	}

	response := utils.Map(gameTypes, func(gt *models.GameType) gameTypeAPI {
		return dbToAPI(gt)
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) createGameType(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}

	var gt gameTypeAPI
	if err := json.NewDecoder(r.Body).Decode(&gt); err != nil {
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

	apiToDb(gt, &gameTypeDb)

	if err := validateGameType(&gameTypeDb); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "validation error")
		return
	}

	if err := h.gameTypeRepository.Create(r.Context(), &gameTypeDb); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error storing gametype %v", gameTypeDb)
		return
	}

	response := dbToAPI(&gameTypeDb)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error encoding response %v", response)
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

	res := dbToAPI(gameType)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error encoding response %v", res)
	}
}

func (h *Handler) updateGameType(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}

	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error parsing incoming request, invalid code/id")
		return
	}

	var gt gameTypeAPI
	if err := json.NewDecoder(r.Body).Decode(&gt); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error parsing incoming request payload")
		return
	}

	gameTypeDb, err := h.gameTypeRepository.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error reading gametype")
		return
	}
	if gameTypeDb == nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, nil, "gametype not found")
		return
	}

	apiToDb(gt, gameTypeDb)
	gameTypeDb.UpdatedAt = time.Now()

	if err := validateGameType(gameTypeDb); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "validation error")
		return
	}

	if err := h.gameTypeRepository.Update(r.Context(), gameTypeDb); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "update error")
		return
	}

	res := dbToAPI(gameTypeDb)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "error encoding response %v", res)
	}
}

func (h *Handler) deleteGameType(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}

	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game type ID/code", http.StatusBadRequest)
		return
	}

	// Перевірка чи це вбудований тип
	gameType, err := h.gameTypeRepository.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if gameType != nil && gameType.BuiltIn {
		http.Error(w, "Cannot delete built-in game type", http.StatusForbidden)
		return
	}

	if err := h.gameTypeRepository.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Validation

func validateGameType(gt *models.GameType) error {
	// Перевірка обов'язкових полів
	if gt.Key == "" && len(gt.Names) == 0 {
		return fmt.Errorf("game type key or name is required")
	}

	// Перевірка min/max players
	if gt.MinPlayers < 1 {
		gt.MinPlayers = 1
	}
	if gt.MaxPlayers < gt.MinPlayers {
		gt.MaxPlayers = gt.MinPlayers
	}

	// Перевірка дублікатів ключів ролей
	roleKeys := make(map[string]bool)
	for _, role := range gt.Roles {
		if role.Key == "" {
			continue
		}
		if roleKeys[role.Key] {
			return fmt.Errorf("duplicate role key: %s", role.Key)
		}
		roleKeys[role.Key] = true
	}

	return nil
}

// Converters

func apiToDb(api gameTypeAPI, db *models.GameType) {
	if api.Key != "" {
		db.Key = api.Key
	}
	if api.Names != nil {
		db.Names = api.Names
	}
	if api.Icon != "" {
		db.Icon = api.Icon
	}
	db.ScoringType = models.ScoringType(api.ScoringType)

	if api.Roles != nil {
		db.Roles = make([]models.Role, len(api.Roles))
		for i, r := range api.Roles {
			db.Roles[i] = models.Role{
				Key:      r.Key,
				Names:    r.Names,
				Color:    r.Color,
				Icon:     r.Icon,
				RoleType: models.RoleType(r.RoleType),
			}
		}
	}

	if api.MinPlayers > 0 {
		db.MinPlayers = api.MinPlayers
	}
	if api.MaxPlayers > 0 {
		db.MaxPlayers = api.MaxPlayers
	}

	// Очищаємо застарілі поля
	db.Labels = nil
	db.Teams = nil
	db.Name = ""
}

func dbToAPI(db *models.GameType) gameTypeAPI {
	roles := make([]roleAPI, len(db.Roles))
	for i, r := range db.Roles {
		roles[i] = roleAPI{
			Key:      r.Key,
			Names:    r.Names,
			Color:    r.Color,
			Icon:     r.Icon,
			RoleType: string(r.RoleType),
		}
	}

	// Підтримка застарілих даних - конвертуємо Labels/Teams в Roles
	if len(roles) == 0 && (len(db.Labels) > 0 || len(db.Teams) > 0) {
		for _, l := range db.Labels {
			roles = append(roles, roleAPI{
				Key:      l.Name,
				Names:    map[string]string{"en": l.Name},
				Color:    l.Color,
				Icon:     l.Icon,
				RoleType: string(models.RoleTypeOptionalOne),
			})
		}
		for _, t := range db.Teams {
			roles = append(roles, roleAPI{
				Key:      t.Name,
				Names:    map[string]string{"en": t.Name},
				Color:    t.Color,
				Icon:     t.Icon,
				RoleType: string(models.RoleTypeMultiple),
			})
		}
	}

	// Підтримка застарілого поля Name
	names := db.Names
	if names == nil {
		names = make(map[string]string)
	}
	if db.Name != "" && names["en"] == "" {
		names["en"] = db.Name
	}

	return gameTypeAPI{
		Code:        utils.IdToCode(db.ID),
		Key:         db.Key,
		Names:       names,
		Icon:        db.Icon,
		ScoringType: string(db.ScoringType),
		Roles:       roles,
		MinPlayers:  db.MinPlayers,
		MaxPlayers:  db.MaxPlayers,
		BuiltIn:     db.BuiltIn,
		Version:     db.Version,
	}
}
