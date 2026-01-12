# API Reference

*[Українська версія](API_REFERENCE.md)*

## Authentication Endpoints

### POST /api/auth/google/callback

Google OAuth callback endpoint. Creates a new session and returns authentication tokens.

**Request:**
- Method: POST
- Query Parameters: OAuth callback parameters from Google
- Body: None

**Response:**
- Status: 200 OK
- Body:
```json
{
  "code": "user_code",
  "external_ids": ["email@example.com"],
  "name": "User Name",
  "avatar": "https://...",
  "alias": "user_alias",
  "names": ["User Name"],
  "avatars": ["https://..."],
  "rotateToken": "base64_encoded_token"
}
```

**Cookies Set:**
- `auth_token`: JWT action token (HTTP-only, Secure, SameSite=Strict)

**Notes:**
- `rotateToken` should be stored in localStorage by the client
- `auth_token` cookie is set automatically

---

### POST /api/auth/refresh

Refreshes the action token. Optionally rotates the rotate token if 12 hours have passed since last rotation.

**Request:**
- Method: POST
- Headers:
  - `Authorization: Bearer <rotateToken>`
- Body: None

**Response:**
- Status: 200 OK
- Body:
```json
{
  "rotateToken": "new_rotate_token"  // Only present if token was rotated
}
```

**Cookies Set:**
- `auth_token`: New JWT action token

**Error Responses:**
- 401 Unauthorized: Invalid or expired rotate token

**Notes:**
- If `rotateToken` is returned, client must update localStorage
- Action token expires after 1 hour
- Rotate token rotates every 12 hours

---

### POST /api/auth/logout

Invalidates the current session and clears authentication cookies.

**Request:**
- Method: POST
- Headers (either):
  - `Authorization: Bearer <rotateToken>`, OR
  - `Content-Type: application/json`
- Body (optional):
```json
{
  "rotateToken": "token_to_invalidate"
}
```

**Response:**
- Status: 200 OK
- Body: None

**Cookies Cleared:**
- `auth_token`: Action token cookie cleared

**Notes:**
- Client should remove `rotateToken` from localStorage after logout

---

## User Endpoints

All user endpoints require authentication (action token cookie).

### GET /api/user

Gets the current user's profile information.

**Request:**
- Method: GET
- Headers: None (uses cookie)
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
{
  "code": "user_code",
  "external_ids": ["email@example.com"],
  "name": "User Name",
  "avatar": "https://...",
  "alias": "user_alias",
  "names": ["User Name"],
  "avatars": ["https://..."]
}
```

**Error Responses:**
- 401 Unauthorized: Invalid or expired token

---

### GET /api/user/sessions

Gets all active sessions for the current user.

**Request:**
- Method: GET
- Query Parameters:
  - `current` (optional): Rotate token to mark as current session
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
[
  {
    "rotate_token": "token_string",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z",
    "last_rotation_at": "2024-01-01T12:00:00Z",
    "expires_at": "2024-01-31T00:00:00Z",
    "is_current": true,
    "geo_info": {
      "country": "United States",
      "country_code": "US",
      "region": "NY",
      "region_name": "New York",
      "city": "New York",
      "timezone": "America/New_York",
      "isp": "ISP Name",
      "ip": "192.168.1.1"
    }
  }
]
```

**Error Responses:**
- 401 Unauthorized: Invalid or expired token

**Notes:**
- `geo_info` may be null if geolocation lookup failed
- `is_current` is true if `current` query param matches the session's rotate token

---

### POST /api/user/alias/exist

Checks if an alias is available.

**Request:**
- Method: POST
- Query Parameters:
  - `alias`: Alias to check
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
{
  "isUnique": true
}
```

---

### PUT /api/user/update

Updates the current user's profile.

**Request:**
- Method: PUT
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "name": "New Name",
  "avatar": "https://...",
  "alias": "new_alias"
}
```

**Response:**
- Status: 200 OK
- Body: None

**Error Responses:**
- 400 Bad Request: Invalid request body
- 401 Unauthorized: Invalid or expired token
- 409 Conflict: Alias already taken

---

## Admin Endpoints

All admin endpoints require authentication and super admin privileges.

### GET /api/admin/diagnostics

Gets system diagnostic information (admin only).

**Request:**
- Method: GET
- Cookies: `auth_token` (required)
- Headers: None

