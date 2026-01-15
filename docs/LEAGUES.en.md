# League System

*[Українська версія](LEAGUES.md)*

## Overview

The League System provides isolated competitive contexts where players can compete, earn points, and track rankings. Each league has its own members, standings, and game rounds.

---

## Concept and Requirements

### League Concept

**League** - is an isolated gaming space where:
- Players are members of a specific league
- All game rounds are created within the league
- When creating a game, only players from the current league can be selected
- Player points and statistics are calculated separately for each league
- Most interfaces work in the context of the selected league
- Game types remain global (not tied to a league)

### Business Rules

#### 1. League Creation and Management
- **League creation**: Only superadmin can create a new league
- **League fields**: Name (required field), description (optional)
- **Administration**: No separate administrators/owners of the league - all players are equal
- **Complaints**: Players can report another player to the administrator

#### 2. League Membership
- **Multiple membership**: A player can be a member of many leagues simultaneously
- **Joining a league**: Through the invitation system
  - Superadmin can generate invitations
  - Any league player can generate invitations
  - Invitation = one-time link with token
- **Invitation process**:
  1. User follows the invitation link
  2. If not logged in → redirect to login → account creation/login
  3. After successful authentication → automatic addition to the league
  4. Invitation link becomes invalid after use (one-time)
- **Leaving league**: User can leave the league
- **Ban**: Only superadmin can ban a player in a specific league

#### 3. Game Rounds in League Context
- Each game round (game round) can be linked to a specific league
- When creating a round, only players from that league are available
- Round list is filtered by selected league

#### 4. Rating System

**Player's total rating in league** consists of:
- **Participation points**: Fixed points per game played (2 points)
- **Position points**: Depending on place in the game
  - 1st place: 10 points
  - 2nd place: 6 points
  - 3rd place: 3 points
  - 4th+ places: 1 point
- **Moderation points**: Additional points if player was moderator (1 point)

**Formula:**
```
Player Rating = Σ (Participation_Points + Position_Points + Moderation_Points)
```

**Standings sorting:**
1. Total points (descending)
2. Number of games (ascending - for tie-breaking)
3. User ID (for consistent ordering)

#### 5. UI/UX
- **League selector**: Interface has ability to select current league
- **Context**: Most operations occur in the context of selected league
- **Game types**: Remain global and available for all leagues

---

## Architecture

### Data Models (Backend)

#### League Model (`league.go`)
```go
type LeagueStatus string

const (
    LeagueActive   LeagueStatus = "active"
    LeagueArchived LeagueStatus = "archived"
)

type League struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Version     int64              `bson:"version"`
    Name        string             `bson:"name"`
    Description string             `bson:"description,omitempty"`
    Status      LeagueStatus       `bson:"status"`      // "active" | "archived"
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
    CreatedBy   primitive.ObjectID `bson:"created_by"`
}
```

#### League Membership Model (`league_membership.go`)
```go
type LeagueMembershipStatus string

const (
    MembershipActive  LeagueMembershipStatus = "active"
    MembershipPending LeagueMembershipStatus = "pending"
    MembershipVirtual LeagueMembershipStatus = "virtual"
    MembershipBanned   LeagueMembershipStatus = "banned"
)

type RecentCoPlayer struct {
    MembershipID primitive.ObjectID `bson:"membership_id"`
    LastPlayedAt time.Time          `bson:"last_played_at"`
}

type LeagueMembership struct {
    ID              primitive.ObjectID     `bson:"_id,omitempty"`
    Version         int64                  `bson:"version"`
    LeagueID        primitive.ObjectID     `bson:"league_id"`
    UserID          primitive.ObjectID     `bson:"user_id,omitempty"` // optional for pending/virtual
    Status          LeagueMembershipStatus `bson:"status"`
    JoinedAt        time.Time              `bson:"joined_at"`
    RecentCoPlayers []RecentCoPlayer       `bson:"recent_co_players,omitempty"` // max 10 items
    LastActivityAt  time.Time              `bson:"last_activity_at,omitempty"`  // last activity timestamp
    CreatedAt       time.Time              `bson:"created_at"`
    UpdatedAt       time.Time              `bson:"updated_at"`
}
```

**Recent Co-Players Cache:**
- Maximum 10 items stored per membership
- Automatically updated when a game round is finalized
- Sorted by `last_played_at` DESC (most recent first)
- When cache is full, oldest entry is removed when adding new players

**Last Activity Tracking:**
The `last_activity_at` field is updated in the following scenarios:

| Event | Updated For |
|-------|-------------|
| Creating named invitation (with alias) | New pending membership (`time.Now()`) |
| Game round finalization | All participants of the game (`time.Now()`) |

