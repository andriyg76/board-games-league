package wizardapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/andriyg76/bgl/utils"
	"github.com/go-chi/chi/v5"
)

type submitBidsRequest struct {
	Bids []int `json:"bids"`
}

type submitResultsRequest struct {
	Results []int `json:"results"`
}

type editRoundRequest struct {
	Bids    *[]int `json:"bids,omitempty"`
	Results *[]int `json:"results,omitempty"`
}

type editRoundResponse struct {
	RoundNumber        int    `json:"round_number"`
	RecalculatedRounds []int  `json:"recalculated_rounds"`
	Message            string `json:"message"`
}

func (h *Handler) submitBids(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	roundStr := chi.URLParam(r, "round")

	roundNumber, err := strconv.Atoi(roundStr)
	if err != nil || roundNumber < 1 {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	var req submitBidsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Validate round number
	if roundNumber > game.MaxRounds {
		http.Error(w, fmt.Sprintf("Round number %d exceeds max rounds %d", roundNumber, game.MaxRounds), http.StatusBadRequest)
		return
	}

	roundIndex := roundNumber - 1

	// Validate bids
	if err := ValidateBids(game, roundNumber, req.Bids); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Invalid bids")
		return
	}

	// Update bids
	for i, bid := range req.Bids {
		game.Rounds[roundIndex].PlayerResults[i].Bid = bid
	}

	// Update round status
	if game.Rounds[roundIndex].Status == "BIDDING" {
		game.Rounds[roundIndex].Status = "PLAYING"
	}

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) submitResults(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	roundStr := chi.URLParam(r, "round")

	roundNumber, err := strconv.Atoi(roundStr)
	if err != nil || roundNumber < 1 {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	var req submitResultsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Validate round number
	if roundNumber > game.MaxRounds {
		http.Error(w, fmt.Sprintf("Round number %d exceeds max rounds %d", roundNumber, game.MaxRounds), http.StatusBadRequest)
		return
	}

	roundIndex := roundNumber - 1

	// Validate results
	if err := ValidateResults(game, roundNumber, req.Results); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Invalid results")
		return
	}

	// Update results
	for i, result := range req.Results {
		game.Rounds[roundIndex].PlayerResults[i].Actual = result
	}

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) completeRound(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	roundStr := chi.URLParam(r, "round")

	roundNumber, err := strconv.Atoi(roundStr)
	if err != nil || roundNumber < 1 {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Validate round number
	if roundNumber > game.MaxRounds {
		http.Error(w, fmt.Sprintf("Round number %d exceeds max rounds %d", roundNumber, game.MaxRounds), http.StatusBadRequest)
		return
	}

	roundIndex := roundNumber - 1

	// Complete round (calculate scores)
	if err := CompleteRound(game, roundIndex); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Error completing round")
		return
	}

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	// Return round results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.Rounds[roundIndex])
}

