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
│   ├── league.go               │   ├── GameroundsList.vue
│   └── player.go               │   ├── GameRoundWizard.vue      # Wizard створення раунду
├── models/                      │   ├── GameroundEdit.vue        # Редагування завершених
│   ├── game_type.go            │   ├── FinalizeGameDialog.vue
│   └── game_round.go           │   ├── types.ts
└── repositories/                │   └── steps/                   # Кроки wizard
    ├── game_type_repository.go │       ├── Step1GameType.vue    # Вибір типу гри
    └── game_round_repository.go│       ├── Step2Players.vue     # Вибір гравців
                                │       ├── Step3Roles.vue       # Призначення ролей
                                │       └── Step4Scoring.vue     # Введення очок
                                ├── wizard/
                                │   └── WizardGameConfig.vue     # Налаштування Wizard гри
                                └── store/
                                    ├── game.ts
                                    └── player.ts
```

## Типи ігор

### Концепція

Тип гри визначає правила та механізм підрахунку очок для конкретної настільної гри. Кожен тип гри має:

- **Ключ (key)**: Унікальний ідентифікатор типу гри
- **Локалізовані назви**: Назви різними мовами (en, uk, et)
- **Тип підрахунку**: Як розраховуються очки та визначаються переможці
- **Ролі**: Визначення ролей/команд/кольорів для гравців
- **Обмеження гравців**: Мінімальна та максимальна кількість гравців
- **Вбудований (built_in)**: Чи є тип вбудованим у систему (захищений від видалення)

### Доступ

**Важливо:** Тільки супер-адміністратори можуть створювати, редагувати та видаляти типи ігор. Типи ігор є глобальними для всієї платформи.

### Вбудовані типи ігор

Система автоматично завантажує вбудовані типи ігор при старті:
- **Mafia** - командна гра з модератором
- **Wizard** - карткова гра
- **Catan** - класична гра з кольоровими фішками
- **Ticket to Ride** - залізнична гра
- **Carcassonne** - гра з мішками
- **Codenames** - командна словесна гра
- **Dixit** - асоціативна гра
- **Uno** - карткова гра
- **Wingspan** - гра про птахів
- **1000 (Тисяча)** - карткова гра

### Типи підрахунку очок

Система підтримує декілька механізмів підрахунку:

| Тип підрахунку | Опис | Приклад використання |
|----------------|------|----------------------|
| `classic` | Традиційний змагальний підрахунок - виграє найвища сума | Catan, Ticket to Ride, Wingspan |
| `cooperative` | Всі гравці виграють або програють разом | Pandemic, Forbidden Island |
| `cooperative_with_moderator` | Кооперативна гра з окремим модератором | Ігри з ведучим |
| `team_vs_team` | Команди змагаються одна з одною | Codenames |
| `mafia` | Команда проти команди з прихованими ролями та модератором | Мафія, Вовкулаки |
| `custom` | Без попередньо визначеної схеми - пряме введення очок | Будь-який власний підрахунок |

### Ролі

Ролі замінюють попередні "мітки" та "команди" і використовуються для:
- Кольорових фішок/фігур у грі (червоний, синій, зелений)
- Ігрових ролей (Мафія, Цивільний, Шериф)
- Команд (Червона команда, Синя команда)

Кожна роль має:

- **Ключ (key)**: Унікальний ідентифікатор ролі
- **Локалізовані назви**: Назви різними мовами
- **Колір**: Колір відображення (формат hex)
- **Іконка**: Назва іконки Material Design
- **Тип ролі**: Обмеження кількості гравців

### Типи ролей (RoleType)

| Тип ролі | Опис | Кількість гравців | Приклад |
|----------|------|-------------------|---------|
| `optional` | Необов'язкова роль | 0+ | Спостерігач |
| `optional_one` | Необов'язкова, максимум один | 0-1 | Кольорова фішка в Catan |
| `exactly_one` | Рівно один гравець | 1 | Шериф, Дон |
| `required` | Обов'язково хоча б один | 1+ | Мафія |
| `multiple` | Обов'язково кілька | 2+ | Цивільні, Команди |
| `moderator` | Модератор гри | 1 | Ведучий |

### Структура моделі

```go
type RoleType string