**Cache Update Logic:**
When a game round is finalized:
1. Update `last_activity_at = now` for all participants
2. For each player in the round:
   - Add all other players to their `recent_co_players` with `last_played_at = now`
   - If player already exists in cache, update `last_played_at`
   - Keep max 10 entries, sorted by `last_played_at` DESC

When creating a virtual player invitation:
1. Set `last_activity_at = now` for the new pending membership
2. Add the new player to creator's `recent_co_players` at the **end** of the list
3. If cache is full (10 entries), remove the oldest entry

#### League Invitation Model (`league_invitation.go`)
```go
type LeagueInvitation struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Version   int64              `bson:"version"`
    LeagueID  primitive.ObjectID `bson:"league_id"`
    CreatedBy primitive.ObjectID `bson:"created_by"` // User who created the invitation
    Token     string             `bson:"token"`      // Unique token for link
    IsUsed    bool               `bson:"is_used"`
    UsedBy    primitive.ObjectID `bson:"used_by,omitempty"`
    UsedAt    time.Time          `bson:"used_at,omitempty"`
    ExpiresAt time.Time          `bson:"expires_at,omitempty"` // Expiry (7 days)
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

#### League Standings (calculated, not stored in DB)
```go
type LeagueStanding struct {
    UserID              primitive.ObjectID `json:"user_id"`
    UserName            string             `json:"user_name"`
    UserAvatar          string             `json:"user_avatar"`
    TotalPoints         int64              `json:"total_points"`
    GamesPlayed         int                `json:"games_played"`
    GamesModerated      int                `json:"games_moderated"`
    FirstPlaceCount     int                `json:"first_place_count"`
    SecondPlaceCount    int                `json:"second_place_count"`
    ThirdPlaceCount     int                `json:"third_place_count"`
    ParticipationPoints int64              `json:"participation_points"`
    PositionPoints      int64              `json:"position_points"`
    ModerationPoints    int64              `json:"moderation_points"`
}
```

#### GameRound Model Update
```go
type GameRound struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    Code             string             `bson:"-"`
    Version          int64              `bson:"version"`
    LeagueID         primitive.ObjectID `bson:"league_id,omitempty"` // Optional
    Name             string             `bson:"name"`
    GameTypeID       primitive.ObjectID `bson:"game_type_id,omitempty"`
    StartTime        time.Time          `bson:"start_time"`
    EndTime          time.Time          `bson:"end_time"`
    Players          []GameRoundPlayer  `bson:"players"`
    TeamScores       []TeamScore        `bson:"team_scores,omitempty"`
    CooperativeScore int64              `bson:"cooperative_score,omitempty"`
    CreatedAt        time.Time          `bson:"created_at"`
    UpdatedAt        time.Time          `bson:"updated_at"`
}
```

### Database Collections

#### leagues
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "name": String,
  "description": String,
  "status": String, // "active" | "archived"
  "created_at": ISODate,
  "updated_at": ISODate,
  "created_by": ObjectId
}
```

**Indexes:**
- `{ "name": 1 }` - unique
- `{ "status": 1 }`

#### league_memberships
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "league_id": ObjectId,
  "user_id": ObjectId, // optional for pending/virtual
  "status": String, // "active" | "pending" | "virtual" | "banned"
  "joined_at": ISODate,
  "recent_co_players": [
    {
      "membership_id": ObjectId,
      "last_played_at": ISODate
    }
  ],
  "last_activity_at": ISODate,
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Indexes:**
- `{ "league_id": 1, "user_id": 1 }` - unique compound index
- `{ "league_id": 1 }`
- `{ "user_id": 1 }`
- `{ "league_id": 1, "status": 1 }`

#### league_invitations
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "league_id": ObjectId,
  "created_by": ObjectId,
  "token": String,
  "is_used": Boolean,
  "used_by": ObjectId,
  "used_at": ISODate,
  "expires_at": ISODate,
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Indexes:**
- `{ "token": 1 }` - unique
- `{ "league_id": 1 }`
- `{ "is_used": 1 }`
- `{ "expires_at": 1 }` - TTL index for automatic deletion of expired invitations

#### game_rounds (update)
Add field:
```json
{
  "league_id": ObjectId, // optional
  // ... existing fields
}
```

**New Index:**
- `{ "league_id": 1, "start_time": -1 }`

---

## API Documentation

### Authentication

All league endpoints require authentication via JWT token in the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

### Endpoints

#### 1. List All Leagues

**Endpoint:** `GET /api/leagues`

**Description:** Returns all active leagues accessible to the authenticated user.

