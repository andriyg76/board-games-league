# Session and Account Management

*[Українська версія](SESSION_MANAGEMENT.md)*

## Overview

The board-games-league application uses a hybrid token-based authentication system with session tracking, geolocation, and diagnostic capabilities. This system provides secure, scalable session management with rotate tokens and short-lived action tokens.

## Architecture

### Token System

The system uses a two-token approach:

1. **Rotate Token** (Long-lived)
   - Stored in browser `localStorage` (client-side)
   - Persisted in MongoDB (server-side)
   - Lifetime: 30 days
   - Rotation interval: Every 12 hours (when action token is refreshed)
   - Purpose: Used only for refreshing action tokens
   - Not sent with every request (stored in localStorage, not cookies)

2. **Action Token** (Short-lived)
   - Stored in HTTP-only, Secure, SameSite cookie
   - Format: JWT (JSON Web Token)
   - Lifetime: 1 hour
   - Purpose: Used for authenticating API requests
   - Automatically refreshed using rotate token when expired

### Token Flow

```
┌─────────┐                ┌──────────┐                ┌─────────┐
│ Client  │                │  Server  │                │   DB    │
└────┬────┘                └────┬─────┘                └────┬────┘
     │                          │                           │
     │  1. Login Request        │                           │
     │─────────────────────────>│                           │
     │                          │                           │
     │                          │  2. Create Session        │
     │                          │──────────────────────────>│
     │                          │                           │
     │  3. Return rotateToken   │                           │
     │     + Set actionToken    │                           │
     │<─────────────────────────│                           │
     │                          │                           │
     │  4. Store rotateToken    │                           │
     │     in localStorage      │                           │
     │                          │                           │
     │  5. API Request          │                           │
     │     (with actionToken)   │                           │
     │─────────────────────────>│                           │
     │                          │  6. Validate JWT          │
     │  7. Response             │                           │
     │<─────────────────────────│                           │
     │                          │                           │
     │  8. Action Token Expired │                           │
     │     (401 error)          │                           │
     │<─────────────────────────│                           │
     │                          │                           │
     │  9. Refresh Request      │                           │
     │     (Authorization:      │                           │
     │      Bearer rotateToken) │                           │
     │─────────────────────────>│                           │
     │                          │  10. Validate rotateToken │
     │                          │──────────────────────────>│
     │                          │                           │
     │                          │  11. Rotate if needed     │
     │                          │──────────────────────────>│
     │  12. New actionToken     │                           │
     │      + New rotateToken   │                           │
     │      (if rotated)        │                           │
     │<─────────────────────────│                           │
```

## Components

### Backend Components

#### Models

**Session Model** (`backend/models/session.go`)
- Stores session information in MongoDB
- Fields:
  - `RotateToken`: Unique token identifier
  - `UserID`: Reference to user
  - `CreatedAt`, `UpdatedAt`, `LastRotationAt`: Timestamps
  - `ExpiresAt`: Session expiration (30 days)
  - `IPAddress`, `UserAgent`: Request tracking
  - `Version`: Optimistic locking for concurrent updates

**User Model** (`backend/models/user.go`)
- Extended with `LastActivity` field to track user activity

**GeoIPInfo Model** (`backend/models/geoip.go`)
- Stores geolocation information
- Fields: Country, Region, City, Timezone, ISP, etc.

#### Services

**SessionService** (`backend/services/session_service.go`)
- `CreateSession()`: Creates new session with rotate and action tokens
- `RefreshActionToken()`: Refreshes action token, rotates rotate token if needed (12h interval)
- `InvalidateSession()`: Deletes session (logout)
- `CleanupExpiredSessions()`: Removes expired sessions (background task)

**RequestService** (`backend/services/request_service.go`)
- `GetClientIP()`: Extracts client IP from headers (X-Forwarded-For, X-Real-IP, RemoteAddr)
- `BuildBaseURL()`: Constructs base URL from request protocol and host
- `IsTrustedOrigin()`: Validates request origin against trusted origins list

**GeoIPService** (`backend/services/geoip_service.go`)
- `GetGeoIPInfo()`: Resolves IP address to geographic information
- Uses ipinfo.io API (requires API token)
- Returns country, region, city, timezone, ISP information

