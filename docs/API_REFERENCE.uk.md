# Довідник API

## Точки доступу аутентифікації

### POST /api/auth/google/callback

Точка доступу callback Google OAuth. Створює нову сесію та повертає токени аутентифікації.

**Запит:**
- Метод: POST
- Параметри запиту: Параметри callback OAuth від Google
- Тіло: Відсутнє

**Відповідь:**
- Статус: 200 OK
- Тіло:
```json
{
  "code": "user_code",
  "external_ids": ["email@example.com"],
  "name": "Ім'я користувача",
  "avatar": "https://...",
  "alias": "user_alias",
  "names": ["Ім'я користувача"],
  "avatars": ["https://..."],
  "rotateToken": "base64_encoded_token"
}
```

**Встановлені Cookies:**
- `auth_token`: JWT токен дії (HTTP-only, Secure, SameSite=Strict)

**Примітки:**
- `rotateToken` повинен бути збережений в localStorage клієнтом
- Cookie `auth_token` встановлюється автоматично

---

### POST /api/auth/refresh

Оновлює токен дії. Опціонально обертає токен обертання якщо пройшло 12 годин з останнього обертання.

**Запит:**
- Метод: POST
- Заголовки:
  - `Authorization: Bearer <rotateToken>`
- Тіло: Відсутнє

**Відповідь:**
- Статус: 200 OK
- Тіло:
```json
{
  "rotateToken": "new_rotate_token"  // Тільки якщо токен був обернутий
}
```

**Встановлені Cookies:**
- `auth_token`: Новий JWT токен дії

**Помилки:**
- 401 Unauthorized: Недійсний або застарілий токен обертання

**Примітки:**
- Якщо `rotateToken` повернуто, клієнт повинен оновити localStorage
- Токен дії закінчується через 1 годину
- Токен обертання обертається кожні 12 годин

---

### POST /api/auth/logout

Деактивує поточну сесію та очищає cookies аутентифікації.

**Запит:**
- Метод: POST
- Заголовки (один з варіантів):
  - `Authorization: Bearer <rotateToken>`, АБО
  - `Content-Type: application/json`
- Тіло (опціонально):
```json
{
  "rotateToken": "token_to_invalidate"
}
```

**Відповідь:**
- Статус: 200 OK
- Тіло: Відсутнє

**Очищені Cookies:**
- `auth_token`: Cookie токена дії очищено

**Примітки:**
- Клієнт повинен видалити `rotateToken` з localStorage після виходу

---

## Користувацькі точки доступу

Всі користувацькі точки доступу потребують аутентифікації (cookie токена дії).

### GET /api/user

Отримує інформацію профілю поточного користувача.

