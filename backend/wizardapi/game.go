package wizardapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createGameRequest struct {
	LeagueID            string   `json:"league_id"`
	GameName            string   `json:"game_name"`
	BidRestriction      string   `json:"bid_restriction"`
	GameVariant         string   `json:"game_variant"`
	FirstDealerIndex    int      `json:"first_dealer_index"`
	PlayerMembershipIDs []string `json:"player_membership_ids"`
}

type createGameResponse struct {
	Code        string           `json:"code"`
	GameRoundID string           `json:"game_round_id"`
	CurrentRound int             `json:"current_round"`
	MaxRounds   int              `json:"max_rounds"`
	Status      string           `json:"status"`
	Players     []playerResponse `json:"players"`
}

type playerResponse struct {
	MembershipID string `json:"membership_id"`
	PlayerName   string `json:"player_name"`
	TotalScore   int    `json:"total_score"`
}

type gameResponse struct {
	Code         string                `json:"code"`
	GameRoundID  string                `json:"game_round_id"`
	Config       models.WizardGameConfig `json:"config"`
	Players      []playerResponse      `json:"players"`
	CurrentRound int                   `json:"current_round"`
	MaxRounds    int                   `json:"max_rounds"`
	Status       string                `json:"status"`
	Rounds       []roundSummary        `json:"rounds,omitempty"`
}

type roundSummary struct {
	RoundNumber int              `json:"round_number"`
	DealerIndex int              `json:"dealer_index"`
	CardsCount  int              `json:"cards_count"`
	Status      string           `json:"status"`
}

// generateGameCode generates a random game code
func generateGameCode() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *Handler) createGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate player count (3-6)
	if len(req.PlayerMembershipIDs) < 3 || len(req.PlayerMembershipIDs) > 6 {
		http.Error(w, "Player count must be between 3 and 6", http.StatusBadRequest)
		return
	}

	// Validate first dealer index
	if req.FirstDealerIndex < 0 || req.FirstDealerIndex >= len(req.PlayerMembershipIDs) {
		http.Error(w, "Invalid first dealer index", http.StatusBadRequest)
		return
	}

	// Parse league ID
	var leagueID primitive.ObjectID
	var err error
	if req.LeagueID != "" {
		leagueID, err = primitive.ObjectIDFromHex(req.LeagueID)
		if err != nil {
			http.Error(w, "Invalid league ID", http.StatusBadRequest)
			return
		}
	}

	// Parse membership IDs and get player names
	membershipIDs := make([]primitive.ObjectID, len(req.PlayerMembershipIDs))
	players := make([]models.WizardPlayer, len(req.PlayerMembershipIDs))

	for i, membershipIDStr := range req.PlayerMembershipIDs {
		membershipID, err := primitive.ObjectIDFromHex(membershipIDStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid membership ID at index %d", i), http.StatusBadRequest)
			return
		}
		membershipIDs[i] = membershipID

		// Get member info from league service
		member, err := h.leagueService.GetMemberByID(r.Context(), membershipID)
		if err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err,
				fmt.Sprintf("Error fetching member info for index %d", i))
			return
		}

		players[i] = models.WizardPlayer{
			MembershipID: membershipID,
			PlayerName:   member.Alias, // Use alias from membership
			TotalScore:   0,
		}
	}

	// Get or create Wizard game type
	gameType, err := h.gameTypeRepo.FindByName(r.Context(), "Wizard")
	if err != nil || gameType == nil {
		// Create Wizard game type if doesn't exist
		gameType = &models.GameType{
			Name:        "Wizard",
			ScoringType: string(models.ScoringTypeClassic),
			Icon:        "mdi-wizard-hat",
			MinPlayers:  3,
			MaxPlayers:  6,
		}
		if err := h.gameTypeRepo.Create(r.Context(), gameType); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error creating Wizard game type")
			return
		}
	}

	// Create GameRound
	gameRoundPlayers := make([]models.GameRoundPlayer, len(membershipIDs))
	for i, membershipID := range membershipIDs {
		gameRoundPlayers[i] = models.GameRoundPlayer{
			MembershipID: membershipID,
			Score:        0,
			Position:     0,
		}
	}

	gameRound := &models.GameRound{
		Name:       req.GameName,
		GameTypeID: gameType.ID,
		LeagueID:   leagueID,
		StartTime:  time.Now(),
		Players:    gameRoundPlayers,
	}

	if err := h.gameRoundRepo.Create(r.Context(), gameRound); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error creating game round")
		return
	}

	// Generate game code
	gameCode, err := generateGameCode()
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error generating game code")
		return
	}

	// Create WizardGame
	config := models.WizardGameConfig{
		BidRestriction:   models.WizardBidRestriction(req.BidRestriction),
		GameVariant:      models.WizardGameVariant(req.GameVariant),
		FirstDealerIndex: req.FirstDealerIndex,
	}

	wizardGame := models.NewWizardGame(gameCode, gameRound.ID, config, players)

	if err := h.wizardRepo.Create(r.Context(), wizardGame); err != nil {
		// Rollback: delete game round
		_ = h.gameRoundRepo.Delete(r.Context(), gameRound.ID)
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error creating wizard game")
		return
	}

	// Build response
	playerResponses := make([]playerResponse, len(players))
	for i, player := range players {
		playerResponses[i] = playerResponse{
			MembershipID: player.MembershipID.Hex(),
			PlayerName:   player.PlayerName,
			TotalScore:   player.TotalScore,
		}
	}

	response := createGameResponse{
		Code:         wizardGame.Code,
		GameRoundID:  wizardGame.GameRoundID.Hex(),
		CurrentRound: wizardGame.CurrentRound,
		MaxRounds:    wizardGame.MaxRounds,
		Status:       string(wizardGame.Status),
		Players:      playerResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) getGame(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "Game code is required", http.StatusBadRequest)
		return
	}

	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	h.respondWithGame(w, game)
}