#### Repositories

**SessionRepository** (`backend/repositories/session_repository.go`)
- MongoDB operations for sessions
- Indexes: `rotate_token` (unique), `user_id`, `expires_at`
- Methods: Create, FindByRotateToken, FindByUserID, Update, Delete, DeleteExpired

#### API Endpoints

**Authentication Endpoints**

- `POST /api/auth/google/callback`
  - Google OAuth callback
  - Creates session
  - Returns user data + `rotateToken` (for localStorage)
  - Sets `actionToken` cookie

- `POST /api/auth/refresh`
  - Refreshes action token
  - Request: `Authorization: Bearer <rotateToken>`
  - Response: New `actionToken` cookie + optional `rotateToken` (if rotated)
  - Returns JSON: `{rotateToken?: string}`

- `POST /api/auth/logout`
  - Invalidates session
  - Accepts `rotateToken` in request body or Authorization header
  - Clears action token cookie

**User Endpoints**

- `GET /api/user/sessions`
  - Returns all active sessions for current user
  - Query param: `?current=<rotateToken>` (optional, marks current session)
  - Response: Array of session info with geo information

**Admin Endpoints**

- `GET /api/admin/diagnostics`
  - Admin-only endpoint
  - Returns server info, request info, and geolocation data
  - Requires super admin privileges

### Frontend Components

#### API Clients

**UserApi** (`frontend/src/api/UserApi.ts`)
- `getUserSessions(currentRotateToken?)`: Fetches user sessions

**DiagnosticsApi** (`frontend/src/api/DiagnosticsApi.ts`)
- `getDiagnostics()`: Fetches diagnostic information (admin-only)

**Auth** (`frontend/src/api/Auth.ts`)
- Updated to store `rotateToken` in localStorage after login

#### Views

**UserView** (`frontend/src/views/UserView.vue`)
- Displays user profile
- Shows active sessions table with:
  - Location (city, country)
  - IP address
  - User agent
  - Created timestamp
  - Last activity
  - Current session indicator

**DiagnosticsView** (`frontend/src/views/DiagnosticsView.vue`)
- Admin-only page
- Displays:
  - Server information (host URL, trusted origins)
  - Request information (IP, base URL, origin, trusted status)
  - Geolocation information

## Configuration

### Environment Variables

**Required**
- `MONGODB_URI`: MongoDB connection string
- `GOOGLE_CLIENT_ID`: Google OAuth client ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth client secret
- `JWT_SECRET`: Secret key for signing JWT tokens
- `SESSION_SECRET`: Secret key for session cookies
- `SUPERADMINS`: Comma-separated list of external IDs (emails) for super admins
- `IPINFO_TOKEN`: API token for ipinfo.io (for geolocation)

**Optional**
- `TRUSTED_ORIGINS`: Comma-separated list of trusted origins for CORS validation
- `HOST_URL`: Base URL of the application (auto-detected if not set)
- `LOG_DIR`: Directory for `server.log`, `access.log`, `debug.log` (logs to console if unset)

### Database

The system creates a new MongoDB collection: `sessions`

Indexes created:
- `rotate_token` (unique)
- `user_id`
- `expires_at`

### Token Lifetimes

- **Action Token**: 1 hour (configurable in `user_profile.CreateAuthTokenWithExpiry`)
- **Rotate Token**: Rotates every 12 hours, expires after 30 days

## Usage

### User Flow

1. **Login**
   - User clicks login → redirects to Google OAuth
   - After authentication, callback creates session
   - `rotateToken` stored in localStorage
   - `actionToken` set as HTTP-only cookie

2. **Making Requests**
   - Client sends requests with `actionToken` cookie automatically
   - Middleware validates token
   - If expired, client receives 401

3. **Token Refresh**
   - Client detects 401 error
   - Calls `/api/auth/refresh` with `rotateToken` from localStorage
   - Receives new `actionToken` cookie
   - Receives new `rotateToken` if rotation occurred (update localStorage)

4. **Viewing Sessions**
   - User navigates to profile page
   - Sessions table shows all active sessions
   - Current session is highlighted

