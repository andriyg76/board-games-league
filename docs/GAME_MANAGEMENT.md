# Game Management

## Overview

The Board Games League application provides comprehensive game management functionality, including game types configuration, game rounds tracking, and player management. This document covers all game-related features implemented in the system.

## Architecture

### Data Models

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    GameType     │     │    GameRound    │     │      User       │
│                 │     │                 │     │    (Player)     │
│ - ID            │◄────│ - GameTypeID    │     │                 │
│ - Name          │     │ - Name          │     │ - ID            │
│ - ScoringType   │     │ - StartTime     │     │ - Alias         │
│ - Labels        │     │ - EndTime       │     │ - Avatar        │
│ - Teams         │     │ - Players[]     │◄────│ - Name          │
│ - MinPlayers    │     │ - TeamScores[]  │     │                 │
│ - MaxPlayers    │     │ - CoopScore     │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### Component Structure

```
Backend                          Frontend
├── gameapi/                     ├── api/
│   ├── handlers.go             │   └── GameApi.ts
│   ├── gametype.go             ├── gametypes/
│   ├── gameround.go            │   ├── ListGameTypes.vue
│   └── player.go               │   ├── GameroundsList.vue
├── models/                      │   ├── GameroundEdit.vue
│   ├── game_type.go            │   ├── FinalizeGameDialog.vue
│   └── game_round.go           │   └── types.ts
└── repositories/                └── store/
    ├── game_type_repository.go     ├── game.ts
    └── game_round_repository.go    └── player.ts
```

## Game Types

### Concept

A Game Type defines the rules and scoring mechanism for a specific board game. Each game type has:

- **Name**: The display name of the game
- **Scoring Type**: How scores are calculated and winners determined
- **Labels**: Optional categorization tags with colors and icons
- **Teams**: Optional team definitions for team-based games
- **Player Limits**: Minimum and maximum number of players

### Scoring Types

The system supports multiple scoring mechanisms:

| Scoring Type | Description | Use Case |
|--------------|-------------|----------|
| `classic` | Traditional competitive scoring - highest score wins | Most board games (Catan, Ticket to Ride) |
| `cooperative` | All players win or lose together | Pandemic, Forbidden Island |
| `cooperative_with_moderator` | Cooperative with a separate moderator | Games with a game master |
| `team_vs_team` | Teams compete against each other | Team-based games |
| `mafia` | Team vs Team with hidden roles and moderator | Mafia, Werewolf |
| `custom` | No predefined scheme - raw score entry | Any custom scoring |

### Labels

Labels are used to categorize players within a game (e.g., roles, starting positions). Each label has:

- **Name**: Label identifier
- **Color**: Display color (hex format)
- **Icon**: Material Design icon name

Labels are available for scoring types: `classic`, `custom`

### Teams

Teams group players for team-based games. Each team has:

- **Name**: Team identifier
- **Color**: Team color (hex format)
- **Icon**: Team icon

Teams are available for scoring types: `mafia`, `custom`, `team_vs_team`

### Model Structure

```go
type GameType struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Version     int64              `bson:"version"`
    Name        string             `bson:"name"`
    ScoringType string             `bson:"scoring_type"`
    Icon        string             `bson:"icon"`
    Labels      []Label            `bson:"labels"`
    Teams       []Label            `bson:"teams"`
    MinPlayers  int                `bson:"min_players"`
    MaxPlayers  int                `bson:"max_players"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}

type Label struct {
    Name  string `bson:"name"`
    Color string `bson:"color"`
    Icon  string `bson:"icon"`
}
```

## Game Rounds

### Concept

A Game Round represents a single play session of a game type. It tracks:

- Which game type is being played
- When the game started and ended
- Who participated and their positions/scores
- Team scores (for team games)
- Cooperative score (for cooperative games)

### Round Lifecycle

```
┌─────────┐     ┌───────────┐     ┌───────────┐     ┌──────────┐
│  Start  │────>│  Active   │────>│ Finalize  │────>│ Completed│
│  Game   │     │  (Playing)│     │  (Scores) │     │          │
└─────────┘     └───────────┘     └───────────┘     └──────────┘
     │                │                  │                │
     │                │                  │                │
  Create          Update             Set final        EndTime
  round           scores             positions        is set
```

### Model Structure

```go
type GameRound struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    Version          int64              `bson:"version"`
    Name             string             `bson:"name"`
    GameTypeID       primitive.ObjectID `bson:"game_type_id"`
    StartTime        time.Time          `bson:"start_time"`
    EndTime          time.Time          `bson:"end_time"`
    Players          []GameRoundPlayer  `bson:"players"`
    TeamScores       []TeamScore        `bson:"team_scores"`
    CooperativeScore int64              `bson:"cooperative_score"`
    CreatedAt        time.Time          `bson:"created_at"`
    UpdatedAt        time.Time          `bson:"updated_at"`
}

type GameRoundPlayer struct {
    PlayerID    primitive.ObjectID `bson:"player_id"`
    Position    int                `bson:"position"`
    Score       int64              `bson:"score"`
    IsModerator bool               `bson:"is_moderator"`
    TeamName    string             `bson:"team_name"`
    LabelName   string             `bson:"label_name"`
}

