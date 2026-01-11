# Керування іграми

*[English version](GAME_MANAGEMENT.en.md)*

## Огляд

Додаток Board Games League надає комплексний функціонал керування іграми, включаючи налаштування типів ігор, відстеження ігрових раундів та керування гравцями. Цей документ охоплює всі функції, пов'язані з іграми, реалізовані в системі.

## Архітектура

### Моделі даних

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    GameType     │     │    GameRound    │     │      User       │
│   (Тип гри)     │     │  (Ігровий раунд)│     │    (Гравець)    │
│                 │     │                 │     │                 │
│ - ID            │◄────│ - GameTypeID    │     │ - ID            │
│ - Name          │     │ - Name          │     │ - Alias         │
│ - ScoringType   │     │ - StartTime     │     │ - Avatar        │
│ - Labels        │     │ - EndTime       │     │ - Name          │
│ - Teams         │     │ - Players[]     │◄────│                 │
│ - MinPlayers    │     │ - TeamScores[]  │     │                 │
│ - MaxPlayers    │     │ - CoopScore     │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### Структура компонентів

```
Backend                          Frontend
├── gameapi/                     ├── api/
│   ├── handlers.go             │   └── GameApi.ts
│   ├── gametype.go             ├── gametypes/
│   ├── gameround.go            │   ├── ListGameTypes.vue
│   └── player.go               │   ├── GameroundsList.vue
├── models/                      │   ├── GameroundEdit.vue
│   ├── game_type.go            │   ├── FinalizeGameDialog.vue
│   └── game_round.go           │   └── types.ts
└── repositories/                └── store/
    ├── game_type_repository.go     ├── game.ts
    └── game_round_repository.go    └── player.ts
```

## Типи ігор

### Концепція

Тип гри визначає правила та механізм підрахунку очок для конкретної настільної гри. Кожен тип гри має:

- **Назва**: Відображувана назва гри
- **Тип підрахунку**: Як розраховуються очки та визначаються переможці
- **Мітки**: Опціональні теги категоризації з кольорами та іконками
- **Команди**: Опціональні визначення команд для командних ігор
- **Обмеження гравців**: Мінімальна та максимальна кількість гравців

### Типи підрахунку очок

Система підтримує декілька механізмів підрахунку:

| Тип підрахунку | Опис | Приклад використання |
|----------------|------|----------------------|
| `classic` | Традиційний змагальний підрахунок - виграє найвища сума | Більшість настільних ігор (Catan, Ticket to Ride) |
| `cooperative` | Всі гравці виграють або програють разом | Pandemic, Forbidden Island |
| `cooperative_with_moderator` | Кооперативна гра з окремим модератором | Ігри з ведучим |
| `team_vs_team` | Команди змагаються одна з одною | Командні ігри |
| `mafia` | Команда проти команди з прихованими ролями та модератором | Мафія, Вовкулаки |
| `custom` | Без попередньо визначеної схеми - пряме введення очок | Будь-який власний підрахунок |

### Мітки

Мітки використовуються для категоризації гравців у грі (наприклад, ролі, стартові позиції). Кожна мітка має:

- **Назва**: Ідентифікатор мітки
- **Колір**: Колір відображення (формат hex)
- **Іконка**: Назва іконки Material Design

Мітки доступні для типів підрахунку: `classic`, `custom`

### Команди

Команди групують гравців для командних ігор. Кожна команда має:

- **Назва**: Ідентифікатор команди
- **Колір**: Колір команди (формат hex)
- **Іконка**: Іконка команди

Команди доступні для типів підрахунку: `mafia`, `custom`, `team_vs_team`

### Структура моделі

```go
type GameType struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Version     int64              `bson:"version"`
    Name        string             `bson:"name"`
    ScoringType string             `bson:"scoring_type"`
    Icon        string             `bson:"icon"`
    Labels      []Label            `bson:"labels"`
    Teams       []Label            `bson:"teams"`
    MinPlayers  int                `bson:"min_players"`
    MaxPlayers  int                `bson:"max_players"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}

