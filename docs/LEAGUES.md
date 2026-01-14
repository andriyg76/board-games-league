# Система Ліг

*[English version](LEAGUES.en.md)*

## Огляд

Система Ліг надає ізольовані конкурентні контексти, де гравці можуть змагатися, заробляти очки та відстежувати рейтинги. Кожна ліга має своїх членів, таблиці лідерів та ігрові раунди.

---

## Концепція та Вимоги

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
- **Поля ліги**: Назва (обов'язкове поле), опис (опціонально)
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
- **Вихід з ліги**: Користувач може покинути лігу
- **Бан**: Тільки суперадмін може забанити гравця в конкретній лізі

#### 3. Ігрові кола в контексті ліги
- Кожне ігрове коло (game round) може бути прив'язане до конкретної ліги
- При створенні раунду доступні тільки гравці з цієї ліги
- Список раундів фільтрується по обраній лізі

#### 4. Рейтингова система

**Загальний рейтинг гравця в лізі** складається з:
- **Бали за участь**: Фіксовані бали за кожну зіграну гру (2 бали)
- **Бали за позицію**: Залежно від місця в грі
  - 1-ше місце: 10 балів
  - 2-ге місце: 6 балів
  - 3-тє місце: 3 бали
  - 4-те+ місця: 1 бал
- **Бали за модерацію**: Додаткові бали, якщо гравець був модератором (1 бал)

**Формула:**
```
Рейтинг гравця = Σ (Бали_за_участь + Бали_за_позицію + Бали_за_модерацію)
```

**Сортування таблиці лідерів:**
1. Загальні бали (за спаданням)
2. Кількість ігор (за зростанням - для розв'язання нічиїх)
3. User ID (для консистентного порядку)

#### 5. UI/UX
- **Селектор ліги**: В інтерфейсі є можливість вибору поточної ліги
- **Контекст**: Більшість операцій відбуваються в контексті обраної ліги
- **Типи ігор**: Залишаються глобальними і доступні для всіх ліг

---

## Архітектура

### Моделі даних (Backend)

#### Модель Ліги (`league.go`)
```go
type LeagueStatus string

const (
    LeagueActive   LeagueStatus = "active"
    LeagueArchived LeagueStatus = "archived"
)

type League struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Version     int64              `bson:"version"`
    Name        string             `bson:"name"`
    Description string             `bson:"description,omitempty"`
    Status      LeagueStatus       `bson:"status"`      // "active" | "archived"
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
    CreatedBy   primitive.ObjectID `bson:"created_by"`
}
```

#### Модель Членства в Лізі (`league_membership.go`)
```go
type LeagueMembershipStatus string

const (
    MembershipActive  LeagueMembershipStatus = "active"
    MembershipPending LeagueMembershipStatus = "pending"
    MembershipVirtual LeagueMembershipStatus = "virtual"
    MembershipBanned   LeagueMembershipStatus = "banned"
)

type RecentCoPlayer struct {
    MembershipID primitive.ObjectID `bson:"membership_id"`
    LastPlayedAt time.Time          `bson:"last_played_at"`
}

type LeagueMembership struct {
    ID              primitive.ObjectID     `bson:"_id,omitempty"`
    Version         int64                  `bson:"version"`
    LeagueID        primitive.ObjectID     `bson:"league_id"`
    UserID          primitive.ObjectID     `bson:"user_id,omitempty"` // optional for pending/virtual
    Status          LeagueMembershipStatus `bson:"status"`
    JoinedAt        time.Time              `bson:"joined_at"`
    RecentCoPlayers []RecentCoPlayer       `bson:"recent_co_players,omitempty"` // max 10 items
    LastActivityAt  time.Time              `bson:"last_activity_at,omitempty"`  // last activity timestamp
    CreatedAt       time.Time              `bson:"created_at"`
    UpdatedAt       time.Time              `bson:"updated_at"`
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

#### Модель Запрошення до Ліги (`league_invitation.go`)
```go
type LeagueInvitation struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Version   int64              `bson:"version"`
    LeagueID  primitive.ObjectID `bson:"league_id"`
    CreatedBy primitive.ObjectID `bson:"created_by"` // Користувач що створив запрошення
    Token     string             `bson:"token"`      // Унікальний токен для посилання
    IsUsed    bool               `bson:"is_used"`
    UsedBy    primitive.ObjectID `bson:"used_by,omitempty"`
    UsedAt    time.Time          `bson:"used_at,omitempty"`
    ExpiresAt time.Time          `bson:"expires_at,omitempty"` // Термін дії (7 днів)
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

#### Рейтинг Ліги (обчислюється, не зберігається в БД)
```go
type LeagueStanding struct {
    UserID              primitive.ObjectID `json:"user_id"`
    UserName            string             `json:"user_name"`
    UserAvatar          string             `json:"user_avatar"`
    TotalPoints         int64              `json:"total_points"`
    GamesPlayed         int                `json:"games_played"`
    GamesModerated      int                `json:"games_moderated"`
    FirstPlaceCount     int                `json:"first_place_count"`
    SecondPlaceCount    int                `json:"second_place_count"`
    ThirdPlaceCount     int                `json:"third_place_count"`
    ParticipationPoints int64              `json:"participation_points"`
    PositionPoints      int64              `json:"position_points"`
    ModerationPoints    int64              `json:"moderation_points"`
}
```

#### Оновлення моделі GameRound
```go
type GameRound struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    Code             string             `bson:"-"`
    Version          int64              `bson:"version"`
    LeagueID         primitive.ObjectID `bson:"league_id,omitempty"` // Опціонально
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

### Колекції бази даних

#### leagues
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "name": String,
  "description": String,
  "status": String, // "active" | "archived"
  "created_at": ISODate,
  "updated_at": ISODate,
  "created_by": ObjectId
}
```

**Індекси:**
- `{ "name": 1 }` - унікальний
- `{ "status": 1 }`

#### league_memberships
```json
{
  "_id": ObjectId,
  "version": NumberLong,
  "league_id": ObjectId,
  "user_id": ObjectId, // optional for pending/virtual
  "status": String, // "active" | "pending" | "virtual" | "banned"
  "joined_at": ISODate,
  "recent_co_players": [
    {
      "membership_id": ObjectId,
      "last_played_at": ISODate
    }
  ],
  "last_activity_at": ISODate,
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
  "league_id": ObjectId, // optional
  // ... існуючі поля
}
```

**Новий Індекс:**
- `{ "league_id": 1, "start_time": -1 }`

---

## API Документація

### Аутентифікація

Всі endpoints ліг потребують аутентифікації через JWT токен в заголовку `Authorization`:
```
Authorization: Bearer <jwt_token>
```

### Endpoints

#### 1. Список всіх ліг

**Endpoint:** `GET /api/leagues`

**Опис:** Повертає всі активні ліги, доступні автентифікованому користувачу.

**Відповідь:**
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

**Статус коди:**
- `200 OK` - Успіх
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `500 Internal Server Error` - Помилка сервера

---

#### 2. Отримати деталі ліги

**Endpoint:** `GET /api/leagues/{code}`

**Опис:** Повертає детальну інформацію про конкретну лігу.

**URL параметри:**
- `code` - Код ліги (base64url encoded ObjectID)

**Відповідь:**
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

**Статус коди:**
- `200 OK` - Успіх
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є членом цієї ліги
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

---

#### 3. Створити лігу

**Endpoint:** `POST /api/leagues`

**Опис:** Створює нову лігу. **Потребує прав суперадміна.**

**Тіло запиту:**
```json
{
  "name": "Summer Championship 2026",
  "description": "Competitive league for summer season"
}
```

**Відповідь:**
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

**Статус коди:**
- `201 Created` - Ліга успішно створена
- `400 Bad Request` - Недійсне тіло запиту
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є суперадміном
- `500 Internal Server Error` - Помилка сервера

---

#### 4. Оновити лігу

**Endpoint:** `PUT /api/leagues/{code}`

**Опис:** Оновлює деталі ліги. **Потребує прав суперадміна.**

**URL параметри:**
- `code` - Код ліги

**Тіло запиту:**
```json
{
  "name": "Updated League Name",
  "description": "Updated description"
}
```

**Відповідь:**
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

**Статус коди:**
- `200 OK` - Успіх
- `400 Bad Request` - Недійсне тіло запиту
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є суперадміном
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

---

#### 5. Архівувати/Розархівувати лігу

**Endpoint:** `PUT /api/leagues/{code}/status`

**Опис:** Змінює статус ліги (active/archived). **Потребує прав суперадміна.**

**URL параметри:**
- `code` - Код ліги

**Тіло запиту:**
```json
{
  "status": "archived"
}
```

**Допустимі значення статусу:** `active`, `archived`

**Відповідь:**
```json
{
  "league": {
    "id": "507f1f77bcf86cd799439011",
    "code": "ABC123",
    "status": "archived"
  }
}
```

**Статус коди:**
- `200 OK` - Успіх
- `400 Bad Request` - Недійсне значення статусу
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є суперадміном
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

---

#### 6. Отримати таблицю лідерів

**Endpoint:** `GET /api/leagues/{code}/standings`

**Опис:** Повертає поточні рейтинги для всіх членів ліги.

**URL параметри:**
- `code` - Код ліги

**Відповідь:**
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

**Сортування:**
1. Загальні бали (за спаданням)
2. Кількість ігор (за зростанням - для розв'язання нічиїх)
3. User ID (для консистентного порядку)

**Статус коди:**
- `200 OK` - Успіх
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є членом цієї ліги
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

---

#### 7. Отримати членів ліги

**Endpoint:** `GET /api/leagues/{code}/members`

**Опис:** Повертає список всіх членів ліги з їх статусом.

**URL параметри:**
- `code` - Код ліги

**Відповідь:**
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

**Значення статусу члена:**
- `active` - Член може брати участь в іграх
- `pending` - Створено через запрошення, очікує прийняття користувачем
- `virtual` - Гравець брав участь в іграх, але ніколи не входив в систему
- `banned` - Користувач заблокований в лізі

**Статус коди:**
- `200 OK` - Успіх
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є членом цієї ліги
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

---

#### 8. Заблокувати/Розблокувати члена

**Endpoint:** `PUT /api/leagues/{code}/members/{userId}/status`

**Опис:** Змінює статус члена (ban/unban). **Потребує прав суперадміна.**

**URL параметри:**
- `code` - Код ліги
- `userId` - ID користувача (hex string)

**Тіло запиту:**
```json
{
  "status": "banned"
}
```

**Допустимі значення статусу:** `active`, `banned`

**Відповідь:**
```json
{
  "success": true
}
```

**Статус коди:**
- `200 OK` - Успіх
- `400 Bad Request` - Недійсне значення статусу
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є суперадміном
- `404 Not Found` - Ліга або член не знайдені
- `500 Internal Server Error` - Помилка сервера

---

#### 9. Отримати рекомендованих гравців

**Endpoint:** `GET /api/leagues/{code}/suggested-players`

**Опис:** Повертає рекомендованих гравців для створення ігрового раунду, на основі історії спільних ігор та активності членства в лізі. Цей endpoint надає інтелектуальні рекомендації гравців для wizard налаштування гри.

**URL параметри:**
- `code` - Код ліги

**Відповідь:**
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

**Структура відповіді:**
- `current_player`: Членство автентифікованого користувача (null якщо суперадмін без membership)
- `recent_players`: До 10 гравців, з якими недавно грали, відсортовані за `last_played_at` DESC
- `other_players`: Інші члени ліги (виключаючи current + recent), відсортовані за `last_activity_at` DESC

**Для членів ліги:**
- `current_player`: 1 елемент (якщо membership існує)
- `recent_players`: До 10 елементів з кешу `recent_co_players`
- `other_players`: До 10 елементів, відсортовані за активністю

**Для суперадміна без membership:**
- `current_player`: null
- `recent_players`: Порожній масив
- `other_players`: До 20 елементів, відсортовані за активністю

**Статус коди:**
- `200 OK` - Успіх
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є членом цієї ліги і не є суперадміном
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

**Примітки:**
- Останні гравці кешуються в `LeagueMembership.recent_co_players` (максимум 10 елементів)
- Кеш автоматично оновлюється при фіналізації ігрового раунду
- Віртуальні гравці (створені через запрошення) включені в відповідь

---

#### 10. Створити запрошення

**Endpoint:** `POST /api/leagues/{code}/invitations`

**Опис:** Створює одноразове запрошення для ліги. Може також створити віртуального гравця, надавши alias. Дійсне 7 днів.

**URL параметри:**
- `code` - Код ліги

**Тіло запиту (опціонально):**
```json
{
  "alias": "NewPlayer"
}
```

Якщо `alias` надано:
- Створює віртуального гравця (pending membership) з вказаним alias
- Перевіряє унікальність alias серед активних членів та активних запрошень
- Створює `LeagueMembership` зі статусом `pending`
- Створює `LeagueInvitation` з токеном
- Оновлює кеш `recent_co_players` поточного користувача (додає нового гравця в кінець)
- Встановлює `last_activity_at` для нового pending membership

**Відповідь:**
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

**Статус коди:**
- `201 Created` - Запрошення успішно створено
- `400 Bad Request` - Недійсний alias або alias вже існує
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є членом цієї ліги
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

**Примітки:**
- Якщо alias надано, віртуальний гравець одразу додається в кеш `recent_co_players` створювача
- Якщо кеш повний (10 елементів), найстаріший запис видаляється
- Віртуальні гравці можуть брати участь в іграх до прийняття запрошення

---

#### 11. Перегляд запрошення (Публічний)

**Endpoint:** `GET /api/leagues/join/{token}/preview`

**Опис:** Повертає публічну інформацію про запрошення. Цей endpoint не потребує аутентифікації і може використовуватися для відображення деталей запрошення перед входом.

**URL параметри:**
- `token` - Токен запрошення

**Відповідь:**
```json
{
  "league_name": "Summer Championship 2026",
  "inviter_alias": "John",
  "player_alias": "NewPlayer",
  "expires_at": "2026-01-15T00:00:00Z",
  "status": "valid"
}
```

**Значення поля status:**
- `valid` - Запрошення може бути прийняте
- `expired` - Запрошення прострочене
- `used` - Запрошення вже використане

**Статус коди:**
- `200 OK` - Успіх
- `404 Not Found` - Запрошення не знайдено

---

#### 12. Прийняти запрошення

**Endpoint:** `POST /api/leagues/join/{token}`

**Опис:** Приймає запрошення і додає користувача до ліги. Pending membership, створений з запрошенням, активується і прив'язується до користувача.

**URL параметри:**
- `token` - Токен запрошення

**Тіло запиту:** Відсутнє

**Відповідь (Успіх):**
```json
{
  "code": "ABC123",
  "name": "Summer Championship 2026",
  "status": "active",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**Відповідь (Вже член - 409 Conflict):**
```json
{
  "error": "user is already a member of this league",
  "league_code": "ABC123"
}
```

**Статус коди:**
- `200 OK` - Запрошення успішно прийняте
- `400 Bad Request` - Недійсне запрошення (прострочене, використане, власне запрошення)
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `404 Not Found` - Запрошення не знайдено
- `409 Conflict` - Користувач вже є членом (включає league_code для редиректу)

**Примітки:**
- Кожне запрошення може бути використане тільки один раз
- Запрошення прострочуються через 7 днів
- Користувачі не можуть приймати свої власні запрошення
- Помилка "вже член" (409) включає `league_code` для редиректу на фронтенді

---

#### 13. Покинути лігу

**Endpoint:** `DELETE /api/leagues/{code}/members/me`

**Опис:** Видаляє автентифікованого користувача з ліги.

**URL параметри:**
- `code` - Код ліги

**Тіло запиту:** Відсутнє

**Відповідь:**
```json
{
  "success": true
}
```

**Статус коди:**
- `204 No Content` - Користувача успішно видалено
- `401 Unauthorized` - Відсутня або недійсна аутентифікація
- `403 Forbidden` - Користувач не є членом цієї ліги
- `404 Not Found` - Ліга не знайдена
- `500 Internal Server Error` - Помилка сервера

---

## Система очок

### Бали за участь
- **2 бали** за кожну зіграну гру (надаються всім гравцям)

### Бали за позицію
На основі фінальної позиції в кожній грі:
- **1-ше місце:** 10 балів
- **2-ге місце:** 6 балів
- **3-тє місце:** 3 бали
- **4-те+ місця:** 1 бал

### Бали за модерацію
- **1 бал** за кожну гру, де гравець був модератором

### Розрахунок загальних балів
```
Загальні бали = Бали за участь + Бали за позицію + Бали за модерацію
```

### Розрахунок таблиці лідерів
Таблиця лідерів перераховується з історії ігор:
1. Отримати всі ігрові раунди для ліги
2. Для кожного ігрового раунду розрахувати бали для кожного гравця
3. Агрегувати бали по гравцях
4. Сортувати за загальними балами (за спаданням), потім за кількістю ігор (за зростанням)

---

## Ігрові раунди в лігах

При створенні ігрового раунду можна опціонально пов'язати його з лігою:

**Endpoint:** `POST /api/games`

**Тіло запиту:**
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

Коли ігровий раунд фіналізується, таблиця лідерів ліги автоматично оновлюється.

---

## Важливі примітки

### Міграція
- **Міграція не потрібна**: Порожні масиви `recent_co_players` автоматично заповняться при фіналізації нових ігрових раундів
- Порожні значення `last_activity_at` сортуються в кінець списків (NULLS LAST)

### Доступ суперадміна
- Суперадміни можуть створювати ігрові раунди без membership, але не будуть автоматично додані до списку гравців
- Суперадміни без membership отримують до 20 `other_players` у відповіді рекомендованих гравців (проти 10 для звичайних членів)

### Віртуальні гравці
- При створенні через запрошення з alias, віртуальні гравці одразу додаються в кеш `recent_co_players` створювача в **кінець** списку
- Якщо кеш повний (10 елементів), найстаріший запис видаляється
- Віртуальні гравці можуть брати участь в іграх до прийняття запрошення

### Endpoint запрошень
- Endpoint `POST /api/leagues/{code}/invitations` використовує **існуючу** функціональність, розширену оновленням кешу
- Перевірка унікальності перевіряє alias серед:
  - Активних членів ліги
  - Pending членів (активні запрошення)
  - Віртуальних членів

### Потік вибору гравців
Система рекомендованих гравців надає інтелектуальні рекомендації на основі:
1. **Поточний гравець**: Членство автентифікованого користувача (якщо існує)
2. **Останні гравці**: До 10 гравців з кешу `recent_co_players`, відсортовані за `last_played_at` DESC
3. **Інші гравці**: Решта членів ліги (виключаючи current + recent), відсортовані за `last_activity_at` DESC NULLS LAST

---

## Frontend Інтеграція

### League Store (Pinia)

Frontend використовує Pinia store (`src/store/league.ts`) для керування станом:

```typescript
import { useLeagueStore } from '@/store/league'

const leagueStore = useLeagueStore()

// Отримати всі ліги
await leagueStore.fetchLeagues()

// Отримати конкретну лігу
await leagueStore.fetchLeague('ABC123')

// Створити запрошення
const result = await leagueStore.createInvitation()
console.log(result.invitation_link)

// Створити запрошення з віртуальним гравцем
const resultWithPlayer = await leagueStore.createInvitation({ alias: "NewPlayer" })

// Отримати рекомендованих гравців для налаштування гри
const suggested = await leagueStore.getSuggestedPlayers('ABC123')
// Повертає: { current_player, recent_players, other_players }

// Прийняти запрошення
await leagueStore.acceptInvitation('token123')
```

### Доступні дії Store

- `fetchLeagues()` - Завантажити всі доступні ліги
- `fetchLeague(code)` - Завантажити деталі конкретної ліги
- `createLeague(data)` - Створити нову лігу (тільки суперадмін)
- `updateLeague(code, data)` - Оновити лігу (тільки суперадмін)
- `updateLeagueStatus(code, status)` - Архівувати/розархівувати лігу (тільки суперадмін)
- `fetchStandings(code)` - Завантажити таблицю лідерів ліги
- `fetchMembers(code)` - Завантажити членів ліги
- `banMember(userId)` - Заблокувати члена (тільки суперадмін)
- `unbanMember(userId)` - Розблокувати члена (тільки суперадмін)
- `createInvitation(data?)` - Створити запрошення (опціонально з alias віртуального гравця)
- `getSuggestedPlayers(code)` - Отримати рекомендованих гравців для налаштування гри
- `acceptInvitation(token)` - Прийняти запрошення
- `leaveLeague(code)` - Покинути лігу

---

## Маршрути Frontend

- `/ui/leagues` - Список всіх ліг
- `/ui/leagues/:code` - Деталі ліги (таблиця лідерів, члени, запрошення)
- `/ui/leagues/join/:token` - Сторінка прийняття запрошення

---

## Безпека

1. **Аутентифікація обов'язкова**: Всі endpoints потребують валідного JWT токена
2. **Членство в лізі**: Більшість endpoints перевіряють, що користувач є активним членом
3. **Права суперадміна**: Створення, оновлення, архівування ліг та блокування членів потребують ролі суперадміна
4. **Термін дії запрошень**: Запрошення прострочуються через 7 днів для запобігання зловживанням
5. **Одноразове використання**: Запрошення можуть бути використані тільки один раз
6. **Забезпечення бану**: Заблоковані члени не можуть брати участь в іграх в межах ліги

---

## Покращення UX

### Пріоритет 1: Критичні проблеми UX

#### 1.1 Повідомлення про помилки для користувачів

**Поточна проблема:**
```typescript
catch (error) {
  console.error('Error creating league:', error);  // Користувач не бачить це!
}
```

**Рекомендація:**
Додати toast-сповіщення або inline alerts для невдалих дій.

**Реалізація:**
```typescript
// Використовуючи Vuetify snackbar або v-alert
const errorMessage = ref<string | null>(null);

try {
  await leagueStore.createLeague(newLeagueName.value);
} catch (error) {
  errorMessage.value = error instanceof Error
    ? error.message
    : 'Не вдалося створити лігу';
  console.error('Error creating league:', error);
}
```

**Затронуті компоненти:**
- `LeagueList.vue` - дія createLeague
- `LeagueDetails.vue` - дії archiveLeague, unarchiveLeague, banMember
- Всі компоненти з API викликами

#### 1.2 Стани завантаження для дій

**Поточна проблема:**
Дії архівування/розархівування та бану/розбану не мають індикаторів завантаження.

**Рекомендація:**
Додати стани завантаження до кнопок дій.

**Реалізація в LeagueDetails.vue:**
```typescript
const archiving = ref(false);

const archiveLeague = async () => {
  archiving.value = true;
  try {
    await leagueStore.archiveLeague(currentLeague.value.code);
    showSuccess('Лігу успішно архівовано');
  } catch (error) {
    showError('Помилка архівування ліги');
  } finally {
    archiving.value = false;
  }
};
```

**Затронуті компоненти:**
- `LeagueDetails.vue` - дії archive, unarchive, ban
- `LeagueInvitation.vue` - вже має стан завантаження ✓

#### 1.3 Кращі діалоги підтвердження

**Поточна проблема:**
Нативні діалоги `confirm()` не стилізовані і не відповідають дизайну додатку.

**Рекомендація:**
Використовувати Vuetify діалоги для підтверджень.

**Реалізація:**
```vue
<template>
  <!-- Діалог підтвердження -->
  <v-dialog v-model="showConfirmDialog" max-width="400">
    <v-card>
      <v-card-title>{{ confirmTitle }}</v-card-title>
      <v-card-text>{{ confirmMessage }}</v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="showConfirmDialog = false">Скасувати</v-btn>
        <v-btn
          color="error"
          :loading="actionInProgress"
          @click="confirmAction"
        >
          {{ confirmButtonText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
const showConfirmDialog = ref(false);
const confirmTitle = ref('');
const confirmMessage = ref('');
const confirmAction = ref<(() => void) | null>(null);

const requestBan = (member: LeagueMember) => {
  confirmTitle.value = 'Заблокувати користувача?';
  confirmMessage.value = `Ви впевнені, що хочете заблокувати ${member.user_name}? Користувач не зможе брати участь в іграх цієї ліги.`;
  confirmButtonText.value = 'Заблокувати';
  confirmAction.value = () => banMember(member);
  showConfirmDialog.value = true;
};
</script>
```

**Затронуті компоненти:**
- `LeagueDetails.vue` - дії ban/unban, archive/unarchive
- `LeagueList.vue` - дія leave league (якщо реалізовано)

---

### Пріоритет 2: Покращений зворотний зв'язок

#### 2.1 Toast-сповіщення

**Рекомендація:**
Додати глобальний компонент toast/snackbar для повідомлень про успіх/помилку.

**Реалізація:**

Створити `composables/useToast.ts`:
```typescript
import { ref } from 'vue';

const message = ref('');
const type = ref<'success' | 'error' | 'info' | 'warning'>('info');
const show = ref(false);

export function useToast() {
  const showSuccess = (msg: string) => {
    message.value = msg;
    type.value = 'success';
    show.value = true;
  };

  const showError = (msg: string) => {
    message.value = msg;
    type.value = 'error';
    show.value = true;
  };

  const showInfo = (msg: string) => {
    message.value = msg;
    type.value = 'info';
    show.value = true;
  };

  return {
    message,
    type,
    show,
    showSuccess,
    showError,
    showInfo,
  };
}
```

Додати в `App.vue`:
```vue
<template>
  <v-app>
    <!-- ... контент додатку ... -->

    <v-snackbar
      v-model="toast.show"
      :color="toast.type"
      :timeout="3000"
      location="top"
    >
      {{ toast.message }}
      <template v-slot:actions>
        <v-btn icon="mdi-close" @click="toast.show = false" />
      </template>
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { useToast } from '@/composables/useToast';
const toast = useToast();
</script>
```

**Використання в компонентах:**
```typescript
import { useToast } from '@/composables/useToast';

const { showSuccess, showError } = useToast();

try {
  await leagueStore.createLeague(name);
  showSuccess('Лігу успішно створено!');
} catch (error) {
  showError('Не вдалося створити лігу');
}
```

#### 2.2 Оптимістичні оновлення UI

**Рекомендація:**
Оновлювати UI одразу для кращої сприйнятої продуктивності, потім відкотити якщо API не вдався.

**Приклад - Бан члена:**
```typescript
const banMember = async (member: LeagueMember) => {
  // Оптимістично оновити UI
  member.status = 'banned';

  try {
    await leagueStore.banUser(currentLeague.value.code, member.user_id);
    showSuccess(`${member.user_name} заблоковано`);
  } catch (error) {
    // Відкотити при помилці
    member.status = 'active';
    showError('Не вдалося заблокувати користувача');
  }
};
```

---

### Пріоритет 3: Доступність та полірування

#### 3.1 Навігація з клавіатури

**Рекомендація:**
Переконатися, що всі інтерактивні елементи доступні з клавіатури.

**Чеклист:**
- [ ] Всі кнопки фокусовані
- [ ] Порядок Tab логічний
- [ ] Клавіша Enter працює на сфокусованих кнопках
- [ ] Escape закриває діалоги
- [ ] ARIA labels на іконкових кнопках

**Реалізація:**
```vue
<v-btn
  icon="mdi-dots-vertical"
  aria-label="Керування лігою"
  @click="showManageMenu = !showManageMenu"
/>
```

#### 3.2 Скелетони завантаження

**Рекомендація:**
Замінити спінери на skeleton loaders для кращого UX.

**Реалізація:**
```vue
<v-skeleton-loader
  v-if="loading"
  type="list-item-avatar-three-line@3"
/>
```

**Затронуті компоненти:**
- `LeagueList.vue` - картки ліг
- `LeagueDetails.vue` - таблиця лідерів, члени
- `LeagueStandings.vue` - рядки таблиці

#### 3.3 Анімації переходів

**Рекомендація:**
Додати плавні переходи для кращого візуального зворотного зв'язку.

**Реалізація:**
```vue
<transition-group name="list" tag="div">
  <league-card
    v-for="league in activeLeagues"
    :key="league.code"
    :league="league"
  />
</transition-group>

<style scoped>
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from {
  opacity: 0;
  transform: translateY(-30px);
}

.list-leave-to {
  opacity: 0;
  transform: translateY(30px);
}
</style>
```

#### 3.4 Покращення порожніх станів

**Поточний стан:**
Простий текстовий меседж.

**Рекомендація:**
Додати ілюстрації та призив до дії.

**Реалізація:**
```vue
<v-card-text v-if="activeLeagues.length === 0" class="text-center py-8">
  <v-icon size="80" color="grey-lighten-1" class="mb-4">
    mdi-trophy-outline
  </v-icon>
  <div class="text-h6 mb-2">Немає активних ліг</div>
  <div class="text-body-2 text-medium-emphasis mb-4">
    Створіть нову лігу або приєднайтесь через запрошення
  </div>
  <div class="d-flex justify-center gap-2">
    <v-btn
      v-if="canCreateLeague"
      color="primary"
      @click="showCreateDialog = true"
    >
      Створити лігу
    </v-btn>
  </div>
</v-card-text>
```

---

### Пріоритет 4: Розширені функції

#### 4.1 Пошук та фільтрація

**Рекомендація:**
Додати пошук для списків ліг та членів.

**Реалізація:**
```vue
<v-text-field
  v-model="searchQuery"
  prepend-inner-icon="mdi-magnify"
  label="Пошук ліг"
  clearable
  hide-details
  class="mb-4"
/>

<script setup>
const searchQuery = ref('');

const filteredLeagues = computed(() => {
  if (!searchQuery.value) return activeLeagues.value;

  return activeLeagues.value.filter(league =>
    league.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});
</script>
```

#### 4.2 Пагінація для великих списків

**Рекомендація:**
Додати пагінацію коли члени/таблиця лідерів перевищують 50 елементів.

**Реалізація:**
```vue
<v-data-table
  :items="standings"
  :headers="headers"
  :items-per-page="25"
  :page.sync="page"
/>
```

#### 4.3 Оновлення в реальному часі

**Рекомендація:**
Додати WebSocket або polling для оновлень таблиці лідерів в реальному часі.

**Реалізація (Polling):**
```typescript
import { useIntervalFn } from '@vueuse/core';

onMounted(() => {
  // Оновлювати таблицю лідерів кожні 30 секунд
  useIntervalFn(() => {
    if (activeTab.value === 'standings') {
      leagueStore.refreshStandings(currentLeague.value.code);
    }
  }, 30000);
});
```

#### 4.4 Скасування дій

**Рекомендація:**
Дозволити скасування для критичних дій як бан/вихід.

**Реалізація:**
```typescript
const showUndoBan = ref(false);
let undoTimeout: ReturnType<typeof setTimeout> | null = null;

const banMember = async (member: LeagueMember) => {
  try {
    await leagueStore.banUser(code, member.user_id);

    showUndoBan.value = true;
    undoTimeout = setTimeout(() => {
      showUndoBan.value = false;
    }, 5000);

    showSuccess('Користувача заблоковано');
  } catch (error) {
    showError('Помилка блокування');
  }
};

const undoBan = async (member: LeagueMember) => {
  if (undoTimeout) clearTimeout(undoTimeout);
  showUndoBan.value = false;

  try {
    await leagueStore.unbanUser(code, member.user_id);
    showSuccess('Блокування скасовано');
  } catch (error) {
    showError('Помилка скасування');
  }
};
```

---

## Покращення валідації форм

### LeagueList.vue - Діалог створення

**Поточний стан:**
Базова валідація обов'язкового поля.

**Рекомендація:**
Додати комплексну валідацію.

**Реалізація:**
```typescript
const nameRules = [
  (v: string) => !!v || 'Назва обов\'язкова',
  (v: string) => v.length >= 3 || 'Мінімум 3 символи',
  (v: string) => v.length <= 50 || 'Максимум 50 символів',
  (v: string) => /^[а-яА-ЯёЁіІїЇєЄa-zA-Z0-9\s-]+$/.test(v) || 'Недопустимі символи',
];

const descriptionRules = [
  (v: string) => !v || v.length <= 200 || 'Максимум 200 символів',
];
```

---

## Гайдлайни повідомлень про помилки

### Зрозумілі повідомлення про помилки

**Погано:**
```
Error: MongoError: connection refused
```

**Добре:**
```
Не вдалося підключитися до сервера. Перевірте з'єднання та спробуйте ще раз.
```

### Маппінг повідомлень про помилки

Створити утиліту для маппінгу API помилок на зрозумілі повідомлення:

```typescript
// utils/errorMessages.ts
export function getUserFriendlyError(error: Error): string {
  const errorMap: Record<string, string> = {
    'Network Error': 'Помилка мережі. Перевірте з\'єднання.',
    'unauthorized': 'Необхідна авторизація',
    'forbidden': 'Недостатньо прав доступу',
    'not found': 'Ресурс не знайдено',
    'already exists': 'Така ліга вже існує',
    'invalid token': 'Невірне запрошення',
    'token expired': 'Запрошення прострочене',
  };

  const message = error.message.toLowerCase();

  for (const [key, value] of Object.entries(errorMap)) {
    if (message.includes(key)) {
      return value;
    }
  }

  return 'Сталася помилка. Спробуйте ще раз.';
}
```

**Використання:**
```typescript
import { getUserFriendlyError } from '@/utils/errorMessages';

catch (error) {
  const message = error instanceof Error
    ? getUserFriendlyError(error)
    : 'Невідома помилка';
  showError(message);
}
```

---

## Чеклист реалізації

### Фаза 1: Критичні виправлення
- [ ] Додати систему toast-сповіщень
- [ ] Додати повідомлення про помилки до всіх обробників дій
- [ ] Додати стани завантаження до кнопок дій
- [ ] Замінити нативний confirm() на Vuetify діалоги

### Фаза 2: Покращений зворотний зв'язок
- [ ] Додати повідомлення про успіх для всіх дій
- [ ] Реалізувати оптимістичні оновлення UI
- [ ] Додати маппінг зрозумілих повідомлень про помилки

### Фаза 3: Полірування
- [ ] Додати skeleton loaders
- [ ] Додати анімації переходів
- [ ] Покращити порожні стани
- [ ] Додати ARIA labels для доступності

### Фаза 4: Розширені (Опціонально)
- [ ] Додати пошук та фільтрацію
- [ ] Додати пагінацію
- [ ] Реалізувати оновлення в реальному часі
- [ ] Додати функціональність скасування

---

## Тестування після реалізації

Для кожного покращеного компонента протестувати:

1. **Сценарії помилок:**
   - Збій мережі
   - Недійсний ввід
   - Неавторизований доступ
   - Помилки сервера

2. **Стани завантаження:**
   - Перевірити що спінер з'являється
   - Перевірити що кнопки вимикаються під час завантаження
   - Перевірити плавні переходи

3. **Зворотний зв'язок про успіх:**
   - Перевірити що повідомлення про успіх з'являються
   - Перевірити що UI оновлюється правильно
   - Перевірити що редиректи працюють

4. **Доступність:**
   - Tab через всі інтерактивні елементи
   - Тестувати з screen reader
   - Тестувати клавіатурні скорочення

---

## Висновок

Ці покращення значно покращать користувацький досвід системи ліг шляхом:

1. **Надання чіткого зворотного зв'язку** для всіх дій користувача
2. **Зменшення фрустрації** з кращими повідомленнями про помилки
3. **Покращення сприйнятої продуктивності** зі станами завантаження
4. **Забезпечення доступності** для всіх користувачів
5. **Додавання полірування** з анімаціями та переходами

Реалізувати по фазах на основі пріоритету, ретельно тестувати після кожної фази.

