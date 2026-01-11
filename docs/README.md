# Documentation Index

## Session and Account Management

- **[SESSION_MANAGEMENT.md](SESSION_MANAGEMENT.md)** - Comprehensive guide to the session management system, including:
  - Architecture overview
  - Token system (rotate tokens and action tokens)
  - Component documentation
  - Configuration
  - Usage examples
  - Security features
  - Troubleshooting

- **[API_REFERENCE.md](API_REFERENCE.md)** - Complete API documentation:
  - Authentication endpoints
  - User endpoints
  - Admin endpoints
  - Request/response formats
  - Error codes
  - Authentication flow

## Key Features

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

## Getting Started

1. Configure environment variables (see [SESSION_MANAGEMENT.md](SESSION_MANAGEMENT.md#configuration))
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
