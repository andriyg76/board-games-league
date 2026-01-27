package gameapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/hexerr"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) startGame(w http.ResponseWriter, r *http.Request) {
	// Get league ID from context (set by middleware)
	leagueID, ok := r.Context().Value("leagueID").(primitive.ObjectID)
	if !ok {
		http.Error(w, "League not found in context", http.StatusInternalServerError)
		return
	}

	var req startGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	players := make([]models.GameRoundPlayer, 0, len(req.Players))
	for _, p := range req.Players {
		player := models.GameRoundPlayer{
			Position:    p.Position,
			IsModerator: p.IsModerator,
			TeamName:    p.TeamName,
		}

		// MembershipCode is required
		if p.MembershipCode == "" {
			http.Error(w, "membership_code is required for each player", http.StatusBadRequest)
			return
		}

		// Convert membership code to ID
		membershipIdAndCode, err := h.idCodeCache.GetByCode(p.MembershipCode)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid membership code: %s", p.MembershipCode), http.StatusBadRequest)
			return
		}
		player.MembershipID = membershipIdAndCode.ID

		players = append(players, player)
	}

	// Try to find game type by code first (IdAndCode), then fallback to key
	var gameType *models.GameType
	var err error

	idAndCode, codeErr := h.idCodeCache.GetByCode(req.Type)
	if codeErr == nil && idAndCode != nil {
		// Found by code, use FindByID
		gameType, err = h.gameTypeRepository.FindByID(r.Context(), idAndCode.ID)
	} else {
		// Not found by code, try FindByKey (for backward compatibility)
		gameType, err = h.gameTypeRepository.FindByKey(r.Context(), req.Type)
	}

	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "database error fetching game type: %s", req.Type)
		return
	}
	if gameType == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest, nil, "game type not found: %s", req.Type)
		return
	}

	// Знаходимо ролі типу team (multiple) для команд
	var teamScores []models.TeamScore
	var teamRoles []models.Role
	for i, role := range gameType.Roles {
		if role.RoleType == models.RoleTypeMultiple {
			teamScores = append(teamScores, models.TeamScore{
				Name:     role.Key,
				Position: i + 1,
			})
			teamRoles = append(teamRoles, role)
		}
	}

	// If game type has team roles, validate team assignments
	if len(teamRoles) > 0 {
		teamAssignments := make(map[string]bool)
		for _, role := range teamRoles {
			teamAssignments[role.Key] = false
		}

		// Check if each team has at least one player
		for _, p := range req.Players {
			if p.TeamName != "" {
				teamAssignments[p.TeamName] = true
			}
		}

		// Verify all teams have players
		for teamName, hasPlayers := range teamAssignments {
			if !hasPlayers {
				utils.LogAndWriteHTTPError(r, w, http.StatusBadRequest,
					hexerr.New(fmt.Sprintf("team %s has no players assigned", teamName)),
					"invalid team assignments")
				return
			}
		}
	}

	round := &models.GameRound{
		Name:       req.Name,
		GameTypeID: gameType.ID,
		LeagueID:   leagueID, // Always set from context
		Status:     models.StatusPlayersSelected,
		StartTime:  req.StartTime,
		Players:    players,
		TeamScores: teamScores,
	}

	if err := h.gameRoundRepository.Create(r.Context(), round); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error creating game round")
		return
	}

	utils.WriteJSON(r, w, round, http.StatusCreated)
}

func (h *Handler) getUserInfo(context context.Context, ID primitive.ObjectID) (*models.User, error) {
	return h.userService.FindByID(context, ID)
}

func (h *Handler) listGameRounds(w http.ResponseWriter, r *http.Request) {
	// Get league ID from context (set by middleware)
	leagueID, ok := r.Context().Value("leagueID").(primitive.ObjectID)
	if !ok {
		http.Error(w, "League not found in context", http.StatusInternalServerError)
		return
	}

	// Check for status filter
	statusFilter := r.URL.Query().Get("status")
	activeOnly := r.URL.Query().Get("active") == "true"

	var rounds []*models.GameRound
	var err error

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
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game rounds")
		return
	}

	// Fill Code, GameType, and MembershipCode for each round
	for _, round := range rounds {
		if idAndCode := h.idCodeCache.GetByID(round.ID); idAndCode != nil {
			round.Code = idAndCode.Code
		}

		// Fill GameType
		if !round.GameTypeID.IsZero() {
			gameType, err := h.gameTypeRepository.FindByID(r.Context(), round.GameTypeID)
			if err == nil && gameType != nil {
				if idAndCode := h.idCodeCache.GetByID(gameType.ID); idAndCode != nil {
					round.GameType = idAndCode.Code
				} else {
					round.GameType = gameType.Key
				}
			}
		}

		// Fill MembershipCode for each player
		for i := range round.Players {
			if !round.Players[i].MembershipID.IsZero() {
				if idAndCode := h.idCodeCache.GetByID(round.Players[i].MembershipID); idAndCode != nil {
					round.Players[i].MembershipCode = idAndCode.Code
				}
			}
		}
	}

	utils.WriteJSON(r, w, rounds, http.StatusOK)
}

