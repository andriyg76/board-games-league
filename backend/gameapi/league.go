package gameapi

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

// POST /api/leagues - Create league (superadmin only)
func (h *Handler) createLeague(w http.ResponseWriter, r *http.Request) {
	// Check if user is superadmin
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !auth.IsSuperAdmin(user) {
		http.Error(w, "Forbidden: superadmin only", http.StatusForbidden)
		return
	}

	// Parse request
	var req createLeagueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create league
	league, err := h.leagueService.CreateLeague(r.Context(), req.Name)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to create league")
		return
	}

	utils.WriteJSON(w, leagueToResponse(league), http.StatusCreated)
}

// GET /api/leagues - List all leagues
func (h *Handler) listLeagues(w http.ResponseWriter, r *http.Request) {
	leagues, err := h.leagueService.ListLeagues(r.Context())
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to list leagues")
		return
	}

	response := make([]leagueResponse, 0, len(leagues))
	for _, league := range leagues {
		response = append(response, leagueToResponse(league))
	}

	utils.WriteJSON(w, response, http.StatusOK)
}

// GET /api/leagues/:code - Get league details
func (h *Handler) getLeague(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	league, err := h.leagueService.GetLeague(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "league not found")
		return
	}

	utils.WriteJSON(w, leagueToResponse(league), http.StatusOK)
}

// GET /api/leagues/:code/members - Get league members
func (h *Handler) getLeagueMembers(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	members, err := h.leagueService.GetLeagueMembers(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to get league members")
		return
	}

	response := make([]playerResponse, 0, len(members))
	for _, member := range members {
		response = append(response, playerResponse{
			Code:   utils.IdToCode(member.ID),
			Alias:  member.Alias,
			Avatar: member.Avatar,
			Name:   member.Name,
		})
	}

	utils.WriteJSON(w, response, http.StatusOK)
}

// GET /api/leagues/:code/standings - Get league standings
func (h *Handler) getLeagueStandings(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	standings, err := h.leagueService.GetLeagueStandings(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to get standings")
		return
	}

	response := make([]standingResponse, 0, len(standings))
	for _, standing := range standings {
		response = append(response, standingResponse{
			UserID:              utils.IdToCode(standing.UserID),
			UserName:            standing.UserName,
			UserAvatar:          standing.UserAvatar,
			TotalPoints:         standing.TotalPoints,
			GamesPlayed:         standing.GamesPlayed,
			GamesModerated:      standing.GamesModerated,
			FirstPlaceCount:     standing.FirstPlaceCount,
			SecondPlaceCount:    standing.SecondPlaceCount,
			ThirdPlaceCount:     standing.ThirdPlaceCount,
			ParticipationPoints: standing.ParticipationPoints,
			PositionPoints:      standing.PositionPoints,
			ModerationPoints:    standing.ModerationPoints,
		})
	}

	utils.WriteJSON(w, response, http.StatusOK)
}

// POST /api/leagues/:code/invitations - Create invitation
func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Get current user
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := utils.CodeToID(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	// Check if user is a member of the league
	isMember, err := h.leagueService.IsUserMember(r.Context(), leagueID, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to check membership")
		return
	}
	if !isMember {
		http.Error(w, "Forbidden: not a member of this league", http.StatusForbidden)
		return
	}

	// Create invitation
	invitation, err := h.leagueService.CreateInvitation(r.Context(), leagueID, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to create invitation")
		return
	}

	utils.WriteJSON(w, invitationToResponse(invitation), http.StatusCreated)
}

// POST /api/leagues/join/:token - Accept invitation
func (h *Handler) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "Invalid invitation token", http.StatusBadRequest)
		return
	}

	// Get current user
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := utils.CodeToID(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	// Accept invitation
	league, err := h.leagueService.AcceptInvitation(r.Context(), token, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "failed to accept invitation")
		return
	}

	utils.WriteJSON(w, leagueToResponse(league), http.StatusOK)
}

// POST /api/leagues/:code/ban/:userCode - Ban user from league (superadmin only)
func (h *Handler) banUserFromLeague(w http.ResponseWriter, r *http.Request) {
	// Check if user is superadmin
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !auth.IsSuperAdmin(user) {
		http.Error(w, "Forbidden: superadmin only", http.StatusForbidden)
		return
	}

	// Get league and user IDs
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	userID, err := utils.GetIDFromChiURL(r, "userCode")
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	// Ban user
	if err := h.leagueService.BanUserFromLeague(r.Context(), leagueID, userID); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to ban user")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST /api/leagues/:code/archive - Archive league (superadmin only)
func (h *Handler) archiveLeague(w http.ResponseWriter, r *http.Request) {
	// Check if user is superadmin
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !auth.IsSuperAdmin(user) {
		http.Error(w, "Forbidden: superadmin only", http.StatusForbidden)
		return
	}

	// Get league ID
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Archive league
	if err := h.leagueService.ArchiveLeague(r.Context(), leagueID); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to archive league")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST /api/leagues/:code/unarchive - Unarchive league (superadmin only)
func (h *Handler) unarchiveLeague(w http.ResponseWriter, r *http.Request) {
	// Check if user is superadmin
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !auth.IsSuperAdmin(user) {
		http.Error(w, "Forbidden: superadmin only", http.StatusForbidden)
		return
	}

	// Get league ID
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Unarchive league
	if err := h.leagueService.UnarchiveLeague(r.Context(), leagueID); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to unarchive league")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Request/Response types

type createLeagueRequest struct {
	Name string `json:"name"`
}

type leagueResponse struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type playerResponse struct {
	Code   string `json:"code"`
	Alias  string `json:"alias"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

type standingResponse struct {
	UserID              string `json:"user_id"`
	UserName            string `json:"user_name"`
	UserAvatar          string `json:"user_avatar"`
	TotalPoints         int64  `json:"total_points"`
	GamesPlayed         int    `json:"games_played"`
	GamesModerated      int    `json:"games_moderated"`
	FirstPlaceCount     int    `json:"first_place_count"`
	SecondPlaceCount    int    `json:"second_place_count"`
	ThirdPlaceCount     int    `json:"third_place_count"`
	ParticipationPoints int64  `json:"participation_points"`
	PositionPoints      int64  `json:"position_points"`
	ModerationPoints    int64  `json:"moderation_points"`
}

type invitationResponse struct {
	Code      string `json:"code"`
	Token     string `json:"token"`
	LeagueID  string `json:"league_id"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// Helper functions

func leagueToResponse(league *models.League) leagueResponse {
	return leagueResponse{
		Code:      utils.IdToCode(league.ID),
		Name:      league.Name,
		Status:    string(league.Status),
		CreatedAt: league.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: league.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func invitationToResponse(inv *models.LeagueInvitation) invitationResponse {
	return invitationResponse{
		Code:      utils.IdToCode(inv.ID),
		Token:     inv.Token,
		LeagueID:  utils.IdToCode(inv.LeagueID),
		ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// getIDFromURL extracts and validates ObjectID from URL parameter
func getIDFromURL(r *http.Request, paramName string) (primitive.ObjectID, error) {
	code := chi.URLParam(r, paramName)
	if code == "" {
		return primitive.NilObjectID, fmt.Errorf("missing %s parameter", paramName)
	}

	id, err := utils.CodeToID(code)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid %s: %w", paramName, err)
	}

	return id, nil
}