type Label struct {
    Name  string `bson:"name"`
    Color string `bson:"color"`
    Icon  string `bson:"icon"`
}
```

## Ігрові раунди

### Концепція

Ігровий раунд представляє одну ігрову сесію певного типу гри. Він відстежує:

- Який тип гри грається
- Коли гра почалася та закінчилася
- Хто брав участь та їхні позиції/очки
- Командні очки (для командних ігор)
- Кооперативні очки (для кооперативних ігор)

### Життєвий цикл раунду

```
┌─────────┐     ┌───────────┐     ┌───────────┐     ┌──────────┐
│  Старт  │────>│  Активна  │────>│ Завершення│────>│Завершено │
│   гри   │     │  (Грається)│    │  (Очки)   │     │          │
└─────────┘     └───────────┘     └───────────┘     └──────────┘
     │                │                  │                │
     │                │                  │                │
  Створити        Оновити          Встановити        EndTime
  раунд           очки             фінальні          встановлено
                                   позиції
```

### Структура моделі

```go
type GameRound struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    Version          int64              `bson:"version"`
    Name             string             `bson:"name"`
    GameTypeID       primitive.ObjectID `bson:"game_type_id"`
    StartTime        time.Time          `bson:"start_time"`
    EndTime          time.Time          `bson:"end_time"`
    Players          []GameRoundPlayer  `bson:"players"`
    TeamScores       []TeamScore        `bson:"team_scores"`
    CooperativeScore int64              `bson:"cooperative_score"`
    CreatedAt        time.Time          `bson:"created_at"`
    UpdatedAt        time.Time          `bson:"updated_at"`
}

type GameRoundPlayer struct {
    PlayerID    primitive.ObjectID `bson:"player_id"`
    Position    int                `bson:"position"`
    Score       int64              `bson:"score"`
    IsModerator bool               `bson:"is_moderator"`
    TeamName    string             `bson:"team_name"`
    LabelName   string             `bson:"label_name"`
}

type TeamScore struct {
    Name     string `bson:"name"`
    Score    int64  `bson:"score"`
    Position int    `bson:"position"`
}
```

### Валідація команд

Для ігор з визначеними командами:
- Кожна команда повинна мати принаймні одного призначеного гравця
- Гравці повинні бути призначені тільки до існуючих команд
- Валідація відбувається при створенні нового ігрового раунду

## Гравці

### Концепція

Гравці - це користувачі, які беруть участь в ігрових раундах. Система гравців надає:

- Список всіх зареєстрованих гравців
- Пошук окремого гравця за кодом
- Інформацію про поточного автентифікованого гравця

### Інформація про гравця

```go
type Player struct {
    Code   string `json:"code"`
    Alias  string `json:"alias"`
    Avatar string `json:"avatar,omitempty"`
}
```

`code` - це унікальний ідентифікатор, отриманий з MongoDB ObjectID користувача.

## API точки доступу

### Типи ігор

| Метод | Точка доступу | Опис |
|-------|---------------|------|
| GET | `/api/game_types` | Список всіх типів ігор |
| POST | `/api/game_types` | Створити новий тип гри |
| GET | `/api/game_types/{code}` | Отримати конкретний тип гри |
| PUT | `/api/game_types/{code}` | Оновити тип гри |
| DELETE | `/api/game_types/{code}` | Видалити тип гри |

### Ігрові раунди

| Метод | Точка доступу | Опис |
|-------|---------------|------|
| GET | `/api/game_rounds` | Список всіх ігрових раундів |
| POST | `/api/game_rounds` | Розпочати новий ігровий раунд |
| GET | `/api/game_rounds/{code}` | Отримати конкретний ігровий раунд |
| PUT | `/api/game_rounds/{code}` | Оновити ігровий раунд |
| PUT | `/api/game_rounds/{code}/finalize` | Завершити ігровий раунд |
| PUT | `/api/game_rounds/{code}/players/{userId}/score` | Оновити очки гравця |

### Гравці

| Метод | Точка доступу | Опис |
|-------|---------------|------|
| GET | `/api/players` | Список всіх гравців |
| GET | `/api/players/{code}` | Отримати конкретного гравця |
| GET | `/api/players/i_am` | Отримати поточного автентифікованого гравця |

## Клієнтські компоненти

### Види та компоненти

#### ListGameTypes.vue

Адміністративний інтерфейс для керування типами ігор:
- Перегляд списку всіх типів ігор
- Створення нових типів ігор
- Редагування існуючих типів ігор
- Видалення типів ігор
- Налаштування міток та команд з вибором кольорів

#### GameroundsList.vue

Головний вид керування ігровими раундами:
- Відображення списку всіх ігрових раундів
- Показ статусу раунду (активний/завершений)
- Редагування активних раундів
- Завершення ігор
- Створення нових раундів

#### GameroundEdit.vue

Форма для створення та редагування ігрових раундів:
- Вибір типу гри
- Додавання/видалення гравців
- Призначення команд та міток
- Встановлення статусу модератора
- Оновлення очок гравців

#### FinalizeGameDialog.vue

Діалог для завершення ігрового раунду:
- Введення фінальних очок гравців
- Введення командних очок (якщо застосовно)
- Введення кооперативних очок (якщо застосовно)
- Розрахунок фінальних позицій

### Pinia сховища

#### game.ts

Керування станом для ігрових даних:
- `gameTypes`: Кешований список типів ігор
- `activeRounds`: Поточні активні ігрові раунди
- `loadGameTypes()`: Отримати типи ігор з API
- `addActiveRound()`: Створити новий ігровий раунд
- `updateRound()`: Оновити існуючий раунд
- `getGameTypeByCode`: Знайти тип гри за кодом

#### player.ts

Керування станом для даних гравців:
- `players`: Кешований список гравців
- `currentPlayer`: Поточний автентифікований гравець
- `allPlayers`: Отримати всіх гравців (з лінивим завантаженням)
- `getPlayerByCode`: Знайти гравця за кодом

## Приклади використання

### Створення типу гри

```typescript
const gameType: GameType = {
    name: "Catan",
    scoring_type: "classic",
    labels: [
        { name: "Перший гравець", color: "#FF0000", icon: "mdi-flag" }
    ],
    teams: [],
    min_players: 3,
    max_players: 4
};

