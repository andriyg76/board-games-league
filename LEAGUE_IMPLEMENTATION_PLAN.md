# План Імплементації Ліг

## Вимоги

### Концепція Ліги
**Ліга** - це ізольований простір для гри, де:
- Гравці є членами конкретної ліги
- Всі ігрові кола створюються в межах ліги
- При створенні гри можна вибирати тільки гравців з поточної ліги
- Очки та статистика гравців підраховуються окремо для кожної ліги
- Більшість інтерфейсів працюють в контексті обраної ліги
- Типи ігор залишаються глобальними (не прив'язані до ліги)

### Бізнес-правила

#### 1. Створення та управління лігою
- **Створення ліги**: Тільки суперадмін може створити нову лігу
- **Поля ліги**: Назва (обов'язкове поле)
- **Адміністрування**: Немає окремих адміністраторів/власників ліги - всі гравці рівні
- **Скарги**: Гравці можуть поскаржитись на іншого гравця адміністратору

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

#### 3. Ігрові кола в контексті ліги
- Кожне ігрове коло (game round) прив'язане до конкретної ліги
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

Де для кожного ігрового кола:
- `Бали_за_участь` = константа (наприклад, 1 бал)
- `Бали_за_позицію` = функція від позиції (наприклад: 1-ше=10, 2-ге=7, 3-тє=5, 4-те=3, 5-те=1)
- `Бали_за_модерацію` = константа якщо гравець модератор (наприклад, 2 бали)

**Примітка**: Конкретні значення балів мають бути налаштовуваними (можливо, на рівні ліги або глобально)

#### 5. UI/UX
- **Селектор ліги**: В інтерфейсі є можливість вибору поточної ліги
- **Контекст**: Більшість операцій відбуваються в контексті обраної ліги
- **Типи ігор**: Залишаються глобальними і доступні для всіх ліг

---

## Архітектура

### Backend Компоненти

#### 1. Моделі (backend/models/)

##### Модель Ліги (`league.go`)
```go
type League struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Version     int64              `bson:"version"`
    Name        string             `bson:"name"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}
```

##### Модель Членства в Лізі (`league_membership.go`)
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

##### Модель Запрошення до Ліги (`league_invitation.go`)
```go
type LeagueInvitation struct {
    ID           primitive.ObjectID `bson:"_id,omitempty"`
    Version      int64              `bson:"version"`
    LeagueID     primitive.ObjectID `bson:"league_id"`
    CreatedBy    primitive.ObjectID `bson:"created_by"` // Користувач що створив запрошення
    Token        string             `bson:"token"`      // Унікальний токен для посилання
    IsUsed       bool               `bson:"is_used"`
    UsedBy       primitive.ObjectID `bson:"used_by,omitempty"`
    UsedAt       time.Time          `bson:"used_at,omitempty"`
    ExpiresAt    time.Time          `bson:"expires_at,omitempty"` // Опціонально: термін дії
    CreatedAt    time.Time          `bson:"created_at"`
    UpdatedAt    time.Time          `bson:"updated_at"`
}
```

##### Рейтинг Ліги (обчислюється, не зберігається в БД)
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

##### Оновлення моделі GameRound
```go
type GameRound struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    Code             string             `bson:"-"`
    Version          int64              `bson:"version"`
    LeagueID         primitive.ObjectID `bson:"league_id"`        // НОВЕ ПОЛЕ
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

#### 2. Репозиторії (backend/repositories/)

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
    // Перевірка чи є користувач активним членом ліги
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
// Додати метод для пошуку раундів по лізі
FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*GameRound, error)
```

#### 3. Сервіси (backend/services/)

##### LeagueService (`league_service.go`)
```go
type LeagueService interface {
    // Створення ліги (тільки суперадмін)
    CreateLeague(ctx context.Context, name string) (*League, error)

    // Отримання інформації про лігу
    GetLeague(ctx context.Context, leagueID primitive.ObjectID) (*League, error)
    ListLeagues(ctx context.Context) ([]*League, error)

    // Управління членством
    GetLeagueMembers(ctx context.Context, leagueID primitive.ObjectID) ([]*User, error)
    IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
    BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error

    // Запрошення
    CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID) (*LeagueInvitation, error)
    AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*League, error)

    // Рейтинг
    GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueStanding, error)
}
```

##### Калькулятор Рейтингу (`standings_calculator.go`)
```go
// Конфігурація для підрахунку балів
type PointsConfig struct {
    ParticipationPoints int64
    ModerationPoints    int64
    PositionPoints      map[int]int64 // позиція -> бали
}

