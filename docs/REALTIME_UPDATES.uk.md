# Real-time оновлення гри (SSE)

## Що це?

Коли кілька гравців одночасно дивляться на одну гру Wizard, вони всі бачать оновлення в реальному часі. Якщо один гравець вводить ставки - всі інші бачать це миттєво без перезавантаження сторінки.

## Як це працює?

### Підключення

1. Гравець відкриває сторінку гри (`/ui/wizard/:code`)
2. Frontend автоматично відкриває SSE з'єднання до сервера
3. Сервер реєструє цього гравця як "підписника" на цю гру
4. В header гри з'являється індикатор "Live (N)" де N - кількість підключених

### Оновлення

1. Гравець 1 вводить ставки і натискає "Submit"
2. Backend зберігає дані в БД
3. Backend надсилає подію `bids_submitted` всім підписникам цієї гри
4. Frontend всіх підписників отримує нові дані і оновлює UI
5. Гравці 2, 3, ... бачать оновлену інформацію

### Відключення

1. Гравець закриває сторінку або переходить на іншу
2. Frontend відписується від подій
3. Сервер видаляє гравця зі списку підписників

## Типи подій

| Подія | Що сталося |
|-------|------------|
| `connected` | Підключились до гри |
| `heartbeat` | Сервер перевіряє з'єднання (кожні 30 сек) |
| `bids_submitted` | Хтось ввів ставки |
| `results_submitted` | Хтось ввів результати |
| `round_completed` | Раунд завершено |
| `round_restarted` | Раунд перезапущено |
| `round_edited` | Раунд відредаговано |
| `next_round` | Перехід до наступного раунду |
| `prev_round` | Перехід до попереднього раунду |
| `game_finalized` | Гра завершена |

## UI індикатор

В header гри відображається:
- **Live (3)** - зелений тег, означає що ви підключені і всього 3 гравці дивляться на гру
- Якщо тега немає - з'єднання втрачено, система спробує перепідключитися

## Як тестувати?

### Базовий тест

1. Відкрийте гру в двох різних вкладках браузера
2. Переконайтеся, що обидві вкладки показують "Live (2)"
3. В одній вкладці введіть ставки
4. Перевірте, що друга вкладка оновилася автоматично

### Тест reconnect

1. Відкрийте гру
2. Вимкніть мережу на кілька секунд
3. Увімкніть мережу
4. Переконайтеся, що з'єднання відновилося (знову з'явився "Live")

### Тест через curl

```bash
# Замініть AUTH_COOKIE на ваш cookie
curl -N \
  -H "Cookie: auth_token=YOUR_TOKEN" \
  "http://localhost:8080/api/leagues/LEAGUE_CODE/wizard/games/GAME_CODE/events"
```

Ви побачите потік подій:
```
event: connected
data: {"type":"connected","game_code":"abc123","timestamp":"...","data":{"client_id":"...","subscribers":1}}

event: heartbeat
data: {"type":"heartbeat","game_code":"abc123","timestamp":"..."}
```

## Технічні деталі

### Endpoint

```
GET /api/leagues/{leagueCode}/wizard/games/{gameCode}/events
```

### Headers відповіді

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

### Формат SSE

```
event: bids_submitted
data: {"type":"bids_submitted","game_code":"abc123",...}

```

(подвійний перенос рядка після data)

## Обмеження

1. **Один instance** - поточна реалізація працює в рамках одного сервера. При горизонтальному масштабуванні потрібен Redis Pub/Sub.

2. **Тільки Wizard** - SSE реалізовано тільки для гри Wizard. Інші ігри поки не підтримуються.

3. **Тільки перегляд** - SSE показує оновлення, але не блокує одночасне редагування. Якщо двоє спробують ввести ставки одночасно - переможе останній.

## Файли

### Backend
- `backend/services/game_event_hub.go` - центральний хаб для SSE
- `backend/wizardapi/sse.go` - SSE endpoint
- `backend/wizardapi/round.go` - broadcast після операцій

### Frontend
- `frontend/src/wizard/types.ts` - типи для SSE подій
- `frontend/src/api/WizardApi.ts` - `subscribeToEvents()`
- `frontend/src/store/wizard.ts` - обробка SSE в store
- `frontend/src/wizard/WizardGamePlay.vue` - UI з Live індикатором
