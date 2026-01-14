# Game Management

*[Українська версія](GAME_MANAGEMENT.md)*

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
- Player recommendations based on co-play history (for leagues)

### Player Recommendation System

For leagues, the system provides intelligent player recommendations based on co-play history:

- **Recent players cache**: Stored in `LeagueMembership.recent_co_players` (maximum 10 entries)
- **Cache updates**: Automatically updated when a game round is finalized
- **Sorting**: Players sorted by last co-play date (`last_played_at` DESC)
- **Activity**: Other players sorted by `last_activity_at` (NULLS LAST)

**API endpoint**: `GET /api/leagues/{code}/suggested-players`

Returns three categories:
- `current_player`: Current user (if membership exists)
- `recent_players`: Up to 10 players from recent co-play cache
- `other_players`: Other league members, sorted by activity

#### TypeScript Types

```typescript
interface SuggestedPlayer {
  membership_id: string;
  alias: string;
  avatar?: string;
  last_played_at?: string;
  is_virtual: boolean;
}

interface SuggestedPlayersResponse {
  current_player: SuggestedPlayer | null;
  recent_players: SuggestedPlayer[];
  other_players: SuggestedPlayer[];
}
```

#### Auto-fill Logic

When creating a new round, the system automatically fills player slots based on recommendations:

```typescript
function autoFillPlayers(gameType: GameType, suggestedPlayers: SuggestedPlayersResponse) {
  const selected: SuggestedPlayer[] = [];
  const maxSlots = gameType.max_players;
  
  // 1. Add current player (if membership exists)
  if (suggestedPlayers.current_player) {
    selected.push(suggestedPlayers.current_player);
  }
  
  // 2. Fill from recent_players
  for (const player of suggestedPlayers.recent_players) {
    if (selected.length >= maxSlots) break;
    if (!selected.find(p => p.membership_id === player.membership_id)) {
      selected.push(player);
    }
  }
  
  // 3. Fill from other_players if space remains
  for (const player of suggestedPlayers.other_players) {
    if (selected.length >= maxSlots) break;
    if (!selected.find(p => p.membership_id === player.membership_id)) {
      selected.push(player);
    }
  }
  
  return selected;
}
```

**Fill Priority:**
1. Current player (if membership exists)
2. Recent co-players (from `recent_co_players` cache)
3. Other league members (sorted by activity)

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

### Leagues (Player Recommendations)

| Method | Endpoint | Description |
|-------|---------------|------|
| GET | `/api/leagues/{code}/suggested-players` | Get recommended players for creating a round |
| POST | `/api/leagues/{code}/invitations` | Create invitation (can include alias for virtual player) |

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

#### GameRoundWizard.vue

Main wizard for creating and continuing game rounds:
- Multi-step interface (stepper)
- Dynamic step determination based on game type
- State saved to server after each step
- Automatic redirect of completed rounds to edit page

#### Wizard Steps (gametypes/steps/)

**Step1GameType.vue** - Game type selection:
- List of available game types
- Localized names and icons
- Player count information

**Step2Players.vue** - Player selection:
- Integration with PlayerSelector
- Auto-fill based on recommended players (`GET /api/leagues/{code}/suggested-players`)
- Moderator support for applicable games (ability to select player for moderator role)
- Min/max count validation
- Create virtual players "on the fly" via invitations
- Two-panel interface: available players (left) and selected players (right)

**Player Selection Flow:**
1. On step load, calls `GET /api/leagues/{code}/suggested-players`
2. System automatically fills slots based on priority (current → recent → other)
3. User can manually add/remove players
4. Can select one player for moderator assignment (for games with moderator)
5. When creating virtual player, automatically added to selected list

**Step3Roles.vue** - Role assignment (for standard games):
- Round name
- Player table with role selection
- Moderator checkbox
- "Save" and "Next" buttons

**Step4Scoring.vue** - Score entry:
- Player table with score and position fields
- "Save" and "Finish Game" buttons

#### PlayerSelector.vue (components/game/)

Two-panel player selection component:
- **Left panel**: Available players
  - Single list (recent + other mixed by priority)
  - Search/filter
  - "Add" button next to each player
