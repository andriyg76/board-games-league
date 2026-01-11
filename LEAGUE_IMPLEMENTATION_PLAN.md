# League Implementation Plan

## Вимоги

### Концепція Ліги
**Ліга** - це ізольований простір для гри, де:
- Гравці є членами конкретної ліги
- Всі game rounds створюються в межах ліги
- При створенні гри можна вибирати тільки гравців з поточної ліги
- Очки та статистика гравців підраховуються окремо для кожної ліги
- Більшість інтерфейсів працюють в контексті обраної ліги
- Game types залишаються глобальними (не прив'язані до ліги)

### Бізнес-правила

#### 1. Створення та управління лігою
- **Створення ліги**: Тільки суперадмін може створити нову лігу
- **Поля ліги**: Назва (обов'язкове поле)
- **Адміністрування**: Немає окремих адміністраторів/власників ліги - всі гравці рівні

#### 2. Членство в лізі
- **Множинне членство**: Гравець може бути членом багатьох ліг одночасно
- **Приєднання до ліги**: Через систему запрошень
  - Суперадмін може генерувати запрошення
  - Будь-який гравець ліги може генерувати запрошення
  - Запрошення = одноразовий лінк з токеном
- **Процес запрошення**:
  1. Користувач переходить по лінку запрошення
  2. Якщо не залогінений → редирект на login → створення/вхід в акаунт
  3. Після успішної автентифікації → автоматичне додавання в лігу
  4. Лінк запрошення стає недійсним після використання (одноразовий)
- **Вихід з ліги**: Немає функції виходу з ліги
- **Бан**: Тільки суперадмін може забанити гравця в конкретній лізі

#### 3. Game Rounds в контексті ліги
- Кожен game round прив'язаний до конкретної ліги
- При створенні раунду доступні тільки гравці з цієї ліги
- Список раундів фільтрується по обраній лізі

#### 4. Рейтингова система
**Загальний рейтинг гравця в лізі** складається з:
- **Бали за участь**: Фіксовані бали за кожну зіграну гру
- **Бали за позицію**: Залежно від місця в грі (1-ше місце = більше балів)
- **Бали за модерацію**: Додаткові бали, якщо гравець був модератором

**Формула** (приклад, може бути налаштована):
```
Рейтинг гравця = Σ (Бали_за_участь + Бали_за_позицію + Бали_за_модерацію)
```

Де для кожного game round:
- `Бали_за_участь` = константа (наприклад, 1 бал)
- `Бали_за_позицію` = функція від позиції (наприклад: 1-ше=10, 2-ге=7, 3-тє=5, 4-те=3, 5-те=1)
- `Бали_за_модерацію` = константа якщо гравець модератор (наприклад, 2 бали)

**Примітка**: Конкретні значення балів мають бути налаштовуваними (можливо, на рівні ліги або глобально)

#### 5. UI/UX
- **Селектор ліги**: В інтерфейсі є можливість вибору поточної ліги
- **Контекст**: Більшість операцій відбуваються в контексті обраної ліги
- **Game Types**: Залишаються глобальними і доступні для всіх ліг

---

## Архітектура

### Backend Components

#### 1. Models (backend/models/)

##### League Model (`league.go`)
```go
type League struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Version     int64              `bson:"version"`
    Name        string             `bson:"name"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}
```

##### LeagueMembership Model (`league_membership.go`)
```go
type LeagueMembershipStatus string

const (
    MembershipActive  LeagueMembershipStatus = "active"
    MembershipBanned  LeagueMembershipStatus = "banned"
)

type LeagueMembership struct {
    ID        primitive.ObjectID     `bson:"_id,omitempty"`
    Version   int64                  `bson:"version"`
    LeagueID  primitive.ObjectID     `bson:"league_id"`
    UserID    primitive.ObjectID     `bson:"user_id"`
    Status    LeagueMembershipStatus `bson:"status"`
    JoinedAt  time.Time              `bson:"joined_at"`
    CreatedAt time.Time              `bson:"created_at"`
    UpdatedAt time.Time              `bson:"updated_at"`
}
```

##### LeagueInvitation Model (`league_invitation.go`)
```go
type LeagueInvitation struct {
    ID           primitive.ObjectID `bson:"_id,omitempty"`
    Version      int64              `bson:"version"`
    LeagueID     primitive.ObjectID `bson:"league_id"`
    CreatedBy    primitive.ObjectID `bson:"created_by"` // User who created invitation
    Token        string             `bson:"token"`      // Unique token for invitation link
    IsUsed       bool               `bson:"is_used"`
    UsedBy       primitive.ObjectID `bson:"used_by,omitempty"`
    UsedAt       time.Time          `bson:"used_at,omitempty"`
    ExpiresAt    time.Time          `bson:"expires_at,omitempty"` // Optional expiration
    CreatedAt    time.Time          `bson:"created_at"`
    UpdatedAt    time.Time          `bson:"updated_at"`
}
```

##### LeagueStanding (computed, не зберігається в БД)
```go
type LeagueStanding struct {
    UserID             primitive.ObjectID `json:"user_id"`
    UserName           string             `json:"user_name"`
    UserAvatar         string             `json:"user_avatar"`
    TotalPoints        int64              `json:"total_points"`
    GamesPlayed        int                `json:"games_played"`
    GamesModerated     int                `json:"games_moderated"`
    FirstPlaceCount    int                `json:"first_place_count"`
    SecondPlaceCount   int                `json:"second_place_count"`
    ThirdPlaceCount    int                `json:"third_place_count"`
    ParticipationPoints int64             `json:"participation_points"`
    PositionPoints     int64              `json:"position_points"`
    ModerationPoints   int64              `json:"moderation_points"`
}
```

##### Оновлення GameRound Model
```go
type GameRound struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    Code             string             `bson:"-"`
    Version          int64              `bson:"version"`
    LeagueID         primitive.ObjectID `bson:"league_id"`        // NEW FIELD
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

#### 2. Repositories (backend/repositories/)

##### LeagueRepository (`league_repository.go`)
```go
type LeagueRepository interface {
    Create(ctx context.Context, league *League) error
    FindByID(ctx context.Context, id primitive.ObjectID) (*League, error)
    FindAll(ctx context.Context) ([]*League, error)
    Update(ctx context.Context, league *League) error
}
```

##### LeagueMembershipRepository (`league_membership_repository.go`)
```go
type LeagueMembershipRepository interface {
    Create(ctx context.Context, membership *LeagueMembership) error
    FindByID(ctx context.Context, id primitive.ObjectID) (*LeagueMembership, error)
    FindByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*LeagueMembership, error)
    FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueMembership, error)
    FindByUser(ctx context.Context, userID primitive.ObjectID) ([]*LeagueMembership, error)
    Update(ctx context.Context, membership *LeagueMembership) error
    // Check if user is active member of league
    IsActiveMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
}
```

##### LeagueInvitationRepository (`league_invitation_repository.go`)
```go
type LeagueInvitationRepository interface {
    Create(ctx context.Context, invitation *LeagueInvitation) error
    FindByToken(ctx context.Context, token string) (*LeagueInvitation, error)
    MarkAsUsed(ctx context.Context, id primitive.ObjectID, usedBy primitive.ObjectID) error
}
```

##### Оновлення GameRoundRepository
```go
// Add method to find rounds by league
FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*GameRound, error)
```

#### 3. Services (backend/services/)

##### LeagueService (`league_service.go`)
```go
type LeagueService interface {
    // Create league (superadmin only)
    CreateLeague(ctx context.Context, name string) (*League, error)

    // Get league info
    GetLeague(ctx context.Context, leagueID primitive.ObjectID) (*League, error)
    ListLeagues(ctx context.Context) ([]*League, error)

    // Membership management
    GetLeagueMembers(ctx context.Context, leagueID primitive.ObjectID) ([]*User, error)
    IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
    BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error

    // Invitations
    CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID) (*LeagueInvitation, error)
    AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*League, error)

    // Standings
    GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueStanding, error)
}
```

##### StandingsCalculator (`standings_calculator.go`)
```go
// Configuration for points calculation
type PointsConfig struct {
    ParticipationPoints int64
    ModerationPoints    int64
    PositionPoints      map[int]int64 // position -> points
}

// Calculate standings for a league
func CalculateStandings(
    ctx context.Context,
    rounds []*GameRound,
    members []*LeagueMembership,
    users []*User,
    config PointsConfig,
) ([]*LeagueStanding, error)
```

#### 4. API Handlers (backend/gameapi/)

##### League Handlers (`league.go`)
```go
// POST /api/leagues - Create league (superadmin only)
func (h *Handler) createLeague(w http.ResponseWriter, r *http.Request)

// GET /api/leagues - List all leagues
func (h *Handler) listLeagues(w http.ResponseWriter, r *http.Request)

// GET /api/leagues/:code - Get league details
func (h *Handler) getLeague(w http.ResponseWriter, r *http.Request)

// GET /api/leagues/:code/members - Get league members
func (h *Handler) getLeagueMembers(w http.ResponseWriter, r *http.Request)

// GET /api/leagues/:code/standings - Get league standings
func (h *Handler) getLeagueStandings(w http.ResponseWriter, r *http.Request)

// POST /api/leagues/:code/invitations - Create invitation (member or superadmin)
func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request)

// POST /api/leagues/join/:token - Accept invitation
func (h *Handler) acceptInvitation(w http.ResponseWriter, r *http.Request)

// POST /api/leagues/:code/ban/:userCode - Ban user (superadmin only)
func (h *Handler) banUserFromLeague(w http.ResponseWriter, r *http.Request)
```

##### Оновлення Game Round Handlers
```go
// Update startGame to require leagueID
// Update listGameRounds to filter by leagueID
// Update getGameRound to verify league access
```

#### 5. Middleware

##### League Access Middleware (`league_middleware.go`)
```go
// Verify user has access to league
func RequireLeagueMembership(leagueRepo LeagueRepository, membershipRepo LeagueMembershipRepository) func(http.Handler) http.Handler

// Verify user is superadmin
func RequireSuperAdmin(userService UserService) func(http.Handler) http.Handler
```

---

### Frontend Components

#### 1. TypeScript Types (frontend/src/gametypes/types.ts)

```typescript
export interface League {
  code: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface LeagueMembership {
  code: string;
  league_id: string;
  user_id: string;
  status: 'active' | 'banned';
  joined_at: string;
}

export interface LeagueInvitation {
  code: string;
  league_id: string;
  token: string;
  is_used: boolean;
  expires_at?: string;
  created_at: string;
}

export interface LeagueStanding {
  user_id: string;
  user_name: string;
  user_avatar: string;
  total_points: number;
  games_played: number;
  games_moderated: number;
  first_place_count: number;
  second_place_count: number;
  third_place_count: number;
  participation_points: number;
  position_points: number;
  moderation_points: number;
}
```

#### 2. API Client (frontend/src/api/LeagueApi.ts)

```typescript
class LeagueApi {
  // Leagues
  static async createLeague(name: string): Promise<League>
  static async listLeagues(): Promise<League[]>
  static async getLeague(code: string): Promise<League>

  // Members
  static async getLeagueMembers(leagueCode: string): Promise<User[]>

  // Standings
  static async getLeagueStandings(leagueCode: string): Promise<LeagueStanding[]>

  // Invitations
  static async createInvitation(leagueCode: string): Promise<LeagueInvitation>
  static async acceptInvitation(token: string): Promise<League>

  // Admin
  static async banUser(leagueCode: string, userCode: string): Promise<void>
}
```

#### 3. Vuex Store (frontend/src/store/league.ts)

```typescript
interface LeagueState {
  leagues: League[];
  currentLeague: League | null;
  currentLeagueMembers: User[];
  currentLeagueStandings: LeagueStanding[];
  loading: boolean;
}

// Actions:
// - fetchLeagues
// - selectLeague
// - fetchLeagueMembers
// - fetchLeagueStandings
// - createLeague
// - createInvitation
// - acceptInvitation
// - banUser
```

#### 4. Vue Components

##### ListLeagues.vue
- Список всіх ліг
- Кнопка створення нової ліги (для суперадмінів)
- Кнопка приєднання до ліги
- Вибір поточної ліги

##### LeagueDetails.vue
- Інформація про лігу
- Список членів
- Таблиця standings (рейтинг)
- Кнопка генерації запрошення
- Функції бану (для суперадмінів)

##### LeagueSelector.vue
- Dropdown для вибору поточної ліги
- Показується в header/navbar
- Зберігає вибір в localStorage або Vuex

##### InvitationAccept.vue
- Сторінка для прийняття запрошення
- Обробка токену з URL
- Редирект на login якщо потрібно
- Показ успіху/помилки

##### CreateLeagueDialog.vue
- Діалог створення нової ліги
- Поле для введення назви
- Тільки для суперадмінів

##### LeagueStandingsTable.vue
- Таблиця з рейтингом гравців
- Сортування по різним колонкам
- Показ деталей (участь, позиції, модерація)

#### 5. Router Updates (frontend/src/router/index.ts)

```typescript
{
  path: '/ui/leagues',
  name: 'Leagues',
  component: () => import('@/gametypes/ListLeagues.vue')
},
{
  path: '/ui/leagues/:code',
  name: 'LeagueDetails',
  component: () => import('@/gametypes/LeagueDetails.vue'),
  props: true
},
{
  path: '/ui/leagues/join/:token',
  name: 'AcceptInvitation',
  component: () => import('@/gametypes/InvitationAccept.vue'),
  props: true
},
```

#### 6. Updates to Existing Components

##### GameroundsList.vue
- Додати фільтр по обраній лізі
- Показувати тільки раунди з поточної ліги

##### GameroundEdit.vue
- При створенні game round: вибір/передача leagueID
- При виборі гравців: показувати тільки членів поточної ліги
- Валідація що всі гравці належать лізі

##### App.vue / Navigation
- Додати LeagueSelector в header
- Додати лінк на сторінку ліг

##### HomeView.vue
- Фільтрувати статистику по обраній лізі
- Показувати топ гравців поточної ліги

---

## Database Schema

### Collections

#### leagues
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "name": String,
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Indexes:**
- `{ "name": 1 }` - unique

#### league_memberships
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "league_id": ObjectId,
  "user_id": ObjectId,
  "status": String, // "active" | "banned"
  "joined_at": ISODate,
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
- `{ "expires_at": 1 }` - TTL index для автоматичного видалення застарілих запрошень

#### game_rounds (update)
Додати поле:
```json
{
  "league_id": ObjectId,
  // ... existing fields
}
```

**New Index:**
- `{ "league_id": 1, "start_time": -1 }`

---

## Implementation Plan

### Phase 1: Backend Foundation (Models & Repositories)

#### Step 1.1: Create Models
- [ ] Create `backend/models/league.go`
- [ ] Create `backend/models/league_membership.go`
- [ ] Create `backend/models/league_invitation.go`
- [ ] Update `backend/models/game_round.go` - add LeagueID field

#### Step 1.2: Create Repositories
- [ ] Create `backend/repositories/league_repository.go`
- [ ] Create `backend/repositories/league_membership_repository.go`
- [ ] Create `backend/repositories/league_invitation_repository.go`
- [ ] Update `backend/repositories/game_round_repository.go` - add FindByLeague method

#### Step 1.3: Database Migrations
- [ ] Add script/code to create indexes for new collections
- [ ] Add migration to add league_id to existing game_rounds (optional, може бути NULL спочатку)

#### Step 1.4: Unit Tests
- [ ] Write tests for league repository
- [ ] Write tests for membership repository
- [ ] Write tests for invitation repository

### Phase 2: Backend Services & Business Logic

#### Step 2.1: Standings Calculator
- [ ] Create `backend/services/standings_calculator.go`
- [ ] Implement PointsConfig structure
- [ ] Implement CalculateStandings function
- [ ] Write unit tests for standings calculation with different scenarios

#### Step 2.2: League Service
- [ ] Create `backend/services/league_service.go`
- [ ] Implement CreateLeague
- [ ] Implement GetLeague, ListLeagues
- [ ] Implement membership methods (GetLeagueMembers, IsUserMember)
- [ ] Implement BanUserFromLeague
- [ ] Implement invitation methods (CreateInvitation, AcceptInvitation)
- [ ] Implement GetLeagueStandings
- [ ] Write unit tests for league service

#### Step 2.3: Authorization Logic
- [ ] Create `backend/auth/roles.go` - define superadmin role
- [ ] Update User model if needed to support roles
- [ ] Implement role checking functions

### Phase 3: Backend API Handlers

#### Step 3.1: League Handlers
- [ ] Create `backend/gameapi/league.go`
- [ ] Implement createLeague (POST /api/leagues)
- [ ] Implement listLeagues (GET /api/leagues)
- [ ] Implement getLeague (GET /api/leagues/:code)
- [ ] Implement getLeagueMembers (GET /api/leagues/:code/members)
- [ ] Implement getLeagueStandings (GET /api/leagues/:code/standings)
- [ ] Implement createInvitation (POST /api/leagues/:code/invitations)
- [ ] Implement acceptInvitation (POST /api/leagues/join/:token)
- [ ] Implement banUserFromLeague (POST /api/leagues/:code/ban/:userCode)

#### Step 3.2: Update Game Round Handlers
- [ ] Update `backend/gameapi/gameround.go::startGame` - require leagueID
- [ ] Update `backend/gameapi/gameround.go::listGameRounds` - filter by leagueID (optional query param)
- [ ] Add validation to check user membership when creating game round

#### Step 3.3: Middleware
- [ ] Create `backend/middleware/league_middleware.go`
- [ ] Implement RequireLeagueMembership middleware
- [ ] Implement RequireSuperAdmin middleware

#### Step 3.4: Route Registration
- [ ] Update `backend/main.go` or route configuration
- [ ] Register league routes with appropriate middleware
- [ ] Update game round routes to use league middleware where needed

#### Step 3.5: API Tests
- [ ] Write integration tests for league endpoints
- [ ] Write tests for invitation flow
- [ ] Write tests for standings calculation
- [ ] Write tests for access control (membership, superadmin)

### Phase 4: Frontend Foundation

#### Step 4.1: TypeScript Types
- [ ] Update `frontend/src/gametypes/types.ts`
- [ ] Add League, LeagueMembership, LeagueInvitation, LeagueStanding interfaces
- [ ] Update GameRoundView to include league_id

#### Step 4.2: API Client
- [ ] Create `frontend/src/api/LeagueApi.ts`
- [ ] Implement all API methods (createLeague, listLeagues, etc.)
- [ ] Add error handling

#### Step 4.3: Vuex Store
- [ ] Create `frontend/src/store/league.ts`
- [ ] Define state interface
- [ ] Implement mutations
- [ ] Implement actions (fetchLeagues, selectLeague, etc.)
- [ ] Implement getters
- [ ] Register store module in main store

### Phase 5: Frontend Components

#### Step 5.1: League Selector
- [ ] Create `frontend/src/components/LeagueSelector.vue`
- [ ] Implement dropdown with league list
- [ ] Connect to Vuex store
- [ ] Store current league in localStorage
- [ ] Add to App.vue navigation

#### Step 5.2: League List
- [ ] Update `frontend/src/gametypes/ListLeagues.vue`
- [ ] Display list of all leagues
- [ ] Add "Create League" button (for superadmins)
- [ ] Add navigation to league details
- [ ] Add loading states and error handling

#### Step 5.3: League Details
- [ ] Create `frontend/src/gametypes/LeagueDetails.vue`
- [ ] Display league info
- [ ] Display members list
- [ ] Display standings table
- [ ] Add "Generate Invitation" button
- [ ] Add "Ban User" functionality (for superadmins)
- [ ] Add copy invitation link functionality

#### Step 5.4: Create League Dialog
- [ ] Create `frontend/src/components/CreateLeagueDialog.vue`
- [ ] Form with name input
- [ ] Validation
- [ ] Submit to API
- [ ] Show success/error messages

#### Step 5.5: Invitation Accept
- [ ] Create `frontend/src/gametypes/InvitationAccept.vue`
- [ ] Extract token from route params
- [ ] Check authentication status
- [ ] Redirect to login if needed with return URL
- [ ] Accept invitation after login
- [ ] Show success message and redirect to league

#### Step 5.6: Standings Table
- [ ] Create `frontend/src/components/LeagueStandingsTable.vue`
- [ ] Display standings in table format
- [ ] Add sorting by columns
- [ ] Show detailed breakdown (participation, position, moderation points)
- [ ] Highlight current user

### Phase 6: Integration & Updates

#### Step 6.1: Update Game Rounds UI
- [ ] Update `frontend/src/gametypes/GameroundsList.vue`
  - Add league filter
  - Show league name for each round
  - Filter by current league from LeagueSelector
- [ ] Update `frontend/src/gametypes/GameroundEdit.vue`
  - Add league selection (or use current league)
  - Filter player list by league membership
  - Add validation

#### Step 6.2: Update Home View
- [ ] Update `frontend/src/views/HomeView.vue`
  - Add league context to statistics
  - Show current league name
  - Filter data by selected league

#### Step 6.3: Update Navigation
- [ ] Update `frontend/src/App.vue`
  - Add LeagueSelector to header/navbar
  - Add link to Leagues page
  - Update mobile navigation

#### Step 6.4: Router Updates
- [ ] Update `frontend/src/router/index.ts`
- [ ] Add routes for league details, invitation accept
- [ ] Add navigation guards if needed

### Phase 7: Testing & Polish

#### Step 7.1: End-to-End Testing
- [ ] Test complete invitation flow (create → share → accept)
- [ ] Test league creation (superadmin)
- [ ] Test game round creation with league context
- [ ] Test standings calculation with real data
- [ ] Test ban functionality
- [ ] Test league switching in UI

#### Step 7.2: Error Handling
- [ ] Add proper error messages for all failure scenarios
- [ ] Handle expired invitations
- [ ] Handle banned users trying to access league
- [ ] Handle non-member trying to access league resources

#### Step 7.3: UI/UX Polish
- [ ] Add loading states to all components
- [ ] Add empty states (no leagues, no members, etc.)
- [ ] Add confirmation dialogs for destructive actions
- [ ] Improve responsive design
- [ ] Add internationalization (i18n) for new components

#### Step 7.4: Performance Optimization
- [ ] Add pagination for large standings tables
- [ ] Add caching for frequently accessed data
- [ ] Optimize standings calculation for large number of rounds
- [ ] Add database indexes verification

#### Step 7.5: Documentation
- [ ] Add API documentation (Swagger/OpenAPI)
- [ ] Update README with league features
- [ ] Add user guide for league management
- [ ] Document points calculation formula
- [ ] Add developer documentation for extending standings calculation

### Phase 8: Deployment

#### Step 8.1: Migration Plan
- [ ] Create database migration script for production
- [ ] Plan for handling existing game_rounds without league_id
- [ ] Create default league for existing data (if needed)
- [ ] Backup plan and rollback strategy

#### Step 8.2: Configuration
- [ ] Add environment variables for points configuration
- [ ] Add feature flags if doing gradual rollout
- [ ] Configure superadmin users

#### Step 8.3: Deploy
- [ ] Deploy backend changes
- [ ] Run database migrations
- [ ] Deploy frontend changes
- [ ] Smoke test in production
- [ ] Monitor for errors

---

## Points Calculation Details

### Default Points Configuration

```go
DefaultPointsConfig = PointsConfig{
    ParticipationPoints: 1,
    ModerationPoints: 2,
    PositionPoints: map[int]int64{
        1: 10,
        2: 7,
        3: 5,
        4: 3,
        5: 1,
        // For positions > 5, could use formula: max(0, 11 - position)
    },
}
```

### Calculation Algorithm

Для кожного гравця в лізі:
1. Знайти всі завершені game rounds в лізі де гравець брав участь
2. Для кожного раунду:
   - `participation_points += ParticipationPoints`
   - `position_points += PositionPoints[player.Position]`
   - `moderation_points += ModerationPoints` (якщо is_moderator)
3. `total_points = participation_points + position_points + moderation_points`
4. Також зберегти статистику: games_played, games_moderated, first_place_count, etc.
5. Сортувати гравців по total_points (desc)

### Edge Cases

- **Незавершені раунди**: Не враховуються в standings
- **Гравець без позиції**: Якщо position == 0, бали за позицію = 0
- **Cooperative games**: Можливо потрібна окрема логіка (всі отримують однакові бали?)
- **Team games**: Позиція команди чи індивідуальна позиція?

**TODO**: Уточнити як рахувати бали для різних scoring types (cooperative, team_vs_team)

---

## Security Considerations

1. **Superadmin Verification**:
   - Необхідно мати надійний спосіб перевірки superadmin ролі
   - Зберігати список superadmin user IDs в конфігурації або БД

2. **Invitation Tokens**:
   - Використовувати криптографічно стійкий генератор (crypto/rand)
   - Достатня довжина токену (мінімум 32 байти)
   - Токени повинні бути унікальними

3. **Authorization Checks**:
   - Перевіряти членство в лізі перед доступом до league resources
   - Перевіряти що banned користувачі не можуть виконувати дії в лізі
   - Перевіряти права при створенні запрошень

4. **Rate Limiting**:
   - Обмежити кількість запрошень що може створити один користувач
   - Захист від спаму при прийнятті запрошень

5. **Input Validation**:
   - Валідація назви ліги (довжина, допустимі символи)
   - Валідація всіх ObjectID перед використанням

---

## Open Questions / Decisions Needed

1. **Points Configuration**:
   - Де зберігати конфігурацію балів? (Environment variables, database, hardcoded?)
   - Чи потрібна можливість налаштування балів per-league?
   - Як рахувати бали для cooperative та team-based games?

2. **Default League**:
   - Що робити з існуючими game_rounds без league_id?
   - Створити "Default League" для старих даних?
   - Чи дозволяти створювати game rounds без ліги?

3. **Superadmin Management**:
   - Як визначити хто є superadmin? (Список в config? Таблиця roles в БД?)
   - Чи потрібен UI для управління superadmin ролями?

4. **Invitation Expiration**:
   - Чи повинні запрошення мати термін дії?
   - Який термін за замовчуванням? (24 години? 7 днів? Безстроково?)

5. **League Archival**:
   - Чи потрібна можливість "архівувати" лігу (зробити read-only)?
   - Чи можна видалити лігу? (Що робити з game_rounds?)

6. **Multiple Invitations**:
   - Чи може бути кілька активних запрошень для однієї ліги?
   - Чи потрібно показувати список активних запрошень?
   - Чи можна "відкликати" запрошення?

7. **Notification System**:
   - Чи потрібні нотифікації при прийнятті запрошення?
   - Чи потрібні нотифікації при бані?

8. **Leaderboard Features**:
   - Чи потрібна історія standings (зміни рейтингу в часі)?
   - Чи потрібні досягнення/badges?
   - Чи потрібна статистика per game type в межах ліги?

---

## Future Enhancements (Out of Scope for v1)

- **Seasons**: Розділення ліги на сезони з окремими standings
- **Tournaments**: Окремий режим для турнірів з bracket системою
- **Achievements**: Система досягнень та badges
- **Statistics**: Детальна аналітика гравців (win rate, favorite games, etc.)
- **Social Features**: Chat, comments on rounds
- **Advanced Permissions**: Різні ролі в лізі (moderator, member, viewer)
- **League Settings**: Налаштування балів per-league
- **Export**: Експорт standings в CSV/PDF
- **Public/Private Leagues**: Контроль видимості ліги
- **Federation**: Об'єднання кількох ліг для спільних турнірів

---

## Success Criteria

Імплементація вважається успішною коли:

1. ✅ Superadmin може створювати ліги
2. ✅ Гравець ліги може генерувати запрошення
3. ✅ Новий користувач може приєднатись по запрошенню (з автоматичним створенням акаунту)
4. ✅ Існуючий користувач може приєднатись по запрошенню
5. ✅ При створенні game round доступні тільки гравці з ліги
6. ✅ Game rounds відображаються в контексті ліги
7. ✅ Standings відображають правильний рейтинг (участь + позиція + модерація)
8. ✅ Superadmin може забанити гравця в лізі
9. ✅ Banned гравець не може виконувати дії в лізі
10. ✅ UI дозволяє легко переключатись між лігами
11. ✅ Всі тести проходять успішно
12. ✅ Документація оновлена
