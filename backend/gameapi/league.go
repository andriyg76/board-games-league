package gameapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to create league")
		return
	}

	utils.WriteJSON(r, w, h.leagueToResponse(league), http.StatusCreated)
}

// GET /api/leagues - List all leagues
func (h *Handler) listLeagues(w http.ResponseWriter, r *http.Request) {
	leagues, err := h.leagueService.ListLeagues(r.Context())
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to list leagues")
		return
	}

	response := make([]leagueResponse, 0, len(leagues))
	for _, league := range leagues {
		response = append(response, h.leagueToResponse(league))
	}

	utils.WriteJSON(r, w, response, http.StatusOK)
}

// GET /api/leagues/:code - Get league details
func (h *Handler) getLeague(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	league, err := h.leagueService.GetLeague(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, err, "league not found")
		return
	}

	utils.WriteJSON(r, w, h.leagueToResponse(league), http.StatusOK)
}

// GET /api/leagues/:code/members - Get league members
func (h *Handler) getLeagueMembers(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	members, err := h.leagueService.GetLeagueMemberships(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to get league members")
		return
	}

	response := make([]memberResponse, 0, len(members))
	for _, member := range members {
		membershipIdAndCode := h.idCodeCache.GetByID(member.MembershipID)
		userIdAndCode := h.idCodeCache.GetByID(member.UserID)
		resp := memberResponse{
			Code:       membershipIdAndCode.Code,
			UserID:     userIdAndCode.Code,
			UserName:   member.UserName,
			UserAvatar: member.UserAvatar,
			Alias:      member.UserAlias,
			Status:     string(member.Status),
			JoinedAt:   member.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if member.InvitationToken != "" {
			resp.InvitationToken = member.InvitationToken
		}
		response = append(response, resp)
	}

	utils.WriteJSON(r, w, response, http.StatusOK)
}

// GET /api/leagues/:code/standings - Get league standings
func (h *Handler) getLeagueStandings(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	standings, err := h.leagueService.GetLeagueStandings(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to get standings")
		return
	}

	response := make([]standingResponse, 0, len(standings))
	for _, standing := range standings {
		userIdAndCode := h.idCodeCache.GetByID(standing.UserID)
		response = append(response, standingResponse{
			UserID:              userIdAndCode.Code,
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

	utils.WriteJSON(r, w, response, http.StatusOK)
}

// CreateInvitationRequest represents the request body for creating an invitation
type CreateInvitationRequest struct {
	Alias string `json:"alias"`
}

// POST /api/leagues/:code/invitations - Create invitation
func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Alias == "" {
		http.Error(w, "Alias is required", http.StatusBadRequest)
		return
	}

	// Get current user
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	// Check if user is a member of the league or superadmin
	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	isSuperAdmin := auth.IsSuperAdmin(user)
	if !isSuperAdmin {
		isMember, err := h.leagueService.IsUserMember(r.Context(), leagueID, userID)
		if err != nil {
			utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to check membership")
			return
		}
		if !isMember {
			http.Error(w, "Forbidden: not a member of this league", http.StatusForbidden)
			return
		}
	}

	// Create invitation with alias
	invitation, err := h.leagueService.CreateInvitation(r.Context(), leagueID, userID, req.Alias)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to create invitation")
		return
	}

	utils.WriteJSON(r, w, h.invitationToResponse(invitation), http.StatusCreated)
}

// GET /api/leagues/:code/invitations - List my active invitations
func (h *Handler) listMyInvitations(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
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

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	// Get my invitations
	invitations, err := h.leagueService.ListMyInvitations(r.Context(), leagueID, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to list invitations")
		return
	}

	response := make([]invitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response = append(response, h.invitationToResponse(inv))
	}

	utils.WriteJSON(r, w, response, http.StatusOK)
}

// POST /api/leagues/:code/invitations/:token/cancel - Cancel invitation by token
func (h *Handler) cancelInvitation(w http.ResponseWriter, r *http.Request) {
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

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	// Cancel the invitation
	if err := h.leagueService.CancelInvitation(r.Context(), token, userID); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, err, "failed to cancel invitation")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /api/leagues/:code/invitations/expired - List my expired invitations
func (h *Handler) listMyExpiredInvitations(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
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

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	invitations, err := h.leagueService.ListMyExpiredInvitations(r.Context(), leagueID, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to list expired invitations")
		return
	}

	response := make([]invitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response = append(response, h.invitationToResponse(inv))
	}

	utils.WriteJSON(r, w, response, http.StatusOK)
}

// POST /api/leagues/:code/invitations/:token/extend - Extend invitation by 7 days
func (h *Handler) extendInvitation(w http.ResponseWriter, r *http.Request) {
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

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	invitation, err := h.leagueService.ExtendInvitation(r.Context(), token, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, err, "failed to extend invitation")
		return
	}

	utils.WriteJSON(r, w, h.invitationToResponse(invitation), http.StatusOK)
}

// UpdatePendingMemberAliasRequest represents the request body for updating a pending member's alias
type UpdatePendingMemberAliasRequest struct {
	Alias string `json:"alias"`
}

// PUT /api/leagues/:code/members/:memberCode/alias - Edit pending member alias
func (h *Handler) updatePendingMemberAlias(w http.ResponseWriter, r *http.Request) {
	memberCode := chi.URLParam(r, "memberCode")
	if memberCode == "" {
		http.Error(w, "Invalid member code", http.StatusBadRequest)
		return
	}

	membershipIdAndCode, err := h.idCodeCache.GetByCode(memberCode)
	if err != nil {
		http.Error(w, "Invalid member code", http.StatusBadRequest)
		return
	}
	membershipID := membershipIdAndCode.ID

	// Parse request body
	var req UpdatePendingMemberAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Alias == "" {
		http.Error(w, "Alias is required", http.StatusBadRequest)
		return
	}

	// Get current user
	profile, err := user_profile.GetUserProfile(r)
	if err != nil || profile == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	if err := h.leagueService.UpdatePendingMemberAlias(r.Context(), membershipID, userID, req.Alias); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, err, "failed to update alias")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// InvitationPreviewResponse represents public invitation preview data
type InvitationPreviewResponse struct {
	LeagueName   string `json:"league_name"`
	InviterAlias string `json:"inviter_alias"`
	PlayerAlias  string `json:"player_alias"`
	ExpiresAt    string `json:"expires_at"`
	Status       string `json:"status"` // valid, expired, used
}

// GET /api/leagues/join/:token/preview - Preview invitation (public, no auth required)
func (h *Handler) previewInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "Invalid invitation token", http.StatusBadRequest)
		return
	}

	preview, err := h.leagueService.PreviewInvitation(r.Context(), token)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, err, "invitation not found")
		return
	}

	response := InvitationPreviewResponse{
		LeagueName:   preview.LeagueName,
		InviterAlias: preview.InviterAlias,
		PlayerAlias:  preview.PlayerAlias,
		ExpiresAt:    preview.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		Status:       preview.Status,
	}

	utils.WriteJSON(r, w, response, http.StatusOK)
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

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	// Accept invitation
	league, err := h.leagueService.AcceptInvitation(r.Context(), token, userID)
	if err != nil {
		// Check if user is already a member - return league code for redirect
		if leagueCode, ok := services.IsAlreadyMemberError(err); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			response := map[string]string{
				"error":       err.Error(),
				"league_code": leagueCode,
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, err, "failed to accept invitation")
		return
	}

	utils.WriteJSON(r, w, h.leagueToResponse(league), http.StatusOK)
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
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	userID, err := h.getIDFromChiURL(r, "userCode")
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	// Prevent superadmin from banning themselves
	if user.ID == userID {
		http.Error(w, "Cannot ban yourself", http.StatusBadRequest)
		return
	}

	// Ban user
	if err := h.leagueService.BanUserFromLeague(r.Context(), leagueID, userID); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to ban user")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST /api/leagues/:code/unban/:userCode - Unban user from league (superadmin only)
