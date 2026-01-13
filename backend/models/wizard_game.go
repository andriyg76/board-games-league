package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// WizardBidRestriction defines bid restriction rules
type WizardBidRestriction string

const (
	BidRestrictionNone        WizardBidRestriction = "NO_RESTRICTIONS"
	BidRestrictionCannotMatch WizardBidRestriction = "CANNOT_MATCH_CARDS"
	BidRestrictionMustMatch   WizardBidRestriction = "MUST_MATCH_CARDS"
)

// WizardGameVariant defines game variant
type WizardGameVariant string

const (
	GameVariantStandard    WizardGameVariant = "STANDARD"
	GameVariantAnniversary WizardGameVariant = "ANNIVERSARY" // Future
)

// WizardGameStatus defines game status
type WizardGameStatus string

const (
	WizardStatusSetup      WizardGameStatus = "SETUP"
	WizardStatusInProgress WizardGameStatus = "IN_PROGRESS"
	WizardStatusCompleted  WizardGameStatus = "COMPLETED"
)

// WizardRoundStatus defines round status
type WizardRoundStatus string

const (
	RoundStatusBidding   WizardRoundStatus = "BIDDING"
	RoundStatusPlaying   WizardRoundStatus = "PLAYING"
	RoundStatusCompleted WizardRoundStatus = "COMPLETED"
)

// WizardGameConfig holds game configuration
type WizardGameConfig struct {
	BidRestriction   WizardBidRestriction `bson:"bid_restriction" json:"bid_restriction"`
	GameVariant      WizardGameVariant    `bson:"game_variant" json:"game_variant"`
	FirstDealerIndex int                  `bson:"first_dealer_index" json:"first_dealer_index"`
}

// WizardPlayer represents a player in the game
type WizardPlayer struct {
	MembershipID primitive.ObjectID `bson:"membership_id" json:"membership_id"`
	PlayerName   string             `bson:"player_name" json:"player_name"`
	TotalScore   int                `bson:"total_score" json:"total_score"`
}

// WizardPlayerResult represents player's result in a round
type WizardPlayerResult struct {
	Bid        int `bson:"bid" json:"bid"`                   // -1 if not set
	Actual     int `bson:"actual" json:"actual"`             // -1 if not set
	Score      int `bson:"score" json:"score"`               // Round score
	Delta      int `bson:"delta" json:"delta"`               // Change from previous round
	TotalScore int `bson:"total_score" json:"total_score"`   // Cumulative score after this round
}

// WizardRound represents a single round in the game
type WizardRound struct {
	RoundNumber   int                  `bson:"round_number" json:"round_number"`
	DealerIndex   int                  `bson:"dealer_index" json:"dealer_index"`
	CardsCount    int                  `bson:"cards_count" json:"cards_count"`
	PlayerResults []WizardPlayerResult `bson:"player_results" json:"player_results"`
	Status        WizardRoundStatus    `bson:"status" json:"status"`
	CompletedAt   time.Time            `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// WizardGame represents a complete Wizard game
type WizardGame struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code        string             `bson:"code" json:"code"`
	GameRoundID primitive.ObjectID `bson:"game_round_id" json:"game_round_id"`

	Config  WizardGameConfig `bson:"config" json:"config"`
	Players []WizardPlayer   `bson:"players" json:"players"`
	Rounds  []WizardRound    `bson:"rounds" json:"rounds"`

	CurrentRound int              `bson:"current_round" json:"current_round"`
	MaxRounds    int              `bson:"max_rounds" json:"max_rounds"`
	Status       WizardGameStatus `bson:"status" json:"status"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// NewWizardGame creates a new Wizard game with initialized rounds
func NewWizardGame(
	code string,
	gameRoundID primitive.ObjectID,
	config WizardGameConfig,
	players []WizardPlayer,
) *WizardGame {
	playerCount := len(players)
	maxRounds := 60 / playerCount

	// Initialize all rounds
	rounds := make([]WizardRound, maxRounds)
	for i := 0; i < maxRounds; i++ {
		roundNumber := i + 1
		dealerIndex := (config.FirstDealerIndex + i) % playerCount

		// Initialize player results with -1 (not set)
		playerResults := make([]WizardPlayerResult, playerCount)
		for j := 0; j < playerCount; j++ {
			playerResults[j] = WizardPlayerResult{
				Bid:        -1,
				Actual:     -1,
				Score:      0,
				Delta:      0,
				TotalScore: 0,
			}
		}

		rounds[i] = WizardRound{
			RoundNumber:   roundNumber,
			DealerIndex:   dealerIndex,
			CardsCount:    roundNumber,
			PlayerResults: playerResults,
			Status:        RoundStatusBidding,
		}
	}

	now := time.Now()

	return &WizardGame{
		Code:         code,
		GameRoundID:  gameRoundID,
		Config:       config,
		Players:      players,
		Rounds:       rounds,
		CurrentRound: 1,
		MaxRounds:    maxRounds,
		Status:       WizardStatusInProgress,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
