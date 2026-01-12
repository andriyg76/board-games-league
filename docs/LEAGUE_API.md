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

### 9. Create Invitation

**Endpoint:** `POST /api/leagues/{code}/invitations`

**Description:** Creates a one-time invitation link for the league. Valid for 7 days.

**URL Parameters:**
- `code` - League code

**Request Body:** None

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
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - User is not a member of this league
- `404 Not Found` - League not found
- `500 Internal Server Error` - Server error

---

### 10. Accept Invitation

**Endpoint:** `POST /api/leagues/invitations/{token}/accept`

**Description:** Accepts an invitation and adds the user to the league.

**URL Parameters:**
- `token` - Invitation token

**Request Body:** None

**Response:**
```json
{
  "league": {
    "id": "507f1f77bcf86cd799439011",
    "code": "ABC123",
    "name": "Summer Championship 2026",
    "description": "Competitive league for summer season",
    "status": "active"
  }
}
```

**Status Codes:**
- `200 OK` - Invitation accepted successfully
- `400 Bad Request` - User is already a member
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Invitation not found or expired
- `500 Internal Server Error` - Server error

**Notes:**
- Each invitation can only be used once
- Invitations expire after 7 days
- Users who are already members will receive a 400 error

---

### 11. Leave League

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
- `createInvitation()` - Create invitation link
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
- `user_id` - ObjectID (user reference)
- `status` - Enum: "active", "banned"
- `joined_at` - DateTime

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

## Testing

See [LEAGUE_TEST_PLAN.md](./LEAGUE_TEST_PLAN.md) for comprehensive testing procedures.
