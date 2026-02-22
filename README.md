# Praxis Backend MVP

Backend for mood checkins and macro logging using:
- Go + Fiber REST API
- Postgres (Supabase-compatible schema)
- Docker-first local workflow

## Layout
- `app/main.go`: API entrypoint
- `app/api`: HTTP router + handlers (split by root path)
- `app/lib`: business logic (models, validation, store)
- `app/db`: migrations and SQL query files

## Prerequisites
- Docker
- Docker Compose

## Environment
Copy `.env.example` values into your runtime environment:
- `HTTP_ADDR` (default `:8080`)
- `DATABASE_URL`
- `DEV_USER_ID` (UUID for single-user dev mode)

Default dev UUID:
`00000000-0000-0000-0000-000000000001`

## Local Run (Docker)
1. Start database:
`make docker-up`
2. Apply migrations:
`make migrate`
3. Run API container from source:
`make run`

API health:
`GET http://localhost:8080/v1/health`

## Docker Compose App + DB
Run both services:
`docker compose up --build`

## Web Client (React)
A mobile-first React web client lives under `/website` and is intentionally decoupled from the Go backend.

Run it separately (example):
`cd website && python3 -m http.server 4173`

By default the client calls `http://localhost:8080/v1`. You can override API host by setting `localStorage.praxis_api_base` in the browser console (for example, `localStorage.praxis_api_base = "https://api.example.com"`).

## SQLC
Generate typed query package:
`make sqlc`

## Tests
Run tests via Dockerized Go toolchain:
`make test`

## API Endpoints
- `POST /v1/mood-checkins`
- `GET /v1/mood-checkins?from&to&mood_type&limit&cursor`
- `PATCH /v1/mood-checkins/{id}`
- `DELETE /v1/mood-checkins/{id}`
- `POST /v1/nutrition-entries`
- `GET /v1/nutrition-entries?from&to&meal_tag&limit&cursor`
- `PATCH /v1/nutrition-entries/{id}`
- `DELETE /v1/nutrition-entries/{id}`
- `GET /v1/daily-summaries?from&to` (`YYYY-MM-DD`)
- `GET /v1/trends?window=7d|30d`
- `GET /v1/health`
