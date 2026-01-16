package gameapi

import (
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"net/http"
)

func (h *Handler) listPlayers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.FindAll(r.Context())
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching users")
		return
	}

	players := make([]player, 0, len(users))
	for _, u := range users {
		players = append(players, player{
			Code:   utils.IdToCode(u.ID),
			Alias:  u.Alias,
			Avatar: u.Avatar,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(players); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error encoding response")
	}
}

func (h *Handler) getPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, err, "invalid player code")
		return
	}

	user, err := h.userService.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching user")
		return
	}
	writePlayer(w, user, utils.IdToCode(id))
}

func writePlayer(w http.ResponseWriter, user *models.User, code string) {
	if user == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	p := player{
		Code:   code,
		Alias:  user.Alias,
		Avatar: user.Avatar,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(p); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error encoding response")
	}
}

type player struct {
	Code   string `json:"code"`
	Alias  string `json:"alias"`
	Avatar string `json:"avatar,omitempty"`
}

func (h *Handler) iAm(w http.ResponseWriter, r *http.Request) {
	profile, err := user_profile.GetUserProfile(r)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching user")
		return
	}
	if profile == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching user")
		return
	}
	writePlayer(w, user, profile.Code)
}
