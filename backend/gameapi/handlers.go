package gameapi

import (
	"encoding/json"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sort"
	"time"
)

type GameRoundHandler struct {
	repo repositories.GameRoundRepository
}

type updateScoreRequest struct {
	Score float64 `json:"score" validate:"required"`
}

type finalizeGameRequest struct {
	PlayerScores map[string]float64 `json:"player_scores"`
	TeamScores   map[string]float64 `json:"team_scores,omitempty"`
}

type startGameRequest struct {
	Name      string          `json:"name" validate:"required"`
	Type      models.GameType `json:"type" validate:"required"`
	StartTime time.Time       `json:"start_time"`
	Players   []playerSetup   `json:"players" validate:"required,min=1"`
}

type playerSetup struct {
	UserID      primitive.ObjectID `json:"user_id" validate:"required"`
	Order       int                `json:"order" validate:"required"`
	IsModerator bool               `json:"is_moderator"`
	TeamName    *string            `json:"team_name,omitempty"`
	TeamColor   *string            `json:"team_color,omitempty"`
}

func (h *GameRoundHandler) RegisterRoutes(r chi.Router) {
	r.Route("/games/rounds", func(r chi.Router) {
		r.Post("/", h.startGame)
		r.Get("/{id}", h.getGameRound)
		r.Put("/{id}/finalize", h.finalizeGame)
		r.Put("/{id}/players/{userId}/score", h.updatePlayerScore)
	})
}

func (h *GameRoundHandler) startGame(w http.ResponseWriter, r *http.Request) {
	var req startGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	players := make([]models.GameRoundPlayer, 0, len(req.Players))
	for _, p := range req.Players {
		user, err := getUserInfo(r.Context(), p.UserID) // Implement this to get user details
		if err != nil {
			http.Error(w, "Error fetching user info", http.StatusInternalServerError)
			return
		}

		players = append(players, models.GameRoundPlayer{
			UserID:      p.UserID,
			Name:        user.Name,
			Avatar:      user.Avatar,
			Order:       p.Order,
			IsModerator: p.IsModerator,
			TeamName:    p.TeamName,
			TeamColor:   p.TeamColor,
		})
	}

	round := &models.GameRound{
		Name:      req.Name,
		Type:      req.Type,
		StartTime: req.StartTime,
		Players:   players,
	}

	if err := h.repo.Create(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, round, http.StatusCreated)
}

func (h *GameRoundHandler) getGameRound(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	round, err := h.repo.FindByID(r.Context(), id)
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

func (h *GameRoundHandler) updatePlayerScore(w http.ResponseWriter, r *http.Request) {
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

	round, err := h.repo.FindByID(r.Context(), id)
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
		if round.Players[i].UserID == userID {
			round.Players[i].Score = req.Score
			playerFound = true
			break
		}
	}

	if !playerFound {
		http.Error(w, "Player not found in game", http.StatusNotFound)
		return
	}

	if err := h.repo.Update(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *GameRoundHandler) finalizeGame(w http.ResponseWriter, r *http.Request) {
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

	round, err := h.repo.FindByID(r.Context(), id)
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
		if score, ok := req.PlayerScores[round.Players[i].UserID.Hex()]; ok {
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
			score float64
		}, 0)

		for name, score := range req.TeamScores {
			teams = append(teams, struct {
				name  string
				score float64
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

	if err := h.repo.Update(r.Context(), round); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
