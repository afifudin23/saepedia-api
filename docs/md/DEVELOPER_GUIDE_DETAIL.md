# Developer Guide Detail — SEAPEDIA API

> Panduan lengkap developer. Ringkasan cepat: [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md).
> Module path: `github.com/afifudin23/saepedia-api`

## Daftar Isi

1. [Overview Arsitektur](#1-overview-arsitektur)
2. [Setup Development](#2-setup-development)
3. [Struktur Direktori](#3-struktur-direktori)
4. [Penjelasan Layer](#4-penjelasan-layer)
5. [Shared Packages (pkg/)](#5-shared-packages-pkg)
6. [Transaksi Lintas-Repo (pkg/tx)](#6-transaksi-lintas-repo-pkgtx)
7. [Database & Migrasi](#7-database--migrasi)
8. [Konfigurasi & Versi](#8-konfigurasi--versi)
9. [Auth, Multi-Role & RBAC](#9-auth-multi-role--rbac)
10. [Menambah Modul Baru](#10-menambah-modul-baru)
11. [Error Handling](#11-error-handling)
12. [Logging](#12-logging)
13. [Swagger](#13-swagger)
14. [Troubleshooting](#14-troubleshooting)

---

## 1. Overview Arsitektur

**Clean Architecture** + dependency injection manual:

```
Handler (Gin)   → HTTP concern: bind/validate, panggil usecase, format response
Usecase         → business logic: aturan, orkestrasi, transaksi (pkg/tx)
Repository      → data access (GORM), TX-aware via pkg/tx — TANPA keputusan bisnis
Domain          → entity + interface + error sentinel (hanya std lib)
```

Dependency hanya mengalir ke dalam. Business logic ada di **usecase**, bukan repository.

---

## 2. Setup Development

Prasyarat: Go 1.25+, PostgreSQL 14+, golang-migrate CLI, Air, swag.

```bash
git clone https://github.com/afifudin23/saepedia-api.git
cd saepedia-api
make setup                                  # go mod tidy + install migrate, air, swag
cp .env.example .env                         # isi DB_PASSWORD, ACCESS_KEY (tanpa kutip)
psql -U postgres -c "CREATE DATABASE seapedia_db;"
make migrate-up
make seed
make run        # atau make air
```

Verifikasi: `curl http://localhost:5000/ping` →
`{"status":true,"data":{"app":"SEAPEDIA API","version":"0.1.0"},"message":"pong"}`.

---

## 3. Struktur Direktori

```
saepedia-api/
├── cmd/api/main.go            # entry point + anotasi umum Swagger
├── internal/
│   ├── auth/                  # register, login, logout (denylist), select-role, me, balance-summary
│   ├── user/                  # entity User + roles + repository
│   ├── review/                # public app reviews
│   ├── store/                 # seller store
│   ├── product/               # product CRUD + katalog publik (+ image_url)
│   ├── wallet/                # wallet, top-up, transaksi
│   ├── address/               # alamat pengiriman buyer
│   ├── cart/                  # cart (single-store rule)
│   ├── discount/              # voucher & promo
│   ├── order/                 # checkout, order, status history, seller process, report, overdue
│   ├── delivery/              # driver job: find/take/complete + earning
│   ├── admin/                 # monitoring + simulasi waktu + trigger overdue
│   ├── setting/               # app_settings (offset simulasi waktu)
│   └── router/router.go       # rakit semua module
├── pkg/                       # response, jwt, helper, middleware, pagination, logger, clock, tx
├── config/                    # config.go (env) + version.go (Version, APIVersion)
├── database/postgres.go
├── migrations/                # *.up.sql / *.down.sql
├── scripts/seed/              # seeder data demo (package main)
├── scripts/release/           # release manager
├── docs/swagger/              # OpenAPI hasil `make swag`
└── Makefile · .env.example · .air.toml · go.mod
```

---

## 4. Penjelasan Layer

### Domain (`internal/{module}/domain/`)
Entity + interface + error sentinel. Hanya std lib. ID memakai `string` (UUID dari Postgres).

```go
type User struct {
    ID        string
    Email     string
    Password  string
    IsAdmin   bool
    Roles     []string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserRepository interface {
    Create(ctx context.Context, user *User, roles []string) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    AddRole(ctx context.Context, userID, role string) error
    CountAll(ctx context.Context) (int64, error)
}

var (
    ErrUserNotFound = errors.New("user not found")
    ErrEmailExists  = errors.New("email already exists")
)
```

### Repository (`internal/{module}/repository/`)
`model.go` (GORM model + konversi ke domain) + `repository.go`. **TX-aware**: pakai
`tx.DB(ctx, r.db)` (bukan `r.db.WithContext(ctx)` langsung) agar ikut transaksi bila ada.

```go
func (r *productRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
    var m ProductModel
    err := tx.DB(ctx, r.db).Where("id = ?", id).First(&m).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, domain.ErrProductNotFound
    }
    if err != nil {
        return nil, err
    }
    return m.toDomain(), nil
}
```

### Usecase (`internal/{module}/usecase/`)
Business logic murni — tahu domain & pkg, tidak tahu HTTP/GORM. Di sinilah keputusan & validasi
bisnis + orkestrasi transaksi.

### Handler (`internal/{module}/handler/`)
HTTP concern + anotasi Swagger. **Selalu** pakai `pkg/response`.

```go
func (h *Handler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, err)
        return
    }
    res, err := h.uc.Login(c.Request.Context(), domain.LoginInput{Email: req.Email, Password: req.Password})
    if err != nil {
        if errors.Is(err, domain.ErrInvalidCredentials) {
            response.Unauthorized(c, err.Error())
            return
        }
        response.InternalServerError(c)
        return
    }
    response.Success(c, http.StatusOK, "login success", dto.ToAuthResponse(res))
}
```

---

## 5. Shared Packages (`pkg/`)

### `pkg/response`
```go
response.Success(c, http.StatusOK, "message", data)               // single object
response.List(c, page, perPage, total, "message", listData)       // list + pagination
response.BadRequest(c, "msg")           // 400
response.Unauthorized(c, "msg")         // 401
response.Forbidden(c, "msg")            // 403
response.NotFound(c, "msg")             // 404
response.Conflict(c, "msg")             // 409
response.UnprocessableEntity(c, "msg")  // 422
response.InternalServerError(c)         // 500 (tanpa bocorkan detail)
response.ValidationError(c, err)        // 400 field-level
```

### `pkg/jwt`
```go
// Token membawa uid + ACTIVE ROLE + jti (untuk denylist logout). TTL 24 jam.
token, err := jwt.Generate(userID, activeRole, accessKey)
claims, err := jwt.Verify(tokenString, accessKey)   // claims.UID, claims.ActiveRole, claims.ID(jti)
```

### `pkg/middleware`
```go
// Pasang Auth dulu, lalu RequireRole.
r := rg.Group("/seller/products")
r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleSeller))

// Di handler:
uid    := middleware.UserID(c)
active := middleware.ActiveRole(c)
```
`middleware.Auth()` membaca `ACCESS_KEY` dari config, memverifikasi token, mengecek denylist logout,
lalu menaruh `userID`, `activeRole`, `jti`, `tokenExp` ke context. `CORS()` juga di sini.

### `pkg/helper`
```go
hashed, err := helper.Hash(plaintext)   // argon2id
ok := helper.Verify(plaintext, hashed)  // bool
```

### `pkg/logger`
```go
logger.Info("server started", logger.String("port", "5000"))
logger.Error("failed", logger.Err(err))
```

### `pkg/pagination`
```go
p := pagination.Parse(c)   // p.Page, p.PerPage, p.Offset()
```

### `pkg/clock`
Virtual now untuk simulasi waktu (overdue). Semua logika waktu pakai `clock.Now()`, bukan `time.Now()`.

---

## 6. Transaksi Lintas-Repo (`pkg/tx`)

Agar business logic tetap di usecase tapi tetap atomik, usecase membungkus beberapa operasi repo
dalam satu transaksi:

```go
err := uc.txMgr.Do(ctx, func(ctx context.Context) error {
    if err := uc.productRepo.DecrementStock(ctx, productID, qty); err != nil { return err }
    w, err := uc.walletRepo.GetForUpdate(ctx, userID)
    if err != nil { return err }
    if w.Balance < total { return orderdomain.ErrInsufficientFunds }
    if err := uc.walletRepo.UpdateBalance(ctx, w.ID, w.Balance-total); err != nil { return err }
    return uc.orderRepo.Create(ctx, order)
})
```

Repo membaca handle DB via `tx.DB(ctx, r.db)` → otomatis pakai transaksi saat di dalam `Do`,
atau koneksi biasa di luar `Do`. Contoh nyata: checkout & overdue refund di modul `order`.

---

## 7. Database & Migrasi

Koneksi di `database/postgres.go` (`database.DB`). Pool: max open 25, idle 10, lifetime 1 jam.
Memakai **golang-migrate** (bukan AutoMigrate).

```bash
make migrate-create name=create_table_orders   # buat pasangan up/down
make migrate-up                                 # jalankan
make migrate-down n=2                           # mundur 2 langkah
make migrate-drop                               # hapus semua (ke nol)
make db-reset                                   # drop → up → seed
```

Setiap migrasi idempoten (`CREATE TABLE IF NOT EXISTS ...` / `DROP TABLE IF EXISTS ...`).

---

## 8. Konfigurasi & Versi

`config/config.go` — `LoadConfig()` baca `.env` (godotenv) ke `config.AppConfig`
(AppName, AppPort, DB*, AccessKey).

`config/version.go`:
```go
const Version = "0.1.0"            // versi rilis (semver) — di-bump make release, dipakai /ping & Swagger
const APIVersion = "v1"            // segmen path → /api/v1 & /docs/v1 (manual)
func APIBasePath() string { return "/api/" + APIVersion }
```

> `make release` hanya menyentuh `Version` (regex di-anchor ke `const Version`), `APIVersion` aman.

---

## 9. Auth, Multi-Role & RBAC

- Satu akun (email) bisa punya banyak role non-admin.
- Login mengembalikan token + `active_role` + `need_role_selection`. Bila >1 role → `active_role`
  kosong, wajib `POST /auth/select-role` (terbitkan token baru ber-active-role).
- Otorisasi server-side mengikuti **active role** di JWT (`middleware.RequireRole`), bukan klaim frontend.
- **Logout** memasukkan `jti` token ke tabel `revoked_tokens`; `middleware.Auth()` menolak token
  yang sudah di-denylist (401).
- **Ownership**: tiap resource difilter per pemilik (produk vs toko seller, alamat/order/wallet per
  user, job per driver).

Wiring auth module:
```go
authModule := auth.New(userRepo, walletRepo /* WalletReader */, revocationRepo, accessKey)
```

---

## 10. Menambah Modul Baru

Contoh modul `example`:

```
internal/example/
├── domain/example.go        # entity + ExampleRepository interface + error sentinel
├── dto/dto.go               # request + response + mapper
├── repository/repository.go # model GORM + impl (pakai tx.DB(ctx, r.db))
├── usecase/usecase.go       # business logic
├── handler/handler.go       # HTTP + anotasi swagger, pakai pkg/response
└── routes.go                # Module + RegisterRoutes
```

`routes.go`:
```go
type Module struct{ handler *handler.Handler }

func New(repo domain.ExampleRepository) *Module {
    return &Module{handler: handler.New(usecase.New(repo))}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
    r := rg.Group("/buyer/examples")
    r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleBuyer))
    {
        r.GET("", m.handler.List)
        r.POST("", m.handler.Create)
    }
}
```

Lalu di `internal/router/router.go`: buat repo, `exampleModule := example.New(repo)`,
`exampleModule.RegisterRoutes(api)`. Buat migrasi tabelnya, dan `make swag` untuk update docs.

---

## 11. Error Handling

```go
// usecase: kembalikan error sentinel domain
if !user.HasRole(role) { return nil, authdomain.ErrRoleNotOwned }

// handler: map ke HTTP
switch {
case errors.Is(err, domain.ErrRoleNotOwned):
    response.Forbidden(c, err.Error())
case errors.Is(err, userdomain.ErrUserNotFound):
    response.NotFound(c, err.Error())
default:
    response.InternalServerError(c)
}
```
Jangan bocorkan detail internal (SQL, stack trace) ke client.

---

## 12. Logging

Zap via `pkg/logger`. Info untuk lifecycle, Error selalu sertakan `logger.Err(err)`.
Field helper: `String`, `Int`, `Int64`, `Uint`, `Bool`, `Err`, `Any`.

---

## 13. Swagger

Anotasi di tiap handler (`// @Summary`, `@Tags`, `@Param`, `@Success`, `@Router`, `@Security BearerAuth`).
Generate: `make swag` → `docs/swagger/`. UI: `/docs/v1/index.html`. Versi `info.version` diisi dari
`config.Version` saat runtime (`cmd/api/main.go`).

---

## 14. Troubleshooting

**Server tak start** — cek `.env` lengkap, PostgreSQL jalan, DB `seapedia_db` ada, `make migrate-up` sudah dijalankan.

**`bind: address already in use` (port 5000)** — ada instance lama:
```bash
netstat -ano | grep ':5000' | grep LISTENING   # ambil PID kolom terakhir
taskkill //PID <PID> //F
```

**`Dirty database version N`**:
```bash
make migrate-version
make migrate-force version=<N-1>
make migrate-up
```

**JWT ditolak** — format `Authorization: Bearer <token>`, `ACCESS_KEY` sama dengan saat generate,
token belum expired (24 jam) & belum logout. Kalau butuh role aktif: `POST /auth/select-role` dulu.

**Build error setelah tambah modul** — `go build ./...` & `go mod tidy`.

---

_Dokumen ini sesuai project versi `0.1.0`._
