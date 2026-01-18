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
	GameName              string   `json:"game_name"`
	BidRestriction        string   `json:"bid_restriction"`
	GameVariant           string   `json:"game_variant"`
	FirstDealerIndex      int      `json:"first_dealer_index"`
	PlayerMembershipCodes []string `json:"player_membership_codes"`
}

type createGameResponse struct {
	Code          string           `json:"code"`
	GameRoundCode string           `json:"game_round_code"`
	CurrentRound  int              `json:"current_round"`
	MaxRounds     int              `json:"max_rounds"`
	Status        string           `json:"status"`
	Players       []playerResponse `json:"players"`
}

type playerResponse struct {
	MembershipCode string `json:"membership_code"`
	PlayerName     string `json:"player_name"`
	TotalScore     int    `json:"total_score"`
}

type gameResponse struct {
	Code          string                  `json:"code"`
	GameRoundCode string                  `json:"game_round_code"`
	Config        models.WizardGameConfig `json:"config"`
	Players       []playerResponse        `json:"players"`
	CurrentRound  int                     `json:"current_round"`
	MaxRounds     int                     `json:"max_rounds"`
	Status        string                  `json:"status"`
	Rounds        []roundSummary          `json:"rounds,omitempty"`
}

type roundSummary struct {
	RoundNumber int    `json:"round_number"`
	DealerIndex int    `json:"dealer_index"`
	CardsCount  int    `json:"cards_count"`
	Status      string `json:"status"`
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

	// Get league ID from context (set by middleware)
	leagueID, ok := r.Context().Value("leagueID").(primitive.ObjectID)
	if !ok {
		http.Error(w, "League not found in context", http.StatusInternalServerError)
		return
	}

	// Validate player count (3-6)
	if len(req.PlayerMembershipCodes) < 3 || len(req.PlayerMembershipCodes) > 6 {
		http.Error(w, "Player count must be between 3 and 6", http.StatusBadRequest)
		return
	}

	// Validate first dealer index
	if req.FirstDealerIndex < 0 || req.FirstDealerIndex >= len(req.PlayerMembershipCodes) {
		http.Error(w, "Invalid first dealer index", http.StatusBadRequest)
		return
	}

	// Validate membership codes are not empty
	for i, code := range req.PlayerMembershipCodes {
		if code == "" {
			http.Error(w, fmt.Sprintf("Empty membership code at index %d", i), http.StatusBadRequest)
			return
		}
	}

	// Parse membership codes and get player names
	membershipIDs := make([]primitive.ObjectID, len(req.PlayerMembershipCodes))
	players := make([]models.WizardPlayer, len(req.PlayerMembershipCodes))

	for i, membershipCode := range req.PlayerMembershipCodes {
		// Convert membership code to ID
		membershipIdAndCode, err := h.idCodeCache.GetByCode(membershipCode)
		if err != nil {
			utils.LogAndWriteHTTPError(
				r,
				w,
				http.StatusBadRequest,
				err,
				"%s",
				fmt.Sprintf("Invalid membership code at index %d: %s", i, membershipCode),
			)
			return
		}
		if membershipIdAndCode == nil {
			http.Error(w, fmt.Sprintf("Membership code at index %d returned nil: %s", i, membershipCode), http.StatusBadRequest)
			return
		}
		membershipID := membershipIdAndCode.ID
		membershipIDs[i] = membershipID

		// Get member info from league service
		member, err := h.leagueService.GetMemberByID(r.Context(), membershipID)
		if err != nil {
			utils.LogAndWriteHTTPError(
				r,
				w,
				http.StatusBadRequest,
				err,
				"%s",
				fmt.Sprintf("Error fetching member info for membership code %s at index %d", membershipCode, i),
			)
			return
		}
		if member == nil {
			http.Error(w, fmt.Sprintf("Membership not found for code %s at index %d", membershipCode, i), http.StatusBadRequest)
			return
		}

		players[i] = models.WizardPlayer{
			MembershipID: membershipID,
			PlayerName:   member.Alias, // Use alias from membership
			TotalScore:   0,
		}
	}

	// Get or create Wizard game type
	gameType, err := h.gameTypeRepo.FindByKey(r.Context(), "wizard")
	if err != nil {
		// Database error - log but continue to create
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "error fetching wizard game type")
		return
	}
	if gameType == nil {
		// Create Wizard game type if doesn't exist
		gameType = &models.GameType{
			Key:         "wizard",
			Names:       map[string]string{"en": "Wizard", "uk": "Візард"},
			ScoringType: models.ScoringTypeClassic,
			Icon:        "mdi-cards-playing",
			MinPlayers:  3,
			MaxPlayers:  6,
		}
		if err := h.gameTypeRepo.Create(r.Context(), gameType); err != nil {
			utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "Error creating Wizard game type")
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
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "Error creating game round")
		return
	}

	// Generate game code
	gameCode, err := generateGameCode()
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "Error generating game code")
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
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "Error creating wizard game")
		return
	}

	// Build response
	playerResponses := make([]playerResponse, len(players))
	for i, player := range players {
		// Convert membership ID to code
		membershipIdAndCode := h.idCodeCache.GetByID(player.MembershipID)
		if membershipIdAndCode == nil {
			utils.LogAndWriteHTTPError(
				r,
				w,
				http.StatusInternalServerError,
				fmt.Errorf("failed to get membership code for ID %s", player.MembershipID.Hex()),
				"%s",
				fmt.Sprintf("Error converting membership ID to code at index %d", i),
			)
			return
		}
		playerResponses[i] = playerResponse{
			MembershipCode: membershipIdAndCode.Code,
			PlayerName:     player.PlayerName,
			TotalScore:     player.TotalScore,
		}
	}

	// Convert GameRoundID to code
	gameRoundIdAndCode := h.idCodeCache.GetByID(wizardGame.GameRoundID)
	if gameRoundIdAndCode == nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, fmt.Errorf("failed to get game round code for ID %s", wizardGame.GameRoundID.Hex()),
			"Error converting game round ID to code")
		return
	}
	response := createGameResponse{
		Code:          wizardGame.Code,
		GameRoundCode: gameRoundIdAndCode.Code,
		CurrentRound:  wizardGame.CurrentRound,
		MaxRounds:     wizardGame.MaxRounds,
		Status:        string(wizardGame.Status),
		Players:       playerResponses,
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
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, err, "Game not found")
		return
	}

	h.respondWithGame(w, game)
}