func (h *Handler) getGameByRoundID(w http.ResponseWriter, r *http.Request) {
	gameRoundIDStr := chi.URLParam(r, "gameRoundId")
	if gameRoundIDStr == "" {
		http.Error(w, "Game round ID is required", http.StatusBadRequest)
		return
	}

	gameRoundID, err := primitive.ObjectIDFromHex(gameRoundIDStr)
	if err != nil {
		http.Error(w, "Invalid game round ID", http.StatusBadRequest)
		return
	}

	game, err := h.wizardRepo.FindByGameRoundID(r.Context(), gameRoundID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	h.respondWithGame(w, game)
}

func (h *Handler) deleteGame(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "Game code is required", http.StatusBadRequest)
		return
	}

	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Delete game round first
	if err := h.gameRoundRepo.Delete(r.Context(), game.GameRoundID); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error deleting game round")
		return
	}

	// Delete wizard game
	if err := h.wizardRepo.DeleteByCode(r.Context(), code); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error deleting wizard game")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) respondWithGame(w http.ResponseWriter, game *models.WizardGame) {
	playerResponses := make([]playerResponse, len(game.Players))
	for i, player := range game.Players {
		playerResponses[i] = playerResponse{
			MembershipID: player.MembershipID.Hex(),
			PlayerName:   player.PlayerName,
			TotalScore:   player.TotalScore,
		}
	}

	rounds := make([]roundSummary, len(game.Rounds))
	for i, round := range game.Rounds {
		rounds[i] = roundSummary{
			RoundNumber: round.RoundNumber,
			DealerIndex: round.DealerIndex,
			CardsCount:  round.CardsCount,
			Status:      string(round.Status),
		}
	}

	response := gameResponse{
		Code:         game.Code,
		GameRoundID:  game.GameRoundID.Hex(),
		Config:       game.Config,
		Players:      playerResponses,
		CurrentRound: game.CurrentRound,
		MaxRounds:    game.MaxRounds,
		Status:       string(game.Status),
		Rounds:       rounds,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