5. **Logout**
   - Client calls `/api/auth/logout` with `rotateToken`
   - Server invalidates session
   - Client removes `rotateToken` from localStorage
   - Cookie is cleared by server

### Admin Flow

1. **Access Diagnostics**
   - Admin navigates to `/ui/admin/diagnostics`
   - System displays diagnostic information
   - Non-admins receive 403 Forbidden

### Developer Flow

#### Adding New Protected Routes

```go
r.Group(func(r chi.Router) {
    r.Use(authHandler.Middleware)
    r.Get("/your-route", yourHandler)
})
```

#### Accessing User in Handler

```go
profile, ok := r.Context().Value("user").(*user_profile.UserProfile)
if !ok || profile == nil {
    // Handle unauthorized
}
```

#### Checking Admin Status

```go
if !auth.IsSuperAdmin(profile.ExternalIDs) {
    // Handle unauthorized
}
```

## Security Features

1. **HTTP-Only Cookies**: Action tokens stored in HTTP-only cookies prevent XSS attacks
2. **Secure Cookies**: Cookies marked as Secure (HTTPS only)
3. **SameSite Strict**: Prevents CSRF attacks
4. **Token Rotation**: Rotate tokens rotate every 12 hours
5. **Session Expiration**: Sessions expire after 30 days
6. **Optimistic Locking**: Prevents race conditions during token rotation
7. **IP Tracking**: Sessions track IP addresses for security monitoring
8. **Trusted Origins**: Optional origin validation for CORS

## Background Tasks

**Session Cleanup**
- Runs every 1 hour
- Removes expired sessions from database
- Logs cleanup activity

## Geolocation

The system uses ipinfo.io API for geolocation:
- Requires API token (`IPINFO_TOKEN`)
- Rate limits apply (check ipinfo.io documentation)
- Falls back gracefully if service unavailable
- Geo information is optional and non-blocking

## Troubleshooting

### Session Not Working

1. Check that `rotateToken` is stored in localStorage
2. Verify `actionToken` cookie is set (check browser DevTools)
3. Check server logs for authentication errors
4. Verify JWT_SECRET is set correctly

### Geolocation Not Showing

1. Check internet connectivity (service uses external API)
2. Verify `IPINFO_TOKEN` is configured
3. Verify ipinfo.io service is accessible
4. Check rate limits on ipinfo.io
4. Geo info is optional - sessions work without it

### Admin Access Denied

1. Verify user's external ID is in `SUPERADMINS` environment variable
2. Check that external IDs match exactly (case-sensitive)
3. User must re-authenticate after being added to SUPERADMINS

## Migration Notes

### From Old System

- Old JWT tokens will be invalid after deployment
- Users must re-authenticate
- No database migration needed (new collection)
- Existing user data is preserved

### Backward Compatibility

- Old authentication endpoints are replaced
- Frontend must be updated to use new token system
- Session information is new feature (backward compatible)

## API Examples

### Refresh Token

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Authorization: Bearer <rotateToken>" \
  -H "Cookie: auth_token=<old_action_token>"
```

Response:
```json
{
  "rotateToken": "new_rotate_token_if_rotated"
}
```

### Get User Sessions

```bash
curl http://localhost:8080/api/user/sessions?current=<rotateToken> \
  -H "Cookie: auth_token=<action_token>"
```

Response:
```json
[
  {
    "rotate_token": "...",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z",
    "last_rotation_at": "2024-01-01T12:00:00Z",
    "expires_at": "2024-01-31T00:00:00Z",
    "is_current": true,
    "geo_info": {
      "country": "United States",
      "city": "New York",
      "region": "NY",
      "timezone": "America/New_York"
    }
  }
]
```

### Diagnostics (Admin Only)

```bash
curl http://localhost:8080/api/admin/diagnostics \
  -H "Cookie: auth_token=<action_token>"
```

Response:
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
      "city": "New York"
    }
  }
}
```

## Future Enhancements

Potential improvements:
- Session revocation UI (revoke specific sessions)
- Session limits (max concurrent sessions per user)
- More granular session permissions
- Local geolocation database (reduce external API dependency)
- Session activity logs
- Multi-factor authentication support
