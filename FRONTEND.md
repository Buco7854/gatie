# GATIE — Frontend Requirements

## Existing Backend API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check (postgres + valkey status) |
| `GET` | `/api/setup/status` | Returns `{ needs_setup: bool }` |
| `POST` | `/api/setup` | Create first admin account (username, password) → tokens |
| `POST` | `/api/auth/login` | Login (username, password) → tokens |
| `POST` | `/api/auth/refresh` | Refresh tokens (refresh_token) → new tokens |
| `POST` | `/api/auth/logout` | Revoke refresh token |

### Auth response shape

```json
{
  "access_token": "jwt...",
  "refresh_token": "opaque-string",
  "member_id": "uuid",
  "role": "ADMIN",
  "username": "alice"
}
```

Refresh token is also set as an `HttpOnly` cookie on `/auth` path.

## Pages to build

### 1. Setup (`/setup`)
- Check `GET /api/setup/status` on load
- If `needs_setup: false` → show message + redirect to login
- Otherwise → form: username, password, confirm password
- On success → store tokens, redirect to dashboard

### 2. Login (`/login`)
- Form: username, password
- On success → store tokens, redirect to dashboard
- Link to `/setup` for first-time users

### 3. Dashboard (`/`)
- Protected route (redirect to `/login` if not authenticated)
- Header with app name, theme toggle, logout
- Show current user info (username, role)
- Placeholder for future gate management

## Auth flow

- Store access token in memory (not localStorage)
- On page load, attempt silent refresh via `POST /api/auth/refresh`
- Attach `Authorization: Bearer <token>` to API requests
- On 401 → try refresh → if fails → redirect to login

## Stack (from CLAUDE.md)

- Vite + React 19 + TypeScript
- TanStack Router + TanStack Query
- Tailwind CSS
- i18next
- React Hook Form + Zod