func (h *Handler) getGameRound(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game code", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game round")
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Fill Code
	if idAndCode := h.idCodeCache.GetByID(round.ID); idAndCode != nil {
		round.Code = idAndCode.Code
	}

	// Fill GameType (key or code)
	if !round.GameTypeID.IsZero() {
		gameType, err := h.gameTypeRepository.FindByID(r.Context(), round.GameTypeID)
		if err == nil && gameType != nil {
			// Try to get code from cache, fallback to key
			if idAndCode := h.idCodeCache.GetByID(gameType.ID); idAndCode != nil {
				round.GameType = idAndCode.Code
			} else {
				round.GameType = gameType.Key
			}
		}
	}

	// Fill MembershipCode for each player
	for i := range round.Players {
		if !round.Players[i].MembershipID.IsZero() {
			if idAndCode := h.idCodeCache.GetByID(round.Players[i].MembershipID); idAndCode != nil {
				round.Players[i].MembershipCode = idAndCode.Code
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(round); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateGameRound(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var req updateGameRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game round")
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Update basic fields
	round.Name = req.Name

	// Update players if provided
	if len(req.Players) > 0 {
		players := make([]models.GameRoundPlayer, 0, len(req.Players))
		for _, p := range req.Players {
			if p.MembershipCode == "" {
				http.Error(w, "membership_code is required for each player", http.StatusBadRequest)
				return
			}

			// Convert membership code to ID
			membershipIdAndCode, err := h.idCodeCache.GetByCode(p.MembershipCode)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid membership code: %s", p.MembershipCode), http.StatusBadRequest)
				return
			}

			players = append(players, models.GameRoundPlayer{
				MembershipID: membershipIdAndCode.ID,
				Position:     p.Position,
				Score:        p.Score,
				IsModerator:  p.IsModerator,
				TeamName:     p.TeamName,
			})
		}
		round.Players = players
	}

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error updating game round")
		return
	}

	utils.WriteJSON(r, w, round, http.StatusOK)
}

func (h *Handler) updatePlayerScore(w http.ResponseWriter, r *http.Request) {
	// Get game round code and player code from URL
	gameRoundCode := chi.URLParam(r, "gameRoundCode")
	if gameRoundCode == "" {
		http.Error(w, "Game round code is required", http.StatusBadRequest)
		return
	}

	playerCode := chi.URLParam(r, "playerCode")
	if playerCode == "" {
		http.Error(w, "Player code is required", http.StatusBadRequest)
		return
	}

	// Convert game round code to ID
	gameRoundIdAndCode, err := h.idCodeCache.GetByCode(gameRoundCode)
	if err != nil {
		http.Error(w, "Invalid game round code", http.StatusBadRequest)
		return
	}

	var req updateScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), gameRoundIdAndCode.ID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game round")
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Convert player code to membership ID
	playerIdAndCode, err := h.idCodeCache.GetByCode(playerCode)
	if err != nil {
		http.Error(w, "Invalid player code", http.StatusBadRequest)
		return
	}
	playerMembershipID := playerIdAndCode.ID

	playerFound := false
	for i := range round.Players {
		if round.Players[i].MembershipID == playerMembershipID {
			round.Players[i].Score = req.Score
			playerFound = true
			break
		}
	}

	if !playerFound {
		http.Error(w, "Player not found in game", http.StatusNotFound)
		return
	}

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error updating player score")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) finalizeGame(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game code", http.StatusBadRequest)
		return
	}

	var req finalizeGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game round")
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Update player scores and calculate positions
	for i := range round.Players {
		if score, ok := req.PlayerScores[utils.IdToCode(round.Players[i].PlayerID)]; ok {
			round.Players[i].Score = score
		}
	}

	// Sort players by score and assign positions
	sort.Slice(round.Players, func(i, j int) bool {
		return round.Players[i].Score > round.Players[j].Score
	})

	for i := range round.Players {
		round.Players[i].Position = i + 1
	}

	// Handle team scores if present
	if len(req.TeamScores) > 0 {
		round.TeamScores = make([]models.TeamScore, 0)
		teams := make([]struct {
			name  string
			score int64
		}, 0)

		for name, score := range req.TeamScores {
			teams = append(teams, struct {
				name  string
				score int64
			}{name: name, score: score})
		}

		sort.Slice(teams, func(i, j int) bool {
			return teams[i].score > teams[j].score
		})

		for i, team := range teams {
			round.TeamScores = append(round.TeamScores, models.TeamScore{
				Name:     team.name,
				Score:    team.score,
				Position: i + 1,
			})
		}
	}

	round.EndTime = time.Now()
	round.Status = models.StatusCompleted

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error finalizing game round")
		return
	}

	// Update recent co-players cache for all players (if game is in a league)
	if !round.LeagueID.IsZero() {
		playerMembershipIDs := make([]primitive.ObjectID, 0, len(round.Players))
		for _, player := range round.Players {
			if !player.MembershipID.IsZero() {
				playerMembershipIDs = append(playerMembershipIDs, player.MembershipID)
			}
		}
		if len(playerMembershipIDs) > 0 {
			// Update in background, don't fail the request if this fails
			go func() {
				ctx := context.Background()
				if err := h.leagueService.UpdatePlayersAfterGame(ctx, playerMembershipIDs); err != nil {
					// Log error but don't fail the request
					fmt.Printf("Failed to update players after game: %v\n", err)
				}
			}()
		}
	}

	w.WriteHeader(http.StatusOK)
}

