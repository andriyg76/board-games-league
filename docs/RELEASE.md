# Процес релізу

*[English version](RELEASE.en.md)*

## Огляд

Процес релізу автоматизований через GitHub Actions workflow. Він дозволяє створити новий реліз з автоматичним bump версії, білдом Docker образу, створенням тегу, GitHub Release та автоматичним деплоєм на production сервер.

## Запуск релізу

### Через GitHub UI

1. Перейдіть на сторінку Actions: https://github.com/andriyg76/board-games-league/actions/workflows/release.yml
2. Натисніть "Run workflow"
3. Виберіть гілку (`main` або `release/v*`)
4. Виберіть тип bump версії:
   - **major** - збільшує major версію (1.0.0 → 2.0.0)
   - **minor** - збільшує minor версію (1.0.0 → 1.1.0)
   - **patch** - збільшує patch версію (1.0.0 → 1.0.1)
5. Натисніть "Run workflow"

### Вимоги до гілки

Workflow можна запускати тільки з:
- Гілки `main`
- Гілок формату `release/v*` (наприклад, `release/v1.0`)

Якщо спробувати запустити з іншої гілки, workflow завершиться з помилкою.

## Що відбувається під час релізу

1. **Валідація гілки** - перевірка що workflow запущено з правильної гілки
2. **Отримання останньої версії** - знаходить останній тег формату `v*.*.*` в історії поточної гілки
3. **Розрахунок нової версії** - застосовує обраний bump тип до останньої версії
4. **Docker build/push** - білдить та пушить Docker образ з тегом нової версії (наприклад, `v1.2.3`)
5. **Створення тегу** - створює git тег з новою версією та пушить його в репозиторій
6. **Генерація changelog** - генерує список змін між попереднім тегом та поточним HEAD
7. **Створення GitHub Release** - створює GitHub Release з changelog
8. **Deployment на production** - автоматично:
   - Оновлює `BACKEND_VERSION` в `.env` файлі на production сервері
   - Виконує `docker compose pull` для завантаження нового образу
   - Виконує `docker compose up -d` для перезапуску контейнерів
9. **Перевірка здоров'я** - перевіряє що контейнер backend працює коректно

## Налаштування середовища

### Необхідні GitHub Secrets

Для роботи workflow потрібно налаштувати наступні secrets в GitHub:

#### Docker Hub
- `DOCKER_USERNAME` - ім'я користувача Docker Hub
- `DOCKER_TOKEN` - access token Docker Hub

#### Production сервер (SSH)
- `SSH_HOST` - hostname або IP адреса production сервера
- `SSH_USER` - ім'я користувача для SSH підключення
- `SSH_PRIVATE_KEY` - приватний SSH ключ для доступу до сервера

### Як налаштувати Secrets

1. Перейдіть в Settings репозиторію
2. Виберіть "Secrets and variables" → "Actions"
3. Натисніть "New repository secret"
4. Додайте кожен secret з назвою та значенням

### Підготовка production сервера

На production сервері має бути:

1. **Docker та Docker Compose** встановлені та налаштовані
2. **Директорія `~/board-games-league`** з файлами:
   - `.env` - файл з `BACKEND_VERSION` (буде автоматично оновлюватися)
   - `.backend.env` - конфігурація backend (створюється вручну або через інші workflows)
   - `.mongo.env` - конфігурація MongoDB (створюється вручну або через інші workflows)
   - `docker-compose.yaml` - файл docker-compose з сервісами backend та mongo

3. **SSH доступ** налаштований з GitHub Actions runner

Приклад `docker-compose.yaml`:
```yaml
services:
  backend:
    image: {docker-registry}/{image-name}:${BACKEND_VERSION}
    ports:
      - "20032:8080"
    env_file:
      - .backend.env
    depends_on:
      - mongo

  mongo:
    image: mongodb/mongodb-community-server:8.2-ubuntu2204
    env_file:
      - .mongo.env
    volumes:
      - mongo_data:/data/db
      - mongo_configdb:/data/configdb

volumes:
  mongo_data: {}
  mongo_configdb: {}
```

## Версіонування

### Формат версій

Версії використовують формат [Semantic Versioning](https://semver.org/):
- `vMAJOR.MINOR.PATCH` (наприклад, `v1.2.3`)

### Визначення останньої версії

Workflow знаходить останній тег формату `v*.*.*`, який є в історії поточної гілки. Це означає, що якщо ви працюєте на гілці `release/v1.0`, workflow буде шукати теги тільки в історії цієї гілки, а не всі теги репозиторію.

### BACKEND_VERSION

Після релізу `BACKEND_VERSION` в `.env` файлі на production сервері оновлюється на повну версію з префіксом `v`:
```
BACKEND_VERSION=v1.2.3
```

## Перевірка здоров'я після деплою

Після деплою workflow автоматично перевіряє стан контейнера backend:

- Контейнер має бути в стані `running`
- Контейнер не повинен бути в станах: `exited`, `dead`, або `restarting`

Якщо контейнер не працює коректно, workflow завершиться з помилкою, що запобігає залишенню production в неробочому стані.

## Усунення проблем

### Workflow не запускається

- Перевірте що ви запускаєте з гілки `main` або `release/v*`
- Перевірте що всі необхідні secrets налаштовані

### Docker build/push не працює

- Перевірте що `DOCKER_USERNAME` та `DOCKER_TOKEN` встановлені правильно
- Перевірте що у користувача є права на push в Docker registry

### Deployment не працює

- Перевірте що SSH secrets налаштовані правильно
- Перевірте що SSH ключ має доступ до production сервера
- Перевірте що на сервері встановлені Docker та Docker Compose
- Перевірте що директорія `~/board-games-league` існує та має правильні права

### Контейнер не запускається після деплою

- Перевірте логи: `docker compose logs backend` на production сервері
- Перевірте що образ з правильною версією існує в Docker registry
- Перевірте що `BACKEND_VERSION` в `.env` оновлено правильно
- Перевірте конфігурацію в `.backend.env` та `.mongo.env`

### Тег не створюється

- Перевірте що workflow має permission `contents: write`
- Перевірте що гілка має права на push тегів

## Посилання

- [GitHub Actions Workflow](https://github.com/andriyg76/board-games-league/actions/workflows/release.yml)
- [Semantic Versioning](https://semver.org/)