// Розрахунок рейтингу для ліги
func CalculateStandings(
    ctx context.Context,
    rounds []*GameRound,
    members []*LeagueMembership,
    users []*User,
    config PointsConfig,
) ([]*LeagueStanding, error)
```

#### 4. API Обробники (backend/gameapi/)

##### Обробники Ліги (`league.go`)
```go
// POST /api/leagues - Створити лігу (тільки суперадмін)
func (h *Handler) createLeague(w http.ResponseWriter, r *http.Request)

// GET /api/leagues - Список всіх ліг
func (h *Handler) listLeagues(w http.ResponseWriter, r *http.Request)

// GET /api/leagues/:code - Отримати деталі ліги
func (h *Handler) getLeague(w http.ResponseWriter, r *http.Request)

// GET /api/leagues/:code/members - Отримати членів ліги
func (h *Handler) getLeagueMembers(w http.ResponseWriter, r *http.Request)

// GET /api/leagues/:code/standings - Отримати рейтинг ліги
func (h *Handler) getLeagueStandings(w http.ResponseWriter, r *http.Request)

// POST /api/leagues/:code/invitations - Створити запрошення (член або суперадмін)
func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request)

// POST /api/leagues/join/:token - Прийняти запрошення
func (h *Handler) acceptInvitation(w http.ResponseWriter, r *http.Request)

// POST /api/leagues/:code/ban/:userCode - Забанити користувача (тільки суперадмін)
func (h *Handler) banUserFromLeague(w http.ResponseWriter, r *http.Request)
```

##### Оновлення обробників Ігрових Кіл
```go
// Оновити startGame - додати обов'язкове поле leagueID
// Оновити listGameRounds - фільтрувати по leagueID
// Оновити getGameRound - перевіряти доступ до ліги
```

#### 5. Middleware

##### Middleware для доступу до ліги (`league_middleware.go`)
```go
// Перевірка що користувач має доступ до ліги
func RequireLeagueMembership(leagueRepo LeagueRepository, membershipRepo LeagueMembershipRepository) func(http.Handler) http.Handler

// Перевірка що користувач є суперадміном
func RequireSuperAdmin(userService UserService) func(http.Handler) http.Handler
```

---

### Frontend Компоненти

#### 1. TypeScript Типи (frontend/src/gametypes/types.ts)

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

#### 2. API Клієнт (frontend/src/api/LeagueApi.ts)

```typescript
class LeagueApi {
  // Ліги
  static async createLeague(name: string): Promise<League>
  static async listLeagues(): Promise<League[]>
  static async getLeague(code: string): Promise<League>

  // Члени
  static async getLeagueMembers(leagueCode: string): Promise<User[]>

  // Рейтинг
  static async getLeagueStandings(leagueCode: string): Promise<LeagueStanding[]>

  // Запрошення
  static async createInvitation(leagueCode: string): Promise<LeagueInvitation>
  static async acceptInvitation(token: string): Promise<League>

  // Адміністрування
  static async banUser(leagueCode: string, userCode: string): Promise<void>
}
```

#### 3. Vuex Сховище (frontend/src/store/league.ts)

```typescript
interface LeagueState {
  leagues: League[];
  currentLeague: League | null;
  currentLeagueMembers: User[];
  currentLeagueStandings: LeagueStanding[];
  loading: boolean;
}