**Response:**
- Status: 200 OK
- Body:
```json
{
  "server_info": {
    "host_url": "http://localhost:8080",
    "trusted_origins": ["http://localhost:3000"]
  },
  "request_info": {
    "ip_address": "192.168.1.1",
    "base_url": "http://localhost:8080",
    "user_agent": "Mozilla/5.0...",
    "origin": "http://localhost:3000",
    "is_trusted": true,
    "geo_info": {
      "country": "United States",
      "country_code": "US",
      "region": "NY",
      "region_name": "New York",
      "city": "New York",
      "timezone": "America/New_York",
      "isp": "ISP Name",
      "ip": "192.168.1.1"
    }
  }
}
```

**Error Responses:**
- 401 Unauthorized: Not authenticated
- 403 Forbidden: Not a super admin

**Notes:**
- `geo_info` may be null if geolocation lookup failed
- `trusted_origins` comes from `TRUSTED_ORIGINS` environment variable

---

### PUT /api/admin/user/create

Creates a new user (admin only).

**Request:**
- Method: PUT
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "external_ids": ["email@example.com"]
}
```

**Response:**
- Status: 201 Created
- Body: "User created successfully"

**Error Responses:**
- 400 Bad Request: Missing or invalid external_ids
- 401 Unauthorized: Not authenticated
- 403 Forbidden: Not a super admin
- 409 Conflict: User already exists

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
Invalid request format or missing required parameters.

### 401 Unauthorized
Authentication required or token invalid/expired.

### 403 Forbidden
Authenticated but lacks required permissions (admin endpoints).

### 404 Not Found
Resource not found.

### 409 Conflict
Resource conflict (e.g., alias already taken).

### 500 Internal Server Error
Server error. Check server logs for details.

## Authentication

Most endpoints require authentication via the `auth_token` cookie (JWT action token).

**Token Refresh Flow:**
1. Client receives 401 Unauthorized
2. Client calls `/api/auth/refresh` with `rotateToken` from localStorage
3. Server returns new `actionToken` cookie
4. Client retries original request

**Token Rotation:**
- Rotate tokens automatically rotate every 12 hours
- When rotated, new `rotateToken` is returned in refresh response
- Client must update localStorage with new token

## Game Type Endpoints

All game type endpoints require authentication (action token cookie).

### GET /api/game_types

Gets all game types.

**Request:**
- Method: GET
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
[
  {
    "code": "abc123",
    "name": "Catan",
    "scoring_type": "classic",
    "version": 1,
    "labels": [
      { "name": "First Player", "color": "#FF0000", "icon": "mdi-flag" }
    ],
    "teams": []
  }
]
```

---

### POST /api/game_types

Creates a new game type.

**Request:**
- Method: POST
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "name": "Catan",
  "scoring_type": "classic",
  "labels": [
    { "name": "First Player", "color": "#FF0000", "icon": "mdi-flag" }
  ],
  "teams": []
}
```

**Response:**
- Status: 201 Created
- Body:
```json
{
  "code": "abc123",
  "name": "Catan",
  "scoring_type": "classic",
  "version": 1,
  "labels": [...],
  "teams": []
}
```

**Error Responses:**
- 400 Bad Request: Invalid request body or duplicate label/team names
- 401 Unauthorized: Invalid or expired token

**Notes:**
- `scoring_type` must be one of: `classic`, `mafia`, `custom`, `cooperative`, `cooperative_with_moderator`, `team_vs_team`
- Label and team names must be unique within their respective arrays

---

### GET /api/game_types/{code}

Gets a specific game type by code.

**Request:**
- Method: GET
- URL Parameters: `code` - Game type code
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body: Same as single item in list response

**Error Responses:**
- 400 Bad Request: Invalid code format
- 404 Not Found: Game type not found

---

### PUT /api/game_types/{code}

Updates an existing game type.

**Request:**
- Method: PUT
- URL Parameters: `code` - Game type code
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body: Same as POST request body

**Response:**
- Status: 200 OK
- Body: Updated game type object

**Error Responses:**
- 400 Bad Request: Invalid request body or duplicate label/team names
- 404 Not Found: Game type not found

---

### DELETE /api/game_types/{code}

Deletes a game type.

**Request:**
- Method: DELETE
- URL Parameters: `code` - Game type code
- Cookies: `auth_token` (required)

**Response:**
- Status: 204 No Content

**Error Responses:**
- 400 Bad Request: Invalid code format
- 404 Not Found: Game type not found

---

## Game Round Endpoints

All game round endpoints require authentication (action token cookie).

### GET /api/game_rounds

Gets all game rounds.

**Request:**
- Method: GET
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
[
  {
    "code": "xyz789",
    "name": "Friday Night Game",
    "game_type_id": "...",
    "start_time": "2024-01-01T19:00:00Z",
    "end_time": "2024-01-01T22:00:00Z",
    "players": [
      {
        "player_id": "...",
        "position": 1,
        "score": 10,
        "is_moderator": false,
        "team_name": ""
      }
    ],
    "team_scores": [],
    "cooperative_score": 0
  }
]
```