const (
    RoleTypeOptional    RoleType = "optional"      // 0+
    RoleTypeOptionalOne RoleType = "optional_one"  // 0-1
    RoleTypeExactlyOne  RoleType = "exactly_one"   // 1
    RoleTypeRequired    RoleType = "required"      // 1+
    RoleTypeMultiple    RoleType = "multiple"      // 2+
    RoleTypeModerator   RoleType = "moderator"     // 1 (модератор)
)

type Role struct {
    Key      string            `bson:"key"`       // унікальний ключ
    Names    map[string]string `bson:"names"`     // {"en": "Red", "uk": "Червоний"}
    Color    string            `bson:"color"`     // "#F44336"
    Icon     string            `bson:"icon"`      // "mdi-account"
    RoleType RoleType          `bson:"role_type"` // тип ролі
}

type GameType struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Key         string             `bson:"key"`          // унікальний ключ
    Names       map[string]string  `bson:"names"`        // локалізовані назви
    Icon        string             `bson:"icon"`         // іконка гри
    ScoringType ScoringType        `bson:"scoring_type"` // тип підрахунку
    Roles       []Role             `bson:"roles"`        // ролі гравців
    MinPlayers  int                `bson:"min_players"`  // мін. гравців
    MaxPlayers  int                `bson:"max_players"`  // макс. гравців
    BuiltIn     bool               `bson:"built_in"`     // вбудований тип
    Version     int64              `bson:"version"`      // версія
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
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
┌──────────────────┐     ┌──────────────┐     ┌───────────┐     ┌──────────┐
│ players_selected │────>│ in_progress  │────>│  scoring  │────>│completed │
│  (Гравці обрані) │     │  (Грається)  │     │  (Очки)   │     │          │
└──────────────────┘     └──────────────┘     └───────────┘     └──────────┘
        │                       │                   │                │
        │                       │                   │                │
   Крок 2:                  Крок 3:             Крок 4:          Фіналізація
   Вибір гравців           Ролі призначені     Введення очок     EndTime
```

#### Статуси раунду

| Статус | Опис | Wizard крок |
|--------|------|-------------|
| `players_selected` | Гравці обрані, очікує налаштування | Після кроку 2 |
| `in_progress` | Гра активна, ролі призначені | Крок 3 |
| `scoring` | Гра завершена, введення очок | Крок 4 |
| `completed` | Раунд завершено | Після фіналізації |

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
- Кнопка "Продовжити" для активних раундів
- Кнопка "Переглянути" для завершених раундів
- Створення нових раундів

#### GameRoundWizard.vue

Головний wizard для створення та продовження ігрових раундів:
- Багатокроковий інтерфейс (stepper)
- Динамічне визначення кроків в залежності від типу гри
- Збереження стану на сервері після кожного кроку
- Автоматичний редирект завершених раундів на сторінку редагування

#### Кроки Wizard (gametypes/steps/)

**Step1GameType.vue** - Вибір типу гри:
- Список доступних типів ігор
- Локалізовані назви та іконки
- Інформація про кількість гравців

**Step2Players.vue** - Вибір гравців:
- Інтеграція з PlayerSelector
- Підтримка модератора для відповідних ігор
- Валідація мінімальної/максимальної кількості

**Step3Roles.vue** - Призначення ролей (для стандартних ігор):
- Назва раунду
- Таблиця гравців з вибором ролей
- Чекбокс модератора
- Кнопки "Зберегти" та "Далі"

**Step4Scoring.vue** - Введення очок:
- Таблиця гравців з полями для очок та позицій
- Кнопки "Зберегти" та "Завершити гру"

#### WizardGameConfig.vue (wizard/)

Специфічні налаштування для гри Wizard:
- Обмеження ставок (bid restriction)
- Вибір першого дилера
- Підсумок гри (кількість раундів, гравці)
- Кнопка "Почати гру"

#### GameroundEdit.vue

Спрощена форма для редагування **завершених** ігрових раундів:
- Редагування назви раунду
- Перегляд та редагування очок гравців
- Редагування позицій

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
const gameType: Partial<GameType> = {
    key: "my_custom_game",
    names: {
        en: "My Custom Game",
        uk: "Моя власна гра"
    },
    icon: "mdi-dice-6",
    scoring_type: "classic",
    roles: [
        {
            key: "red",
            names: { en: "Red", uk: "Червоний" },
            color: "#F44336",
            icon: "",
            role_type: "optional_one"
        },
        {
            key: "blue",
            names: { en: "Blue", uk: "Синій" },
            color: "#2196F3",
            icon: "",
            role_type: "optional_one"
        }
    ],
    min_players: 2,
    max_players: 4
};

await GameApi.createGameType(gameType);
```

