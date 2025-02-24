package gameapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sort"
	"time"
)

func (h *Handler) startGame(w http.ResponseWriter, r *http.Request) {
	var req startGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	players := make([]models.GameRoundPlayer, 0, len(req.Players))
	for _, p := range req.Players {
		user, err := h.getUserInfo(r.Context(), p.UserID)
		if err != nil {
			http.Error(w, "Error fetching user info", http.StatusInternalServerError)
			return
		}

		players = append(players, models.GameRoundPlayer{
			PlayerID:    user.ID,
			Position:    p.Position,
			IsModerator: p.IsModerator,
			TeamName:    p.TeamName,
		})
	}

	gameType, err := h.gameTypeRepository.FindByName(r.Context(), req.Type)
	if gameType == nil || err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "error fetching game type")
	}

	var teamScores []models.TeamScore
	for i, team := range gameType.Teams {
		teamScores = append(teamScores, models.TeamScore{
			Name:     team.Name,
			Position: i + 1,
		})
	}

	// If game type has teams, validate team assignments
	if len(gameType.Teams) > 0 {
		teamAssignments := make(map[string]bool)
		for _, team := range gameType.Teams {
			teamAssignments[team.Name] = false
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
		StartTime:  req.StartTime,
		Players:    players,
		TeamScores: teamScores,
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

func (h *Handler) getGameRound(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
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

	if err := h.gameRoundRepository.Update(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

type startGameRequest struct {
	Name      string        `json:"name" validate:"required"`
	Type      string        `json:"type" validate:"required"`
	StartTime time.Time     `json:"start_time"`
	Players   []playerSetup `json:"players" validate:"required,min=1"`
}

type playerSetup struct {
	UserID      primitive.ObjectID `json:"user_id" validate:"required"`
	Position    int                `json:"position" validate:"required"`
	IsModerator bool               `json:"is_moderator"`
	TeamName    string             `json:"team_name,omitempty"`
	TeamColor   string             `json:"team_color,omitempty"`
}
