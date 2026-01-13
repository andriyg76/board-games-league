# Wizard Game Implementation Plan

## Overview
Додавання повноцінного інтерфейсу для гри Wizard в Board Games League систему.

## Аналіз з скріншотів

### Функціональність:

#### 1. Game Setup (Налаштування гри)
- **Players**: 3-6 гравців
- **First Dealer**:
  - Randomize toggle
  - Manual selection
- **Bid Restrictions**:
  - `NO_RESTRICTIONS` - ставки не обмежені
  - `CANNOT_MATCH_CARDS` - сума ставок не може дорівнювати кількості карт
  - `MUST_MATCH_CARDS` - сума ставок повинна дорівнювати кількості карт
- **Game Variant**:
  - Standard/Fantasy
  - Anniversary (Bomb, Cloud)

#### 2. Game Flow (Ігровий процес)
- **12 rounds** (раундів)
- **Round number** = **Cards dealt** (1 карта в раунді 1, 12 карт у раунді 12)
- **Dealer rotation** (дилер змінюється кожен раунд)

**Кожен раунд складається з:**
1. **Bid Phase**: гравці вводять свої ставки (bid)
2. **Play Phase**: гра відбувається офлайн
3. **Result Phase**: введення фактичних результатів (actual tricks)
4. **Scoring**: автоматичний підрахунок очок

#### 3. Scoring Rules (Правила підрахунку)
```
IF bid == actual:
  points = 20 + (10 × actual)

IF bid ≠ actual:
  points = -10 × |bid - actual|
```

**Приклади:**
- Bid: 0, Actual: 0 → **+20 points**
- Bid: 1, Actual: 1 → **+30 points** (20 + 10)
- Bid: 3, Actual: 3 → **+50 points** (20 + 30)
- Bid: 0, Actual: 1 → **-10 points**
- Bid: 4, Actual: 5 → **-10 points**
- Bid: 1, Actual: 0 → **-10 points**

#### 4. UI Components

**Main Game Screen:**
- Header: `{current_round} / 12` | Game Icon | `difference`
- Current dealer highlighted
- Player rows showing:
  - Player name
  - Current round score with delta
  - `actual / bid` display
  - Checkmark when confirmed

**Dialogs:**
- **Bid Input Dialog**: Number picker with validation
- **Result Input Dialog**: Number picker (0 to round number)
- **Scoreboard**: Full game table with color-coded results

**Color Coding:**
- **Blue background**: High score in round
- **Green number**: Bid matched (bid == actual)
- **Orange number**: Small miss (|difference| = 1)
- **Red number**: Big miss (|difference| > 1)

#### 5. Actions
- Start Round
- Next Round
- Previous Rounds (history)
- Restart Round
- Show Scoreboard
- Main Menu

---

## Technical Implementation Plan

### Phase 1: Database Schema

#### New Collections

**1. `wizard_games`**
```go
type WizardGame struct {
    ID                primitive.ObjectID   `bson:"_id,omitempty"`
    LeagueID          primitive.ObjectID   `bson:"league_id,omitempty"`
    GameRoundID       primitive.ObjectID   `bson:"game_round_id,omitempty"` // Link to main game_rounds

    // Configuration
    BidRestriction    string               `bson:"bid_restriction"` // NO_RESTRICTIONS, CANNOT_MATCH_CARDS, MUST_MATCH_CARDS
    GameVariant       string               `bson:"game_variant"`    // STANDARD, ANNIVERSARY
    RandomizeDealer   bool                 `bson:"randomize_dealer"`
    FirstDealerID     primitive.ObjectID   `bson:"first_dealer_id,omitempty"`

    // Players
    Players           []WizardPlayer       `bson:"players"`

    // Game state
    CurrentRound      int                  `bson:"current_round"` // 1-12
    Status            string               `bson:"status"`        // SETUP, IN_PROGRESS, COMPLETED

    // Rounds data
    Rounds            []WizardRound        `bson:"rounds"`

    CreatedAt         time.Time            `bson:"created_at"`
    UpdatedAt         time.Time            `bson:"updated_at"`
}

type WizardPlayer struct {
    MembershipID      primitive.ObjectID   `bson:"membership_id"`
    PlayerName        string               `bson:"player_name"`
    TotalScore        int64                `bson:"total_score"`
    Position          int                  `bson:"position"` // Final position (1-6)
}

type WizardRound struct {
    RoundNumber       int                  `bson:"round_number"` // 1-12
    DealerID          primitive.ObjectID   `bson:"dealer_id"`
    CardsCount        int                  `bson:"cards_count"`  // == round_number

    PlayerRounds      []WizardPlayerRound  `bson:"player_rounds"`

    Status            string               `bson:"status"` // BIDDING, PLAYING, COMPLETED
    CompletedAt       time.Time            `bson:"completed_at,omitempty"`
}

type WizardPlayerRound struct {
    MembershipID      primitive.ObjectID   `bson:"membership_id"`
    Bid               int                  `bson:"bid"`           // -1 if not set
    Actual            int                  `bson:"actual"`        // -1 if not set
    Score             int64                `bson:"score"`         // Round score
    TotalScore        int64                `bson:"total_score"`   // Cumulative score
    Delta             int64                `bson:"delta"`         // Change from previous round
}
```