// Дії:
// - fetchLeagues
// - selectLeague
// - fetchLeagueMembers
// - fetchLeagueStandings
// - createLeague
// - createInvitation
// - acceptInvitation
// - banUser
```

#### 4. Vue Компоненти

##### ListLeagues.vue
- Список всіх ліг
- Кнопка створення нової ліги (для суперадмінів)
- Кнопка приєднання до ліги
- Вибір поточної ліги

##### LeagueDetails.vue
- Інформація про лігу
- Список членів
- Таблиця рейтингу
- Кнопка генерації запрошення
- Функції бану (для суперадмінів)

##### LeagueSelector.vue
- Випадаючий список для вибору поточної ліги
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

#### 5. Оновлення Роутера (frontend/src/router/index.ts)

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

#### 6. Оновлення Існуючих Компонентів

##### GameroundsList.vue
- Додати фільтр по обраній лізі
- Показувати тільки раунди з поточної ліги

##### GameroundEdit.vue
- При створенні ігрового кола: вибір/передача leagueID
- При виборі гравців: показувати тільки членів поточної ліги
- Валідація що всі гравці належать лізі

##### App.vue / Навігація
- Додати LeagueSelector в header
- Додати лінк на сторінку ліг

##### HomeView.vue
- Фільтрувати статистику по обраній лізі
- Показувати топ гравців поточної ліги

---

## Схема Бази Даних

### Колекції

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

**Індекси:**
- `{ "name": 1 }` - унікальний

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

**Індекси:**
- `{ "league_id": 1, "user_id": 1 }` - унікальний складений індекс
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

**Індекси:**
- `{ "token": 1 }` - унікальний
- `{ "league_id": 1 }`
- `{ "is_used": 1 }`
- `{ "expires_at": 1 }` - TTL індекс для автоматичного видалення застарілих запрошень

#### game_rounds (оновлення)
Додати поле:
```json
{
  "league_id": ObjectId,
  // ... існуючі поля
}
```

**Новий Індекс:**
- `{ "league_id": 1, "start_time": -1 }`

---

## План Імплементації

### Фаза 1: Backend Основа (Моделі та Репозиторії)

#### Крок 1.1: Створення Моделей
- [ ] Створити `backend/models/league.go`
- [ ] Створити `backend/models/league_membership.go`
- [ ] Створити `backend/models/league_invitation.go`
- [ ] Оновити `backend/models/game_round.go` - додати поле LeagueID

#### Крок 1.2: Створення Репозиторіїв
- [ ] Створити `backend/repositories/league_repository.go`
- [ ] Створити `backend/repositories/league_membership_repository.go`
- [ ] Створити `backend/repositories/league_invitation_repository.go`
- [ ] Оновити `backend/repositories/game_round_repository.go` - додати метод FindByLeague

#### Крок 1.3: Міграції Бази Даних
- [ ] Додати скрипт/код для створення індексів для нових колекцій
- [ ] Додати міграцію для додавання league_id до існуючих game_rounds (опціонально, може бути NULL спочатку)

#### Крок 1.4: Юніт Тести
- [ ] Написати тести для league repository
- [ ] Написати тести для membership repository
- [ ] Написати тести для invitation repository

### Фаза 2: Backend Сервіси та Бізнес-Логіка

#### Крок 2.1: Калькулятор Рейтингу
- [ ] Створити `backend/services/standings_calculator.go`
- [ ] Реалізувати структуру PointsConfig
- [ ] Реалізувати функцію CalculateStandings
- [ ] Написати юніт тести для розрахунку рейтингу з різними сценаріями

#### Крок 2.2: League Service
- [ ] Створити `backend/services/league_service.go`
- [ ] Реалізувати CreateLeague
- [ ] Реалізувати GetLeague, ListLeagues
- [ ] Реалізувати методи членства (GetLeagueMembers, IsUserMember)
- [ ] Реалізувати BanUserFromLeague
- [ ] Реалізувати методи запрошень (CreateInvitation, AcceptInvitation)
- [ ] Реалізувати GetLeagueStandings
- [ ] Написати юніт тести для league service

#### Крок 2.3: Логіка Авторизації
- [ ] Створити `backend/auth/roles.go` - визначити роль суперадміна
- [ ] Оновити модель User якщо потрібно для підтримки ролей
- [ ] Реалізувати функції перевірки ролей

### Фаза 3: Backend API Обробники

#### Крок 3.1: Обробники Ліги
- [ ] Створити `backend/gameapi/league.go`
- [ ] Реалізувати createLeague (POST /api/leagues)
- [ ] Реалізувати listLeagues (GET /api/leagues)
- [ ] Реалізувати getLeague (GET /api/leagues/:code)
- [ ] Реалізувати getLeagueMembers (GET /api/leagues/:code/members)
- [ ] Реалізувати getLeagueStandings (GET /api/leagues/:code/standings)
- [ ] Реалізувати createInvitation (POST /api/leagues/:code/invitations)
- [ ] Реалізувати acceptInvitation (POST /api/leagues/join/:token)
- [ ] Реалізувати banUserFromLeague (POST /api/leagues/:code/ban/:userCode)

#### Крок 3.2: Оновлення Обробників Ігрових Кіл
- [ ] Оновити `backend/gameapi/gameround.go::startGame` - додати обов'язкове поле leagueID
- [ ] Оновити `backend/gameapi/gameround.go::listGameRounds` - фільтрувати по leagueID (опціональний параметр запиту)
- [ ] Додати валідацію для перевірки членства користувача при створенні ігрового кола

#### Крок 3.3: Middleware
- [ ] Створити `backend/middleware/league_middleware.go`
- [ ] Реалізувати RequireLeagueMembership middleware
- [ ] Реалізувати RequireSuperAdmin middleware

#### Крок 3.4: Реєстрація Маршрутів
- [ ] Оновити `backend/main.go` або конфігурацію маршрутів
- [ ] Зареєструвати маршрути ліги з відповідним middleware
- [ ] Оновити маршрути ігрових кіл для використання league middleware де потрібно

#### Крок 3.5: API Тести
- [ ] Написати інтеграційні тести для endpoints ліги
- [ ] Написати тести для процесу запрошень
- [ ] Написати тести для розрахунку рейтингу
- [ ] Написати тести для контролю доступу (членство, суперадмін)

### Фаза 4: Frontend Основа

#### Крок 4.1: TypeScript Типи
- [ ] Оновити `frontend/src/gametypes/types.ts`
- [ ] Додати інтерфейси League, LeagueMembership, LeagueInvitation, LeagueStanding
- [ ] Оновити GameRoundView для включення league_id

#### Крок 4.2: API Клієнт
- [ ] Створити `frontend/src/api/LeagueApi.ts`
- [ ] Реалізувати всі методи API (createLeague, listLeagues, тощо)
- [ ] Додати обробку помилок

#### Крок 4.3: Vuex Сховище
- [ ] Створити `frontend/src/store/league.ts`
- [ ] Визначити інтерфейс стану
- [ ] Реалізувати мутації
- [ ] Реалізувати дії (fetchLeagues, selectLeague, тощо)
- [ ] Реалізувати геттери
- [ ] Зареєструвати модуль сховища в головному store

### Фаза 5: Frontend Компоненти

#### Крок 5.1: Селектор Ліги
- [ ] Створити `frontend/src/components/LeagueSelector.vue`
- [ ] Реалізувати випадаючий список з лігами
- [ ] Підключити до Vuex store
- [ ] Зберігати поточну лігу в localStorage
- [ ] Додати в навігацію App.vue

#### Крок 5.2: Список Ліг
- [ ] Оновити `frontend/src/gametypes/ListLeagues.vue`
- [ ] Відображати список всіх ліг
- [ ] Додати кнопку "Створити Лігу" (для суперадмінів)
- [ ] Додати навігацію до деталей ліги
- [ ] Додати стани завантаження та обробку помилок

#### Крок 5.3: Деталі Ліги
- [ ] Створити `frontend/src/gametypes/LeagueDetails.vue`
- [ ] Відображати інформацію про лігу
- [ ] Відображати список членів
- [ ] Відображати таблицю рейтингу
- [ ] Додати кнопку "Згенерувати Запрошення"
- [ ] Додати функціонал "Забанити Користувача" (для суперадмінів)
- [ ] Додати функціонал копіювання посилання запрошення

#### Крок 5.4: Діалог Створення Ліги
- [ ] Створити `frontend/src/components/CreateLeagueDialog.vue`
- [ ] Форма з полем для введення назви
- [ ] Валідація
- [ ] Відправка до API
- [ ] Показ повідомлень успіху/помилки

#### Крок 5.5: Прийняття Запрошення
- [ ] Створити `frontend/src/gametypes/InvitationAccept.vue`
- [ ] Витягнути токен з параметрів маршруту
- [ ] Перевірити статус автентифікації
- [ ] Редирект на login якщо потрібно з URL повернення
- [ ] Прийняти запрошення після login
- [ ] Показати повідомлення успіху та редирект до ліги

#### Крок 5.6: Таблиця Рейтингу
- [ ] Створити `frontend/src/components/LeagueStandingsTable.vue`
- [ ] Відображати рейтинг в форматі таблиці
- [ ] Додати сортування по колонкам
- [ ] Показувати детальний розклад (бали за участь, позиції, модерацію)
- [ ] Виділяти поточного користувача

### Фаза 6: Інтеграція та Оновлення

#### Крок 6.1: Оновлення UI Ігрових Кіл
- [ ] Оновити `frontend/src/gametypes/GameroundsList.vue`
  - Додати фільтр по лізі
  - Показувати назву ліги для кожного раунду
  - Фільтрувати по поточній лізі з LeagueSelector
- [ ] Оновити `frontend/src/gametypes/GameroundEdit.vue`
  - Додати вибір ліги (або використовувати поточну лігу)
  - Фільтрувати список гравців по членству в лізі
  - Додати валідацію

#### Крок 6.2: Оновлення Головної Сторінки
- [ ] Оновити `frontend/src/views/HomeView.vue`
  - Додати контекст ліги до статистики
  - Показувати назву поточної ліги
  - Фільтрувати дані по обраній лізі

#### Крок 6.3: Оновлення Навігації
- [ ] Оновити `frontend/src/App.vue`
  - Додати LeagueSelector в header/navbar
  - Додати лінк на сторінку Ліг
  - Оновити мобільну навігацію

#### Крок 6.4: Оновлення Роутера
- [ ] Оновити `frontend/src/router/index.ts`
- [ ] Додати маршрути для деталей ліги, прийняття запрошення
- [ ] Додати navigation guards якщо потрібно

### Фаза 7: Тестування та Доопрацювання

#### Крок 7.1: End-to-End Тестування
- [ ] Протестувати повний процес запрошення (створити → поділитись → прийняти)
- [ ] Протестувати створення ліги (суперадмін)
- [ ] Протестувати створення ігрового кола з контекстом ліги
- [ ] Протестувати розрахунок рейтингу з реальними даними
- [ ] Протестувати функціонал бану
- [ ] Протестувати переключення між лігами в UI

#### Крок 7.2: Обробка Помилок
- [ ] Додати відповідні повідомлення про помилки для всіх сценаріїв збоїв
- [ ] Обробити застарілі запрошення
- [ ] Обробити забанених користувачів що намагаються отримати доступ до ліги
- [ ] Обробити не-членів що намагаються отримати доступ до ресурсів ліги

#### Крок 7.3: Доопрацювання UI/UX
- [ ] Додати стани завантаження до всіх компонентів
- [ ] Додати порожні стани (немає ліг, немає членів, тощо)
- [ ] Додати діалоги підтвердження для деструктивних дій
- [ ] Покращити responsive дизайн
- [ ] Додати інтернаціоналізацію (i18n) для нових компонентів

#### Крок 7.4: Оптимізація Продуктивності
- [ ] Додати пагінацію для великих таблиць рейтингу
- [ ] Додати кешування для часто використовуваних даних
- [ ] Оптимізувати розрахунок рейтингу для великої кількості раундів
- [ ] Перевірити індекси бази даних

#### Крок 7.5: Документація
- [ ] Додати документацію API (Swagger/OpenAPI)
- [ ] Оновити README з функціями ліг
- [ ] Додати керівництво користувача для управління лігами
- [ ] Задокументувати формулу розрахунку балів
- [ ] Додати документацію для розробників щодо розширення розрахунку рейтингу

### Фаза 8: Розгортання

#### Крок 8.1: План Міграції
- [ ] Створити скрипт міграції бази даних для production
- [ ] Спланувати обробку існуючих game_rounds без league_id
- [ ] Створити ліга за замовчуванням для існуючих даних (якщо потрібно)
- [ ] План резервного копіювання та стратегія відкату

#### Крок 8.2: Конфігурація
- [ ] Додати змінні оточення для конфігурації балів
- [ ] Додати feature flags якщо робимо поступове розгортання
- [ ] Налаштувати користувачів суперадміна

#### Крок 8.3: Розгортання
- [ ] Розгорнути зміни backend
- [ ] Запустити міграції бази даних
- [ ] Розгорнути зміни frontend
- [ ] Smoke test в production
- [ ] Моніторинг помилок

---

## Деталі Розрахунку Балів

### Конфігурація Балів за Замовчуванням

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
        // Для позицій > 5, можна використати формулу: max(0, 11 - position)
    },
}
```

