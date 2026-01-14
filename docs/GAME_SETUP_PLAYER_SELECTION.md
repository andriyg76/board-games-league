# План імплементації: Налаштування ігор з вибором гравців

## 📋 Загальний огляд

### Функціональність
Wizard для створення ігрового раунду з інтелектуальним вибором гравців на основі історії спільних ігор.

### Ключові особливості
- Двопанельний інтерфейс вибору гравців
- Кеш "останніх гравців" на сервері (в `LeagueMembership`)
- Автозаповнення слотів на основі історії
- Створення віртуальних гравців "на льоту"
- Двоетапний wizard (вибір → налаштування)

---

## 🎯 Вимоги

| # | Питання | Рішення |
|---|---------|---------|
| 1 | Джерело гравців | Члени ліги + віртуальні гравці |
| 2 | "Останні гравці" | Сортування за датою останньої спільної гри, топ-10 |
| 3a | Початкові слоти | Кількість = розмір кешу (min ≤ slots ≤ max) |
| 3b | Динамічні слоти | Так, можна додавати/видаляти |
| 3c | Автозаповнення | До max_players з кешу |
| 4/10 | Поточний гравець | Якщо є membership → додати; якщо ні → запитати створити |
| 10a | Видалити себе | Так, можна |
| 10b | Хто може створювати | Члени ліги + суперадмін |
| 5 | Кешування | На сервері, оновлювати по закінченню раунду |
| 6 | UI | Двопанельний інтерфейс |
| 7 | Віртуальні гравці | Можна створити з UI (аліас → унікальність → інвайт) |
| 8 | Де кеш | На сервері (в LeagueMembership) |
| 9 | API | `GET /api/leagues/{code}/suggested-players` |
| 11 | Структура кешу | В `LeagueMembership.recent_co_players` |
| 12 | Інвайт віртуального | Toast + копіювання в буфер + сповіщення в лізі |
| 13 | Flow | Двоетапний wizard |
| 14 | Мінімум гравців | Попередження, але дозволити продовжити |

---

## 🏗️ Архітектура

### Зміни в моделях (Backend)

#### 1. `LeagueMembership` - додати поле кешу

```go
type RecentCoPlayer struct {
    MembershipID primitive.ObjectID `bson:"membership_id"`
    LastPlayedAt time.Time          `bson:"last_played_at"`
}

type LeagueMembership struct {
    // ... existing fields ...
    RecentCoPlayers []RecentCoPlayer `bson:"recent_co_players,omitempty"` // max 10 items
}
```

### Нові API endpoints

| Method | Endpoint | Опис |
|--------|----------|------|
| GET | `/api/leagues/{code}/suggested-players` | Рекомендовані гравці для раунду |
| POST | `/api/leagues/{code}/virtual-players` | Створити віртуального гравця |

### Структура відповіді `suggested-players`

```json
{
  "current_player": {
    "membership_id": "...",
    "alias": "...",
    "avatar": "...",
    "is_member": true
  },
  "recent_players": [
    {
      "membership_id": "...",
      "alias": "...",
      "avatar": "...",
      "last_played_at": "2026-01-10T...",
      "is_virtual": false
    }
  ],
  "other_players": [
    {
      "membership_id": "...",
      "alias": "...",
      "avatar": "...",
      "is_virtual": true
    }
  ]
}
```

---

## 📝 Детальний план задач

### Фаза 1: Backend - Модель та репозиторій (2-3 дні)

#### 1.1 Оновити модель `LeagueMembership`
- [ ] Додати структуру `RecentCoPlayer`
- [ ] Додати поле `RecentCoPlayers []RecentCoPlayer` (max 10)
- [ ] Міграція існуючих даних (поле буде пустим)

#### 1.2 Оновити `LeagueMembershipRepository`
- [ ] Метод `UpdateRecentCoPlayers(ctx, membershipID, coPlayers []RecentCoPlayer)`
- [ ] Метод `GetMembershipWithRecentPlayers(ctx, leagueID, userID)`

#### 1.3 Сервіс оновлення кешу
- [ ] Створити функцію `UpdateCoPlayersCache(ctx, gameRound)` 
- [ ] Викликати при фіналізації раунду (`finalizeGameRound`)
- [ ] Логіка: для кожного гравця раунду оновити його `recent_co_players`:
  - Додати всіх інших гравців раунду з `last_played_at = now`
  - Якщо гравець вже є в списку - оновити `last_played_at`
  - Тримати max 10 записів, сортованих за `last_played_at` desc

