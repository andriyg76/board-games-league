# Wizard Game Implementation Plan

## Overview
Додавання повноцінного інтерфейсу для гри Wizard в Board Games League систему.
Wizard інтегрується як спеціалізований інтерфейс для звичайної гри - результати автоматично відображаються в league standings.

## Аналіз з скріншотів

### Функціональність:

#### 1. Game Setup (Налаштування гри)
- **Players**: 3-6 гравців
- **First Dealer**: Manual selection (index 0-N)
- **Bid Restrictions**:
  - `NO_RESTRICTIONS` - ставки не обмежені
  - `CANNOT_MATCH_CARDS` - сума ставок не може дорівнювати кількості карт
  - `MUST_MATCH_CARDS` - сума ставок повинна дорівнювати кількості карт
- **Game Variant**:
  - Standard (поки що)
  - Anniversary (майбутнє розширення)

#### 2. Game Flow (Ігровий процес)
- **Динамічна кількість раундів**: `60 / кількість_гравців`
  - 3 гравці → 20 раундів
  - 4 гравці → 15 раундів
  - 5 гравців → 12 раундів
  - 6 гравців → 10 раундів
- **Round number** = **Cards dealt** (1 карта в раунді 1, N карт у раунді N)
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

#### Інтеграція з існуючою системою

**Wizard використовує існуючі таблиці:**
- `game_types` - тип гри "Wizard" (scoring_type: "classic")
- `game_rounds` - основний запис гри для league standings

**Нова колекція для деталей гри:**

**1. `wizard_games` - детальний стан Wizard гри**
```go
type WizardGame struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Code         string             `bson:"code" json:"code"` // Унікальний код для URL

    // Прив'язка до основного game_round
    GameRoundID  primitive.ObjectID `bson:"game_round_id" json:"game_round_id"` // ОБОВ'ЯЗКОВО

    // Конфігурація гри
    Config struct {
        BidRestriction   string `bson:"bid_restriction" json:"bid_restriction"` // NO_RESTRICTIONS, CANNOT_MATCH_CARDS, MUST_MATCH_CARDS
        GameVariant      string `bson:"game_variant" json:"game_variant"`       // STANDARD (поки що)
        FirstDealerIndex int    `bson:"first_dealer_index" json:"first_dealer_index"` // 0-based індекс у масиві Players
    } `bson:"config" json:"config"`

    // Гравці (фіксований порядок, індекс відповідає GameRound.Players)
    Players []struct {
        MembershipID primitive.ObjectID `bson:"membership_id" json:"membership_id"`
        PlayerName   string             `bson:"player_name" json:"player_name"`
        TotalScore   int                `bson:"total_score" json:"total_score"` // Кумулятивний рахунок
    } `bson:"players" json:"players"`

    // Раунди (масив, length = 60 / len(Players))
    Rounds []struct {
        RoundNumber int `bson:"round_number" json:"round_number"` // 1 до MaxRounds
        DealerIndex int `bson:"dealer_index" json:"dealer_index"` // Індекс у Players
        CardsCount  int `bson:"cards_count" json:"cards_count"`   // = RoundNumber

        // Результати гравців (індекс відповідає індексу в Players)
        PlayerResults []struct {
            Bid        int `bson:"bid" json:"bid"`              // -1 = не встановлено
            Actual     int `bson:"actual" json:"actual"`        // -1 = не встановлено
            Score      int `bson:"score" json:"score"`          // Очки за раунд
            Delta      int `bson:"delta" json:"delta"`          // Зміна від попереднього раунду
            TotalScore int `bson:"total_score" json:"total_score"` // Кумулятивний рахунок після цього раунду
        } `bson:"player_results" json:"player_results"`

        Status      string    `bson:"status" json:"status"` // BIDDING, PLAYING, COMPLETED
        CompletedAt time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
    } `bson:"rounds" json:"rounds"`

    // Стан гри
    CurrentRound int    `bson:"current_round" json:"current_round"` // 1 до MaxRounds
    MaxRounds    int    `bson:"max_rounds" json:"max_rounds"`       // 60 / len(Players)
    Status       string `bson:"status" json:"status"`               // SETUP, IN_PROGRESS, COMPLETED

    // Метадані
    CreatedAt time.Time `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