### Алгоритм Розрахунку

Для кожного гравця в лізі:
1. Знайти всі завершені ігрові кола в лізі де гравець брав участь
2. Для кожного раунду:
   - `participation_points += ParticipationPoints`
   - `position_points += PositionPoints[player.Position]`
   - `moderation_points += ModerationPoints` (якщо is_moderator)
3. `total_points = participation_points + position_points + moderation_points`
4. Також зберегти статистику: games_played, games_moderated, first_place_count, тощо
5. Сортувати гравців по total_points (desc)

### Крайні Випадки

- **Незавершені раунди**: Не враховуються в рейтингу
- **Гравець без позиції**: Якщо position == 0, бали за позицію = 0
- **Кооперативні ігри**: Можливо потрібна окрема логіка (всі отримують однакові бали?)
- **Командні ігри**: Позиція команди чи індивідуальна позиція?

**TODO**: Уточнити як рахувати бали для різних типів підрахунку (cooperative, team_vs_team)

---

## Міркування Безпеки

1. **Перевірка Суперадміна**:
   - Необхідно мати надійний спосіб перевірки ролі суперадміна
   - Зберігати список ID користувачів суперадміна в конфігурації або БД

2. **Токени Запрошень**:
   - Використовувати криптографічно стійкий генератор (crypto/rand)
   - Достатня довжина токену (мінімум 32 байти)
   - Токени повинні бути унікальними

