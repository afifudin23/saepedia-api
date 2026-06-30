# Developer Guide — SEAPEDIA API

> Ringkasan untuk developer. Detail lengkap: [DEVELOPER_GUIDE_DETAIL.md](DEVELOPER_GUIDE_DETAIL.md).
> Daftar endpoint: [API_ENDPOINTS.md](API_ENDPOINTS.md). Format response: [API_RESPONSE.md](API_RESPONSE.md).

Module path: `github.com/afifudin23/saepedia-api`

---

## Tech Stack

| Layer | Library |
|---|---|
| HTTP Framework | [Gin](https://github.com/gin-gonic/gin) v1.10 |
| ORM | [GORM](https://gorm.io) + driver PostgreSQL |
| Auth | JWT (`golang-jwt/jwt/v5`) + Argon2id (`alexedwards/argon2id`) |
| Logging | Zap (`go.uber.org/zap`) |
| Validation | `go-playground/validator/v10` |
| Migration | [golang-migrate](https://github.com/golang-migrate/migrate) (CLI) |
| API Docs | [swaggo](https://github.com/swaggo/swag) + gin-swagger |
| Hot Reload | [Air](https://github.com/air-verse/air) |

---

## Arsitektur

**Clean Architecture** dengan dependency injection manual di `internal/router/router.go`.

```
cmd/api/main.go        → Entry point (load config, connect DB, set swagger version, run)
internal/{module}/
  ├── domain/          → Entity + interface repository/usecase + error sentinel
  ├── dto/             → Request & response struct (+ mapper)
  ├── repository/      → Data access (GORM) — TX-aware via pkg/tx
  ├── usecase/         → Business logic (orkestrasi, aturan, transaksi)
  ├── handler/         → HTTP handler (Gin) + anotasi Swagger
  └── routes.go        → Registrasi route + middleware (Auth, RequireRole)
internal/router/       → Rakit semua module
pkg/                   → response, jwt, helper, middleware, pagination, logger, clock, tx
config/                → LoadConfig (.env) + version.go (Version, APIVersion)
database/              → Koneksi PostgreSQL (GORM)
migrations/            → SQL migrasi (golang-migrate)
scripts/seed/          → Seeder data demo (package main)
scripts/release/       → Release manager interaktif
docs/swagger/          → File OpenAPI hasil `make swag`
```

Aturan dependency: **Handler → Usecase → Repository → Domain**. Business logic ada di **usecase**;
repository hanya operasi data. Operasi lintas-repo yang harus atomik dibungkus `pkg/tx.Manager`.

### Modul

`auth`, `user`, `review`, `store`, `product`, `wallet`, `address`, `cart`, `discount`, `order`,
`delivery`, `admin`, `setting`.

---

## Quick Start

```bash
git clone https://github.com/afifudin23/saepedia-api.git
cd saepedia-api
make setup                 # go mod tidy + install migrate, air, swag

cp .env.example .env       # isi DB_PASSWORD & ACCESS_KEY
psql -U postgres -c "CREATE DATABASE seapedia_db;"

make migrate-up            # buat tabel
make seed                  # data demo
make run                   # atau: make air (hot reload)
```

Swagger UI: `http://localhost:5000/docs/v1/index.html`.

---

## Environment Variables

```env
APP_NAME=SEAPEDIA API
APP_PORT=5000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=seapedia_db
ACCESS_KEY=your_jwt_secret
```

> Nilai **tanpa tanda kutip** (Makefile membaca `.env` apa adanya). `DB_PASSWORD` & `ACCESS_KEY` wajib.

---

## Makefile Commands

| Command | Deskripsi |
|---|---|
| `make setup` | go mod tidy + install migrate, air, swag |
| `make run` / `make air` | Jalankan server (biasa / hot reload) |
| `make build` | Build binary ke `bin/` |
| `make swag` | Generate Swagger ke `docs/swagger` |
| `make seed` | Jalankan semua seeder |
| `make seed-one name=user` | Jalankan seeder satuan (`user`/`discount`/`review`/`address`/`order`) |
| `make release` | Release manager (bump version + tag + push) |
| `make migrate-up` | Jalankan semua migrasi |
| `make migrate-down [n=N]` | Mundur N langkah (default 1) |
| `make migrate-drop` | Hapus semua tabel (ke nol) |
| `make migrate-reset` | Drop → up (skema fresh) |
| `make db-reset` | Drop → up → seed (reset penuh + data demo) |
| `make migrate-create name=xxx` | Buat file migrasi baru |
| `make migrate-version` / `migrate-force version=N` | Cek / paksa versi migrasi |

---

## Konvensi Kode

- **Response**: selalu lewat `pkg/response` — `response.Success(c, http.StatusOK, "msg", data)` /
  `response.List(c, page, perPage, total, "msg", listData)` / `response.BadRequest`,
  `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `UnprocessableEntity`, `InternalServerError`,
  `ValidationError`. Jangan `c.JSON(...)` langsung.
- **Auth**: `middleware.Auth()` (validasi JWT) lalu `middleware.RequireRole("buyer"/"seller"/"driver"/"admin")`.
  Ambil context: `middleware.UserID(c)`, `middleware.ActiveRole(c)`.
- **Error**: error sentinel di `domain/` (mis. `ErrStoreNotFound`), bandingkan `errors.Is`, map ke
  HTTP di handler.
- **Transaksi**: bungkus dengan `txMgr.Do(ctx, func(ctx) error { ... })`; repo membaca handle via
  `tx.DB(ctx, r.db)` agar ikut transaksi yang sama.
- **Logging**: `pkg/logger` (zap), bukan `fmt.Println`.
- **Keamanan**: konten user-generated di-`html.EscapeString`; query selalu parameterized.

---

## Menambah Modul Baru

Buat `internal/<modul>/` dengan `domain/`, `dto/`, `repository/`, `usecase/`, `handler/`, `routes.go`,
lalu daftarkan di `internal/router/router.go`. Detail langkah: lihat [DEVELOPER_GUIDE_DETAIL.md](DEVELOPER_GUIDE_DETAIL.md) bagian "Menambah Modul Baru".
