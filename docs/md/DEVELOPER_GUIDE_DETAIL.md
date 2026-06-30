# Developer Guide Detail — Putra Sunda Trans API

> Panduan lengkap untuk developer. Untuk ringkasan cepat lihat [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md).

---

## Daftar Isi

1. [Overview Arsitektur](#1-overview-arsitektur)
2. [Setup Development Environment](#2-setup-development-environment)
3. [Struktur Direktori Lengkap](#3-struktur-direktori-lengkap)
4. [Penjelasan Layer per Layer](#4-penjelasan-layer-per-layer)
5. [Shared Packages (pkg/)](#5-shared-packages-pkg)
6. [Database &amp; Migrasi](#6-database--migrasi)
7. [Konfigurasi](#7-konfigurasi)
8. [Request Lifecycle](#8-request-lifecycle)
9. [Panduan Menambah Modul Baru](#9-panduan-menambah-modul-baru)
10. [Konvensi &amp; Aturan Kode](#10-konvensi--aturan-kode)
11. [Error Handling](#11-error-handling)
12. [Authentication &amp; Authorization](#12-authentication--authorization)
13. [Logging](#13-logging)
14. [Troubleshooting Umum](#14-troubleshooting-umum)

---

## 1. Overview Arsitektur

Project ini mengimplementasikan **Clean Architecture** dengan prinsip utama:

- **Dependency Rule**: Dependency hanya boleh mengalir ke dalam (ke layer yang lebih dalam). Layer domain tidak boleh bergantung pada layer manapun.
- **Dependency Injection Manual**: Semua dependency di-inject di `cmd/api/main.go`, tidak ada global singleton tersembunyi.
- **Interface-driven**: Setiap modul berkomunikasi lewat interface yang didefinisikan di `domain/`, bukan lewat implementasi konkret.

```
┌──────────────────────────────────────────────────┐
│  Handler (Gin HTTP)                              │
│  Menerima HTTP request, validasi input,          │
│  panggil usecase, format response                │
├──────────────────────────────────────────────────┤
│  Usecase (Business Logic)                        │
│  Orkestrasi alur bisnis, tidak tahu HTTP         │
│  Bergantung pada Repository Interface            │
├──────────────────────────────────────────────────┤
│  Repository (Data Access)                        │
│  Query ke database via GORM                      │
│  Mengimplementasikan interface dari domain       │
├──────────────────────────────────────────────────┤
│  Domain (Entity + Interfaces)                    │
│  Pure Go struct, tidak ada external dependency   │
│  Sumber kebenaran (source of truth) bisnis       │
└──────────────────────────────────────────────────┘
```

### Alur Dependency Injection

```
main.go
  └─ database.Connect()          → *gorm.DB
  └─ router.Setup(db, accessKey)
       └─ userrepo.New(db)       → UserRepository (implementasi)
       └─ auth.New(userRepo, accessKey)
            └─ authUC            → AuthUsecase
            └─ authHandler       → AuthHandler
            └─ authHandler.RegisterRoutes(api)
```

---

## 2. Setup Development Environment

### Prasyarat

- Go 1.25+
- PostgreSQL 14+
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [Air](https://github.com/air-verse/air) (hot reload)

### Langkah Setup

```bash
# 1. Clone repository
git clone https://github.com/afifudin23/putra-sunda-trans-api.git
cd putra-sunda-trans-api

# 2. Install semua tools dan dependencies
make setup
# Equivalent dengan:
#   go mod tidy
#   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
#   go install github.com/air-verse/air@latest

# 3. Buat file .env dari template
cp .env.example .env

# 4. Edit .env sesuai konfigurasi lokal
#    Minimal yang harus diubah: DB_PASSWORD, ACCESS_KEY

# 5. Buat database PostgreSQL
psql -U postgres -c "CREATE DATABASE putra_sunda_db;"

# 6. Jalankan migrasi
make migrate-up

# 7. Jalankan server
make air   # dengan hot reload
make run   # tanpa hot reload
```

### Verifikasi Setup

```bash
# Server berhasil jalan jika muncul log seperti:
# {"level":"info","msg":"server started","port":"5000"}

# Test health check
curl http://localhost:5000/ping
# Response: {"message":"pong"}
```

---

## 3. Struktur Direktori Lengkap

```
putra-sunda-trans-api/
│
├── cmd/
│   └── api/
│       └── main.go                 # Entry point, wiring semua dependency
│
├── internal/                       # Kode internal, tidak bisa diimport dari luar module
│   ├── auth/
│   │   ├── domain/
│   │   │   └── auth.go            # AuthUsecase interface, AuthToken struct, error sentinels
│   │   ├── dto/
│   │   │   ├── request.go         # LoginRequest, RegisterRequest
│   │   │   └── response.go        # AuthResponse, UserData
│   │   ├── handler/
│   │   │   └── handler.go         # Handler struct, Register(), Login()
│   │   ├── usecase/
│   │   │   └── usecase.go         # Implementasi AuthUsecase
│   │   └── routes.go              # Registrasi route /auth/*
│   │
│   ├── user/
│   │   ├── domain/
│   │   │   └── user.go            # User entity, UserRepository interface
│   │   ├── dto/
│   │   │   ├── request.go         # (reserved)
│   │   │   └── response.go        # (reserved)
│   │   ├── handler/
│   │   │   └── handler.go         # (reserved)
│   │   ├── repository/
│   │   │   ├── model.go           # UserModel (GORM model dengan tag)
│   │   │   └── repository.go      # Implementasi UserRepository
│   │   ├── usecase/
│   │   │   └── usecase.go         # (reserved)
│   │   └── routes.go              # (reserved)
│   │
│   └── router/
│       └── router.go              # Setup Gin engine, mount semua module
│
├── pkg/                           # Shared utilities, bisa digunakan semua layer
│   ├── response/
│   │   ├── response.go            # Struct Response, error codes
│   │   ├── success.go             # Success(), List() helper
│   │   └── error.go               # Error(), ValidationError(), dll
│   ├── jwt/
│   │   └── jwt.go                 # Generate(), Verify()
│   ├── middleware/
│   │   └── auth.go                # AuthMiddleware (JWT validation)
│   ├── helper/
│   │   └── hashing.go             # Hash(), Verify() — Argon2id
│   └── logger/
│       └── logger.go              # Init(), Info(), Error(), Warn(), Debug()
│
├── database/
│   └── postgres.go                # Connect(), global DB *gorm.DB
│
├── migrations/
│   ├── YYYYMMDDHHMMSS_*.up.sql    # Migrasi maju
│   └── YYYYMMDDHHMMSS_*.down.sql  # Migrasi mundur (rollback)
│
├── config/
│   └── config.go                  # LoadConfig(), Config struct, AppConfig global
│
├── docs/
│   └── md/
│       ├── API_RESPONSE.md        # Format standar response API
│       ├── CHANGELOG_GUIDE.md     # Panduan penulisan changelog
│       ├── DEVELOPER_GUIDE.md     # Ringkasan developer guide (file ini)
│       └── DEVELOPER_GUIDE_DETAIL.md
│
├── .air.toml                      # Konfigurasi hot reload Air
├── .env.example                   # Template environment variables
├── CHANGELOG.md                   # History perubahan versi
├── Dockerfile                     # (belum diisi)
├── docker-compose.yml             # (belum diisi)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 4. Penjelasan Layer per Layer

### 4.1 Domain Layer (`internal/{module}/domain/`)

Layer paling dalam. Tidak boleh mengimport apapun selain standard library Go.

Berisi:

- **Entity struct** — representasi bisnis (bukan model database)
- **Repository Interface** — kontrak yang harus diimplementasikan oleh repository
- **Usecase Interface** — kontrak yang harus diimplementasikan oleh usecase
- **Error sentinels** — error yang bisa dibandingkan dengan `errors.Is()`

Contoh (`internal/user/domain/user.go`):

```go
type User struct {
    ID        uuid.UUID
    Email     string
    Password  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
)
```

### 4.2 Repository Layer (`internal/{module}/repository/`)

Implementasi konkret dari Repository Interface. Hanya layer ini yang boleh menggunakan GORM.

Berisi dua file:

- `model.go` — GORM model dengan struct tag (`gorm:"column:..."`, `gorm:"primaryKey"`, dll)
- `repository.go` — struct yang mengimplementasikan interface dari domain

```go
// model.go — GORM model (berbeda dari domain entity)
type UserModel struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (UserModel) TableName() string { return "users" }

// repository.go — Konversi antara domain entity dan GORM model
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var model UserModel
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    return toDomain(&model), nil
}
```

**Penting**: Repository harus mengkonversi GORM model ke domain entity sebelum return. Jangan ekspose model GORM ke luar repository.

### 4.3 Usecase Layer (`internal/{module}/usecase/`)

Business logic murni. Tidak boleh tahu tentang HTTP, GORM, atau detail infrastruktur.

```go
type authUsecase struct {
    userRepo userdomain.UserRepository
    accessKey string
}

func (u *authUsecase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
    // 1. Cari user
    user, err := u.userRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        if errors.Is(err, userdomain.ErrUserNotFound) {
            return nil, domain.ErrInvalidCredentials
        }
        return nil, err
    }
    // 2. Verifikasi password
    match, err := helper.Verify(req.Password, user.Password)
    if err != nil || !match {
        return nil, domain.ErrInvalidCredentials
    }
    // 3. Generate JWT
    token, err := jwt.Generate(user.ID.String(), u.accessKey)
    if err != nil {
        return nil, err
    }
    return &dto.AuthResponse{Token: token, User: toUserData(user)}, nil
}
```

### 4.4 Handler Layer (`internal/{module}/handler/`)

Bertanggung jawab untuk HTTP concern saja: parsing request, validasi input, format response.

```go
func (h *authHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, err)
        return
    }

    result, err := h.usecase.Login(c.Request.Context(), &req)
    if err != nil {
        if errors.Is(err, domain.ErrInvalidCredentials) {
            response.Unauthorized(c, "invalid credentials")
            return
        }
        response.InternalServerError(c, err)
        return
    }

    response.Success(c, result, "login success")
}
```

### 4.5 Router (`internal/router/router.go`)

Orchestrator yang meng-wire semua module dan mendaftarkan routes.

```go
func Setup(db *gorm.DB, accessKey string) *gin.Engine {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    userRepo := userrepo.New(db)
    authModule := auth.New(userRepo, accessKey)

    api := r.Group("/api/v1")
    authModule.RegisterRoutes(api)

    return r
}
```

---

## 5. Shared Packages (`pkg/`)

### 5.1 `pkg/response`

Semua HTTP response harus menggunakan package ini. Jangan pernah tulis `c.JSON(...)` langsung di handler.

```go
// Success — single object
response.Success(c, data, "message")
// → {"status": true, "data": {...}, "message": "..."}

// List — dengan pagination
response.List(c, listData, pagination, "message")
// → {"status": true, "list_data": [...], "message": "...", "pagination": {...}}

// Error responses
response.BadRequest(c, "pesan error")          // 400
response.Unauthorized(c, "pesan error")        // 401
response.Forbidden(c, "pesan error")           // 403
response.NotFound(c, "pesan error")            // 404
response.Conflict(c, "pesan error")            // 409
response.InternalServerError(c, err)           // 500
response.ValidationError(c, err)               // 400 dengan field-level errors
```

Lihat [API_RESPONSE.md](API_RESPONSE.md) untuk format lengkap setiap response.

### 5.2 `pkg/jwt`

```go
// Generate token (expired 24 jam)
token, err := jwt.Generate(userID, accessKey)

// Verify token, return claims
claims, err := jwt.Verify(tokenString, accessKey)
userID := claims.UID
```

### 5.3 `pkg/middleware`

```go
// Gunakan di routes yang memerlukan auth
protected := api.Group("/").Use(middleware.Auth(accessKey))
protected.GET("/profile", handler.GetProfile)

// Di handler, ambil userID dari context
userID := c.GetString("userID")
```

### 5.4 `pkg/helper`

```go
// Hash password (Argon2id)
hashed, err := helper.Hash(plaintext)

// Verifikasi password
match, err := helper.Verify(plaintext, hashed)
```

### 5.5 `pkg/logger`

```go
logger.Info("pesan", logger.String("key", "value"), logger.Int("count", 5))
logger.Error("error terjadi", logger.Err(err))
logger.Warn("peringatan")
logger.Debug("debug info")
```

---

## 6. Database & Migrasi

### Koneksi Database

Database dikoneksikan di `database/postgres.go` dan disimpan di variable global `database.DB`. Konfigurasi connection pool:

- Max open connections: 25
- Max idle connections: 10
- Connection max lifetime: 1 jam

### Manajemen Migrasi

Project menggunakan **golang-migrate** (bukan GORM AutoMigrate) agar migrasi lebih terkontrol.

```bash
# Buat file migrasi baru
make migrate-create name=create_table_orders
# Akan membuat: migrations/YYYYMMDDHHMMSS_create_table_orders.{up,down}.sql

# Jalankan semua migrasi yang belum dijalankan
make migrate-up

# Rollback satu langkah
make migrate-down

# Cek versi migrasi saat ini
make migrate-version

# Force ke versi tertentu (jika ada dirty state)
make migrate-force version=1
```

### Konvensi File Migrasi

Setiap pasang file migrasi harus **idempoten**:

```sql
-- Up: gunakan IF NOT EXISTS
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ...
);

-- Down: harus membatalkan persis apa yang dilakukan Up
DROP TABLE IF EXISTS orders;
```

---

## 7. Konfigurasi

Semua konfigurasi dibaca dari environment variables (via `.env` menggunakan `godotenv`).

Config struct di `config/config.go`:

```go
type Config struct {
    AppName   string // APP_NAME
    AppPort   string // APP_PORT (default: 5000)
    DBHost    string // DB_HOST
    DBPort    string // DB_PORT
    DBUser    string // DB_USER
    DBPassword string // DB_PASSWORD
    DBName    string // DB_NAME
    AccessKey string // ACCESS_KEY (JWT secret, wajib diisi)
}
```

`config.LoadConfig()` dipanggil pertama kali di `main.go`. Setelah itu, gunakan `config.AppConfig` dari mana saja (sudah di-pass sebagai parameter, bukan global langsung).

---

## 8. Request Lifecycle

Berikut alur lengkap dari satu HTTP request masuk hingga response keluar:

```
HTTP Request
    │
    ▼
Gin Router (internal/router/router.go)
    │  Route matching
    ▼
Middleware (pkg/middleware/auth.go)        ← Jika route protected
    │  Validasi JWT, inject userID ke context
    ▼
Handler (internal/{module}/handler/handler.go)
    │  1. Bind & validate request body → ShouldBindJSON
    │  2. Panggil usecase
    │  3. Map error ke HTTP response
    │  4. Return response via pkg/response
    ▼
Usecase (internal/{module}/usecase/usecase.go)
    │  1. Business logic
    │  2. Panggil repository
    │  3. Return domain object atau error
    ▼
Repository (internal/{module}/repository/repository.go)
    │  1. Query database via GORM
    │  2. Konversi model ke domain entity
    │  3. Return domain entity atau error
    ▼
Database (PostgreSQL)
```

---

## 9. Panduan Menambah Modul Baru

Contoh: menambahkan modul `order`.

### Step 1: Domain

Buat `internal/order/domain/order.go`:

```go
package domain

import (
    "context"
    "errors"
    "time"
    "github.com/google/uuid"
)

type Order struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Status    string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
}

type OrderUsecase interface {
    CreateOrder(ctx context.Context, req *dto.CreateOrderRequest, userID string) (*dto.OrderResponse, error)
}

var (
    ErrOrderNotFound = errors.New("order not found")
)
```

### Step 2: DTO

Buat `internal/order/dto/request.go` dan `response.go`:

```go
// request.go
type CreateOrderRequest struct {
    // Isi field sesuai kebutuhan
    Status string `json:"status" binding:"required,oneof=pending confirmed cancelled"`
}

// response.go
type OrderResponse struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Step 3: Repository

Buat `internal/order/repository/model.go`:

```go
type OrderModel struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    Status    string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (OrderModel) TableName() string { return "orders" }
```

Buat `internal/order/repository/repository.go`:

```go
type orderRepository struct {
    db *gorm.DB
}

func New(db *gorm.DB) domain.OrderRepository {
    return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
    model := fromDomain(order)
    return r.db.WithContext(ctx).Create(model).Error
}
```

### Step 4: Usecase

Buat `internal/order/usecase/usecase.go`:

```go
type orderUsecase struct {
    orderRepo domain.OrderRepository
}

func New(orderRepo domain.OrderRepository) domain.OrderUsecase {
    return &orderUsecase{orderRepo: orderRepo}
}

func (u *orderUsecase) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest, userID string) (*dto.OrderResponse, error) {
    // implementasi business logic
}
```

### Step 5: Handler

Buat `internal/order/handler/handler.go`:

```go
type OrderHandler struct {
    usecase domain.OrderUsecase
}

func New(usecase domain.OrderUsecase) *OrderHandler {
    return &OrderHandler{usecase: usecase}
}

func (h *OrderHandler) Create(c *gin.Context) {
    var req dto.CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, err)
        return
    }
    userID := c.GetString("userID")
    result, err := h.usecase.CreateOrder(c.Request.Context(), &req, userID)
    if err != nil {
        response.InternalServerError(c, err)
        return
    }
    response.Success(c, result, "order created")
}
```

### Step 6: Routes

Buat `internal/order/routes.go`:

```go
package order

import (
    "github.com/gin-gonic/gin"
    "github.com/afifudin23/putra-sunda-trans-api/internal/order/handler"
    "github.com/afifudin23/putra-sunda-trans-api/pkg/middleware"
)

type Module struct {
    handler   *handler.OrderHandler
    accessKey string
}

func New(handler *handler.OrderHandler, accessKey string) *Module {
    return &Module{handler: handler, accessKey: accessKey}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
    orders := rg.Group("/orders")
    protected := orders.Use(middleware.Auth(m.accessKey))
    {
        protected.POST("", m.handler.Create)
    }
}
```

### Step 7: Daftarkan di Router

Edit `internal/router/router.go`:

```go
import orderrepo "github.com/afifudin23/putra-sunda-trans-api/internal/order/repository"
import orderuc "github.com/afifudin23/putra-sunda-trans-api/internal/order/usecase"
import orderhandler "github.com/afifudin23/putra-sunda-trans-api/internal/order/handler"
import order "github.com/afifudin23/putra-sunda-trans-api/internal/order"

// Di dalam Setup():
orderRepo := orderrepo.New(db)
orderUC := orderuc.New(orderRepo)
orderH := orderhandler.New(orderUC)
orderModule := order.New(orderH, accessKey)
orderModule.RegisterRoutes(api)
```

### Step 8: Buat Migrasi

```bash
make migrate-create name=create_table_orders
# Edit migrations/YYYYMMDDHHMMSS_create_table_orders.up.sql
make migrate-up
```

---

## 10. Konvensi & Aturan Kode

### Naming

| Item                | Konvensi                | Contoh                                         |
| ------------------- | ----------------------- | ---------------------------------------------- |
| Package             | lowercase, singkat      | `handler`, `usecase`, `repository`       |
| Interface           | Noun/Noun+er            | `UserRepository`, `AuthUsecase`            |
| Implementasi struct | lowercase + nama domain | `userRepository`, `authUsecase`            |
| Constructor         | `New()`               | `func New(db *gorm.DB) UserRepository`       |
| Error sentinel      | `Err` + PascalCase    | `ErrUserNotFound`, `ErrEmailAlreadyExists` |
| DTO Request         | Nama +`Request`       | `LoginRequest`, `CreateOrderRequest`       |
| DTO Response        | Nama +`Response`      | `AuthResponse`, `OrderResponse`            |

### Import Order

```go
import (
    // 1. Standard library
    "context"
    "errors"

    // 2. External packages
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 3. Internal packages
    "github.com/afifudin23/putra-sunda-trans-api/pkg/response"
    userdomain "github.com/afifudin23/putra-sunda-trans-api/internal/user/domain"
)
```

### Validation Tags di DTO

```go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8,max=72"`
    Name     string `json:"name" binding:"required,min=2,max=100"`
}
```

Tag validation yang umum digunakan:

- `required` — field wajib ada
- `email` — format email valid
- `min=N`, `max=N` — panjang minimum/maximum
- `uuid` — format UUID valid
- `oneof=a b c` — harus salah satu dari nilai yang ditentukan
- `eqfield=Field` — harus sama dengan field lain (untuk konfirmasi password)

---

## 11. Error Handling

### Pattern Error di Usecase

```go
func (u *authUsecase) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
    _, err := u.userRepo.FindByEmail(ctx, req.Email)
    if err == nil {
        // User ditemukan → email sudah ada
        return nil, domain.ErrEmailAlreadyExists
    }
    if !errors.Is(err, userdomain.ErrUserNotFound) {
        // Error lain yang tidak diharapkan → propagate
        return nil, err
    }
    // Lanjutkan proses registrasi...
}
```

### Pattern Error di Handler

```go
func (h *authHandler) Register(c *gin.Context) {
    // ...
    result, err := h.usecase.Register(c.Request.Context(), &req)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrEmailAlreadyExists):
            response.Conflict(c, "email already exists")
        case errors.Is(err, domain.ErrInvalidCredentials):
            response.Unauthorized(c, "invalid credentials")
        default:
            // Unexpected error: log dan return 500
            logger.Error("register failed", logger.Err(err))
            response.InternalServerError(c, err)
        }
        return
    }
    response.Success(c, result, "register success")
}
```

**Aturan**: Jangan pernah ekspose detail error internal (pesan error dari database, stack trace) ke client. Hanya gunakan error message yang user-friendly.

---

## 12. Authentication & Authorization

### Flow Auth

```
POST /api/v1/auth/register  →  Return JWT token
POST /api/v1/auth/login     →  Return JWT token

Protected routes:
  Request Header: Authorization: Bearer <token>
  Middleware.Auth() → verify token → inject userID ke gin.Context
  Handler → c.GetString("userID") → gunakan untuk query
```

### JWT Claims

```go
type Claims struct {
    UID string `json:"uid"`  // User UUID
    jwt.RegisteredClaims     // exp, iat, dll
}
// Expiry: 24 jam dari waktu generate
```

### Menambah Protected Route

```go
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
    group := rg.Group("/resource")

    // Public routes
    group.GET("/:id", m.handler.GetByID)

    // Protected routes (butuh JWT)
    protected := group.Use(middleware.Auth(m.accessKey))
    {
        protected.POST("", m.handler.Create)
        protected.PUT("/:id", m.handler.Update)
        protected.DELETE("/:id", m.handler.Delete)
    }
}
```

---

## 13. Logging

Project menggunakan **Zap** (structured logging) via wrapper di `pkg/logger`.

### Penggunaan

```go
// Info
logger.Info("user registered", logger.String("email", user.Email), logger.String("id", user.ID.String()))

// Error (selalu include err)
logger.Error("failed to create user", logger.Err(err), logger.String("email", req.Email))

// Warn
logger.Warn("deprecated endpoint called", logger.String("path", c.Request.URL.Path))
```

### Field Helpers

| Helper                  | Tipe        | Contoh                            |
| ----------------------- | ----------- | --------------------------------- |
| `logger.String(k, v)` | string      | `logger.String("email", email)` |
| `logger.Int(k, v)`    | int         | `logger.Int("count", 5)`        |
| `logger.Int64(k, v)`  | int64       | `logger.Int64("size", 1024)`    |
| `logger.Uint(k, v)`   | uint        | `logger.Uint("page", 1)`        |
| `logger.Bool(k, v)`   | bool        | `logger.Bool("success", true)`  |
| `logger.Err(err)`     | error       | `logger.Err(err)`               |
| `logger.Any(k, v)`    | interface{} | `logger.Any("data", obj)`       |

### Kapan Logging

- **Info**: Lifecycle event (server start, request penting, user action)
- **Warn**: Kondisi tidak ideal tapi masih bisa lanjut
- **Error**: Error yang perlu perhatian, selalu include `logger.Err(err)`
- **Debug**: Detail untuk debugging, tidak dipakai di production

---

## 14. Troubleshooting Umum

### Server tidak mau start

```
Pastikan:
1. File .env ada dan lengkap
2. PostgreSQL berjalan dan bisa diakses
3. Database "putra_sunda_db" sudah dibuat
4. Migrasi sudah dijalankan (make migrate-up)
```

### `migrate: error: Dirty database version N`

```bash
# Cek versi saat ini
make migrate-version

# Force ke versi sebelumnya
make migrate-force version=N-1

# Jalankan ulang
make migrate-up
```

### Build error setelah menambah module baru

```bash
# Pastikan semua import path benar
go build ./...

# Jika ada dependency baru
go mod tidy
```

### JWT token tidak diterima middleware

Pastikan:

1. Header format benar: `Authorization: Bearer <token>` (ada spasi setelah "Bearer")
2. `ACCESS_KEY` di `.env` sama persis dengan yang digunakan saat generate token
3. Token belum expired (TTL 24 jam)

### Validation error tidak muncul detail field

Pastikan menggunakan `response.ValidationError(c, err)` di handler, bukan `response.BadRequest(c, err.Error())`.

---

_Versi dokumen ini sesuai dengan project versi `0.1.0`._