3. **Перевірки Авторизації**:
   - Перевіряти членство в лізі перед доступом до ресурсів ліги
   - Перевіряти що забанені користувачі не можуть виконувати дії в лізі
   - Перевіряти права при створенні запрошень

4. **Rate Limiting**:
   - Обмежити кількість запрошень що може створити один користувач
   - Захист від спаму при прийнятті запрошень

5. **Валідація Введення**:
   - Валідація назви ліги (довжина, допустимі символи)
   - Валідація всіх ObjectID перед використанням

---

## Відкриті Питання / Необхідні Рішення

1. **Конфігурація Балів**:
   - Де зберігати конфігурацію балів? (Змінні оточення, база даних, hardcoded?)
   - Чи потрібна можливість налаштування балів per-league?
   - Як рахувати бали для кооперативних та командних ігор?

2. **Ліга за Замовчуванням**:
   - Що робити з існуючими game_rounds без league_id?
   - Створити "Лігу за Замовчуванням" для старих даних?
   - Чи дозволяти створювати game rounds без ліги?

3. **Управління Суперадміном**:
   - Як визначити хто є суперадміном? (Список в config? Таблиця roles в БД?)
   - Чи потрібен UI для управління ролями суперадміна?

4. **Термін Дії Запрошень**:
   - Чи повинні запрошення мати термін дії?
   - Який термін за замовчуванням? (24 години? 7 днів? Безстроково?)