**Indexes:**
```go
- wizard_games.league_id
- wizard_games.game_round_id
- wizard_games.status
```

---

### Phase 2: Backend API (Go)

#### New Package: `/backend/wizardapi/`

**Files:**
- `handlers.go` - Route registration
- `game.go` - Game CRUD operations
- `round.go` - Round management
- `scoring.go` - Scoring logic

#### API Endpoints:

```
POST   /api/wizard/games                          - Create new Wizard game
GET    /api/wizard/games/:id                      - Get game details
PUT    /api/wizard/games/:id/config               - Update game configuration
DELETE /api/wizard/games/:id                      - Delete game

POST   /api/wizard/games/:id/start                - Start game (move to IN_PROGRESS)
GET    /api/wizard/games/:id/current-round        - Get current round details

PUT    /api/wizard/games/:id/rounds/:round/bids   - Submit all bids for round
PUT    /api/wizard/games/:id/rounds/:round/bid    - Submit single player bid
PUT    /api/wizard/games/:id/rounds/:round/results - Submit all results for round
PUT    /api/wizard/games/:id/rounds/:round/result  - Submit single player result

POST   /api/wizard/games/:id/rounds/:round/complete - Complete round & calculate scores
POST   /api/wizard/games/:id/rounds/:round/restart  - Restart round (clear bids/results)

GET    /api/wizard/games/:id/scoreboard           - Get full scoreboard
POST   /api/wizard/games/:id/finalize             - Finalize game & update league standings
```

#### Request/Response Examples:

**Create Game:**
```json
POST /api/wizard/games
{
  "league_id": "...",
  "bid_restriction": "NO_RESTRICTIONS",
  "game_variant": "STANDARD",
  "randomize_dealer": true,
  "first_dealer_id": null,
  "players": [
    {"membership_id": "..."},
    {"membership_id": "..."},
    {"membership_id": "..."}
  ]
}

Response:
{
  "id": "...",
  "current_round": 1,
  "status": "SETUP",
  ...
}
```

**Submit Bids:**
```json
PUT /api/wizard/games/:id/rounds/1/bids
{
  "bids": {
    "membership_id_1": 0,
    "membership_id_2": 1,
    "membership_id_3": 0,
    "membership_id_4": 0
  }
}
```

**Submit Results:**
```json
PUT /api/wizard/games/:id/rounds/1/results
{
  "results": {
    "membership_id_1": 0,
    "membership_id_2": 0,
    "membership_id_3": 1,
    "membership_id_4": 0
  }
}
```

**Scoreboard Response:**
```json
GET /api/wizard/games/:id/scoreboard
{
  "rounds": [
    {
      "round_number": 1,
      "dealer_id": "...",
      "player_rounds": [
        {
          "membership_id": "...",
          "player_name": "Andrij",
          "bid": 0,
          "actual": 0,
          "score": 20,
          "total_score": 20,
          "delta": 20
        },
        ...
      ]
    },
    ...
  ],
  "final_standings": [
    {
      "membership_id": "...",
      "player_name": "Lena",
      "total_score": 320,
      "position": 1
    },
    ...
  ]
}
```

#### Scoring Logic (`scoring.go`):

