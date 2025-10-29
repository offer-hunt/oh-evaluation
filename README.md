# oh-evaluation

## Содержание

* [Требования](#требования)
* [Быстрый старт](#быстрый-старт)
* [Переменные окружения](#переменные-окружения)
* [Локальный запуск](#локальный-запуск)
* [Тесты](#тесты)
* [Миграции](#миграции)
* [API](#api)
* [Аутентификация и права](#аутентификация-и-права)
* [CI/CD](#cicd)
* [Docker-образ](#docker-образ)
* [Архитектура/директории](#архитектурадиректории)
* [Troubleshooting](#troubleshooting)

---

## Требования

* Go **1.24.1**
* Docker / Docker Desktop (для локального Testcontainers и docker build)
* Make (для удобного запуска миграций — опционально)

---

## Быстрый старт

```bash
# 1) Поднять локальную БД (postgres:16)
docker-compose up -d postgres

# 2) Прогнать миграции (опционально; можно и тестами применить)
make migrate-up

# 3) Запустить сервис
go run ./cmd/evaluation-service
# сервис слушает :8080
```

Проверка публичного эндпоинта:

```bash
curl -s http://localhost:8080/api/ping
# {"message":"pong"}
```

Проверка защищённого эндпоинта (нужен валидный JWT с scope `evaluation.read`):

```bash
curl -sH "Authorization: Bearer <token>" http://localhost:8080/api/secure/ping
# {"message":"pong-secure"}
```

---

## Переменные окружения

Можно задать через `.env` (в проект уже добавлен `godotenv`).

| Переменная         | По умолчанию                                  | Описание              |
| ------------------ | --------------------------------------------- | --------------------- |
| `AUTH_ISSUER`      | `http://localhost:8999`                       | `iss` ожидаемый в JWT |
| `AUTH_JWKS_URL`    | `http://localhost:8999/.well-known/jwks.json` | JWKS endpoint         |
| `AUTH_AUDIENCE`    | `evaluation-service`                          | `aud` ожидаемый в JWT |
| `HTTP_SERVER_PORT` | `8080`                                        | Порт сервера          |
| `DB_HOST`          | `localhost`                                   | Хост PostgreSQL       |
| `DB_PORT`          | `5432`                                        | Порт PostgreSQL       |
| `DB_USER`          | `evaluation_user`                             | Пользователь          |
| `DB_PASS`          | `evaluation_password`                         | Пароль                |
| `DB_NAME`          | `evaluation_db`                               | Имя БД                |
| `DB_SCHEMA`        | `evaluation`                                  | Схема сервиса         |

Пример `.env`:

```env
AUTH_ISSUER=http://localhost:8999
AUTH_JWKS_URL=http://localhost:8999/.well-known/jwks.json
AUTH_AUDIENCE=evaluation-service

HTTP_SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=evaluation_user
DB_PASS=evaluation_password
DB_NAME=evaluation_db
DB_SCHEMA=evaluation
```

---

## Локальный запуск

### Вариант A — только сервис (без контейнеров)

1. Подними PostgreSQL:

```bash
docker-compose up -d postgres
```

2. Прогони миграции:

```bash
make migrate-up
```

3. Запусти сервис:

```bash
go run ./cmd/evaluation-service
```

### Вариант B — тестовый запуск (Testcontainers сам поднимает БД)

```bash
go test ./... -v
```

> Если Docker не запущен локально — DB smoke-тест будет **пропущен**.
> В CI этот тест **обязателен** и упадёт, если docker недоступен.

---

## Тесты

Запуск всех тестов:

```bash
go test ./... -v
```

Что проверяется:

* **Security-тесты** (`internal/middleware/auth_test.go`):

    * 200 при валидном токене и `scope=evaluation.read`
    * 401 без токена
    * 401 при неверной подписи
    * 403 при отсутствии нужного `scope`
* **DB-smoke** (`db_smoke_test.go`):

    * Поднимается `postgres:16-alpine` в контейнере
    * Применяются миграции
    * Проверяется наличие таблицы `evaluation.__migration_probe`
    * Инсерт строки и возврат `id`

---

## Миграции

Миграции лежат в `db/migrations`. Используется `golang-migrate`.

* `000000_restrict_public_schema.*` — ограничение прав на `public`, доступ только к своей схеме.
* `000001_initial_schema.*` — создание схемы `evaluation`, выдача прав пользователю.
* `000002_add_probe_table.*` — тех.таблица `evaluation.__migration_probe`.

Команды:

```bash
make migrate-up
make migrate-down
```

---

## API

* `GET /api/ping` — публично, `200 {"message":"pong"}`.
* `GET /api/secure/ping` — требует валидный JWT **и** `scope=evaluation.read`, `200 {"message":"pong-secure"}`.

Пример запросов:

```bash
# публичный
curl -i http://localhost:8080/api/ping

# защищённый
curl -i -H "Authorization: Bearer <jwt>" http://localhost:8080/api/secure/ping
```

---

## Аутентификация и права

* JWT валидируется по JWKS (`AUTH_JWKS_URL`), проверяются `iss` и `aud`.
* Требуемый скоуп для защищённой ручки: **`evaluation.read`** (claim `scp` — массив строк).
* В коде: `internal/middleware/auth.go`.

**Как получить токен для проверки?**

В тестах мы генерим ключ/токен через `testutil` (см. `testutil/jwks.go`, `testutil/jwt.go`).
Для ручного теста можно:

* поднять свой mock JWKS endpoint;
* или использовать внешний IdP, настроив `AUTH_ISSUER/AUTH_JWKS_URL/AUTH_AUDIENCE`.

---

## CI/CD

GitHub Actions (`.github/workflows/ci.yml`):

1. golangci-lint
2. `go build`
3. `go test` (юнит + интеграционные; в CI DB-smoke обязателен)
4. Docker build & push в Docker Hub (на ветке `main`)

---

## Docker-образ

Имена тегов:

* `offerhunt/evaluation-service:<git-sha7>`
* `offerhunt/evaluation-service:latest`

Dockerfile — многослойная сборка: Go builder → `alpine:latest`.

---

## Архитектура/директории

```
.
├─ cmd/evaluation-service/    # main
├─ internal/
│  ├─ app/                    # запуск http-сервера, роутинг
│  ├─ config/                 # конфиг (envconfig)
│  ├─ handler/                # HTTP handlers
│  └─ middleware/             # аутентификация + проверка scope
├─ db/
│  └─ migrations/             # SQL миграции (evaluation schema)
├─ testutil/                  # генерация RSA/JWKS/JWT для тестов
├─ docker-compose.yml         # локальная БД postgres:16
├─ Makefile                   # миграции
├─ Dockerfile                 # сборка сервиса
└─ .github/workflows/ci.yml   # CI pipeline
```

**Bounded context:**

* Схема: `evaluation`
* Пользователь: `evaluation_user`
* Доступ к `public` ограничен (REVOKE), права только на свою схему.

---

## Troubleshooting

**`Cannot connect to the Docker daemon at unix:///var/run/docker.sock`**

* Запусти Docker Desktop / `colima start` / `systemctl start docker` (в зависимости от ОС).
* Локально тест DB-smoke будет пропущен; в CI — обязателен.

**`missing go.sum entry ...`**

* Выполни:

  ```bash
  go mod tidy
  go mod download
  ```

**`401 Invalid token`**

* Проверь `iss/aud`, корректность `kid` и доступность `AUTH_JWKS_URL`.
* Для защищённого эндпоинта нужен `scp` со значением `evaluation.read`.