```

**Indexes:**
```go
- wizard_games.code (unique)
- wizard_games.game_round_id (unique) // Одна wizard_game на один game_round
- wizard_games.status
```

#### Життєвий цикл гри:

**1. Створення гри:**
```
1. Створюємо GameRound (status: in progress, без scores)
2. Створюємо WizardGame (прив'язаний до GameRound)
3. Ініціалізуємо всі раунди з порожніми bid/actual (-1)
```

**2. Процес гри:**
```
- Всі bid/actual зберігаються в WizardGame.Rounds
- GameRound залишається без змін
```

**3. Завершення гри:**
```
1. Розраховуємо фінальні scores в WizardGame
2. Записуємо scores назад в GameRound.Players[].Score
3. Розраховуємо positions в GameRound.Players[].Position
4. Встановлюємо GameRound.EndTime
5. League standings автоматично оновлюються
```

---

### Phase 2: Backend API (Go)

#### New Package: `/backend/wizardapi/`

**Files:**
- `handlers.go` - Route registration
- `game.go` - Game lifecycle (create, get, finalize)
- `round.go` - Round management (bids, results, complete)
- `scoring.go` - Scoring calculation logic

#### API Endpoints:

```
# Game Management
POST   /api/wizard/games                          - Create new Wizard game + GameRound
GET    /api/wizard/games/:code                    - Get game details by code
GET    /api/wizard/games/by-round/:game_round_id  - Get game by GameRound ID
DELETE /api/wizard/games/:code                    - Delete game (also deletes GameRound)

# Round Operations
PUT    /api/wizard/games/:code/rounds/:round/bids     - Submit bids for round (bulk)
PUT    /api/wizard/games/:code/rounds/:round/results  - Submit results for round (bulk)
POST   /api/wizard/games/:code/rounds/:round/complete - Complete round & calculate scores
POST   /api/wizard/games/:code/rounds/:round/restart  - Restart round (clear bids/results)
PUT    /api/wizard/games/:code/rounds/:round/edit     - Edit bid/actual after completion (recalculates all subsequent rounds)

# Game State
GET    /api/wizard/games/:code/scoreboard         - Get full scoreboard (all rounds)
POST   /api/wizard/games/:code/finalize           - Finalize game → update GameRound scores
POST   /api/wizard/games/:code/next-round         - Move to next round
POST   /api/wizard/games/:code/prev-round         - Move to previous round (view only)
```

#### Request/Response Examples:

**Create Game:**
```json
POST /api/wizard/games
{
  "league_id": "67abc123...",           // League ID (required)
  "game_name": "Friday Wizard Night",   // GameRound name
  "bid_restriction": "NO_RESTRICTIONS", // or CANNOT_MATCH_CARDS, MUST_MATCH_CARDS
  "game_variant": "STANDARD",
  "first_dealer_index": 0,              // 0-based index in players array
  "player_membership_ids": [
    "67def456...",
    "67def789...",
    "67defabc...",
    "67defdef..."
  ]
}

Response:
{
  "code": "abc123xyz",                  // Wizard game code (for URL)
  "game_round_id": "67ghi123...",       // Created GameRound ID
  "current_round": 1,
  "max_rounds": 15,                     // 60/4 = 15
  "status": "IN_PROGRESS",
  "players": [
    {
      "membership_id": "67def456...",
      "player_name": "Andrij",
      "total_score": 0
    },
    ...
  ]
}
```

**Submit Bids (by index):**
```json
PUT /api/wizard/games/abc123xyz/rounds/1/bids
{
  "bids": [0, 1, 0, 0]  // Array indices correspond to Players array
}

Response: 200 OK
```

**Submit Results (by index):**
```json
PUT /api/wizard/games/abc123xyz/rounds/1/results
{
  "results": [0, 0, 1, 0]  // Array indices correspond to Players array
}

Response: 200 OK
```

**Complete Round:**
```json
POST /api/wizard/games/abc123xyz/rounds/1/complete

Response:
{
  "round_number": 1,
  "player_results": [
    {
      "bid": 0,
      "actual": 0,
      "score": 20,        // +20 (matched bid)
      "delta": 20,
      "total_score": 20
    },
    {
      "bid": 1,
      "actual": 0,
      "score": -10,       // -10 (missed by 1)
      "delta": -10,
      "total_score": -10
    },
    ...
  ]
}
```

**Edit Round (fix mistakes):**
```json
PUT /api/wizard/games/abc123xyz/rounds/3/edit
{
  "bids": [1, 2, 1, 0],      // New bids (optional)
  "results": [1, 1, 1, 0]    // New results (optional)
}

Response:
{
  "round_number": 3,
  "recalculated_rounds": [3, 4, 5, ...],  // All rounds that were recalculated
  "message": "Round 3 updated, recalculated 7 subsequent rounds"
}

// При редагуванні:
// 1. Оновлюємо bid/actual в раунді 3
// 2. Перераховуємо scores для раунду 3
// 3. Перераховуємо total_score для всіх наступних раундів (4, 5, 6...)
```

**Scoreboard:**
```json
GET /api/wizard/games/abc123xyz/scoreboard

Response:
{
  "game_code": "abc123xyz",
  "current_round": 5,
  "max_rounds": 15,
  "players": [
    {
      "player_name": "Andrij",
      "total_score": 120
    },
    ...
  ],
  "rounds": [
    {
      "round_number": 1,
      "dealer_index": 0,
      "cards_count": 1,
      "status": "COMPLETED",
      "player_results": [
        {
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
  ]
}
```

**Finalize Game:**
```json
POST /api/wizard/games/abc123xyz/finalize

Response:
{
  "wizard_game_code": "abc123xyz",
  "game_round_id": "67ghi123...",
  "final_standings": [
    {
      "player_name": "Lena",
      "total_score": 320,
      "position": 1
    },
    {
      "player_name": "Andrij",
      "total_score": 280,
      "position": 2
    },
    ...
  ]
}

// При цьому автоматично оновлюється GameRound:
// - Players[].Score = final scores
// - Players[].Position = 1, 2, 3, ...
// - EndTime = now()
```

#### Scoring Logic (`scoring.go`):

```go
// CalculateRoundScore обчислює очки за раунд
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

// ValidateBids перевіряє чи ставки відповідають правилам
func ValidateBids(game *WizardGame, roundNumber int, bids []int) error {
    if game.Config.BidRestriction == "NO_RESTRICTIONS" {
        return nil
    }

    totalBids := 0
    for _, bid := range bids {
        if bid < 0 {
            return errors.New("bid cannot be negative")
        }
        totalBids += bid
    }

    cardsCount := roundNumber

    switch game.Config.BidRestriction {
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

// CalculateDealerIndex обчислює індекс дилера для раунду
func CalculateDealerIndex(firstDealerIndex int, roundNumber int, playerCount int) int {
    // Дилер ротується: (firstDealer + roundNumber - 1) % playerCount
    return (firstDealerIndex + roundNumber - 1) % playerCount
}

// CompleteRound розраховує очки для всіх гравців після завершення раунду
func CompleteRound(game *WizardGame, roundIndex int) error {
    if roundIndex < 0 || roundIndex >= len(game.Rounds) {
        return errors.New("invalid round index")
    }

    round := &game.Rounds[roundIndex]

    // Перевіряємо чи всі bid/actual встановлені
    for i, pr := range round.PlayerResults {
        if pr.Bid < 0 || pr.Actual < 0 {
            return fmt.Errorf("player %d has missing bid or actual", i)
        }
    }

    // Обчислюємо очки для кожного гравця
    for i := range round.PlayerResults {
        pr := &round.PlayerResults[i]

        // Очки за раунд
        pr.Score = CalculateRoundScore(pr.Bid, pr.Actual)

        // Попередній total score
        prevTotalScore := 0
        if roundIndex > 0 {
            prevTotalScore = game.Rounds[roundIndex-1].PlayerResults[i].TotalScore
        }

        // Кумулятивний рахунок
        pr.TotalScore = prevTotalScore + pr.Score

        // Delta (зміна)
        pr.Delta = pr.Score

        // Оновлюємо загальний рахунок гравця
        game.Players[i].TotalScore = pr.TotalScore
    }

    round.Status = "COMPLETED"
    round.CompletedAt = time.Now()

    return nil
}

// FinalizeGame завершує гру і записує результати в GameRound
func FinalizeGame(wizardGame *WizardGame, gameRound *GameRound) error {
    // Перевіряємо чи всі раунди завершені
    for _, round := range wizardGame.Rounds {
        if round.Status != "COMPLETED" {
            return errors.New("not all rounds are completed")
        }
    }

    // Сортуємо гравців за очками (від більшого до меншого)
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

    // Оновлюємо GameRound з фінальними результатами
    for position, result := range results {
        gameRound.Players[result.Index].Score = int64(result.TotalScore)
        gameRound.Players[result.Index].Position = position + 1
    }

    gameRound.EndTime = time.Now()
    wizardGame.Status = "COMPLETED"

    return nil
}

// RecalculateFromRound перераховує scores починаючи з вказаного раунду
// Використовується при редагуванні помилок
func RecalculateFromRound(game *WizardGame, fromRoundIndex int) error {
    if fromRoundIndex < 0 || fromRoundIndex >= len(game.Rounds) {
        return errors.New("invalid round index")
    }

    // Перераховуємо всі раунди починаючи з fromRoundIndex
    for roundIdx := fromRoundIndex; roundIdx < len(game.Rounds); roundIdx++ {
        round := &game.Rounds[roundIdx]

        // Перевіряємо чи є bid/actual для всіх гравців
        for i, pr := range round.PlayerResults {
            if pr.Bid < 0 || pr.Actual < 0 {
                // Якщо раунд не завершений - пропускаємо
                continue
            }

            pr := &round.PlayerResults[i]

            // Перераховуємо score для раунду
            pr.Score = CalculateRoundScore(pr.Bid, pr.Actual)

            // Попередній total score
            prevTotalScore := 0
            if roundIdx > 0 {
                prevTotalScore = game.Rounds[roundIdx-1].PlayerResults[i].TotalScore
            }

            // Кумулятивний рахунок
            pr.TotalScore = prevTotalScore + pr.Score
            pr.Delta = pr.Score

            // Оновлюємо загальний рахунок гравця
            game.Players[i].TotalScore = pr.TotalScore
        }
    }

    return nil
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