5. **Архівування Ліги**:
   - Чи потрібна можливість "архівувати" лігу (зробити read-only)?
   - Чи можна видалити лігу? (Що робити з game_rounds?)

6. **Множинні Запрошення**:
   - Чи може бути кілька активних запрошень для однієї ліги?
   - Чи потрібно показувати список активних запрошень?
   - Чи можна "відкликати" запрошення?

7. **Система Нотифікацій**:
   - Чи потрібні нотифікації при прийнятті запрошення?
   - Чи потрібні нотифікації при бані?

8. **Функції Таблиці Лідерів**:
   - Чи потрібна історія рейтингу (зміни рейтингу в часі)?
   - Чи потрібні досягнення/бейджі?
   - Чи потрібна статистика по типу гри в межах ліги?

---

## Майбутні Покращення (Поза Межами v1)

- **Сезони**: Розділення ліги на сезони з окремими рейтингами
- **Турніри**: Окремий режим для турнірів з bracket системою
- **Досягнення**: Система досягнень та бейджів
- **Статистика**: Детальна аналітика гравців (win rate, улюблені ігри, тощо)
- **Соціальні Функції**: Чат, коментарі до раундів
- **Розширені Права**: Різні ролі в лізі (модератор, член, глядач)
- **Налаштування Ліги**: Налаштування балів per-league
- **Експорт**: Експорт рейтингу в CSV/PDF
- **Публічні/Приватні Ліги**: Контроль видимості ліги
- **Федерація**: Об'єднання кількох ліг для спільних турнірів

---

## Критерії Успіху

Імплементація вважається успішною коли:

1. ✅ Суперадмін може створювати ліги
2. ✅ Гравець ліги може генерувати запрошення
3. ✅ Новий користувач може приєднатись по запрошенню (з автоматичним створенням акаунту)
4. ✅ Існуючий користувач може приєднатись по запрошенню
5. ✅ При створенні ігрового кола доступні тільки гравці з ліги
6. ✅ Ігрові кола відображаються в контексті ліги
7. ✅ Рейтинг відображає правильний підрахунок (участь + позиція + модерація)
8. ✅ Суперадмін може забанити гравця в лізі
9. ✅ Забанений гравець не може виконувати дії в лізі
10. ✅ UI дозволяє легко переключатись між лігами
11. ✅ Всі тести проходять успішно
12. ✅ Документація оновлена