```go
func CalculateRoundScore(bid int, actual int) int64 {
    if bid == actual {
        return 20 + (10 * int64(actual))
    }

    difference := int64(bid - actual)
    if difference < 0 {
        difference = -difference
    }

    return -10 * difference
}

func ValidateBids(game *WizardGame, round *WizardRound, bids map[string]int) error {
    if game.BidRestriction == "NO_RESTRICTIONS" {
        return nil
    }

    totalBids := 0
    for _, bid := range bids {
        totalBids += bid
    }

    cardsCount := round.RoundNumber

    switch game.BidRestriction {
    case "CANNOT_MATCH_CARDS":
        if totalBids == cardsCount {
            return errors.New("total bids cannot equal cards count")
        }
    case "MUST_MATCH_CARDS":
        if totalBids != cardsCount {
            return errors.New("total bids must equal cards count")
        }
    }

    return nil
}

func CalculateDealerRotation(game *WizardGame, roundNumber int) primitive.ObjectID {
    dealerIndex := (roundNumber - 1) % len(game.Players)

    if game.RandomizeDealer && roundNumber == 1 {
        dealerIndex = rand.Intn(len(game.Players))
    } else if !game.FirstDealerID.IsZero() && roundNumber == 1 {
        // Find first dealer
        for i, p := range game.Players {
            if p.MembershipID == game.FirstDealerID {
                dealerIndex = i
                break
            }
        }
    }

    // Rotate for subsequent rounds
    if roundNumber > 1 {
        dealerIndex = (dealerIndex + roundNumber - 1) % len(game.Players)
    }

    return game.Players[dealerIndex].MembershipID
}
```

---

### Phase 3: Frontend (Vue 3 + TypeScript)

#### New Directory: `/frontend/src/wizard/`

**Components:**

1. **`WizardGameSetup.vue`**
   - Player selection (integrate with league membership)
   - Bid restriction radio buttons
   - Game variant selection
   - First dealer selection with randomize toggle
   - "Start Game" button

2. **`WizardGamePlay.vue`** (Main game screen)
   - Round header (X / 12, difference indicator)
   - Current dealer display
   - Player list with:
     - Name
     - Score with delta (+/-X)
     - actual / bid display
     - Clickable to enter bid/result
   - Action buttons:
     - Start Round
     - Next Round
     - Show Scoreboard
     - More Actions menu

3. **`WizardBidDialog.vue`**
   - Number picker (0 to round number)
   - Show other players' bids
   - Validation based on bid restriction
   - Submit button

4. **`WizardResultDialog.vue`**
   - Number picker (0 to round number)
   - Show player's bid
   - Submit button

5. **`WizardScoreboard.vue`**
   - Responsive table with horizontal scroll
   - Round columns for each player
   - Color-coded cells:
     - Background: blue for high scores
     - Text: green (match), orange (miss by 1), red (miss > 1)
   - Display: score (large) + tricks (small)
   - Final standings at bottom

6. **`WizardRoundSummary.vue`**
   - Summary after round completion
   - Show all players' bids vs actual
   - Score changes
   - "Continue" button

**TypeScript Types (`/frontend/src/wizard/types.ts`):**

```typescript
export enum BidRestriction {
  NO_RESTRICTIONS = 'NO_RESTRICTIONS',
  CANNOT_MATCH_CARDS = 'CANNOT_MATCH_CARDS',
  MUST_MATCH_CARDS = 'MUST_MATCH_CARDS'
}

export enum GameVariant {
  STANDARD = 'STANDARD',
  ANNIVERSARY = 'ANNIVERSARY'
}

export enum GameStatus {
  SETUP = 'SETUP',
  IN_PROGRESS = 'IN_PROGRESS',
  COMPLETED = 'COMPLETED'
}

export enum RoundStatus {
  BIDDING = 'BIDDING',
  PLAYING = 'PLAYING',
  COMPLETED = 'COMPLETED'
}

export interface WizardPlayer {
  membership_id: string
  player_name: string
  total_score: number
  position: number
}

export interface WizardPlayerRound {
  membership_id: string
  bid: number
  actual: number
  score: number
  total_score: number
  delta: number
}

export interface WizardRound {
  round_number: number
  dealer_id: string
  cards_count: number
  player_rounds: WizardPlayerRound[]
  status: RoundStatus
  completed_at?: string
}

export interface WizardGame {
  id: string
  league_id?: string
  game_round_id?: string

  bid_restriction: BidRestriction
  game_variant: GameVariant
  randomize_dealer: boolean
  first_dealer_id?: string

  players: WizardPlayer[]
  current_round: number
  status: GameStatus
  rounds: WizardRound[]

  created_at: string
  updated_at: string
}

export interface ScoreboardData {
  rounds: WizardRound[]
  final_standings: WizardPlayer[]
}
```

**API Client (`/frontend/src/api/WizardApi.ts`):**

