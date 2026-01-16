---
name: UI-Backend E2E Tests
overview: "Planning - draft"
todos:
  - id: requirements
    content: "Define priority UI->API flows and roles"
    status: pending
  - id: env
    content: "Provision isolated production-like environment"
    status: pending
  - id: auth
    content: "Decide auth strategy (real OAuth vs test login)"
    status: pending
  - id: tooling
    content: "Choose E2E framework and setup runner"
    status: pending
  - id: data
    content: "Define data seeding and cleanup strategy"
    status: pending
  - id: tests
    content: "Implement core E2E scenarios"
    status: pending
  - id: ci
    content: "Add CI pipeline for E2E"
    status: pending
  - id: docs
    content: "Document local and CI E2E usage"
    status: pending
---

# План: UI-backend E2E тести

## Мета

- Перевірити повний шлях UI -> API -> DB у production-like середовищі.
- Отримати стабільний набір регресійних тестів для критичних flows.

## Передумови

- Ізольований інстанс із окремою MongoDB базою.
- Усі обов'язкові змінні середовища:
  - `MONGODB_URI`
  - `GOOGLE_CLIENT_ID`
  - `GOOGLE_CLIENT_SECRET`
  - `JWT_SECRET`
  - `SESSION_SECRET`
  - `SUPERADMINS`
  - `IPINFO_TOKEN`
- Налаштовані `TRUSTED_ORIGINS` і `HOST_URL` якщо UI і API на різних доменах.
- HTTPS для коректної роботи `Secure` cookies.

## Рішення, які потрібно узгодити

1. **Стратегія логіну**
   - A: Реальний Google OAuth (потрібні тестові акаунти).
   - B: Тестовий login flow (тестовий endpoint або feature flag).
   - C: Сідинг сесій напряму в БД і установка cookies.
2. **Розміщення UI і API**
   - Один домен (спрощує cookies).
   - Різні домени (потрібні CORS і trusted origins).
3. **Дані**
   - Сідинг фікстур перед запуском.
   - Повна очистка після прогону.
4. **Фреймворк**
   - Playwright або Cypress.

## Обсяг першої хвилі (P0)

- Логін і відновлення сесії.
- Створення ліги, перегляд списку, перехід у лігу.
- Створення інвайту, прийняття інвайту, поява в списку учасників.
- Створення ігрового раунду і збереження результатів.
- Перевірка таблиці standings після завершення раунду.
- Вихід з акаунту і перевірка доступу до захищених сторінок.

## Дані і ізоляція

- Кожен прогін використовує унікальні назви та маркери.
- Опціонально: окрема база на кожен прогін.
- Відмова від використання реальних прод даних.

## Тестові артефакти

- Скриншоти і відео для падінь.
- Логи бекенду і UI у випадку помилок.

## Критерії готовності

- Документація запуску локально і в CI.
- Стабільність прогону без flaky помилок.
- Покриття P0 сценаріїв.
