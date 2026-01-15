package gameapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/andriyg76/bgl/auth"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/services"
	"github.com/andriyg76/bgl/user_profile"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
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

	members, err := h.leagueService.GetLeagueMemberships(r.Context(), leagueID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to get league members")
		return
	}

	response := make([]memberResponse, 0, len(members))
	for _, member := range members {
		response = append(response, memberResponse{
			Code:       utils.IdToCode(member.MembershipID),
			UserID:     utils.IdToCode(member.UserID),
			UserName:   member.UserName,
			UserAvatar: member.UserAvatar,
			Alias:      member.UserAlias,
			Status:     string(member.Status),
			JoinedAt:   member.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
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

// CreateInvitationRequest represents the request body for creating an invitation
type CreateInvitationRequest struct {
	Alias string `json:"alias"`
}

// POST /api/leagues/:code/invitations - Create invitation
func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
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

	userID, err := utils.CodeToID(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

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
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to check membership")
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
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to create invitation")
		return
	}

	utils.WriteJSON(w, invitationToResponse(invitation), http.StatusCreated)
}

// GET /api/leagues/:code/invitations - List my active invitations
func (h *Handler) listMyInvitations(w http.ResponseWriter, r *http.Request) {
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

	// Get my invitations
	invitations, err := h.leagueService.ListMyInvitations(r.Context(), leagueID, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to list invitations")
		return
	}

	response := make([]invitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response = append(response, invitationToResponse(inv))
	}

	utils.WriteJSON(w, response, http.StatusOK)
}

// POST /api/leagues/:code/invitations/:token/cancel - Cancel invitation by token
func (h *Handler) cancelInvitation(w http.ResponseWriter, r *http.Request) {
	tokenRaw := chi.URLParam(r, "token")
	if tokenRaw == "" {
		http.Error(w, "Invalid invitation token", http.StatusBadRequest)
		return
	}

	// URL decode the token (chi.URLParam may not decode it automatically)
	token, err := url.QueryUnescape(tokenRaw)
	if err != nil {
		// If decoding fails, use raw token
		token = tokenRaw
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

	// Cancel the invitation
	if err := h.leagueService.CancelInvitation(r.Context(), token, userID); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "failed to cancel invitation")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /api/leagues/:code/invitations/expired - List my expired invitations
func (h *Handler) listMyExpiredInvitations(w http.ResponseWriter, r *http.Request) {
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

	invitations, err := h.leagueService.ListMyExpiredInvitations(r.Context(), leagueID, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to list expired invitations")
		return
	}

	response := make([]invitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response = append(response, invitationToResponse(inv))
	}

	utils.WriteJSON(w, response, http.StatusOK)
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

	userID, err := utils.CodeToID(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	invitation, err := h.leagueService.ExtendInvitation(r.Context(), token, userID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "failed to extend invitation")
		return
	}

	utils.WriteJSON(w, invitationToResponse(invitation), http.StatusOK)
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

	membershipID, err := utils.CodeToID(memberCode)
	if err != nil {
		http.Error(w, "Invalid member code", http.StatusBadRequest)
		return
	}

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

	userID, err := utils.CodeToID(profile.Code)
	if err != nil {
		http.Error(w, "Invalid user code", http.StatusBadRequest)
		return
	}

	if err := h.leagueService.UpdatePendingMemberAlias(r.Context(), membershipID, userID, req.Alias); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "failed to update alias")
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
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "invitation not found")
		return
	}

	response := InvitationPreviewResponse{
		LeagueName:   preview.LeagueName,
		InviterAlias: preview.InviterAlias,
		PlayerAlias:  preview.PlayerAlias,
		ExpiresAt:    preview.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		Status:       preview.Status,
	}

	utils.WriteJSON(w, response, http.StatusOK)
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

type memberResponse struct {
	Code       string `json:"code"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
	Alias      string `json:"alias"`
	Status     string `json:"status"`
	JoinedAt   string `json:"joined_at"`
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
	Token        string `json:"token"`
	LeagueID     string `json:"league_id"`
	PlayerAlias  string `json:"player_alias"`
	MembershipID string `json:"membership_id,omitempty"`
	ExpiresAt    string `json:"expires_at"`
	CreatedAt    string `json:"created_at"`
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
	resp := invitationResponse{
		Token:       inv.Token,
		LeagueID:    utils.IdToCode(inv.LeagueID),
		PlayerAlias: inv.PlayerAlias,
		ExpiresAt:   inv.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !inv.MembershipID.IsZero() {
		resp.MembershipID = utils.IdToCode(inv.MembershipID)
	}
	return resp
}

// GET /api/leagues/:code/game_rounds - List game rounds for league
func (h *Handler) listLeagueGameRounds(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
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
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to list game rounds")
		return
	}

	utils.WriteJSON(w, rounds, http.StatusOK)
}

// createLeagueGameRoundRequest - запит на створення раунду в лізі
type createLeagueGameRoundRequest struct {
	Name    string        `json:"name"`
	Type    string        `json:"type" validate:"required"`
	Players []playerSetup `json:"players" validate:"required,min=1"`
}

// POST /api/leagues/:code/game_rounds - Create game round in league
func (h *Handler) createLeagueGameRound(w http.ResponseWriter, r *http.Request) {
	leagueID, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid league code", http.StatusBadRequest)
		return
	}

	var req createLeagueGameRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get game type
	gameType, err := h.gameTypeRepository.FindByKey(r.Context(), req.Type)
	if gameType == nil || err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error fetching game type: "+req.Type)
		return
	}

	// Build players list
	players := make([]models.GameRoundPlayer, 0, len(req.Players))
	for _, p := range req.Players {
		player := models.GameRoundPlayer{
			Position:    p.Position,
			IsModerator: p.IsModerator,
			TeamName:    p.TeamName,
		}

		if !p.MembershipID.IsZero() {
			player.MembershipID = p.MembershipID
		} else {
			http.Error(w, "membership_id is required for each player", http.StatusBadRequest)
			return
		}

		players = append(players, player)
	}

	// Create team scores if game type has team roles
	var teamScores []models.TeamScore
	for i, role := range gameType.Roles {
		if role.RoleType == models.RoleTypeMultiple {
			teamScores = append(teamScores, models.TeamScore{
				Name:     role.Key,
				Position: i + 1,
			})
		}
	}

	round := &models.GameRound{
		Name:       req.Name,
		GameTypeID: gameType.ID,
		LeagueID:   leagueID,
		Status:     models.StatusPlayersSelected,
		StartTime:  time.Now(),
		Players:    players,
		TeamScores: teamScores,
	}

	if err := h.gameRoundRepository.Create(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, round, http.StatusCreated)
}

// GET /api/leagues/:code/suggested-players - Get suggested players for game creation
func (h *Handler) getSuggestedPlayers(w http.ResponseWriter, r *http.Request) {
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
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to check membership")
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
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "failed to get suggested players")
		return
	}

	utils.WriteJSON(w, response, http.StatusOK)
}