await GameApi.createGameType(gameType);
```

### Запуск ігрового раунду

```typescript
const round: GameRound = {
    name: "П'ятнична гра в Catan",
    game_type: "Catan",
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1_code", is_moderator: false },
        { user_id: "player2_code", is_moderator: false },
        { user_id: "player3_code", is_moderator: false }
    ]
};

await GameApi.createGameRound(round);
```

### Завершення гри

```typescript
const finalizationData: FinalizeGameRoundRequest = {
    player_scores: {
        "player1_code": 10,
        "player2_code": 8,
        "player3_code": 7
    }
};

await GameApi.finalizeGameRound(roundCode, finalizationData);
```

### Приклад командної гри

```typescript
// Тип гри з командами
const mafiaGame: GameType = {
    name: "Мафія",
    scoring_type: "mafia",
    labels: [],
    teams: [
        { name: "Мирні", color: "#00FF00", icon: "mdi-account-group" },
        { name: "Мафія", color: "#FF0000", icon: "mdi-skull" }
    ],
    min_players: 6,
    max_players: 15
};

// Запуск гри з командами
const mafiaRound: GameRound = {
    name: "Епічна ніч мафії",
    game_type: "Мафія",
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1", is_moderator: true, team_name: "" },
        { user_id: "player2", is_moderator: false, team_name: "Мирні" },
        { user_id: "player3", is_moderator: false, team_name: "Мафія" },
        // ... більше гравців
    ]
};
```

## Колекції бази даних

### game_types

Зберігає визначення типів ігор:
- Індекс: `name` (унікальний)

### game_rounds

Зберігає дані ігрових раундів:
- Індекс: `game_type_id`
- Індекс: `start_time`
- Індекс: `players.player_id`

## Маршрути клієнтської частини

| Маршрут | Компонент | Опис |
|---------|-----------|------|
| `/ui/admin/game-types` | ListGameTypes.vue | Керування типами ігор |
| `/ui/game-rounds` | GameroundsList.vue | Список ігрових раундів |
| `/ui/game-rounds/new` | GameroundEdit.vue | Створити новий раунд |
| `/ui/game-rounds/:id` | GameroundEdit.vue | Редагувати існуючий раунд |
| `/ui/leagues` | ListLeagues.vue | Вид ліг (майбутнє) |

## Інтернаціоналізація

Клієнтська частина підтримує декілька мов використовуючи vue-i18n:

- `gameTypes.title`: Заголовок "Типи ігор"
- `gameTypes.edit`: Текст кнопки редагування
- `gameTypes.delete`: Текст кнопки видалення
- `gameTypes.create`: Текст кнопки створення
- `scoring.*`: Описи типів підрахунку очок

## Майбутні покращення

Потенційні покращення:
- Статистика та таблиці лідерів
- Система рейтингу гравців (ELO)
- Історія ігор та повтори
- Керування турнірами
- Таблиці ліг
- Система досягнень
- Рекомендації ігор
- Експорт/імпорт ігрових даних
