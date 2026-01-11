# API Reference

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

## Rate Limits

- Geolocation API (ipapi.co): Check ipapi.co documentation for rate limits
- No built-in rate limiting on application endpoints (can be added via middleware)

## CORS

- Configure trusted origins via `TRUSTED_ORIGINS` environment variable
- Format: Comma-separated list of URLs (e.g., `http://localhost:3000,https://example.com`)
- Used by `RequestService.IsTrustedOrigin()` for validation
