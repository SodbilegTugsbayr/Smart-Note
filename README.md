# Smart Note

Full-stack Go + Nuxt app with a Go/Chi backend and a Nuxt 3 + Vuetify frontend.

## Project layout
- `backend/` — Go server (Chi router, Gorm/Postgres, session auth, OAuth2 helpers).
- `backend/confs/` — environment-specific YAML config (`dev.yaml`, `prod.yaml`).
- `frontend/` — Nuxt 3 SPA with Vuetify UI and composables.
- `frontend/env/` — environment files for Nuxt (`local.env`, `prod.env`).
- `host_conf/` — example host/nginx configs for deployment.
- `Makefile` — convenience targets for running and deploying.

## Prerequisites
- Go (1.20+ recommended)
- Node.js + Yarn
- Postgres running and reachable via the DSN in `backend/confs/dev.yaml`

## Setup
1) Backend deps: `cd backend && go mod download`
2) Frontend deps: `cd frontend && yarn install`
3) Configure:
   - `backend/confs/dev.yaml` — update `dsn`, `session_secret`, OAuth client IDs.
   - `frontend/env/local.env` — set `BASE_URL`, domains, etc.

## Run locally
- Start backend (defaults to `backend/confs/dev.yaml`):  
  ```bash
  make web
  # or with live reload if you have nodemon installed
  make web-dev
  ```
- Start frontend (uses `frontend/env/local.env`):  
  ```bash
  make ui
  ```
Backend listens on `:4000` (per `dev.yaml`); frontend on `:3000`.

## Deploy (reference)
- Build and push frontend static output: `make dep-ui`
- Build and push backend binary/config, then restart: `make dep-web`

## Notes
- The backend auto-migrates the `userman.User` schema on startup.
- Sessions are stored via `golangcollege/sessions`; adjust cookie options in `backend/cmd/web/app/app.go` if needed.