**Запит:**
- Метод: GET
- Заголовки: Відсутні (використовує cookie)
- Cookies: `auth_token` (обов'язковий)

**Відповідь:**
- Статус: 200 OK
- Тіло:
```json
{
  "code": "user_code",
  "external_ids": ["email@example.com"],
  "name": "Ім'я користувача",
  "avatar": "https://...",
  "alias": "user_alias",
  "names": ["Ім'я користувача"],
  "avatars": ["https://..."]
}
```

**Помилки:**
- 401 Unauthorized: Недійсний або застарілий токен

---

### GET /api/user/sessions

Отримує всі активні сесії поточного користувача.

**Запит:**
- Метод: GET
- Параметри запиту:
  - `current` (опціональний): Токен обертання для позначення як поточна сесія
- Cookies: `auth_token` (обов'язковий)

**Відповідь:**
- Статус: 200 OK
- Тіло:
```json
[
  {
    "rotate_token": "token_string",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z",
    "last_rotation_at": "2024-01-01T12:00:00Z",
    "expires_at": "2024-01-31T00:00:00Z",
    "is_current": true,
    "geo_info": {
      "country": "United States",
      "country_code": "US",
      "region": "NY",
      "region_name": "New York",
      "city": "New York",
      "timezone": "America/New_York",
      "isp": "ISP Name",
      "ip": "192.168.1.1"
    }
  }
]
```

**Помилки:**
- 401 Unauthorized: Недійсний або застарілий токен

**Примітки:**
- `geo_info` може бути null якщо пошук геолокації не вдався
- `is_current` є true якщо параметр `current` запиту відповідає токену обертання сесії

---

### POST /api/user/alias/exist

Перевіряє чи доступний псевдонім.

**Запит:**
- Метод: POST
- Параметри запиту:
  - `alias`: Псевдонім для перевірки
- Cookies: `auth_token` (обов'язковий)

**Відповідь:**
- Статус: 200 OK
- Тіло:
```json
{
  "isUnique": true
}
```

---

### PUT /api/user/update

Оновлює профіль поточного користувача.

**Запит:**
- Метод: PUT
- Заголовки:
  - `Content-Type: application/json`
- Cookies: `auth_token` (обов'язковий)
- Тіло:
```json
{
  "name": "Нове ім'я",
  "avatar": "https://...",
  "alias": "new_alias"
}
```

**Відповідь:**
- Статус: 200 OK
- Тіло: Відсутнє

**Помилки:**
- 400 Bad Request: Недійсне тіло запиту
- 401 Unauthorized: Недійсний або застарілий токен
- 409 Conflict: Псевдонім вже зайнятий

---

## Адміністративні точки доступу

Всі адміністративні точки доступу потребують аутентифікації та прав супер-адміністратора.

### GET /api/admin/diagnostics

Отримує діагностичну інформацію системи (тільки для адмінів).

**Запит:**
- Метод: GET
- Cookies: `auth_token` (обов'язковий)
- Заголовки: Відсутні

**Відповідь:**
- Статус: 200 OK
- Тіло:
```json
{
  "server_info": {
    "host_url": "http://localhost:8080",
    "trusted_origins": ["http://localhost:3000"]
  },
  "request_info": {
    "ip_address": "192.168.1.1",
    "base_url": "http://localhost:8080",
    "user_agent": "Mozilla/5.0...",
    "origin": "http://localhost:3000",
    "is_trusted": true,
    "geo_info": {
      "country": "United States",
      "country_code": "US",
      "region": "NY",
      "region_name": "New York",
      "city": "New York",
      "timezone": "America/New_York",
      "isp": "ISP Name",
      "ip": "192.168.1.1"
    }
  }
}
```

**Помилки:**
- 401 Unauthorized: Не автентифікований
- 403 Forbidden: Не є супер-адміністратором

**Примітки:**
- `geo_info` може бути null якщо пошук геолокації не вдався
- `trusted_origins` походить зі змінної середовища `TRUSTED_ORIGINS`

---

### PUT /api/admin/user/create

Створює нового користувача (тільки для адмінів).

**Запит:**
- Метод: PUT
- Заголовки:
  - `Content-Type: application/json`
- Cookies: `auth_token` (обов'язковий)
- Тіло:
```json
{
  "external_ids": ["email@example.com"]
}
```

**Відповідь:**
- Статус: 201 Created
- Тіло: "User created successfully"

**Помилки:**
- 400 Bad Request: Відсутні або недійсні external_ids
- 401 Unauthorized: Не автентифікований
- 403 Forbidden: Не є супер-адміністратором
- 409 Conflict: Користувач вже існує

---

## Відповіді про помилки

Всі точки доступу можуть повертати наступні відповіді про помилки:

### 400 Bad Request
Недійсний формат запиту або відсутні обов'язкові параметри.

### 401 Unauthorized
Потрібна аутентифікація або токен недійсний/застарілий.

### 403 Forbidden
Автентифікований але не має необхідних дозволів (адміністративні точки доступу).

### 404 Not Found
Ресурс не знайдено.

### 409 Conflict
Конфлікт ресурсу (наприклад, псевдонім вже зайнятий).

### 500 Internal Server Error
Помилка сервера. Перевірте логи сервера для деталей.

## Аутентифікація

Більшість точок доступу потребують аутентифікації через cookie `auth_token` (JWT токен дії).

**Потік оновлення токена:**
1. Клієнт отримує 401 Unauthorized
2. Клієнт викликає `/api/auth/refresh` з `rotateToken` з localStorage
3. Сервер повертає новий cookie `actionToken`
4. Клієнт повторює оригінальний запит

**Обертання токена:**
- Токени обертання автоматично обертаються кожні 12 годин
- При обертанні новий `rotateToken` повертається у відповіді на оновлення
- Клієнт повинен оновити localStorage новим токеном

## Обмеження швидкості

- API геолокації (ipapi.co): Перевірте документацію ipapi.co для обмежень швидкості
- Вбудованих обмежень швидкості на точках доступу додатку немає (може бути додано через middleware)

## CORS

- Налаштуйте довірені джерела через змінну середовища `TRUSTED_ORIGINS`
- Формат: Список URL через кому (наприклад, `http://localhost:3000,https://example.com`)
- Використовується `RequestService.IsTrustedOrigin()` для валідації