**Response:**
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "code": "ABC123",
    "name": "Summer Championship 2026",
    "description": "Competitive league for summer season",
    "status": "active",
    "created_at": "2026-01-01T00:00:00Z",
    "created_by": "507f1f77bcf86cd799439012",
    "member_count": 15
  }
]
```

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

#### 2. Get League Details

**Endpoint:** `GET /api/leagues/{code}`

**Description:** Returns detailed information about a specific league.

**URL Parameters:**
- `code` - League code (base64url encoded ObjectID)

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "code": "ABC123",
  "name": "Summer Championship 2026",
  "description": "Competitive league for summer season",
  "status": "active",
  "created_at": "2026-01-01T00:00:00Z",
  "created_by": "507f1f77bcf86cd799439012",
  "member_count": 15,
  "game_count": 42
}
```

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

#### 3. Create League

**Endpoint:** `POST /api/leagues`

**Description:** Creates a new league. **Requires superadmin privileges.**

**Request Body:**
```json
{
  "name": "Summer Championship 2026",
  "description": "Competitive league for summer season"
}
```

**Response:**
```json
{
  "league": {
    "id": "507f1f77bcf86cd799439011",
    "code": "ABC123",
    "name": "Summer Championship 2026",
    "description": "Competitive league for summer season",
    "status": "active",
    "created_at": "2026-01-01T00:00:00Z",
    "created_by": "507f1f77bcf86cd799439012"
  }
}
```

**Status Codes:**
- `201 Created` - League created successfully
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a superadmin
- `500 Internal Server Error` - Server error

---

#### 4. Update League

**Endpoint:** `PUT /api/leagues/{code}`

**Description:** Updates league details. **Requires superadmin privileges.**

**URL Parameters:**
- `code` - League code

**Request Body:**
```json
{
  "name": "Updated League Name",
  "description": "Updated description"
}
```

