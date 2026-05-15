# dz-pixel-dashboard

Dashboard BFF (Backend-For-Frontend). REST API consumed by `dz-pixel-ui`.

## Endpoints

| Method | Path | Returns |
|---|---|---|
| GET | `/healthz` | Service + DB status |
| GET | `/api/stats` | Aggregated counts (placeholder) |
| GET | `/api/campaigns` | List of campaigns from Postgres |
| GET | `/api/whoami` | Current user (placeholder for JWT auth) |

## Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `PG_HOST` | `localhost` | Postgres host |
| `PG_PORT` | `5432` | Postgres port |
| `PG_USER` | `pixel_app` | DB user |
| `PG_PASSWORD` | — | DB password (from Secrets Manager) |
| `PG_DB` | `argus_admin` | Database name |
| `CORS_ORIGIN` | `*` | Allowed CORS origin (set to your UI domain in prod) |

## Auto-created schema

On startup, this service ensures the `campaigns` table exists and seeds one demo row.

## Where to extend

- Wire `/api/stats` to ClickHouse for real event counts
- Replace `/api/whoami` placeholder with proper JWT validation
- Add CRUD endpoints for campaigns, segments, audiences
