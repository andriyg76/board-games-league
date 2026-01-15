# Real-time оновлення гри (SSE)

## Огляд

Система real-time оновлень дозволяє всім гравцям, які беруть участь у грі та знаходяться на сторінці гри, отримувати миттєві оновлення стану гри без необхідності перезавантажувати сторінку.

## Технологія

Використовується **Server-Sent Events (SSE)** - однонаправлений канал від сервера до клієнта через HTTP. SSE обрано через:
- Простоту реалізації порівняно з WebSocket
- Автоматичний reconnect на стороні браузера
- Підтримку cookies для автентифікації
- Достатність для нашого use case (сервер → клієнт)

## Архітектура

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Гравець 1     │     │   Гравець 2     │     │   Гравець 3     │
│  (редагує гру)  │     │  (спостерігає)  │     │  (спостерігає)  │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │ POST /bids            │ SSE /events           │ SSE /events
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Backend API                              │
│  ┌─────────────────┐    ┌─────────────────────────────────────┐ │
│  │  Wizard Handler │───▶│          GameEventHub               │ │
│  │  (submitBids)   │    │  ┌─────────────────────────────┐    │ │
│  └─────────────────┘    │  │  clients[gameCode][clientID] │    │ │
│                         │  │  - Гравець 2                 │    │ │
│                         │  │  - Гравець 3                 │    │ │
│                         │  └─────────────────────────────┘    │ │
│                         │            │                         │ │
│                         │   Broadcast("bids_submitted")        │ │
│                         │            │                         │ │
│                         │            ▼                         │ │
│                         │   ┌───────────────────────┐          │ │
│                         │   │ Push to all channels  │          │ │
│                         │   └───────────────────────┘          │ │
│                         └─────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Компоненти

### Backend

#### 1. GameEventHub (`backend/services/game_event_hub.go`)

Центральний хаб для керування SSE з'єднаннями.

```go
type GameEventHub interface {
    // Підписка клієнта на оновлення гри
    Subscribe(gameCode string, clientID string) *GameEventClient
    
    // Відписка клієнта
    Unsubscribe(client *GameEventClient)
    
    // Розсилка події всім підписникам гри
    Broadcast(gameCode string, eventType string, data interface{})
    
    // Кількість підписників гри
    GetSubscriberCount(gameCode string) int
}
```

**Структура даних:**
- `clients map[string]map[string]*GameEventClient` - gameCode → clientID → client
- Кожен клієнт має буферизований канал на 10 повідомлень

#### 2. SSE Endpoint (`backend/wizardapi/sse.go`)

```
GET /api/leagues/{leagueCode}/wizard/games/{gameCode}/events
```

- Перевіряє існування гри
- Встановлює SSE headers
- Генерує унікальний `clientID` (UUID)
- Надсилає `connected` подію з інформацією про кількість підписників
- Слухає канал клієнта та надсилає події
- Heartbeat кожні 30 секунд

#### 3. Broadcast в handlers (`backend/wizardapi/round.go`)

Після кожної успішної операції викликається:
```go
h.broadcastGameUpdate(game, "event_type")
```

### Frontend

#### 1. WizardApi (`frontend/src/api/WizardApi.ts`)

```typescript
subscribeToEvents(
    leagueCode: string,
    gameCode: string,
    onEvent: (event: GameEvent) => void,
    onError?: (error: Event) => void
): GameEventSubscription
```

Повертає об'єкт з методом `unsubscribe()`.

#### 2. Wizard Store (`frontend/src/store/wizard.ts`)

Нові поля стану:
- `eventSubscription` - активна підписка
- `isConnected` - статус з'єднання
- `subscriberCount` - кількість підписників

Нові actions:
- `subscribeToEvents()` - підписка з auto-reconnect
- `unsubscribeFromEvents()` - відписка
- `handleGameEvent()` - обробка подій

#### 3. WizardGamePlay.vue