### Запуск ігрового раунду

```typescript
const round: GameRound = {
    name: "П'ятнична гра в Catan",
    game_type: "catan",  // використовуємо key типу гри
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1_code", is_moderator: false, team_name: "red" },
        { user_id: "player2_code", is_moderator: false, team_name: "blue" },
        { user_id: "player3_code", is_moderator: false, team_name: "white" }
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

### Приклад командної гри (Мафія)

```typescript
// Тип гри Мафія вже вбудований, але ось як він виглядає:
const mafiaGame: GameType = {
    key: "mafia",
    names: {
        en: "Mafia",
        uk: "Мафія",
        et: "Maffia"
    },
    icon: "mdi-account-group",
    scoring_type: "mafia",
    roles: [
        {
            key: "civilian",
            names: { en: "Civilian", uk: "Цивільний" },
            color: "#4CAF50",
            role_type: "multiple"
        },
        {
            key: "mafia",
            names: { en: "Mafia", uk: "Мафія" },
            color: "#F44336",
            role_type: "required"
        },
        {
            key: "sheriff",
            names: { en: "Sheriff", uk: "Шериф" },
            color: "#2196F3",
            role_type: "optional_one"
        },
        {
            key: "moderator",
            names: { en: "Moderator", uk: "Ведучий" },
            color: "#FF9800",
            role_type: "moderator"
        }
    ],
    min_players: 6,
    max_players: 20,
    built_in: true
};

// Запуск гри з ролями
const mafiaRound: GameRound = {
    name: "Епічна ніч мафії",
    game_type: "mafia",
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1", is_moderator: false, team_name: "moderator" },
        { user_id: "player2", is_moderator: false, team_name: "civilian" },
        { user_id: "player3", is_moderator: false, team_name: "mafia" },
        { user_id: "player4", is_moderator: false, team_name: "sheriff" },
        // ... більше гравців
    ]
};
```

### Приклад гри Codenames (команди)

```typescript
// Codenames з двома командами
const codenamesRound: GameRound = {
    name: "Битва розумів",
    game_type: "codenames",
    start_time: new Date().toISOString(),
    players: [
        { user_id: "player1", is_moderator: false, team_name: "red_team" },
        { user_id: "player2", is_moderator: false, team_name: "red_team" },
        { user_id: "player3", is_moderator: false, team_name: "blue_team" },
        { user_id: "player4", is_moderator: false, team_name: "blue_team" },
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
| `/ui/game-rounds/new` | GameRoundWizard.vue | Wizard створення нового раунду |
| `/ui/game-rounds/new?gameType=wizard` | GameRoundWizard.vue | Wizard з передвибраним типом Wizard |
| `/ui/game-rounds/:id` | GameRoundWizard.vue | Продовження активного раунду |
| `/ui/game-rounds/:id/edit` | GameroundEdit.vue | Редагування завершеного раунду |
| `/ui/wizard/:code` | WizardGamePlay.vue | Інтерфейс гри Wizard |
| `/ui/leagues` | ListLeagues.vue | Список ліг |

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