type updateScoreRequest struct {
	Score int64 `json:"score" validate:"required"`
}

type finalizeGameRequest struct {
	PlayerScores     map[string]int64 `json:"player_scores"`
	TeamScores       map[string]int64 `json:"team_scores,omitempty"`
	CooperativeScore int64            `json:"cooperative_score,omitempty"`
}

type updateGameRoundRequest struct {
	Name    string              `json:"name"`
	Players []updatePlayerSetup `json:"players"`
}

type updatePlayerSetup struct {
	MembershipCode string `json:"membership_code" validate:"required"`
	Position       int    `json:"position"`
	Score          int64  `json:"score"`
	IsModerator    bool   `json:"is_moderator"`
	TeamName       string `json:"team_name,omitempty"`
}

type startGameRequest struct {
	Name      string        `json:"name" validate:"required"`
	Type      string        `json:"type" validate:"required"`
	StartTime time.Time     `json:"start_time"`
	Players   []playerSetup `json:"players" validate:"required,min=1"`
}

type playerSetup struct {
	MembershipCode string `json:"membership_code" validate:"required"`
	Position       int    `json:"position" validate:"required"`
	IsModerator    bool   `json:"is_moderator"`
	TeamName       string `json:"team_name,omitempty"`
	TeamColor      string `json:"team_color,omitempty"`
}

// updateRolesRequest - запит на оновлення ролей гравців
type updateRolesRequest struct {
	Players []playerRoleUpdate `json:"players" validate:"required"`
}

type playerRoleUpdate struct {
	MembershipCode string `json:"membership_code" validate:"required"`
	RoleKey        string `json:"role_key,omitempty"`
	TeamName       string `json:"team_name,omitempty"`
	IsModerator    bool   `json:"is_moderator"`
}

// updateScoresRequest - запит на оновлення очок гравців
type updateScoresRequest struct {
	PlayerScores map[string]int64 `json:"player_scores"` // membership_code -> score
	TeamScores   map[string]int64 `json:"team_scores,omitempty"`
}

// updateStatusRequest - запит на зміну статусу раунду
type updateStatusRequest struct {
	Status  models.GameRoundStatus `json:"status" validate:"required"`
	Version int64                  `json:"version" validate:"required"`
}

// updateRoles оновлює ролі гравців (крок 3)
func (h *Handler) updateRoles(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var req updateRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game round")
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Update player roles
	for _, update := range req.Players {
		// Convert membership code to ID
		membershipIdAndCode, err := h.idCodeCache.GetByCode(update.MembershipCode)
		if err != nil {
			continue
		}
		for i := range round.Players {
			if round.Players[i].MembershipID == membershipIdAndCode.ID {
				round.Players[i].TeamName = update.TeamName
				round.Players[i].LabelName = update.RoleKey
				round.Players[i].IsModerator = update.IsModerator
				break
			}
		}
	}

	// Set status to in_progress if it was players_selected
	if round.Status == models.StatusPlayersSelected {
		round.Status = models.StatusInProgress
	}

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error updating game round")
		return
	}

	utils.WriteJSON(r, w, round, http.StatusOK)
}

// updateScores оновлює очки гравців (крок 4)
func (h *Handler) updateScores(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var req updateScoresRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching game round")
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Update player scores
	for membershipCode, score := range req.PlayerScores {
		// Convert membership code to ID
		membershipIdAndCode, err := h.idCodeCache.GetByCode(membershipCode)
		if err != nil {
			continue
		}
		for i := range round.Players {
			if round.Players[i].MembershipID == membershipIdAndCode.ID {
				round.Players[i].Score = score
				break
			}
		}
	}

	// Update team scores
	if len(req.TeamScores) > 0 {
		for i := range round.TeamScores {
			if score, ok := req.TeamScores[round.TeamScores[i].Name]; ok {
				round.TeamScores[i].Score = score
			}
		}
	}

	// Set status to scoring if it was in_progress
	if round.Status == models.StatusInProgress {
		round.Status = models.StatusScoring
	}

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error updating game round")
		return
	}

	utils.WriteJSON(r, w, round, http.StatusOK)
}

// updateRoundStatus змінює статус раунду
func (h *Handler) updateRoundStatus(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "code")
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if !req.Status.IsValidStatus() {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := h.gameRoundRepository.UpdateStatus(r.Context(), id, req.Status, req.Version); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error updating game round status")
		return
	}

	w.WriteHeader(http.StatusOK)
}