- `onMounted`: підписка після завантаження гри
- `onUnmounted`: відписка при виході зі сторінки
- UI індикатор "Live (N)" з кількістю підписників

## Типи подій

| Подія | Опис | Коли надсилається |
|-------|------|-------------------|
| `connected` | Клієнт підключився | Одразу після підключення |
| `heartbeat` | Keep-alive | Кожні 30 секунд |
| `bids_submitted` | Ставки введені | `PUT /{code}/rounds/{round}/bids` |
| `results_submitted` | Результати введені | `PUT /{code}/rounds/{round}/results` |
| `round_completed` | Раунд завершено | `POST /{code}/rounds/{round}/complete` |
| `round_restarted` | Раунд перезапущено | `POST /{code}/rounds/{round}/restart` |
| `round_edited` | Раунд відредаговано | `PUT /{code}/rounds/{round}/edit` |
| `next_round` | Перехід до наступного раунду | `POST /{code}/next-round` |
| `prev_round` | Перехід до попереднього раунду | `POST /{code}/prev-round` |
| `game_finalized` | Гра завершена | `POST /{code}/finalize` |

## Формат події

```json
{
    "type": "bids_submitted",
    "game_code": "abc123",
    "timestamp": "2024-01-15T12:00:00Z",
    "data": {
        "code": "abc123",
        "game_round_code": "xyz789",
        "config": {...},
        "players": [...],
        "rounds": [...],
        "current_round": 3,
        "max_rounds": 12,
        "status": "IN_PROGRESS"
    }
}
```

## SSE Wire Format

```
event: bids_submitted
data: {"type":"bids_submitted","game_code":"abc123",...}

event: heartbeat
data: {"type":"heartbeat","game_code":"abc123","timestamp":"..."}
```

## Обробка помилок

### Втрата з'єднання

1. `EventSource.onerror` спрацьовує
2. Store встановлює `isConnected = false`
3. Через 5 секунд автоматична спроба reconnect
4. Якщо гра все ще відкрита - підписка відновлюється

### Повільний клієнт

- Канал буферизований (10 повідомлень)
- Якщо канал переповнений - повідомлення пропускається
- Логування warning

### Відключення клієнта

- `context.Done()` спрацьовує при закритті HTTP з'єднання
- Клієнт автоматично видаляється з hub
- Якщо це останній клієнт гри - очищається запис гри

## Безпека

- SSE endpoint за authenticated middleware
- Cookies передаються через `withCredentials: true`
- Перевірка існування гри перед підпискою
- Клієнт може підписатися тільки на гри в лігах, до яких має доступ

## Масштабування (TODO)

Поточна реалізація працює в рамках одного instance. Для горизонтального масштабування потрібно:

1. **Redis Pub/Sub** для синхронізації між instances
2. **Sticky sessions** або broadcast через Redis

## Тестування

### Manual testing

1. Відкрити гру в двох браузерах/вкладках
2. Перевірити індикатор "Live (2)"
3. Ввести ставки в одному браузері
4. Переконатися, що другий браузер оновився автоматично

### Curl test

```bash
curl -N -H "Cookie: auth_token=..." \
  "http://localhost:8080/api/leagues/abc/wizard/games/xyz/events"
```

## Діаграма послідовності

```
Гравець 1          Backend           GameEventHub        Гравець 2
    │                 │                    │                 │
    │                 │   Subscribe        │                 │
    │                 │◄───────────────────│◄────────────────│
    │                 │                    │   connected     │
    │                 │────────────────────│────────────────▶│
    │  POST /bids     │                    │                 │
    │────────────────▶│                    │                 │
    │                 │   Update DB        │                 │
    │                 │─────────┐          │                 │
    │                 │◄────────┘          │                 │
    │                 │   Broadcast        │                 │
    │                 │───────────────────▶│                 │
    │                 │                    │  bids_submitted │
    │                 │                    │────────────────▶│
    │◄────────────────│                    │                 │
    │   200 OK        │                    │                 │
```
