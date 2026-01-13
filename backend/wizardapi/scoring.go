package wizardapi

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/andriyg76/bgl/models"
)

// CalculateRoundScore calculates score for a round based on bid and actual tricks
func CalculateRoundScore(bid int, actual int) int {
	if bid == actual {
		return 20 + (10 * actual)
	}

	difference := bid - actual
	if difference < 0 {
		difference = -difference
	}

	return -10 * difference
}

// ValidateBids validates bids according to game rules
func ValidateBids(game *models.WizardGame, roundNumber int, bids []int) error {
	if len(bids) != len(game.Players) {
		return fmt.Errorf("bids count (%d) doesn't match players count (%d)", len(bids), len(game.Players))
	}

	// Check all bids are valid
	totalBids := 0
	for i, bid := range bids {
		if bid < 0 {
			return fmt.Errorf("bid for player %d cannot be negative", i)
		}
		if bid > roundNumber {
			return fmt.Errorf("bid for player %d (%d) cannot exceed cards count (%d)", i, bid, roundNumber)
		}
		totalBids += bid
	}

	// Check restriction
	if game.Config.BidRestriction == models.BidRestrictionNone {
		return nil
	}

	cardsCount := roundNumber

	switch game.Config.BidRestriction {
	case models.BidRestrictionCannotMatch:
		if totalBids == cardsCount {
			return errors.New("total bids cannot equal cards count")
		}
	case models.BidRestrictionMustMatch:
		if totalBids != cardsCount {
			return fmt.Errorf("total bids must equal cards count (got %d, expected %d)", totalBids, cardsCount)
		}
	}

	return nil
}

// ValidateResults validates actual results
func ValidateResults(game *models.WizardGame, roundNumber int, results []int) error {
	if len(results) != len(game.Players) {
		return fmt.Errorf("results count (%d) doesn't match players count (%d)", len(results), len(game.Players))
	}

	totalResults := 0
	for i, result := range results {
		if result < 0 {
			return fmt.Errorf("result for player %d cannot be negative", i)
		}
		if result > roundNumber {
			return fmt.Errorf("result for player %d (%d) cannot exceed cards count (%d)", i, result, roundNumber)
		}
		totalResults += result
	}

	// Total actual tricks must equal cards count
	if totalResults != roundNumber {
		return fmt.Errorf("total actual tricks (%d) must equal cards count (%d)", totalResults, roundNumber)
	}

	return nil
}

// CalculateDealerIndex calculates dealer index for a round
func CalculateDealerIndex(firstDealerIndex int, roundNumber int, playerCount int) int {
	return (firstDealerIndex + roundNumber - 1) % playerCount
}

// CompleteRound calculates scores for all players after round completion
func CompleteRound(game *models.WizardGame, roundIndex int) error {
	if roundIndex < 0 || roundIndex >= len(game.Rounds) {
		return errors.New("invalid round index")
	}

	round := &game.Rounds[roundIndex]

	// Check all bids and actuals are set
	for i, pr := range round.PlayerResults {
		if pr.Bid < 0 || pr.Actual < 0 {
			return fmt.Errorf("player %d has missing bid or actual", i)
		}
	}

	// Calculate scores for each player
	for i := range round.PlayerResults {
		pr := &round.PlayerResults[i]

		// Round score
		pr.Score = CalculateRoundScore(pr.Bid, pr.Actual)

		// Previous total score
		prevTotalScore := 0
		if roundIndex > 0 {
			prevTotalScore = game.Rounds[roundIndex-1].PlayerResults[i].TotalScore
		}

		// Cumulative score
		pr.TotalScore = prevTotalScore + pr.Score

		// Delta (change)
		pr.Delta = pr.Score

		// Update player's total score
		game.Players[i].TotalScore = pr.TotalScore
	}

	round.Status = models.RoundStatusCompleted
	round.CompletedAt = time.Now()

	return nil
}

// RecalculateFromRound recalculates scores starting from specified round
// Used when editing mistakes
func RecalculateFromRound(game *models.WizardGame, fromRoundIndex int) error {
	if fromRoundIndex < 0 || fromRoundIndex >= len(game.Rounds) {
		return errors.New("invalid round index")
	}

	// Recalculate all rounds starting from fromRoundIndex
	for roundIdx := fromRoundIndex; roundIdx < len(game.Rounds); roundIdx++ {
		round := &game.Rounds[roundIdx]

		// Check if round has bid/actual for all players
		hasAllData := true
		for _, pr := range round.PlayerResults {
			if pr.Bid < 0 || pr.Actual < 0 {
				hasAllData = false
				break
			}
		}

		if !hasAllData {
			// Skip incomplete rounds
			continue
		}

		// Recalculate scores for each player
		for i := range round.PlayerResults {
			pr := &round.PlayerResults[i]

			// Recalculate round score
			pr.Score = CalculateRoundScore(pr.Bid, pr.Actual)

			// Previous total score
			prevTotalScore := 0
			if roundIdx > 0 {
				prevTotalScore = game.Rounds[roundIdx-1].PlayerResults[i].TotalScore
			}

			// Cumulative score
			pr.TotalScore = prevTotalScore + pr.Score
			pr.Delta = pr.Score

			// Update player's total score
			game.Players[i].TotalScore = pr.TotalScore
		}
	}

	return nil
}

// FinalizeGame finalizes game and prepares results for GameRound
func FinalizeGame(wizardGame *models.WizardGame, gameRound *models.GameRound) error {
	// Check all rounds are completed
	for i, round := range wizardGame.Rounds {
		if round.Status != models.RoundStatusCompleted {
			return fmt.Errorf("round %d is not completed", i+1)
		}
	}

	// Sort players by score (highest first)
	type PlayerResult struct {
		Index      int
		TotalScore int
	}

	results := make([]PlayerResult, len(wizardGame.Players))
	for i, player := range wizardGame.Players {
		results[i] = PlayerResult{
			Index:      i,
			TotalScore: player.TotalScore,
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	// Update GameRound with final results
	for position, result := range results {
		gameRound.Players[result.Index].Score = int64(result.TotalScore)
		gameRound.Players[result.Index].Position = position + 1
	}

	gameRound.EndTime = time.Now()
	wizardGame.Status = models.WizardStatusCompleted

	return nil
}