```typescript
import axios from 'axios'
import type { WizardGame, ScoreboardData } from '@/wizard/types'

const BASE_URL = '/api/wizard'

export class WizardApi {
  static async createGame(gameData: Partial<WizardGame>): Promise<WizardGame> {
    const response = await axios.post(`${BASE_URL}/games`, gameData)
    return response.data
  }

  static async getGame(gameId: string): Promise<WizardGame> {
    const response = await axios.get(`${BASE_URL}/games/${gameId}`)
    return response.data
  }

  static async startGame(gameId: string): Promise<WizardGame> {
    const response = await axios.post(`${BASE_URL}/games/${gameId}/start`)
    return response.data
  }

  static async submitBids(gameId: string, roundNumber: number, bids: Record<string, number>): Promise<void> {
    await axios.put(`${BASE_URL}/games/${gameId}/rounds/${roundNumber}/bids`, { bids })
  }

  static async submitResults(gameId: string, roundNumber: number, results: Record<string, number>): Promise<void> {
    await axios.put(`${BASE_URL}/games/${gameId}/rounds/${roundNumber}/results`, { results })
  }

  static async completeRound(gameId: string, roundNumber: number): Promise<WizardGame> {
    const response = await axios.post(`${BASE_URL}/games/${gameId}/rounds/${roundNumber}/complete`)
    return response.data
  }

  static async getScoreboard(gameId: string): Promise<ScoreboardData> {
    const response = await axios.get(`${BASE_URL}/games/${gameId}/scoreboard`)
    return response.data
  }

  static async finalizeGame(gameId: string): Promise<WizardGame> {
    const response = await axios.post(`${BASE_URL}/games/${gameId}/finalize`)
    return response.data
  }
}
```

**Pinia Store (`/frontend/src/store/wizard.ts`):**

```typescript
import { defineStore } from 'pinia'
import { WizardApi } from '@/api/WizardApi'
import type { WizardGame, ScoreboardData } from '@/wizard/types'

export const useWizardStore = defineStore('wizard', {
  state: () => ({
    currentGame: null as WizardGame | null,
    scoreboard: null as ScoreboardData | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    currentRound(state): number {
      return state.currentGame?.current_round || 1
    },

    isGameInProgress(state): boolean {
      return state.currentGame?.status === 'IN_PROGRESS'
    },

    currentDealer(state): string | undefined {
      if (!state.currentGame) return undefined
      const round = state.currentGame.rounds[state.currentGame.current_round - 1]
      return round?.dealer_id
    }
  },

  actions: {
    async createGame(gameData: Partial<WizardGame>) {
      this.loading = true
      try {
        this.currentGame = await WizardApi.createGame(gameData)
      } catch (error: any) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async loadGame(gameId: string) {
      this.loading = true
      try {
        this.currentGame = await WizardApi.getGame(gameId)
      } catch (error: any) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async submitBids(bids: Record<string, number>) {
      if (!this.currentGame) return

      await WizardApi.submitBids(
        this.currentGame.id,
        this.currentGame.current_round,
        bids
      )
      await this.loadGame(this.currentGame.id)
    },

    async submitResults(results: Record<string, number>) {
      if (!this.currentGame) return

      await WizardApi.submitResults(
        this.currentGame.id,
        this.currentGame.current_round,
        results
      )
      await this.loadGame(this.currentGame.id)
    },

    async completeRound() {
      if (!this.currentGame) return

      this.currentGame = await WizardApi.completeRound(
        this.currentGame.id,
        this.currentGame.current_round
      )
    },

    async loadScoreboard() {
      if (!this.currentGame) return

      this.scoreboard = await WizardApi.getScoreboard(this.currentGame.id)
    }
  }
})
```

---

### Phase 4: Integration with Existing System

#### 1. Link Wizard to Game Rounds

**Modify `/backend/models/game_round.go`:**

```go
type GameRound struct {
    // ... existing fields ...

    // Link to specialized game data
    WizardGameID  primitive.ObjectID `bson:"wizard_game_id,omitempty"`
}
```

#### 2. Add Wizard Game Type

Create a predefined "Wizard" game type or allow users to create it:

```json
{
  "name": "Wizard",
  "scoring_type": "custom",
  "icon": "mdi-wizard-hat",
  "min_players": 3,
  "max_players": 6
}
```

#### 3. Navigation Integration

Add Wizard menu item to `/frontend/src/router/index.ts`:

```typescript
{
  path: '/wizard',
  children: [
    {
      path: 'new',
      component: () => import('@/wizard/WizardGameSetup.vue')
    },
    {
      path: ':id',
      component: () => import('@/wizard/WizardGamePlay.vue')
    }
  ]
}
```

---

## Implementation Phases

### Phase 1: Backend Foundation (Days 1-2)
- [ ] Create database models
- [ ] Implement basic CRUD endpoints
- [ ] Implement scoring logic
- [ ] Add validation logic
- [ ] Write unit tests