func (h *Handler) unbanUserFromLeague(w http.ResponseWriter, r *http.Request) {
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
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	userID, err := h.getIDFromChiURL(r, "userCode")
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	// Unban user
	if err := h.leagueService.UnbanUserFromLeague(r.Context(), leagueID, userID); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to unban user")
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
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Archive league
	if err := h.leagueService.ArchiveLeague(r.Context(), leagueID); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to archive league")
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
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Unarchive league
	if err := h.leagueService.UnarchiveLeague(r.Context(), leagueID); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to unarchive league")
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

type memberResponse struct {
	Code            string `json:"code"`
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	UserAvatar      string `json:"user_avatar"`
	Alias           string `json:"alias"`
	Status          string `json:"status"`
	JoinedAt        string `json:"joined_at"`
	InvitationToken string `json:"invitation_token,omitempty"` // Token of the invitation if exists
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
	Token          string `json:"token"`
	LeagueCode     string `json:"league_code"`
	PlayerAlias    string `json:"player_alias"`
	MembershipCode string `json:"membership_code,omitempty"`
	ExpiresAt      string `json:"expires_at"`
	CreatedAt      string `json:"created_at"`
}

// Helper functions

// getIDFromChiURL extracts a MongoDB ObjectID from a Chi URL parameter using cache
func (h *Handler) getIDFromChiURL(r *http.Request, codeParam string) (primitive.ObjectID, error) {
	code := chi.URLParam(r, codeParam)
	if code == "" {
		return primitive.NilObjectID, errors.New("code parameter is required")
	}
	idAndCode, err := h.idCodeCache.GetByCode(code)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return idAndCode.ID, nil
}

func (h *Handler) leagueToResponse(league *models.League) leagueResponse {
	leagueIdAndCode := h.idCodeCache.GetByID(league.ID)
	return leagueResponse{
		Code:      leagueIdAndCode.Code,
		Name:      league.Name,
		Status:    string(league.Status),
		CreatedAt: league.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: league.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) invitationToResponse(inv *models.LeagueInvitation) invitationResponse {
	leagueIdAndCode := h.idCodeCache.GetByID(inv.LeagueID)
	resp := invitationResponse{
		Token:       inv.Token,
		LeagueCode:  leagueIdAndCode.Code,
		PlayerAlias: inv.PlayerAlias,
		ExpiresAt:   inv.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !inv.MembershipID.IsZero() {
		membershipIdAndCode := h.idCodeCache.GetByID(inv.MembershipID)
		resp.MembershipCode = membershipIdAndCode.Code
	}
	return resp
}

// GET /api/leagues/:code/game_rounds - List game rounds for league
func (h *Handler) listLeagueGameRounds(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	// Check for status filter
	statusFilter := r.URL.Query().Get("status")
	activeOnly := r.URL.Query().Get("active") == "true"

	var rounds []*models.GameRound

	if activeOnly {
		rounds, err = h.gameRoundRepository.FindActiveByLeague(r.Context(), leagueID)
	} else if statusFilter != "" {
		status := models.GameRoundStatus(statusFilter)
		if !status.IsValidStatus() {
			http.Error(w, "Invalid status filter", http.StatusBadRequest)
			return
		}
		rounds, err = h.gameRoundRepository.FindByLeagueAndStatus(r.Context(), leagueID, []models.GameRoundStatus{status})
	} else {
		rounds, err = h.gameRoundRepository.FindByLeague(r.Context(), leagueID)
	}

	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to list game rounds")
		return
	}

	// Fill Code for each round
	for _, round := range rounds {
		if idAndCode := h.idCodeCache.GetByID(round.ID); idAndCode != nil {
			round.Code = idAndCode.Code
		}
		// Note: If code is missing, it means the round is not in the cache yet.
		// This can happen for newly created rounds. The frontend will handle this gracefully.
	}

	utils.WriteJSON(r, w, rounds, http.StatusOK)
}

// createLeagueGameRound is deprecated - use startGame from gameround.go instead
// This method is kept for backward compatibility but should be removed
// POST /api/leagues/:code/game_rounds - Create game round in league
func (h *Handler) createLeagueGameRound(w http.ResponseWriter, r *http.Request) {
	// Delegate to startGame which handles league from context
	h.startGame(w, r)
}

// GET /api/leagues/:code/suggested-players - Get suggested players for game creation
func (h *Handler) getSuggestedPlayers(w http.ResponseWriter, r *http.Request) {
	leagueID, err := h.getIDFromChiURL(r, "code")
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

	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	// Check if user is a member of the league or superadmin
	user, err := h.userService.FindByCode(r.Context(), profile.Code)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	isSuperAdmin := auth.IsSuperAdmin(user)
	if !isSuperAdmin {
		isMember, err := h.leagueService.IsUserMember(r.Context(), leagueID, userID)
		if err != nil {
			utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to check membership")
			return
		}
		if !isMember {
			http.Error(w, "Forbidden: not a member of this league", http.StatusForbidden)
			return
		}
	}

	// Get suggested players
	response, err := h.leagueService.GetSuggestedPlayers(r.Context(), leagueID, userID, isSuperAdmin)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to get suggested players")
		return
	}

	utils.WriteJSON(r, w, response, http.StatusOK)
}

// POST /api/leagues/:code/memberships - Create membership for superadmin (superadmin only)
func (h *Handler) createMembershipForSuperAdmin(w http.ResponseWriter, r *http.Request) {
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

	// Get league ID from context (set by middleware)
	leagueID, ok := r.Context().Value("leagueID").(primitive.ObjectID)
	if !ok {
		http.Error(w, "League not found in context", http.StatusInternalServerError)
		return
	}

	// Get user ID
	userIdAndCode, err := h.idCodeCache.GetByCode(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}
	userID := userIdAndCode.ID

	// Parse request
	var req createMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create membership
	membership, err := h.leagueService.CreateMembershipForSuperAdmin(r.Context(), leagueID, userID, req.Alias)
	if err != nil {
		if err.Error() == "user is already an active member of this league" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err.Error() == "alias is already taken in this league" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "failed to create membership")
		return
	}

	// Return membership info
	response := membershipResponse{
		MembershipCode: utils.IdToCode(membership.ID),
		Alias:          membership.Alias,
		Status:         string(membership.Status),
		JoinedAt:       membership.JoinedAt.Format(time.RFC3339),
	}

	utils.WriteJSON(r, w, response, http.StatusCreated)
}

type createMembershipRequest struct {
	Alias string `json:"alias,omitempty"`
}

type membershipResponse struct {
	MembershipCode string `json:"membership_code"`
	Alias          string `json:"alias"`
	Status         string `json:"status"`
	JoinedAt       string `json:"joined_at"`
}
