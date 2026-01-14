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
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) startGame(w http.ResponseWriter, r *http.Request) {
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

		// Support both MembershipID and UserID (for backward compatibility)
		if !p.MembershipID.IsZero() {
			// Using membership ID directly (supports pending members)
			player.MembershipID = p.MembershipID
		} else if !p.UserID.IsZero() {
			// Legacy: using user ID directly
			user, err := h.getUserInfo(r.Context(), p.UserID)
			if err != nil {
				http.Error(w, "Error fetching user info", http.StatusInternalServerError)
				return
			}
			player.PlayerID = user.ID
		} else {
			http.Error(w, "Either user_id or membership_id is required for each player", http.StatusBadRequest)
			return
		}

		players = append(players, player)
	}

	gameType, err := h.gameTypeRepository.FindByKey(r.Context(), req.Type)
	if gameType == nil || err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error fetching game type: "+req.Type)
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
				utils.LogAndWriteHTTPError(w, http.StatusBadRequest,
					fmt.Errorf("team %s has no players assigned", teamName),
					"invalid team assignments")
				return
			}
		}
	}

	round := &models.GameRound{
		Name:       req.Name,
		GameTypeID: gameType.ID,
		Status:     models.StatusPlayersSelected,
		StartTime:  req.StartTime,
		Players:    players,
		TeamScores: teamScores,
	}

	// Set league ID if provided
	if req.LeagueID != "" {
		leagueID, err := primitive.ObjectIDFromHex(req.LeagueID)
		if err != nil {
			http.Error(w, "Invalid league ID", http.StatusBadRequest)
			return
		}
		round.LeagueID = leagueID
	}

	if err := h.gameRoundRepository.Create(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, round, http.StatusCreated)
}

func (h *Handler) getUserInfo(context context.Context, ID primitive.ObjectID) (*models.User, error) {
	return h.userService.FindByID(context, ID)
}

func (h *Handler) listGameRounds(w http.ResponseWriter, r *http.Request) {
	var rounds []*models.GameRound
	var err error

	// Check if league filter is provided in query parameters
	leagueIDParam := r.URL.Query().Get("league_id")
	if leagueIDParam != "" {
		leagueID, err := primitive.ObjectIDFromHex(leagueIDParam)
		if err != nil {
			http.Error(w, "Invalid league ID", http.StatusBadRequest)
			return
		}
		rounds, err = h.gameRoundRepository.FindByLeague(r.Context(), leagueID)
	} else {
		rounds, err = h.gameRoundRepository.FindAll(r.Context())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, rounds, http.StatusOK)
}

func (h *Handler) getGameRound(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetIDFromChiURL(r, "id")
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			userID, err := utils.CodeToID(p.UserID)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			players = append(players, models.GameRoundPlayer{
				PlayerID:    userID,
				Position:    p.Position,
				Score:       p.Score,
				IsModerator: p.IsModerator,
				TeamName:    p.TeamName,
			})
		}
		round.Players = players
	}

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, round, http.StatusOK)
}

func (h *Handler) updatePlayerScore(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req updateScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	playerFound := false
	for i := range round.Players {
		if round.Players[i].PlayerID == userID {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) finalizeGame(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var req finalizeGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	round, err := h.gameRoundRepository.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	Name    string                  `json:"name"`
	Players []updatePlayerSetup     `json:"players"`
}

type updatePlayerSetup struct {
	UserID      string `json:"user_id"`
	Position    int    `json:"position"`
	Score       int64  `json:"score"`
	IsModerator bool   `json:"is_moderator"`
	TeamName    string `json:"team_name,omitempty"`
}

type startGameRequest struct {
	Name      string        `json:"name" validate:"required"`
	Type      string        `json:"type" validate:"required"`
	LeagueID  string        `json:"league_id,omitempty"`
	StartTime time.Time     `json:"start_time"`
	Players   []playerSetup `json:"players" validate:"required,min=1"`
}

type playerSetup struct {
	UserID       primitive.ObjectID `json:"user_id,omitempty"`
	MembershipID primitive.ObjectID `json:"membership_id,omitempty"`
	Position     int                `json:"position" validate:"required"`
	IsModerator  bool               `json:"is_moderator"`
	TeamName     string             `json:"team_name,omitempty"`
	TeamColor    string             `json:"team_color,omitempty"`
}

// updateRolesRequest - запит на оновлення ролей гравців
type updateRolesRequest struct {
	Players []playerRoleUpdate `json:"players" validate:"required"`
}

type playerRoleUpdate struct {
	MembershipID string `json:"membership_id" validate:"required"`
	RoleKey      string `json:"role_key,omitempty"`
	TeamName     string `json:"team_name,omitempty"`
	IsModerator  bool   `json:"is_moderator"`
}

// updateScoresRequest - запит на оновлення очок гравців
type updateScoresRequest struct {
	PlayerScores map[string]int64 `json:"player_scores"` // membership_id -> score
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Update player roles
	for _, update := range req.Players {
		membershipID, err := primitive.ObjectIDFromHex(update.MembershipID)
		if err != nil {
			continue
		}
		for i := range round.Players {
			if round.Players[i].MembershipID == membershipID {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, round, http.StatusOK)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if round == nil {
		http.Error(w, "Game round not found", http.StatusNotFound)
		return
	}

	// Update player scores
	for membershipIDStr, score := range req.PlayerScores {
		membershipID, err := primitive.ObjectIDFromHex(membershipIDStr)
		if err != nil {
			continue
		}
		for i := range round.Players {
			if round.Players[i].MembershipID == membershipID {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, round, http.StatusOK)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