---

### POST /api/game_rounds

Starts a new game round.

**Request:**
- Method: POST
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "name": "Friday Night Game",
  "type": "Catan",
  "start_time": "2024-01-01T19:00:00Z",
  "players": [
    {
      "user_id": "player_object_id",
      "position": 1,
      "is_moderator": false,
      "team_name": ""
    }
  ]
}
```

**Response:**
- Status: 201 Created
- Body: Created game round object

**Error Responses:**
- 400 Bad Request: Invalid request body, game type not found, or invalid team assignments
- 401 Unauthorized: Invalid or expired token
- 500 Internal Server Error: Error creating round

**Notes:**
- `type` refers to the game type name (not code)
- For team games, each team defined in the game type must have at least one player assigned

---

### GET /api/game_rounds/{code}

Gets a specific game round by code.

**Request:**
- Method: GET
- URL Parameters: `code` - Game round code
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body: Game round object

**Error Responses:**
- 400 Bad Request: Invalid code format
- 404 Not Found: Game round not found

---

### PUT /api/game_rounds/{code}

Updates an existing game round.

**Request:**
- Method: PUT
- URL Parameters: `code` - Game round code
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "name": "Updated Game Name",
  "players": [
    {
      "user_id": "player_code",
      "position": 1,
      "score": 5,
      "is_moderator": false,
      "team_name": ""
    }
  ]
}
```

**Response:**
- Status: 200 OK
- Body: Updated game round object

**Error Responses:**
- 400 Bad Request: Invalid request body or user ID
- 404 Not Found: Game round not found

---

### PUT /api/game_rounds/{code}/finalize

Finalizes a game round with final scores and positions.

**Request:**
- Method: PUT
- URL Parameters: `code` - Game round code
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "player_scores": {
    "player_code_1": 10,
    "player_code_2": 8,
    "player_code_3": 7
  },
  "team_scores": {
    "Team A": 15,
    "Team B": 12
  },
  "cooperative_score": 0
}
```

**Response:**
- Status: 200 OK

**Error Responses:**
- 400 Bad Request: Invalid request body
- 404 Not Found: Game round not found

**Notes:**
- `team_scores` and `cooperative_score` are optional
- Player positions are automatically calculated based on scores (highest score = position 1)
- Sets the `end_time` to current time

---

### PUT /api/game_rounds/{code}/players/{userId}/score

Updates a specific player's score within a game round.

**Request:**
- Method: PUT
- URL Parameters:
  - `code` - Game round code
  - `userId` - Player user ID
- Headers:
  - `Content-Type: application/json`
- Cookies: `auth_token` (required)
- Body:
```json
{
  "score": 5
}
```

**Response:**
- Status: 200 OK

**Error Responses:**
- 400 Bad Request: Invalid game ID or user ID
- 404 Not Found: Game round not found or player not in game

---

## Player Endpoints

All player endpoints require authentication (action token cookie).

### GET /api/players

Gets all registered players.

**Request:**
- Method: GET
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
[
  {
    "code": "abc123",
    "alias": "player_alias",
    "avatar": "https://..."
  }
]
```

---

### GET /api/players/{code}

Gets a specific player by code.

**Request:**
- Method: GET
- URL Parameters: `code` - Player code
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
{
  "code": "abc123",
  "alias": "player_alias",
  "avatar": "https://..."
}
```

**Error Responses:**
- 400 Bad Request: Invalid player code
- 404 Not Found: Player not found

---

### GET /api/players/i_am

Gets the current authenticated player's information.

**Request:**
- Method: GET
- Cookies: `auth_token` (required)

**Response:**
- Status: 200 OK
- Body:
```json
{
  "code": "abc123",
  "alias": "my_alias",
  "avatar": "https://..."
}
```

**Error Responses:**
- 404 Not Found: Player profile not found
- 500 Internal Server Error: Error fetching user

---

## Rate Limits

- Geolocation API (ipinfo.io): Check ipinfo.io documentation for rate limits
- No built-in rate limiting on application endpoints (can be added via middleware)

## CORS

- Configure trusted origins via `TRUSTED_ORIGINS` environment variable
- Format: Comma-separated list of URLs (e.g., `http://localhost:3000,https://example.com`)
- Used by `RequestService.IsTrustedOrigin()` for validation
