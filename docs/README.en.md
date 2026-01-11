# Documentation Index

*[Українська версія](README.md)*

## Session and Account Management

- **[SESSION_MANAGEMENT.en.md](SESSION_MANAGEMENT.en.md)** - Comprehensive guide to the session management system, including:
  - Architecture overview
  - Token system (rotate tokens and action tokens)
  - Component documentation
  - Configuration
  - Usage examples
  - Security features
  - Troubleshooting

- **[API_REFERENCE.en.md](API_REFERENCE.en.md)** - Complete API documentation:
  - Authentication endpoints
  - User endpoints
  - Admin endpoints
  - Game type endpoints
  - Game round endpoints
  - Player endpoints
  - Request/response formats
  - Error codes
  - Authentication flow

## Game Management

- **[GAME_MANAGEMENT.en.md](GAME_MANAGEMENT.en.md)** - Complete guide to game management features:
  - Game types configuration
  - Scoring types (classic, cooperative, team-based, mafia)
  - Labels and teams
  - Game rounds lifecycle
  - Player management
  - Frontend components
  - API examples
  - Database collections

## Key Features

### Game Management
- Multiple scoring types (classic, cooperative, team vs team, mafia)
- Game type configuration with labels and teams
- Game round tracking with start/end times
- Player score and position tracking
- Team scores for team-based games
- Cooperative scoring support

### Session Management
- Hybrid token system (rotate tokens + action tokens)
- Automatic token rotation every 12 hours
- Session tracking with IP address and user agent
- Geolocation integration
- Multiple concurrent sessions per user

### Security
- HTTP-only cookies for action tokens
- Secure, SameSite cookie settings
- Token expiration and rotation
- Session expiration (30 days)
- Optimistic locking for concurrency

### Diagnostics
- Admin-only diagnostics endpoint
- Request information (IP, origin, user agent)
- Geolocation data
- Server configuration display
- Trusted origins validation

### User Features
- Session information display
- Active sessions list
- Current session indicator
- Geographic location display
- Session activity tracking
- Player profile management

## Getting Started

1. Configure environment variables (see [SESSION_MANAGEMENT.en.md](SESSION_MANAGEMENT.en.md#configuration))
2. Set up MongoDB connection
3. Configure Google OAuth credentials
4. Set super admin external IDs
5. (Optional) Configure trusted origins

## Architecture Overview

```
┌─────────────┐
│   Client    │
│  (Browser)  │
└──────┬──────┘
       │
       │ HTTP Requests
       │ (actionToken cookie)
       │
┌──────▼──────────┐
│   Backend API   │
│  (Go/Chi Router)│
└──────┬──────────┘
       │
       ├──────────┐
       │          │
┌──────▼──────┐  │  ┌──────────────┐
│   MongoDB   │  │  │  GeoIP API   │
│  (Sessions) │  │  │  (ipapi.co)  │
└─────────────┘  │  └──────────────┘
                 │
       ┌─────────┘
       │
┌──────▼──────┐
│   Services  │
│  - Session  │
│  - Request  │
│  - GeoIP    │
└─────────────┘
```

## Related Documentation

- Backend code: `backend/`
- Frontend code: `frontend/`
- Configuration: Environment variables
- Database: MongoDB collections and indexes