- **Right panel**: Selected players
  - Selected list
  - "Remove" button next to each
  - Ability to select one player (click) - for moderator assignment
  - Current player marked (if exists)
- `min_players` / `max_players` indicator
- Warning if less than `min_players`
- Selected player passed as moderator to next step (for games with moderator)

#### CreateVirtualPlayerDialog.vue

Dialog for creating virtual player:
- Alias input field
- Uniqueness check (debounced) - via existing validation
- "Create" button
- Uses `POST /api/leagues/{code}/invitations` with alias
- On success: toast + copy link to clipboard
- Automatically adds to selected list

#### GameroundEdit.vue

Simplified form for editing **completed** game rounds:
- Edit round name
- View and edit player scores
- Edit positions

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

## Game Round Creation Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Create Game Round                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    STEP 1: Select Game Type                   │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  🎲 Mafia              [4-10 players, with moderator]│    │
│  │  🎯 Codenames          [4-8 players, teams]       │    │
│  │  🃏 Poker              [2-6 players]              │    │
│  │  ...                                                │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│                                          [Next →]            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  GET /api/leagues/{code}/suggested-players                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 1. Check permissions (league member / superadmin)   │   │
│  │ 2. Get current user's membership                     │   │
│  │ 3. Get recent_co_players from cache                  │   │
│  │ 4. Get other league members (by last_activity_at)   │   │
│  │ 5. Form response                                     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    STEP 2: Select Players                    │
│  ┌─────────────────────┐    ┌─────────────────────┐         │
│  │   Available Players  │    │   Selected Players   │         │
│  │   ───────────────    │    │   ─────────────     │         │
│  │   • Player A    [+]  │ ←→ │   • Player 1   [x]  │         │
│  │   • Player B    [+]  │    │   • Player 2   [x]  │ ← selected (moderator) │
│  │   • Player C    [+]  │    │   • ...              │         │
│  │   • Player D    [+]  │    │                      │         │
│  │   ...                │    │   [+ Virtual]       │         │
│  └─────────────────────┘    └─────────────────────┘         │
│                                                              │
│  🔍 [Search...]          min: 4 / selected: 5 / max: 6         │
│                                                              │
│  💡 Click on player on right to assign as moderator          │
│                              [← Back]  [Next →]             │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┴─────────────────────┐
        │ [+ Virtual] clicked                        │
        ▼                                           │
┌───────────────────────────────┐                   │
│  POST /api/leagues/{code}/    │                   │
│       invitations             │                   │
│  ┌─────────────────────────┐  │                   │
│  │ { "alias": "New" }      │  │                   │
│  │ → Create membership     │  │                   │
│  │ → Create invitation     │  │                   │
│  │ → Update last_activity  │  │                   │
│  │ → Update cache (end)    │  │                   │
│  │ → Return token          │  │                   │
│  └─────────────────────────┘  │                   │
│  Response: invitation + link  │                   │
└───────────────────────────────┘                   │
        │                                           │
        │ Toast: "Link copied!"                     │
        │ Add player to selected                    │
        └─────────────────────────────────────────→─┤
                                                    │
                              ┌─────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    STEP 3: Configuration                     │
│                                                              │
│  Round name: [___________________]                          │
│                                                              │
│  Players and roles:                                         │
│  ┌────────────────┬──────────────────────────────┐          │
│  │ Player         │ Role                         │          │
│  ├────────────────┼──────────────────────────────┤          │
│  │ Player A       │ [🔴 Mafia      ▼]            │          │
│  │ Player B       │ [👁 Moderator  ▼] ← from step 2│          │
│  │ Player C       │ [🔵 Civilian   ▼]            │          │
│  │ ...            │ ...                          │          │
│  └────────────────┴──────────────────────────────┘          │
│                                                              │
│  💡 Roles are determined by game type (gameType.roles)      │
│                              [← Back]  [Save]               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Round Finalization (later)                      │
│  ───────────────────────────────────────                     │
│  On finalization → update last_activity_at and             │
│  recent_co_players for all round participants                │
└─────────────────────────────────────────────────────────────┘
```

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