---

### Фаза 2: Backend - API endpoints (2-3 дні)

#### 2.1 Endpoint `GET /api/leagues/{code}/suggested-players`
- [ ] Перевірка прав: член ліги АБО суперадмін
- [ ] Отримати поточного користувача та його membership (якщо є)
- [ ] Отримати `recent_co_players` з membership поточного користувача
- [ ] Отримати всіх інших членів ліги
- [ ] Сформувати відповідь з трьома секціями:
  - `current_player` (null якщо немає membership)
  - `recent_players` (до 10, сортовані за `last_played_at`)
  - `other_players` (решта членів ліги)

#### 2.2 Endpoint `POST /api/leagues/{code}/virtual-players`
- [ ] Request body: `{ "alias": "string" }`
- [ ] Перевірка унікальності alias в межах ліги
- [ ] Створити `LeagueMembership` зі статусом `virtual`
- [ ] Створити invitation для цього віртуального гравця
- [ ] Оновити кеш поточного користувача (додати віртуального на перше місце)
- [ ] Response: `{ "membership": {...}, "invitation_link": "..." }`

#### 2.3 Оновити права доступу
- [ ] Створення раунду: член ліги АБО суперадмін
- [ ] Додати middleware перевірку

---

### Фаза 3: Frontend - API клієнт (1 день)

#### 3.1 Оновити `LeagueApi.ts`
- [ ] `getSuggestedPlayers(leagueCode: string): Promise<SuggestedPlayersResponse>`
- [ ] `createVirtualPlayer(leagueCode: string, alias: string): Promise<VirtualPlayerResponse>`

#### 3.2 Типи TypeScript
```typescript
interface SuggestedPlayer {
  membership_id: string;
  alias: string;
  avatar?: string;
  last_played_at?: string;
  is_virtual: boolean;
}

interface SuggestedPlayersResponse {
  current_player: SuggestedPlayer | null;
  recent_players: SuggestedPlayer[];
  other_players: SuggestedPlayer[];
}

interface VirtualPlayerResponse {
  membership: SuggestedPlayer;
  invitation_link: string;
}
```

---

### Фаза 4: Frontend - UI компоненти (3-4 дні)

#### 4.1 Компонент `PlayerSelector.vue` (двопанельний)
- [ ] Ліва панель: "Доступні гравці"
  - Секція "Останні" (recent_players)
  - Секція "Інші" (other_players)
  - Пошук/фільтр
  - Кнопка "Додати" біля кожного гравця
- [ ] Права панель: "Обрані гравці"
  - Список обраних
  - Кнопка "Видалити" біля кожного
  - Поточний гравець позначений (якщо є)
- [ ] Індикатор `min_players` / `max_players`
- [ ] Попередження якщо менше `min_players`

#### 4.2 Компонент `CreateVirtualPlayerDialog.vue`
- [ ] Поле вводу alias
- [ ] Перевірка унікальності (debounced)
- [ ] Кнопка "Створити"
- [ ] При успіху: toast + копіювання посилання в буфер
- [ ] Автоматично додати до списку обраних

#### 4.3 Оновити `GameroundEdit.vue` → Wizard
- [ ] **Крок 1**: Вибір типу гри + `PlayerSelector`
  - Автозаповнення при виборі типу гри
  - Кнопка "Далі"
- [ ] **Крок 2**: Налаштування раунду
  - Назва раунду
  - Призначення команд (якщо є)
  - Призначення модератора (якщо потрібно)
  - Кнопка "Зберегти"

#### 4.4 Логіка автозаповнення
```typescript
function autoFillPlayers(gameType: GameType, suggestedPlayers: SuggestedPlayersResponse) {
  const selected: SuggestedPlayer[] = [];
  const maxSlots = gameType.max_players;
  
  // 1. Додати поточного гравця (якщо є membership)
  if (suggestedPlayers.current_player) {
    selected.push(suggestedPlayers.current_player);
  }
  
  // 2. Заповнити з recent_players
  for (const player of suggestedPlayers.recent_players) {
    if (selected.length >= maxSlots) break;
    if (!selected.find(p => p.membership_id === player.membership_id)) {
      selected.push(player);
    }
  }
  
  // 3. Заповнити з other_players якщо ще є місце
  for (const player of suggestedPlayers.other_players) {
    if (selected.length >= maxSlots) break;
    if (!selected.find(p => p.membership_id === player.membership_id)) {
      selected.push(player);
    }
  }
  
  return selected;
}
```