**Response:**
```json
{
  "league": {
    "id": "507f1f77bcf86cd799439011",
    "code": "ABC123",
    "name": "Updated League Name",
    "description": "Updated description",
    "status": "active",
    "created_at": "2026-01-01T00:00:00Z",
    "created_by": "507f1f77bcf86cd799439012"
  }
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a superadmin
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

#### 5. Archive/Unarchive League

**Endpoint:** `PUT /api/leagues/{code}/status`

**Description:** Changes league status (active/archived). **Requires superadmin privileges.**

**URL Parameters:**
- `code` - League code

**Request Body:**
```json
{
  "status": "archived"
}
```

**Valid status values:** `active`, `archived`

**Response:**
```json
{
  "league": {
    "id": "507f1f77bcf86cd799439011",
    "code": "ABC123",
    "status": "archived"
  }
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid status value
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a superadmin
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

#### 6. Get League Standings

**Endpoint:** `GET /api/leagues/{code}/standings`

**Description:** Returns current rankings for all members of a league.

**URL Parameters:**
- `code` - League code

**Response:**
```json
[
  {
    "user_id": "507f1f77bcf86cd799439012",
    "user_name": "John Doe",
    "user_avatar": "https://example.com/avatar.jpg",
    "total_points": 1250,
    "games_played": 42,
    "games_moderated": 5,
    "participation_points": 420,
    "position_points": 780,
    "moderation_points": 50,
    "first_place_count": 8,
    "second_place_count": 12,
    "third_place_count": 10
  }
]
```

**Standings are sorted by:**
1. Total points (descending)
2. Games played (ascending - tie-breaker)
3. User ID (for consistent ordering)

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

#### 7. Get League Members

**Endpoint:** `GET /api/leagues/{code}/members`

**Description:** Returns list of all members in a league with their status.

**URL Parameters:**
- `code` - League code

**Response:**
```json
[
  {
    "user_id": "507f1f77bcf86cd799439012",
    "user_name": "John Doe",
    "user_avatar": "https://example.com/avatar.jpg",
    "status": "active",
    "joined_at": "2026-01-01T00:00:00Z"
  },
  {
    "user_id": "507f1f77bcf86cd799439013",
    "user_name": "Jane Smith",
    "user_avatar": "https://example.com/avatar2.jpg",
    "status": "banned",
    "joined_at": "2026-01-02T00:00:00Z"
  }
]
```

**Member Status Values:**
- `active` - Member can participate in games
- `pending` - Created via invitation, waiting for user to accept
- `virtual` - Player participated in games but never logged in
- `banned` - User banned from the league

**Note:** Banned users are displayed at the end of the list. Among non-banned users, sorting is by join date (newest first).

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

#### 8. Ban Member

**Endpoint:** `POST /api/leagues/{code}/ban/{userCode}`

**Description:** Bans a league member. **Requires superadmin privileges.** Superadmin cannot ban themselves.

**URL Parameters:**
- `code` - League code
- `userCode` - User code

**Request Body:** Not required

**Response:**
- `200 OK` - User successfully banned

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Cannot ban yourself or user is already banned
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a superadmin
- `404 Not Found` - League or member not found
- `500 Internal Server Error` - Server error

---

#### 9. Unban Member

**Endpoint:** `POST /api/leagues/{code}/unban/{userCode}`

**Description:** Unbans a league member. **Requires superadmin privileges.**

**URL Parameters:**
- `code` - League code
- `userCode` - User code

**Request Body:** Not required

**Response:**
- `200 OK` - User successfully unbanned

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - User is not banned
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a superadmin
- `404 Not Found` - League or member not found
- `500 Internal Server Error` - Server error

**Note:** Banned users are displayed at the end of the league members list.

---

#### 10. Get Suggested Players

**Endpoint:** `GET /api/leagues/{code}/suggested-players`

**Description:** Returns suggested players for creating a game round, based on recent co-players history and league membership activity. This endpoint provides intelligent player recommendations for the game setup wizard.

**URL Parameters:**
- `code` - League code

**Response:**
```json
{
  "current_player": {
    "membership_id": "507f1f77bcf86cd799439012",
    "alias": "John",
    "avatar": "https://example.com/avatar.jpg",
    "is_member": true
  },
  "recent_players": [
    {
      "membership_id": "507f1f77bcf86cd799439013",
      "alias": "Jane",
      "avatar": "https://example.com/avatar2.jpg",
      "last_played_at": "2026-01-10T18:00:00Z",
      "is_virtual": false
    }
  ],
  "other_players": [
    {
      "membership_id": "507f1f77bcf86cd799439014",
      "alias": "Bob",
      "avatar": "https://example.com/avatar3.jpg",
      "is_virtual": true
    }
  ]
}
```

**Response Structure:**
- `current_player`: The authenticated user's membership (null if superadmin without membership)
- `recent_players`: Up to 10 players recently played with, sorted by `last_played_at` DESC
- `other_players`: Other league members (excluding current + recent), sorted by `last_activity_at` DESC

**For League Members:**
- `current_player`: 1 item (if membership exists)
- `recent_players`: Up to 10 items from `recent_co_players` cache
- `other_players`: Up to 10 items sorted by activity

**For Superadmin without Membership:**
- `current_player`: null
- `recent_players`: Empty array
- `other_players`: Up to 20 items sorted by activity

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league and not a superadmin
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

**Notes:**
- Recent players are cached in `LeagueMembership.recent_co_players` (max 10 items)
- Cache is updated automatically when a game round is finalized
- Virtual players (created via invitations) are included in the response

---

#### 11. Create Invitation

**Endpoint:** `POST /api/leagues/{code}/invitations`

**Description:** Creates a one-time invitation link for the league. Can also create a virtual player by providing an alias. Valid for 7 days.

**URL Parameters:**
- `code` - League code

**Request Body (Optional):**
```json
{
  "alias": "NewPlayer"
}
```

If `alias` is provided:
- Creates a virtual player (pending membership) with the specified alias
- Checks uniqueness of alias among active members and pending invitations
- Creates `LeagueMembership` with status `pending`
- Creates `LeagueInvitation` with token
- Updates the current user's `recent_co_players` cache (adds new player at the end)
- Sets `last_activity_at` for the new pending membership

**Response:**
```json
{
  "invitation": {
    "token": "abc123def456",
    "league_id": "507f1f77bcf86cd799439011",
    "created_by": "507f1f77bcf86cd799439012",
    "created_at": "2026-01-12T00:00:00Z",
    "expires_at": "2026-01-19T00:00:00Z",
    "used": false
  },
  "invitation_link": "https://example.com/ui/leagues/join/abc123def456"
}
```

**Status Codes:**
- `201 Created` - Invitation created successfully
- `400 Bad Request` - Invalid alias or alias already exists
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

**Notes:**
- If alias is provided, the virtual player is immediately added to the creator's `recent_co_players` cache
- If cache is full (10 items), the oldest entry is removed
- Virtual players can participate in games before accepting the invitation

---

#### 12. Preview Invitation (Public)

**Endpoint:** `GET /api/leagues/join/{token}/preview`

**Description:** Returns public information about an invitation. This endpoint does not require authentication and can be used to show invitation details before login.

**URL Parameters:**
- `token` - Invitation token

**Response:**
```json
{
  "league_name": "Summer Championship 2026",
  "inviter_alias": "John",
  "player_alias": "NewPlayer",
  "expires_at": "2026-01-15T00:00:00Z",
  "status": "valid"
}
```

**Status field values:**
- `valid` - Invitation can be accepted
- `expired` - Invitation has expired
- `used` - Invitation has already been used

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Invitation not found

---

#### 12. Accept Invitation

**Endpoint:** `POST /api/leagues/join/{token}`

**Description:** Accepts an invitation and adds the user to the league. The pending membership created with the invitation is activated and linked to the user.

**URL Parameters:**
- `token` - Invitation token

**Request Body:** None

**Response (Success):**
```json
{
  "code": "ABC123",
  "name": "Summer Championship 2026",
  "status": "active",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**Response (Already Member - 409 Conflict):**
```json
{
  "error": "user is already a member of this league",
  "league_code": "ABC123"
}
```

**Status Codes:**
- `200 OK` - Invitation accepted successfully
- `400 Bad Request` - Invalid invitation (expired, used, own invitation)
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Invitation not found
- `409 Conflict` - User is already a member (includes league_code for redirect)

**Notes:**
- Each invitation can only be used once
- Invitations expire after 7 days
- Users cannot accept their own invitations
- Already member error (409) includes `league_code` for frontend redirect

---

#### 14. Leave League

**Endpoint:** `DELETE /api/leagues/{code}/members/me`

**Description:** Removes the authenticated user from the league.

**URL Parameters:**
- `code` - League code

**Request Body:** None

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `204 No Content` - User removed successfully
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

## Points System

### Participation Points
- **2 points** per game played (awarded to all players)

### Position Points
Based on final position in each game:
- **1st place:** 10 points
- **2nd place:** 6 points
- **3rd place:** 3 points
- **4th+ places:** 1 point

### Moderation Points
- **1 point** per game moderated

### Total Points Calculation
```
Total Points = Participation Points + Position Points + Moderation Points
```

### Standings Calculation
Standings are recalculated from game history:
1. Fetch all game rounds for the league
2. For each game round, calculate points for each player
3. Aggregate points by player
4. Sort by total points (desc), then by games played (asc)

---

## League Game Rounds

When creating a game round, you can optionally associate it with a league:

**Endpoint:** `POST /api/games`

**Request Body:**
```json
{
  "name": "Game 1",
  "type": "Carcassonne",
  "start_time": "2026-01-12T18:00:00Z",
  "league_id": "507f1f77bcf86cd799439011",
  "players": [
    {
      "user_id": "507f1f77bcf86cd799439012",
      "position": 1,
      "team_name": "Red"
    }
  ]
}
```

When a game round is finalized, league standings are automatically updated.

---

## Important Notes

### Migration
- **No migration required**: Empty `recent_co_players` arrays will be automatically populated when new game rounds are finalized
- Empty `last_activity_at` values are sorted to the end of lists (NULLS LAST)

### Superadmin Access
- Superadmins can create game rounds without membership, but will not be automatically added to the player list
- Superadmins without membership receive up to 20 `other_players` in suggested players response (vs 10 for regular members)

### Virtual Players
- When created via invitation with alias, virtual players are immediately added to the creator's `recent_co_players` cache at the **end** of the list
- If cache is full (10 items), the oldest entry is removed
- Virtual players can participate in games before accepting the invitation
- If an invitation expires or is cancelled, a new invitation can be created for the same virtual player (with the same alias)
- In the UI, virtual users (status = 'virtual' or 'pending') display an "Open Invitation" button if there is an active invitation
- If the invitation has expired, the system offers to extend it for 7 days

### Invitation Endpoint
- The `POST /api/leagues/{code}/invitations` endpoint uses **existing** functionality, extended with cache updates
- Uniqueness check validates alias among:
  - Active league members
  - Pending members (active invitations)
  - Virtual members

### Player Selection Flow
The suggested players system provides intelligent recommendations based on:
1. **Current player**: Authenticated user's membership (if exists)
2. **Recent players**: Up to 10 players from `recent_co_players` cache, sorted by `last_played_at` DESC
3. **Other players**: Remaining league members (excluding current + recent), sorted by `last_activity_at` DESC NULLS LAST

---

## Frontend Integration

### League Store (Pinia)

The frontend uses a Pinia store (`src/store/league.ts`) for state management:

```typescript
import { useLeagueStore } from '@/store/league'

const leagueStore = useLeagueStore()

// Fetch all leagues
await leagueStore.fetchLeagues()

// Fetch specific league
await leagueStore.fetchLeague('ABC123')

// Create invitation
const result = await leagueStore.createInvitation()
console.log(result.invitation_link)

// Create invitation with virtual player
const resultWithPlayer = await leagueStore.createInvitation({ alias: "NewPlayer" })

// Get suggested players for game setup
const suggested = await leagueStore.getSuggestedPlayers('ABC123')
// Returns: { current_player, recent_players, other_players }

// Accept invitation
await leagueStore.acceptInvitation('token123')
```

### Available Store Actions

- `fetchLeagues()` - Load all accessible leagues
- `fetchLeague(code)` - Load specific league details
- `createLeague(data)` - Create new league (superadmin only)
- `updateLeague(code, data)` - Update league (superadmin only)
- `updateLeagueStatus(code, status)` - Archive/unarchive league (superadmin only)
- `fetchStandings(code)` - Load league standings
- `fetchMembers(code)` - Load league members
- `banMember(userId)` - Ban member (superadmin only)
- `unbanMember(userId)` - Unban member (superadmin only)
- `createInvitation(data?)` - Create invitation link (optionally with virtual player alias)
- `getSuggestedPlayers(code)` - Get suggested players for game setup
- `acceptInvitation(token)` - Accept invitation
- `leaveLeague(code)` - Leave league

---

## Frontend Routes

- `/ui/leagues` - List all leagues
- `/ui/leagues/:code` - League details (standings, members, invitation)
- `/ui/leagues/join/:token` - Accept invitation page

---

## Security Considerations

1. **Authentication Required:** All endpoints require valid JWT authentication
2. **League Membership:** Most endpoints verify user is an active member
3. **Superadmin Privileges:** Creating, updating, archiving leagues and banning members requires superadmin role
4. **Invitation Expiry:** Invitations expire after 7 days to prevent abuse
5. **One-Time Use:** Invitations can only be used once
6. **Ban Enforcement:** Banned members cannot participate in games within the league

---

## UX Improvements

### Priority 1: Critical UX Issues

#### 1.1 User-Facing Error Messages

**Current Issue:**
```typescript
catch (error) {
  console.error('Error creating league:', error);  // User doesn't see this!
}
```

**Recommendation:**
Add toast notifications or inline error alerts for failed actions.

**Implementation:**
```typescript
// Using Vuetify's snackbar or v-alert
const errorMessage = ref<string | null>(null);

try {
  await leagueStore.createLeague(newLeagueName.value);
} catch (error) {
  errorMessage.value = error instanceof Error
    ? error.message
    : 'Failed to create league';
  console.error('Error creating league:', error);
}
```

**Affected Components:**
- `LeagueList.vue` - createLeague action
- `LeagueDetails.vue` - archiveLeague, unarchiveLeague, banMember, unbanMember actions
- All components with API calls

#### 1.2 Loading States for Actions

**Current Issue:**
Archive/unarchive and ban/unban actions have no loading indicators.

**Note:** Banned users are displayed at the end of the members list with an unban button.

**Recommendation:**
Add loading states to action buttons.

**Implementation in LeagueDetails.vue:**
```typescript
const archiving = ref(false);

const archiveLeague = async () => {
  archiving.value = true;
  try {
    await leagueStore.archiveLeague(currentLeague.value.code);
    showSuccess('League archived successfully');
  } catch (error) {
    showError('Error archiving league');
  } finally {
    archiving.value = false;
  }
};
```

**Affected Components:**
- `LeagueDetails.vue` - archive, unarchive, ban actions
- `LeagueInvitation.vue` - already has loading state ✓

#### 1.3 Better Confirmation Dialogs

**Current Issue:**
Native `confirm()` dialogs are not styled and don't match the app's design.

**Recommendation:**
Use Vuetify dialogs for confirmations.

**Implementation:**
```vue
<template>
  <!-- Confirmation Dialog -->
  <v-dialog v-model="showConfirmDialog" max-width="400">
    <v-card>
      <v-card-title>{{ confirmTitle }}</v-card-title>
      <v-card-text>{{ confirmMessage }}</v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="showConfirmDialog = false">Cancel</v-btn>
        <v-btn
          color="error"
          :loading="actionInProgress"
          @click="confirmAction"
        >
          {{ confirmButtonText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
const showConfirmDialog = ref(false);
const confirmTitle = ref('');
const confirmMessage = ref('');
const confirmAction = ref<(() => void) | null>(null);

const requestBan = (member: LeagueMember) => {
  confirmTitle.value = 'Ban user?';
  confirmMessage.value = `Are you sure you want to ban ${member.user_name}? The user will not be able to participate in games in this league.`;
  confirmButtonText.value = 'Ban';
  confirmAction.value = () => banMember(member);
  showConfirmDialog.value = true;
};
</script>
```

**Affected Components:**
- `LeagueDetails.vue` - ban/unban actions, archive/unarchive actions
- `LeagueList.vue` - leave league action (if implemented)

---

### Priority 2: Enhanced Feedback

#### 2.1 Toast Notifications

**Recommendation:**
Add a global toast/snackbar component for success/error messages.

**Implementation:**

Create `composables/useToast.ts`:
```typescript
import { ref } from 'vue';

const message = ref('');
const type = ref<'success' | 'error' | 'info' | 'warning'>('info');
const show = ref(false);

export function useToast() {
  const showSuccess = (msg: string) => {
    message.value = msg;
    type.value = 'success';
    show.value = true;
  };

  const showError = (msg: string) => {
    message.value = msg;
    type.value = 'error';
    show.value = true;
  };

  const showInfo = (msg: string) => {
    message.value = msg;
    type.value = 'info';
    show.value = true;
  };

  return {
    message,
    type,
    show,
    showSuccess,
    showError,
    showInfo,
  };
}
```

Add to `App.vue`:
```vue
<template>
  <v-app>
    <!-- ... app content ... -->

    <v-snackbar
      v-model="toast.show"
      :color="toast.type"
      :timeout="3000"
      location="top"
    >
      {{ toast.message }}
      <template v-slot:actions>
        <v-btn icon="mdi-close" @click="toast.show = false" />
      </template>
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { useToast } from '@/composables/useToast';
const toast = useToast();
</script>
```

**Usage in Components:**
```typescript
import { useToast } from '@/composables/useToast';

const { showSuccess, showError } = useToast();

try {
  await leagueStore.createLeague(name);
  showSuccess('League created successfully!');
} catch (error) {
  showError('Failed to create league');
}
```

#### 2.2 Optimistic UI Updates

**Recommendation:**
Update UI immediately for better perceived performance, then rollback if API fails.

**Example - Ban Member:**
```typescript
const banMember = async (member: LeagueMember) => {
  // Optimistically update UI
  member.status = 'banned';

  try {
    await leagueStore.banUser(currentLeague.value.code, member.user_id);
    showSuccess(`${member.user_name} banned`);
  } catch (error) {
    // Rollback on error
    member.status = 'active';
    showError('Failed to ban user');
  }
};
```

---

### Priority 3: Accessibility & Polish

#### 3.1 Keyboard Navigation

**Recommendation:**
Ensure all interactive elements are keyboard accessible.

**Checklist:**
- [ ] All buttons focusable
- [ ] Tab order logical
- [ ] Enter key works on focused buttons
- [ ] Escape closes dialogs
- [ ] ARIA labels on icon buttons

**Implementation:**
```vue
<v-btn
  icon="mdi-dots-vertical"
  aria-label="League management"
  @click="showManageMenu = !showManageMenu"
/>
```

#### 3.2 Loading Skeletons

**Recommendation:**
Replace spinners with skeleton loaders for better UX.

**Implementation:**
```vue
<v-skeleton-loader
  v-if="loading"
  type="list-item-avatar-three-line@3"
/>
```

**Affected Components:**
- `LeagueList.vue` - league cards
- `LeagueDetails.vue` - standings, members
- `LeagueStandings.vue` - table rows

#### 3.3 Transition Animations

**Recommendation:**
Add smooth transitions for better visual feedback.

**Implementation:**
```vue
<transition-group name="list" tag="div">
  <league-card
    v-for="league in activeLeagues"
    :key="league.code"
    :league="league"
  />
</transition-group>

<style scoped>
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from {
  opacity: 0;
  transform: translateY(-30px);
}

.list-leave-to {
  opacity: 0;
  transform: translateY(30px);
}
</style>
```

#### 3.4 Empty State Improvements

**Current:**
Simple text message.

**Recommendation:**
Add illustrations and call-to-action.

**Implementation:**
```vue
<v-card-text v-if="activeLeagues.length === 0" class="text-center py-8">
  <v-icon size="80" color="grey-lighten-1" class="mb-4">
    mdi-trophy-outline
  </v-icon>
  <div class="text-h6 mb-2">No active leagues</div>
  <div class="text-body-2 text-medium-emphasis mb-4">
    Create a new league or join through invitation
  </div>
  <div class="d-flex justify-center gap-2">
    <v-btn
      v-if="canCreateLeague"
      color="primary"
      @click="showCreateDialog = true"
    >
      Create League
    </v-btn>
  </div>
</v-card-text>
```

---

### Priority 4: Advanced Features

#### 4.1 Search and Filtering

**Recommendation:**
Add search for leagues and members lists.

**Implementation:**
```vue
<v-text-field
  v-model="searchQuery"
  prepend-inner-icon="mdi-magnify"
  label="Search leagues"
  clearable
  hide-details
  class="mb-4"
/>

<script setup>
const searchQuery = ref('');

const filteredLeagues = computed(() => {
  if (!searchQuery.value) return activeLeagues.value;

  return activeLeagues.value.filter(league =>
    league.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});
</script>
```

#### 4.2 Pagination for Large Lists

**Recommendation:**
Add pagination when members/standings exceed 50 items.

**Implementation:**
```vue
<v-data-table
  :items="standings"
  :headers="headers"
  :items-per-page="25"
  :page.sync="page"
/>
```

#### 4.3 Real-Time Updates

**Recommendation:**
Add WebSocket or polling for real-time standings updates.

**Implementation (Polling):**
```typescript
import { useIntervalFn } from '@vueuse/core';

onMounted(() => {
  // Poll standings every 30 seconds
  useIntervalFn(() => {
    if (activeTab.value === 'standings') {
      leagueStore.refreshStandings(currentLeague.value.code);
    }
  }, 30000);
});
```

#### 4.4 Undo Actions

**Recommendation:**
Allow undo for critical actions like ban/leave.

**Implementation:**
```typescript
const showUndoBan = ref(false);
let undoTimeout: ReturnType<typeof setTimeout> | null = null;

const banMember = async (member: LeagueMember) => {
  try {
    await leagueStore.banUser(code, member.user_id);

    showUndoBan.value = true;
    undoTimeout = setTimeout(() => {
      showUndoBan.value = false;
    }, 5000);

    showSuccess('User banned');
  } catch (error) {
    showError('Error banning');
  }
};

const undoBan = async (member: LeagueMember) => {
  if (undoTimeout) clearTimeout(undoTimeout);
  showUndoBan.value = false;

  try {
    await leagueStore.unbanUser(code, member.user_id);
    showSuccess('Ban cancelled');
  } catch (error) {
    showError('Error cancelling');
  }
};
```

---

## Form Validation Improvements

### LeagueList.vue - Create Dialog

**Current:**
Basic required validation.

**Recommendation:**
Add comprehensive validation.

**Implementation:**
```typescript
const nameRules = [
  (v: string) => !!v || 'Name is required',
  (v: string) => v.length >= 3 || 'Minimum 3 characters',
  (v: string) => v.length <= 50 || 'Maximum 50 characters',
  (v: string) => /^[a-zA-Z0-9\s-]+$/.test(v) || 'Invalid characters',
];

const descriptionRules = [
  (v: string) => !v || v.length <= 200 || 'Maximum 200 characters',
];
```

---

## Error Message Guidelines

### User-Friendly Error Messages

**Bad:**
```
Error: MongoError: connection refused
```

**Good:**
```
Failed to connect to server. Check your connection and try again.
```

### Error Message Mapping

Create a utility for mapping API errors to user-friendly messages:

```typescript
// utils/errorMessages.ts
export function getUserFriendlyError(error: Error): string {
  const errorMap: Record<string, string> = {
    'Network Error': 'Network error. Check your connection.',
    'unauthorized': 'Authentication required',
    'forbidden': 'Insufficient permissions',
    'not found': 'Resource not found',
    'already exists': 'League already exists',
    'invalid token': 'Invalid invitation',
    'token expired': 'Invitation expired',
  };

  const message = error.message.toLowerCase();

  for (const [key, value] of Object.entries(errorMap)) {
    if (message.includes(key)) {
      return value;
    }
  }

  return 'An error occurred. Please try again.';
}
```

**Usage:**
```typescript
import { getUserFriendlyError } from '@/utils/errorMessages';

catch (error) {
  const message = error instanceof Error
    ? getUserFriendlyError(error)
    : 'Unknown error';
  showError(message);
}
```

---

## Implementation Checklist

### Phase 1: Critical Fixes
- [ ] Add toast notification system
- [ ] Add error messages to all action handlers
- [ ] Add loading states to action buttons
- [ ] Replace native confirm() with Vuetify dialogs

### Phase 2: Enhanced Feedback
- [ ] Add success messages for all actions
- [ ] Implement optimistic UI updates
- [ ] Add user-friendly error message mapping

### Phase 3: Polish
- [ ] Add skeleton loaders
- [ ] Add transition animations
- [ ] Improve empty states
- [ ] Add ARIA labels for accessibility

### Phase 4: Advanced (Optional)
- [ ] Add search and filtering
- [ ] Add pagination
- [ ] Implement real-time updates
- [ ] Add undo functionality

---

## Testing After Implementation

For each improved component, test:

1. **Error Scenarios:**
   - Network failure
   - Invalid input
   - Unauthorized access
   - Server errors

2. **Loading States:**
   - Verify spinner appears
   - Verify buttons disable during loading
   - Verify smooth transitions

3. **Success Feedback:**
   - Verify success messages appear
   - Verify UI updates correctly
   - Verify redirects work

4. **Accessibility:**
   - Tab through all interactive elements
   - Test with screen reader
   - Test keyboard shortcuts

---

## Conclusion

These improvements will significantly enhance the user experience of the League System by:

1. **Providing clear feedback** for all user actions
2. **Reducing frustration** with better error messages
3. **Improving perceived performance** with loading states
4. **Ensuring accessibility** for all users
5. **Adding polish** with animations and transitions

Implement in phases based on priority, testing thoroughly after each phase.