### Phase 2: Backend Game Logic (Days 3-4)
- [ ] Implement dealer rotation
- [ ] Implement bid submission with validation
- [ ] Implement result submission
- [ ] Implement round completion
- [ ] Implement scoreboard generation
- [ ] Write integration tests

### Phase 3: Frontend Setup & Basic Components (Days 5-6)
- [ ] Create TypeScript types
- [ ] Implement API client
- [ ] Create Pinia store
- [ ] Implement WizardGameSetup component
- [ ] Implement basic routing

### Phase 4: Frontend Game Flow (Days 7-9)
- [ ] Implement WizardGamePlay component
- [ ] Implement WizardBidDialog component
- [ ] Implement WizardResultDialog component
- [ ] Implement round progression logic
- [ ] Add animations and transitions

### Phase 5: Scoreboard & Polish (Days 10-11)
- [ ] Implement WizardScoreboard component
- [ ] Add color coding logic
- [ ] Implement responsive design
- [ ] Add loading states and error handling
- [ ] Implement WizardRoundSummary component

### Phase 6: Integration & Testing (Days 12-13)
- [ ] Integrate with existing game system
- [ ] Link to league standings
- [ ] End-to-end testing
- [ ] Bug fixes
- [ ] Performance optimization

### Phase 7: Documentation & Deployment (Day 14)
- [ ] API documentation
- [ ] User guide
- [ ] Deployment preparation
- [ ] Final testing
- [ ] Release

---

## Technical Considerations

### 1. Real-time Updates (Optional Enhancement)
Consider adding WebSocket support for multiplayer real-time experience:
- Players see when others submit bids
- Live scoreboard updates
- Round completion notifications

### 2. Offline Support (Future Enhancement)
- Cache game state in localStorage
- Sync when connection restored
- PWA implementation

### 3. Mobile Optimization
- Touch-friendly number pickers
- Horizontal scrollable scoreboard
- Responsive layout for all screen sizes

### 4. Accessibility
- Keyboard navigation
- Screen reader support
- High contrast mode
- Color-blind friendly color schemes

### 5. Performance
- Lazy load scoreboard for completed rounds
- Optimize table rendering for 12+ rounds
- Consider virtual scrolling for large scoreboards

---

## Testing Strategy

### Backend Tests
```go
// Example test structure
func TestCalculateRoundScore(t *testing.T) {
    tests := []struct {
        bid      int
        actual   int
        expected int64
    }{
        {0, 0, 20},
        {1, 1, 30},
        {3, 3, 50},
        {0, 1, -10},
        {1, 0, -10},
        {5, 2, -30},
    }

    for _, tt := range tests {
        result := CalculateRoundScore(tt.bid, tt.actual)
        assert.Equal(t, tt.expected, result)
    }
}
```

### Frontend Tests
- Component unit tests (Vitest)
- Store tests
- Integration tests (Playwright/Cypress)
- E2E game flow tests

---

## Security Considerations

1. **Authorization**:
   - Only league members can participate
   - Only players in the game can submit bids/results
   - Admins can moderate/restart rounds

2. **Validation**:
   - Backend validation for all inputs
   - Prevent submitting bids out of range
   - Prevent manipulating completed rounds

3. **Rate Limiting**:
   - Prevent spam submissions
   - Throttle API calls

---

## Future Enhancements

1. **Statistics Dashboard**
   - Player win rates
   - Average bid accuracy
   - Best/worst rounds

2. **Game Variants**
   - Anniversary edition rules (Bomb, Cloud special cards)
   - Custom scoring formulas
   - Team play mode

3. **AI Opponent** (Advanced)
   - Practice mode with AI players
   - Difficulty levels

4. **Tournament Mode**
   - Multiple games
   - Brackets
   - Aggregate scoring

---

## Questions / Decisions Needed

1. Should Wizard games count towards league standings like other games?
2. Should we support mid-game player substitution?
3. Should we allow editing past rounds (in case of mistakes)?
4. Should we implement game history/replay functionality?
5. Mobile app vs mobile-responsive web?

---

## Estimated Timeline

**Total: 2-3 weeks** (assuming 1 full-time developer)

- Backend: 4-5 days
- Frontend: 6-7 days
- Integration & Testing: 2-3 days
- Polish & Documentation: 1-2 days

---

## Success Criteria

✅ Users can create and configure Wizard games
✅ All scoring rules are correctly implemented
✅ Scoreboard accurately displays all 12 rounds
✅ Mobile-responsive and touch-friendly
✅ Integration with league system
✅ No data loss during game play
✅ Fast and smooth user experience
✅ Comprehensive error handling