---

### Фаза 5: Інтеграція та тестування (2 дні)

#### 5.1 Backend тести
- [ ] Тест оновлення кешу при фіналізації
- [ ] Тест endpoint `suggested-players`
- [ ] Тест створення віртуального гравця
- [ ] Тест перевірки прав доступу

#### 5.2 Frontend тести
- [ ] Тест автозаповнення
- [ ] Тест wizard flow
- [ ] Тест створення віртуального гравця

#### 5.3 E2E сценарії
- [ ] Новий користувач без історії
- [ ] Користувач з історією ігор
- [ ] Суперадмін без membership
- [ ] Створення віртуального гравця

---

## 🔄 Діаграма потоку даних

```
┌─────────────────────────────────────────────────────────────┐
│                    Створення раунду                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  GET /api/leagues/{code}/suggested-players                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ 1. Перевірка прав (член ліги / суперадмін)           │   │
│  │ 2. Отримати membership поточного користувача          │   │
│  │ 3. Отримати recent_co_players з кешу                  │   │
│  │ 4. Отримати всіх членів ліги                          │   │
│  │ 5. Сформувати відповідь                               │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    КРОК 1: Вибір                             │
│  ┌─────────────────────┐    ┌─────────────────────┐         │
│  │   Доступні гравці   │    │   Обрані гравці     │         │
│  │   ───────────────   │    │   ─────────────     │         │
│  │   [Останні]         │ ←→ │   • Гравець 1  [x]  │         │
│  │   • Player A    [+] │    │   • Гравець 2  [x]  │         │
│  │   • Player B    [+] │    │   • ...             │         │
│  │   [Інші]            │    │                     │         │
│  │   • Player C    [+] │    │   [+ Віртуальний]   │         │
│  └─────────────────────┘    └─────────────────────┘         │
│                                                              │
│  Тип гри: [Вибрати ▼]     min: 4 / обрано: 5 / max: 6       │
│                                          [Далі →]            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    КРОК 2: Налаштування                      │
│                                                              │
│  Назва раунду: [___________________]                         │
│                                                              │
│  Гравці:                                                     │
│  ┌────────────────┬──────────────┬────────────────┐         │
│  │ Гравець        │ Команда      │ Модератор      │         │
│  ├────────────────┼──────────────┼────────────────┤         │
│  │ Player A       │ [Команда 1▼] │ [ ]            │         │
│  │ Player B       │ [Команда 2▼] │ [✓]            │         │
│  │ ...            │ ...          │ ...            │         │
│  └────────────────┴──────────────┴────────────────┘         │
│                                                              │
│                              [← Назад]  [Зберегти]           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Фіналізація раунду (пізніше)                    │
│  ───────────────────────────────────────                     │
│  При фіналізації → оновити кеш recent_co_players            │
│  для всіх учасників раунду                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 📅 Орієнтовний таймлайн

| Фаза | Опис | Тривалість |
|------|------|------------|
| 1 | Backend: Модель та репозиторій | 2-3 дні |
| 2 | Backend: API endpoints | 2-3 дні |
| 3 | Frontend: API клієнт | 1 день |
| 4 | Frontend: UI компоненти | 3-4 дні |
| 5 | Інтеграція та тестування | 2 дні |
| **Всього** | | **10-13 днів** |

---

## ⚠️ Важливі примітки

1. **Міграція**: Поле `recent_co_players` буде пустим для існуючих користувачів. Кеш заповниться автоматично при фіналізації нових раундів.

2. **Суперадмін**: Може створювати раунди без membership, але не буде автоматично доданий до списку гравців.

3. **Віртуальні гравці**: При створенні одразу додаються в кеш `recent_co_players` поточного користувача (замінюють останнього якщо кеш повний).

4. **Інвайт**: Використовує існуючу функціональність інвайтів, toast-повідомлення для всіх в лізі.

---

## 🔗 Пов'язані файли

### Backend
- `backend/models/league_membership.go` - модель membership
- `backend/repositories/league_membership_repository.go` - репозиторій
- `backend/gameapi/handlers.go` - API handlers
- `backend/gameapi/gameround.go` - логіка раундів

### Frontend
- `frontend/src/gametypes/GameroundEdit.vue` - поточний компонент редагування
- `frontend/src/api/LeagueApi.ts` - API клієнт ліг
- `frontend/src/api/GameApi.ts` - API клієнт ігор
- `frontend/src/store/player.ts` - store гравців