type TeamScore struct {
    Name     string `bson:"name"`
    Score    int64  `bson:"score"`
    Position int    `bson:"position"`
}
```

### Team Validation

For games with teams defined:
- Each team must have at least one player assigned
- Players must be assigned to existing teams only
- The validation occurs when starting a new game round

## Players

### Concept

Players are users who participate in game rounds. The player system provides:

- List of all registered players
- Individual player lookup by code
- Current authenticated player info

### Player Information

```go
type Player struct {
    Code   string `json:"code"`
    Alias  string `json:"alias"`
    Avatar string `json:"avatar,omitempty"`
}
```

The `code` is a unique identifier derived from the user's MongoDB ObjectID.

## API Endpoints

### Game Types

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/game_types` | List all game types |
| POST | `/api/game_types` | Create a new game type |
| GET | `/api/game_types/{code}` | Get a specific game type |
| PUT | `/api/game_types/{code}` | Update a game type |
| DELETE | `/api/game_types/{code}` | Delete a game type |

### Game Rounds

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/game_rounds` | List all game rounds |
| POST | `/api/game_rounds` | Start a new game round |
| GET | `/api/game_rounds/{code}` | Get a specific game round |
| PUT | `/api/game_rounds/{code}` | Update a game round |
| PUT | `/api/game_rounds/{code}/finalize` | Finalize a game round |
| PUT | `/api/game_rounds/{code}/players/{userId}/score` | Update player score |

### Players

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/players` | List all players |
| GET | `/api/players/{code}` | Get a specific player |
| GET | `/api/players/i_am` | Get current authenticated player |

## Frontend Components

### Views and Components

#### ListGameTypes.vue

Admin interface for managing game types:
- View list of all game types
- Create new game types
- Edit existing game types
- Delete game types
- Configure labels and teams with color pickers

#### GameroundsList.vue

Main game rounds management view:
- Display list of all game rounds
- Show round status (active/completed)
- Edit active rounds
- Finalize games
- Create new rounds

#### GameroundEdit.vue

Form for creating and editing game rounds:
- Select game type
- Add/remove players
- Assign teams and labels
- Set moderator status
- Update player scores

#### FinalizeGameDialog.vue

Dialog for ending a game round:
- Enter final player scores
- Enter team scores (if applicable)
- Enter cooperative score (if applicable)
- Calculates final positions

### Pinia Stores

#### game.ts

State management for game data:
- `gameTypes`: Cached list of game types
- `activeRounds`: Currently active game rounds
- `loadGameTypes()`: Fetch game types from API
- `addActiveRound()`: Create new game round
- `updateRound()`: Update existing round
- `getGameTypeByCode`: Find game type by code

#### player.ts

State management for player data:
- `players`: Cached list of players
- `currentPlayer`: Currently authenticated player
- `allPlayers`: Get all players (with lazy loading)
- `getPlayerByCode`: Find player by code

## Usage Examples

### Creating a Game Type

```typescript
const gameType: GameType = {
    name: "Catan",
    scoring_type: "classic",
    labels: [
        { name: "First Player", color: "#FF0000", icon: "mdi-flag" }
    ],
    teams: [],
    min_players: 3,
    max_players: 4
};

await GameApi.createGameType(gameType);
```

### Starting a Game Round

```typescript
const round: GameRound = {
    name: "Friday Night Catan",
    game_type: "Catan",
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1_code", is_moderator: false },
        { user_id: "player2_code", is_moderator: false },
        { user_id: "player3_code", is_moderator: false }
    ]
};

await GameApi.createGameRound(round);
```

### Finalizing a Game

```typescript
const finalizationData: FinalizeGameRoundRequest = {
    player_scores: {
        "player1_code": 10,
        "player2_code": 8,
        "player3_code": 7
    }
};

await GameApi.finalizeGameRound(roundCode, finalizationData);
```

### Team Game Example

```typescript
// Game type with teams
const mafiaGame: GameType = {
    name: "Mafia",
    scoring_type: "mafia",
    labels: [],
    teams: [
        { name: "Civilians", color: "#00FF00", icon: "mdi-account-group" },
        { name: "Mafia", color: "#FF0000", icon: "mdi-skull" }
    ],
    min_players: 6,
    max_players: 15
};

// Starting a game with teams
const mafiaRound: GameRound = {
    name: "Epic Mafia Night",
    game_type: "Mafia",
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1", is_moderator: true, team_name: "" },
        { user_id: "player2", is_moderator: false, team_name: "Civilians" },
        { user_id: "player3", is_moderator: false, team_name: "Mafia" },
        // ... more players
    ]
};
```

## Database Collections

### game_types

Stores game type definitions:
- Index: `name` (unique)

### game_rounds

Stores game round data:
- Index: `game_type_id`
- Index: `start_time`
- Index: `players.player_id`

## Frontend Routes

| Route | Component | Description |
|-------|-----------|-------------|
| `/ui/admin/game-types` | ListGameTypes.vue | Game types management |
| `/ui/game-rounds` | GameroundsList.vue | Game rounds list |
| `/ui/game-rounds/new` | GameroundEdit.vue | Create new round |
| `/ui/game-rounds/:id` | GameroundEdit.vue | Edit existing round |
| `/ui/leagues` | ListLeagues.vue | Leagues view (future) |

## Internationalization

The frontend supports multiple languages using vue-i18n:

- `gameTypes.title`: "Game Types" title
- `gameTypes.edit`: Edit button text
- `gameTypes.delete`: Delete button text
- `gameTypes.create`: Create button text
- `scoring.*`: Scoring type descriptions

## Future Enhancements

Potential improvements:
- Statistics and leaderboards
- Player rating system (ELO)
- Game history and replays
- Tournament management
- League standings
- Achievement system
- Game recommendations
- Export/import game data