func (h *Handler) getGameByRoundID(w http.ResponseWriter, r *http.Request) {
	gameRoundCode := chi.URLParam(r, "code")
	if gameRoundCode == "" {
		http.Error(w, "Game round code is required", http.StatusBadRequest)
		return
	}

	// Convert game round code to ID
	gameRoundIdAndCode, err := h.idCodeCache.GetByCode(gameRoundCode)
	if err != nil {
		http.Error(w, "Invalid game round code", http.StatusBadRequest)
		return
	}

	game, err := h.wizardRepo.FindByGameRoundID(r.Context(), gameRoundIdAndCode.ID)
	if err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, err, "Game not found")
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
		utils.LogAndWriteHTTPError(r, w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Delete game round first
	if err := h.gameRoundRepo.Delete(r.Context(), game.GameRoundID); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "Error deleting game round")
		return
	}

	// Delete wizard game
	if err := h.wizardRepo.DeleteByCode(r.Context(), code); err != nil {
		utils.LogAndWriteHTTPError(r, w, http.StatusInternalServerError, err, "Error deleting wizard game")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) respondWithGame(w http.ResponseWriter, game *models.WizardGame) {
	playerResponses := make([]playerResponse, len(game.Players))
	for i, player := range game.Players {
		// Convert membership ID to code
		membershipIdAndCode := h.idCodeCache.GetByID(player.MembershipID)
		playerResponses[i] = playerResponse{
			MembershipCode: membershipIdAndCode.Code,
			PlayerName:     player.PlayerName,
			TotalScore:     player.TotalScore,
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

	// Convert GameRoundID to code
	gameRoundIdAndCode := h.idCodeCache.GetByID(game.GameRoundID)
	response := gameResponse{
		Code:          game.Code,
		GameRoundCode: gameRoundIdAndCode.Code,
		Config:        game.Config,
		Players:       playerResponses,
		CurrentRound:  game.CurrentRound,
		MaxRounds:     game.MaxRounds,
		Status:        string(game.Status),
		Rounds:        rounds,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
