# Developer Guide — Putra Sunda Trans API

> Ringkasan singkat untuk developer yang ingin memahami dan berkontribusi pada project ini.
> Untuk panduan lengkap, lihat [DEVELOPER_GUIDE_DETAIL.md](DEVELOPER_GUIDE_DETAIL.md).

---

## Tech Stack

| Layer | Library |
|---|---|
| HTTP Framework | [Gin](https://github.com/gin-gonic/gin) v1.9.1 |
| ORM | [GORM](https://gorm.io) + driver PostgreSQL |
| Auth | JWT (`golang-jwt/jwt/v5`) + Argon2id (`alexedwards/argon2id`) |
| Logging | Zap (`go.uber.org/zap`) |
| Validation | `go-playground/validator/v10` |
| Migration | [golang-migrate](https://github.com/golang-migrate/migrate) (CLI) |
| Hot Reload | [Air](https://github.com/air-verse/air) |

---

## Arsitektur

Project menggunakan **Clean Architecture** dengan dependency injection manual.

```
cmd/api/main.go        → Entry point, inject semua dependency
internal/{module}/
  ├── domain/          → Entity + Interface (tidak ada external dep)
  ├── dto/             → Request & Response struct
  ├── repository/      → Implementasi data access (GORM)
  ├── usecase/         → Business logic
  ├── handler/         → HTTP handler (Gin)
  └── routes.go        → Registrasi route module
internal/router/       → Orchestrator semua module
pkg/                   → Shared utilities (response, jwt, middleware, logger, helper)
config/                → Load .env ke Config struct
database/              → Koneksi PostgreSQL via GORM
migrations/            → File SQL migrasi (golang-migrate)
```

Aturan dependency: **Handler → Usecase → Repository → Domain**. Layer bawah tidak boleh tahu layer atas.

---

## Quick Start

```bash
# 1. Clone & install dependencies
git clone https://github.com/afifudin23/putra-sunda-trans-api.git
cd putra-sunda-trans-api
make setup

# 2. Siapkan env
cp .env.example .env
# Edit .env sesuai konfigurasi lokal

# 3. Jalankan migrasi
make migrate-up

# 4. Jalankan server (hot reload)
make air
# atau tanpa hot reload
make run
```

---

## Environment Variables

```env
APP_NAME="Putra Sunda Trans API"
APP_PORT=5000
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="your_password"
DB_NAME="putra_sunda_db"
ACCESS_KEY="your_jwt_secret"
```

---

## Makefile Commands

| Command | Deskripsi |
|---|---|
| `make setup` | Install dependencies, migrate CLI, dan air |
| `make run` | Jalankan server (go run) |
| `make air` | Jalankan server dengan hot reload |
| `make build` | Build binary ke `bin/` |
| `make migrate-up` | Jalankan semua migrasi |
| `make migrate-down` | Rollback semua migrasi |
| `make migrate-create name=xxx` | Buat file migrasi baru |
| `make migrate-version` | Cek versi migrasi saat ini |
| `make migrate-force version=N` | Force migrasi ke versi tertentu |

---

## Menambahkan Modul Baru

Struktur yang harus dibuat untuk setiap modul baru (contoh: `order`):

```
internal/order/
├── domain/order.go       → Order struct + OrderRepository interface + error sentinels
├── dto/request.go        → CreateOrderRequest, dll
├── dto/response.go       → OrderResponse, dll
├── repository/
│   ├── model.go          → OrderModel (GORM model)
│   └── repository.go     → Implementasi OrderRepository
├── usecase/usecase.go    → Business logic
├── handler/handler.go    → HTTP handlers
└── routes.go             → Daftarkan routes ke gin.RouterGroup
```

Setelah itu, daftarkan modul di `internal/router/router.go`.

---

## Endpoint Saat Ini

| Method | Path | Deskripsi | Auth |
|---|---|---|---|
| GET | `/ping` | Health check | - |
| POST | `/api/v1/auth/register` | Registrasi user baru | - |
| POST | `/api/v1/auth/login` | Login, mendapatkan JWT | - |

Header auth untuk endpoint terproteksi: `Authorization: Bearer <token>`

---

## Format Response

Semua response menggunakan struktur standar. Lihat [API_RESPONSE.md](API_RESPONSE.md) untuk detail lengkap.

```json
// Success
{ "status": true, "data": {}, "message": "..." }

// Error
{ "status": false, "message": "...", "error": {}, "error_code": "..." }
```

---

## Konvensi Kode

- **Error handling**: Gunakan error sentinels di `domain/` (contoh: `ErrEmailAlreadyExists`)
- **Logging**: Gunakan `pkg/logger`, bukan `fmt.Println`
- **Response**: Selalu gunakan `pkg/response` untuk semua HTTP response
- **Validation**: Gunakan binding tag di DTO request struct
- **Context**: Selalu teruskan `ctx context.Context` ke repository method
