# League System API Documentation

## Overview

The League System provides isolated competitive contexts where players can compete, earn points, and track rankings. Each league has its own members, standings, and game rounds.

## Authentication

All league endpoints require authentication via JWT token in the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### 1. List All Leagues

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

### 2. Get League Details

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

### 3. Create League

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

### 4. Update League

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

### 5. Archive/Unarchive League

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

### 6. Get League Standings

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

### 7. Get League Members

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
- `banned` - Member is banned from the league

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

### 8. Ban/Unban Member

**Endpoint:** `PUT /api/leagues/{code}/members/{userId}/status`

**Description:** Changes member status (ban/unban). **Requires superadmin privileges.**

**URL Parameters:**
- `code` - League code
- `userId` - User ID (hex string)

**Request Body:**
```json
{
  "status": "banned"
}
```

**Valid status values:** `active`, `banned`

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid status value
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a superadmin
- `404 Not Found` - League or member not found
- `500 Internal Server Error` - Server error

---

### 9. Get Suggested Players

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

### 10. Create Invitation

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

### 11. Preview Invitation (Public)

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

### 12. Accept Invitation

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

### 13. Membership Statuses

League members can have the following statuses:

| Status | Description |
|--------|-------------|
| `active` | Regular member with full access |
| `pending` | Created via invitation, waiting for user to accept |
| `virtual` | Player participated in games but never logged in (pending with games after invitation cancelled) |
| `banned` | User banned from the league |

**Virtual status flow:**
- When an invitation is cancelled and the pending member has participated in games, they become `virtual`
- Virtual members can be re-invited using the same alias to preserve their game history
- Virtual members appear in standings if they have games

---

### 14. Leave League

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

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

Common error scenarios:
- **401 Unauthorized:** Missing or invalid JWT token
- **403 Forbidden:** Insufficient permissions (not a member, not a superadmin)
- **404 Not Found:** Resource not found (league, member, invitation)
- **500 Internal Server Error:** Unexpected server error

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

## Routes

### Frontend Routes

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

## Architecture

### Data Models

#### LeagueMembership Model

The `LeagueMembership` model includes fields for tracking player activity and recent co-players:

```go
type RecentCoPlayer struct {
    MembershipID primitive.ObjectID `bson:"membership_id"`
    LastPlayedAt time.Time          `bson:"last_played_at"`
}

type LeagueMembership struct {
    // ... existing fields ...
    RecentCoPlayers []RecentCoPlayer `bson:"recent_co_players,omitempty"` // max 10 items
    LastActivityAt  time.Time        `bson:"last_activity_at,omitempty"`  // last activity timestamp
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

## Database Collections

### leagues
- `_id` - ObjectID
- `name` - String
- `description` - String
- `status` - Enum: "active", "archived"
- `created_at` - DateTime
- `created_by` - ObjectID (user reference)

### league_memberships
- `_id` - ObjectID
- `league_id` - ObjectID (league reference)
- `user_id` - ObjectID (user reference, optional for pending/virtual)
- `status` - Enum: "active", "pending", "virtual", "banned"
- `joined_at` - DateTime
- `recent_co_players` - Array of RecentCoPlayer (max 10 items)
  - `membership_id` - ObjectID (reference to co-player's membership)
  - `last_played_at` - DateTime (timestamp of last game together)
- `last_activity_at` - DateTime (last activity timestamp, updated on game finalization and invitation creation)

### league_invitations
- `_id` - ObjectID
- `token` - String (unique, indexed)
- `league_id` - ObjectID (league reference)
- `created_by` - ObjectID (user reference)
- `created_at` - DateTime
- `expires_at` - DateTime
- `used` - Boolean
- `used_by` - ObjectID (user reference, optional)
- `used_at` - DateTime (optional)

### game_rounds (extended)
- `league_id` - ObjectID (optional, league reference)
- When set, game results contribute to league standings

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

## Testing

See [LEAGUE_TEST_PLAN.md](./LEAGUE_TEST_PLAN.md) for comprehensive testing procedures.