func (h *Handler) restartRound(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	roundStr := chi.URLParam(r, "round")

	roundNumber, err := strconv.Atoi(roundStr)
	if err != nil || roundNumber < 1 {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Validate round number
	if roundNumber > game.MaxRounds {
		http.Error(w, fmt.Sprintf("Round number %d exceeds max rounds %d", roundNumber, game.MaxRounds), http.StatusBadRequest)
		return
	}

	roundIndex := roundNumber - 1

	// Reset round
	for i := range game.Rounds[roundIndex].PlayerResults {
		game.Rounds[roundIndex].PlayerResults[i].Bid = -1
		game.Rounds[roundIndex].PlayerResults[i].Actual = -1
		game.Rounds[roundIndex].PlayerResults[i].Score = 0
		game.Rounds[roundIndex].PlayerResults[i].Delta = 0
		game.Rounds[roundIndex].PlayerResults[i].TotalScore = 0
	}

	game.Rounds[roundIndex].Status = "BIDDING"
	game.Rounds[roundIndex].CompletedAt = time.Time{}

	// Recalculate all subsequent rounds
	if err := RecalculateFromRound(game, roundIndex); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error recalculating rounds")
		return
	}

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) editRound(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	roundStr := chi.URLParam(r, "round")

	roundNumber, err := strconv.Atoi(roundStr)
	if err != nil || roundNumber < 1 {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	var req editRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Validate round number
	if roundNumber > game.MaxRounds {
		http.Error(w, fmt.Sprintf("Round number %d exceeds max rounds %d", roundNumber, game.MaxRounds), http.StatusBadRequest)
		return
	}

	roundIndex := roundNumber - 1

	// Update bids if provided
	if req.Bids != nil {
		if err := ValidateBids(game, roundNumber, *req.Bids); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Invalid bids")
			return
		}
		for i, bid := range *req.Bids {
			game.Rounds[roundIndex].PlayerResults[i].Bid = bid
		}
	}

	// Update results if provided
	if req.Results != nil {
		if err := ValidateResults(game, roundNumber, *req.Results); err != nil {
			utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Invalid results")
			return
		}
		for i, result := range *req.Results {
			game.Rounds[roundIndex].PlayerResults[i].Actual = result
		}
	}

	// Recalculate from this round
	if err := RecalculateFromRound(game, roundIndex); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error recalculating rounds")
		return
	}

	// Count recalculated rounds
	recalculatedRounds := []int{}
	for i := roundIndex; i < len(game.Rounds); i++ {
		if game.Rounds[i].Status == "COMPLETED" {
			recalculatedRounds = append(recalculatedRounds, i+1)
		}
	}

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	// Return response
	response := editRoundResponse{
		RoundNumber:        roundNumber,
		RecalculatedRounds: recalculatedRounds,
		Message:            fmt.Sprintf("Round %d updated, recalculated %d subsequent rounds", roundNumber, len(recalculatedRounds)-1),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) nextRound(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Check if can move to next round
	if game.CurrentRound >= game.MaxRounds {
		http.Error(w, "Already at last round", http.StatusBadRequest)
		return
	}

	// Move to next round
	game.CurrentRound++

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) prevRound(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Check if can move to previous round
	if game.CurrentRound <= 1 {
		http.Error(w, "Already at first round", http.StatusBadRequest)
		return
	}

	// Move to previous round
	game.CurrentRound--

	// Save game
	if err := h.wizardRepo.Update(r.Context(), game); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getScoreboard(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Get game
	game, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Build scoreboard response
	response := map[string]interface{}{
		"game_code":     game.Code,
		"current_round": game.CurrentRound,
		"max_rounds":    game.MaxRounds,
		"players":       game.Players,
		"rounds":        game.Rounds,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) finalizeGame(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Get wizard game
	wizardGame, err := h.wizardRepo.FindByCode(r.Context(), code)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game not found")
		return
	}

	// Get game round
	gameRound, err := h.gameRoundRepo.FindByID(r.Context(), wizardGame.GameRoundID)
	if err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusNotFound, err, "Game round not found")
		return
	}

	// Finalize game
	if err := FinalizeGame(wizardGame, gameRound); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusBadRequest, err, "Error finalizing game")
		return
	}

	// Save wizard game
	if err := h.wizardRepo.Update(r.Context(), wizardGame); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating wizard game")
		return
	}

	// Save game round
	if err := h.gameRoundRepo.Update(r.Context(), gameRound); err != nil {
		utils.LogAndWriteHTTPError(w, http.StatusInternalServerError, err, "Error updating game round")
		return
	}

	// Build final standings
	type FinalStanding struct {
		PlayerName string `json:"player_name"`
		TotalScore int    `json:"total_score"`
		Position   int    `json:"position"`
	}

	standings := make([]FinalStanding, len(wizardGame.Players))
	for i, player := range wizardGame.Players {
		// Find position from game round
		position := 0
		for _, grPlayer := range gameRound.Players {
			if grPlayer.MembershipID == player.MembershipID {
				position = grPlayer.Position
				break
			}
		}

		standings[i] = FinalStanding{
			PlayerName: player.PlayerName,
			TotalScore: player.TotalScore,
			Position:   position,
		}
	}

	// Sort by position
	for i := 0; i < len(standings); i++ {
		for j := i + 1; j < len(standings); j++ {
			if standings[j].Position < standings[i].Position {
				standings[i], standings[j] = standings[j], standings[i]
			}
		}
	}

	// Convert GameRoundID to code
	gameRoundIdAndCode := h.idCodeCache.GetByID(wizardGame.GameRoundID)
	
	// Return response
	response := map[string]interface{}{
		"wizard_game_code": wizardGame.Code,
		"game_round_code":  gameRoundIdAndCode.Code,
		"final_standings":  standings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
