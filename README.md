# Board Games League

A platform for board game players to organize games, manage leagues, and track competitive rankings. Built with Vue.js, Go, and MongoDB, featuring Material Design UI.

## Features

### 🎮 Game Management
- Create and manage board game rounds
- Track player scores and positions
- Support for multiple game types with custom scoring rules
- Team-based games support

### 🏆 League System
- **Isolated Competitive Contexts** - Create separate leagues for different groups
- **Points & Rankings** - Automatic calculation of standings based on game results
- **Invitation System** - Secure, time-limited invitation links
- **Member Management** - Ban/unban players, track participation
- **Statistics** - Detailed player stats including podium finishes and moderation points

**Points System:**
- **Participation:** 2 points per game
- **Positions:** 1st (10 pts), 2nd (6 pts), 3rd (3 pts), 4th+ (1 pt)
- **Moderation:** 1 point per game moderated

### 👥 User Management
- Session-based authentication with JWT
- User profiles with avatars
- Superadmin roles for league creation

### 🌍 Internationalization
- English, Ukrainian, Estonian translations
- Date/time localization

## Architecture

**Backend:**
- Go (Golang) with Chi router
- MongoDB for data persistence
- Repository pattern for data access
- Service layer for business logic

**Frontend:**
- Vue 3 (Composition API)
- Vuetify 3 (Material Design)
- Pinia for state management
- TypeScript for type safety

## Documentation

### English
- [Session Management](docs/SESSION_MANAGEMENT.en.md) - Session and account management guide
- [League System](docs/LEAGUES.en.md) - Complete League System documentation (API, architecture, UX)
- [Wizard Game Scoring](docs/WIZARD_GAME_SCORING.en.md) - Wizard card game scoring system documentation
- [API Reference](docs/API_REFERENCE.en.md) - General API documentation
- [Documentation Index](docs/README.en.md) - Complete documentation index

### Українська
- [Керування сесіями](docs/SESSION_MANAGEMENT.md) - Посібник керування сесіями
- [Система ліг](docs/LEAGUES.md) - Повна документація системи ліг (API, архітектура, UX)
- [Система підрахунку очок Wizard](docs/WIZARD_GAME_SCORING.md) - Документація системи підрахунку очок для гри Wizard
- [Довідник API](docs/API_REFERENCE.md) - Загальна документація API
- [Індекс документації](docs/README.md) - Повний індекс документації

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- MongoDB 6.0+

### Backend Setup
```bash
cd backend
go mod download
go run .
```

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Environment Variables
See `.env.example` for required environment variables.

### Electron Wrapper (Mobile UI)
The Electron shell is intended to always show the mobile UI.

Example environment config:
```
VITE_ELECTRON=true
VITE_API_BASE_URL=https://api.example.com
VITE_PUBLIC_WEB_BASE_URL=https://app.example.com

# Electron runtime options
ELECTRON_PROTOCOL=bgl
ELECTRON_START_URL=http://localhost:5173
ELECTRON_INITIAL_ROUTE=/m?ui=mobile
```

Deep link example:
```
bgl://m/accept-invite/INVITE_TOKEN?ui=mobile
```

## League System Usage

### Creating a League (Superadmin Only)
1. Navigate to `/ui/leagues`
2. Click "Create League"
3. Enter league name and description
4. Creator is automatically added as first member

### Inviting Players
1. Open league details
2. Go to "Invitation" tab
3. Click "Create Invitation"
4. Share the generated link (valid for 7 days)

### Playing Games in a League
1. Start a new game
2. Select the league from dropdown
3. Add players (only active members can be selected)
4. Play and finalize the game
5. Standings automatically update

### Viewing Standings
- Navigate to league details
- Click "Standings" tab
- View rankings, points breakdown, and player statistics

## Project Status

**Current Version:** Phase 7 (Testing & Documentation)

**Completed:**
- ✅ Backend models and repositories
- ✅ League service with points calculation
- ✅ API handlers for all league operations
- ✅ Frontend state management
- ✅ Vue components for all league features
- ✅ Router integration
- ✅ Comprehensive documentation

**In Progress:**
- 🔄 End-to-end testing
- 🔄 UX improvements and polish

**Planned:**
- 📋 Production deployment
- 📋 Database migrations
- 📋 Performance optimization

## Contributing

Contributions are welcome! Please ensure:
1. Code follows existing patterns
2. Tests are included for new features
3. Documentation is updated
4. Ukrainian translations are provided

## License

[Add license information]

## Contact

[Add contact information]