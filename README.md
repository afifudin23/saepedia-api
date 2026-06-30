# SEAPEDIA API

Backend API untuk **SEAPEDIA** — marketplace multi-role (Admin, Seller, Buyer, Driver).
Repository ini mengimplementasikan **Level 1 – Level 7** dari Technical Challenge COMPFEST 18.

> Stack: **Gin** · **GORM** · **golang-migrate** · **zap** · **JWT (HS256)** · **argon2id**
> Pola: **Clean Architecture** — layer terpisah, dependency injection manual di `router.Setup`.

---

## Daftar Isi

1. [Cakupan Level](#cakupan-level)
2. [Cara Menjalankan](#cara-menjalankan)
3. [Environment Variables](#environment-variables)
4. [Akun Demo / Seed Data](#akun-demo--seed-data)
5. [Struktur Folder](#struktur-folder)
6. [Aturan Antar Layer](#aturan-antar-layer)
7. [Standard API Response](#standard-api-response)
8. [Konsep Multi-Role &amp; Active Role](#konsep-multi-role--active-role)
9. [Business Rules Penting](#business-rules-penting)
10. [Daftar Endpoint](#daftar-endpoint)
11. [Contoh Alur Demo End-to-End](#contoh-alur-demo-end-to-end)
12. [Development Commands](#development-commands)

---

## Cakupan Level

| Level       | Fitur                                                                                           | Status |
| ----------- | ----------------------------------------------------------------------------------------------- | ------ |
| **1** | Public catalog (read), Auth + multi-role + active role, Public app reviews                      | ✅     |
| **2** | Seller store management, Product CRUD, Public catalog dari data backend                         | ✅     |
| **3** | Buyer wallet + top-up, Delivery address, Cart (single-store), Checkout + Order + status history | ✅     |
| **4** | Voucher & Promo discount, Seller process order, Buyer & Seller reports                          | ✅     |
| **5** | Delivery job, Driver find/take/complete, Driver earnings & dashboard                            | ✅     |
| **6** | Admin monitoring dashboard, Voucher/Promo management, Overdue auto-refund + time simulation     | ✅     |
| **7** | Logout invalidation (token denylist), RBAC server-side, input validation, XSS/SQLi hardening    | ✅     |

> **Cakupan repo ini: BACKEND / REST API saja.** Technical Challenge SEAPEDIA bersifat
> fullstack, namun submission ini fokus pada sisi API yang menjalankan **seluruh business rule
> Level 1–7**. Tanda ✅ di atas = cakupan backend (logika + endpoint), bukan UI klien. Bagian
> antarmuka (web/mobile) dan bonus UI/deployment belum termasuk. Dokumentasi interaktif endpoint
> tersedia via Swagger UI (`/docs`).

---

## Cara Menjalankan

### 0. Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- **Make** (opsional) & **golang-migrate** CLI

Install tooling (migrate + air) sekaligus:

```bash
make setup
```

Atau install `migrate` manual:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 1. Konfigurasi Environment

```bash
cp .env.example .env
# Edit .env → sesuaikan DB_PASSWORD, DB_NAME, dan ACCESS_KEY
```

Buat database-nya (sekali saja):

```sql
CREATE DATABASE seapedia_db;
```

### 2. Jalankan Migration

```bash
make migrate-up
```

> Tanpa Make:
> `migrate -path migrations -database "postgres://postgres:PASSWORD@localhost:5432/seapedia_db?sslmode=disable" up`

### 3. Seed Data Demo (akun admin/seller/buyer/driver)

```bash
# Jalankan SEMUA seeder
make seed                       # atau: go run ./scripts/seed

# Jalankan seeder SATUAN (pilih: user, discount, review, address, order)
make seed-one name=user         # atau: go run ./scripts/seed data=user
go run ./scripts/seed data=order
```

> Semua file seeder ada di `scripts/seed/` (satu package): `users.go`, `discounts.go`,
> `reviews.go`, `addresses.go`, `orders.go` berisi datanya, `seeder.go` registry-nya,
> `main.go` runner-nya. Tanpa argumen → seed semua; `data=<nama>` → seed satuan.
>
> **Isi seed (±10 baris/tabel):** 10 user (4 toko, 14 produk dengan gambar Unsplash),
> 10 diskon (5 voucher + 5 promo), 10 review, 10 alamat, 10 order (berbagai status +
> item + riwayat status + transaksi wallet, stok & saldo ikut disesuaikan).

### 4. Jalankan Server

```bash
make run       # go run ./cmd/api
# atau hot reload:
make air
```

Server berjalan di `http://localhost:5000`. Health check: `GET /ping`.

### 5. Swagger / OpenAPI

Dokumentasi interaktif (Swagger UI) tersedia setelah server jalan:

```
http://localhost:5000/docs/v1/index.html      (atau cukup /docs → auto-redirect)
```

Spec mentah: `http://localhost:5000/docs/v1/doc.json`. Path UI mengikuti `config.APIVersion`
(jadi otomatis `/docs/v2/...` saat versi dinaikkan). File hasil generate ada di
`docs/swagger/` (`swagger.json`, `swagger.yaml`, `docs.go`). Untuk regenerate setelah
mengubah anotasi handler:

```bash
make swag        # = swag init -g cmd/api/main.go -o docs/swagger --parseInternal
```

> Butuh CLI `swag` (otomatis terpasang via `make setup`, atau
> `go install github.com/swaggo/swag/cmd/swag@latest`). Untuk endpoint yang butuh login,
> klik tombol **Authorize** di Swagger UI lalu isi `Bearer <token>`.

---

## Environment Variables

```env
APP_NAME="SEAPEDIA API"
APP_PORT=5000

DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="your_password"
DB_NAME="seapedia_db"

ACCESS_KEY="long-random-secret-for-jwt"   # WAJIB diisi
```

`DB_PASSWORD` dan `ACCESS_KEY` wajib ada; aplikasi berhenti bila kosong.

---

## Akun Demo / Seed Data

Setelah `make seed`:

Login memakai **email** (bukan username). Password mengikuti pola `Role123`:

| Email                     | Password     | Role                    | Catatan                                                                    |
| ------------------------- | ------------ | ----------------------- | -------------------------------------------------------------------------- |
| `admin@seapedia.test`   | `Admin123`  | Admin                   | role aktif otomatis`admin`                                               |
| `seller1@seapedia.test` | `Seller123` | Seller                  | Toko Sumber Rejeki (4 produk)                                              |
| `seller2@seapedia.test` | `Seller123` | Seller                  | Elektronik Jaya (4 produk)                                                 |
| `seller3@seapedia.test` | `Seller123` | Seller                  | Dapur Sehat (3 produk)                                                     |
| `buyer1@seapedia.test`  | `Buyer123`  | Buyer                   | saldo Rp1.000.000                                                          |
| `buyer2@seapedia.test`  | `Buyer123`  | Buyer                   | saldo Rp500.000                                                            |
| `buyer3@seapedia.test`  | `Buyer123`  | Buyer                   | saldo Rp2.000.000                                                          |
| `driver1@seapedia.test` | `Driver123` | Driver                  | —                                                                         |
| `driver2@seapedia.test` | `Driver123` | Driver                  | —                                                                         |
| `multi1@seapedia.test`  | `Multi123`  | Buyer + Seller + Driver | **multi-role** → wajib pilih active role; Toko Serba Ada (3 produk) |

Seed juga membuat **10 kode diskon** (5 voucher + 5 promo):

- **Voucher** (punya kuota): `SEAPEDIA10`, `GROCERY15`, `NEWUSER20`, `FLASH25`, `HEMAT50K`
- **Promo** (tanpa kuota): `HEMAT5K`, `POTONG10K`, `POTONG20K`, `PROMO10`, `GAJIAN`

> **Membuat admin baru:** admin dibuat lewat seed (`scripts/seed/users.go`, field `IsAdmin: true`).
> Registrasi publik **tidak** bisa membuat admin. Tambahkan entry baru di seeder lalu jalankan ulang `make seed`.

---

## Struktur Folder

```
saepedia-api/
├── cmd/api/main.go                ← entry point + bootstrap
├── internal/
│   ├── auth/                      ← register, login, logout, select-role, me, balance-summary
│   ├── user/                      ← entity User + roles + repository (dipakai auth)
│   ├── review/                    ← public application reviews
│   ├── store/                     ← seller store management
│   ├── product/                  ← product CRUD + public catalog
│   ├── wallet/                    ← buyer wallet + top-up + transactions
│   ├── address/                  ← buyer delivery address
│   ├── cart/                      ← cart (single-store rule)
│   ├── order/                     ← checkout, order, status history, seller process, reports, overdue
│   ├── discount/                  ← voucher & promo (generate/list/validate)
│   ├── delivery/                  ← driver job: find/take/complete + earnings
│   ├── admin/                     ← monitoring dashboard + time simulation + overdue trigger
│   ├── setting/                   ← app_settings (offset simulasi waktu)
│   └── router/router.go           ← rakit semua modul + dependency injection
├── pkg/
│   ├── response/                  ← standard response builder + ErrorCode
│   ├── jwt/                       ← generate & verify JWT (uid + active_role + jti)
│   ├── helper/                    ← argon2id hashing
│   ├── middleware/                ← Auth (+ token denylist), RequireRole, CORS
│   ├── pagination/                ← parse ?page=&per_page=
│   ├── clock/                     ← virtual now untuk simulasi waktu
│   └── logger/                    ← zap wrapper
├── config/config.go               ← load ENV + Version
├── database/postgres.go           ← koneksi GORM
├── migrations/*.sql               ← golang-migrate (up/down)
├── scripts/seed/                   ← seeder (main, seeder, users, discounts, reviews, addresses, orders)
├── Makefile · .air.toml · .env.example · go.mod
```

Setiap modul fitur memakai layering: `domain/` (entity + interface) → `repository/` (GORM) →
`usecase/` (business logic) → `handler/` (HTTP) → `routes.go` (wiring + middleware).

---

## Aturan Antar Layer

```
Handler (Gin)        → tahu: domain, dto, pkg/response, pkg/middleware
Usecase              → tahu: domain, pkg/
Repository (GORM)    → tahu: domain
Domain               → tidak tahu siapa pun (hanya std lib)
```

Layer atas boleh tahu layer bawah, tidak sebaliknya.

---

## Standard API Response

```jsonc
// Sukses — single object
{ "status": true, "data": { ... }, "message": "login success" }

// Sukses — list + pagination
{ "status": true, "list_data": [ ... ], "message": "",
  "pagination": { "page": 1, "per_page": 10, "total": 100, "total_pages": 10 } }

// Validation error — 400
{ "status": false, "message": "Validation failed",
  "error": { "email": "email must be a valid email" }, "error_code": "VALIDATION_ERROR" }

// Unauthorized 401 / Forbidden 403 / Not found 404 / Conflict 409 / Unprocessable 422
{ "status": false, "message": "...", "error_code": "UNAUTHORIZED" }
```

---

## Konsep Multi-Role & Active Role

Satu akun (email) non-admin bisa memiliki banyak role (buyer/seller/driver). Authorization **selalu
mengikuti role aktif** yang tersimpan di JWT, bukan sekadar daftar role yang dimiliki.

Alur:

1. **Register** (`/auth/register`) — pilih satu/lebih role (`roles: ["buyer","seller"]`). Default `buyer`.
2. **Login** (`/auth/login`) — respons mengembalikan:
   - `token` (JWT), `user.roles` (semua role), `active_role`, `need_role_selection`.
   - Jika user hanya punya **1 role** → role itu langsung aktif (`need_role_selection: false`).
   - Jika punya **>1 role** → `active_role: ""` dan `need_role_selection: true`. Token tetap
     diberikan tapi **belum** bisa mengakses dashboard privat.
   - Admin → `active_role: "admin"` otomatis.
3. **Select role** (`/auth/select-role`) — kirim `{ "role": "seller" }`. Server memverifikasi user
   benar memiliki role tsb, lalu mengeluarkan **token baru** dengan `active_role` terisi.
4. Endpoint privat dilindungi `RequireRole(...)` yang mengecek `active_role` di token.

Active role bisa dilihat kapan saja via `GET /auth/me`.

---

## Business Rules Penting

### Single-store checkout

Satu cart hanya boleh berisi produk dari **satu toko**. Saat menambah produk dari toko berbeda,
API menolak dengan `409 CONFLICT` (`"cart can only contain products from one store, please clear the cart first"`). `store_id` cart di-set saat item pertama masuk dan otomatis di-reset ketika cart
dikosongkan (`DELETE /buyer/cart`) atau item terakhir dihapus.

### Perhitungan Checkout & PPN 12%

Urutan & basis perhitungan (semua nominal **integer rupiah**):

```
subtotal      = Σ (harga produk × qty)
discount      = 0            (voucher/promo menyusul di Level 4)
tax (PPN 12%) = (subtotal - discount) × 12 / 100      ← PPN dihitung SETELAH diskon
delivery_fee  = sesuai metode (lihat tabel)
total         = (subtotal - discount) + delivery_fee + tax
```

> **Basis PPN:** PPN 12% dikenakan pada `subtotal - discount` (bukan termasuk ongkir). Pembulatan
> memakai pembagian integer (floor). Field `tax_percent: 12` selalu disertakan di respons checkout.

### Tarif Pengiriman (delivery fee berbeda per metode)

| Metode   | `delivery_method` | Fee      |
| -------- | ------------------- | -------- |
| Instant  | `instant`         | Rp20.000 |
| Next Day | `next_day`        | Rp10.000 |
| Regular  | `regular`         | Rp5.000  |

### Wallet & Stok

- Buyer **tidak bisa** checkout bila saldo wallet < total → `422 UNPROCESSABLE_ENTITY`.
- Checkout berjalan dalam **satu transaksi DB atomik**: kurangi stok (aman, `WHERE stock >= qty`
  sehingga **tidak pernah negatif**), potong saldo wallet, catat `wallet_transaction` (type
  `payment`), buat order + items + status history, lalu kosongkan cart.
- Top-up bersifat **dummy** (langsung menambah saldo, tanpa payment gateway).

### Order Lifecycle

Status awal setelah checkout sukses: **`Sedang Dikemas`**. Status utama yang dipakai sistem:

```
Sedang Dikemas → Menunggu Pengirim → Sedang Dikirim → Pesanan Selesai
                                                     ↘ Dikembalikan
```

Setiap perubahan status dicatat di `order_status_histories` lengkap dengan timestamp (di Level 3
baru status awal `Sedang Dikemas`; transisi lain menyusul di Level 4–6).

### Discount: Voucher & Promo (Level 4)

- Dua jenis diskon dibedakan lewat kolom `kind`: **voucher** (punya kuota `usage_limit` +
  `used_count`) dan **promo** (tanpa kuota). Keduanya punya `expires_at`.
- Tipe potongan: `percent` (opsional `max_discount` sebagai cap) atau `fixed`.
- **Aturan kombinasi:** hanya **satu** kode diskon per checkout (voucher **ATAU** promo, tidak
  bisa digabung). Dikirim lewat field `discount_code`.
- **Posisi terhadap PPN:** diskon mengurangi subtotal **sebelum** PPN. PPN 12% = `(subtotal - discount) × 12%`. Konsisten di preview & checkout.
- Validasi saat checkout: kode tidak ditemukan → 404; kadaluarsa / kuota habis / belum mencapai
  `min_spend` → 422. Kuota voucher dikurangi **atomik** di dalam transaksi checkout (anti
  race / over-use).

### Seller Process Order (Level 4)

- Aksi `POST /seller/orders/:id/process` memindahkan status **Sedang Dikemas → Menunggu Pengirim**
  (atomik, hanya order milik toko seller). Sebelum diproses, order **tidak** muncul sebagai job driver.
- Transisi dicatat di `order_status_histories`.

### Driver Earning & Delivery (Level 5)

- Job tersedia = order berstatus **Menunggu Pengirim** dan belum punya driver.
- **Ambil job** atomik: `UPDATE ... WHERE status='Menunggu Pengirim' AND driver_id IS NULL` →
  mencegah dua driver mengambil order yang sama (job sudah diambil → 409).
- Take → **Sedang Dikirim**, Complete → **Pesanan Selesai** (keduanya catat history + timestamp).
- **Aturan pendapatan driver:** driver mendapat **80% dari ongkir** (`delivery_fee`), dikunci saat
  job selesai (`driver_earning`). Dashboard driver menampilkan job aktif, riwayat, & total earning.

### Overdue, SLA & Simulasi Waktu (Level 6)

- **SLA per metode** (dari `created_at`): Instant **1 hari**, Next Day **2 hari**, Regular **3 hari**.
- Order yang melewati SLA dan **belum** final (masih Sedang Dikemas / Menunggu Pengirim / Sedang
  Dikirim) akan **di-refund otomatis**: status → **Dikembalikan**, dana (full `total`) kembali ke
  wallet buyer + tercatat di riwayat wallet (`type: refund`), **stok dipulihkan**, dan history
  status ditulis. Semua dalam **satu transaksi**.
- **Anti double-refund:** kolom `refunded_at` + guard status memastikan refund hanya sekali.
- **Pendapatan seller** dihitung dari order **Pesanan Selesai** saja (`subtotal - discount`); order
  Dikembalikan otomatis tidak terhitung sebagai income (tidak perlu reversal manual).
- **Simulasi waktu:** offset waktu virtual disimpan di tabel `app_settings` dan dipakai semua
  logika waktu via `pkg/clock`. Admin memajukan waktu dengan `POST /admin/simulate/advance-day`
  (`{"days": N}`) yang langsung menjalankan overdue handling, atau `POST /admin/overdue/run`
  untuk memproses manual. Cek waktu virtual: `GET /admin/simulate/now`.

### Keamanan (Level 7)

- **SQL Injection:** seluruh akses DB lewat GORM / query parameter (`?`); tidak ada string
  concatenation pada query. Input pencarian (`ILIKE ?`) pun parameterized.
- **XSS:** konten user-generated (review, deskripsi toko/produk, alamat) di-`html.EscapeString`
  saat disimpan sehingga ditampilkan sebagai teks biasa, tidak mengeksekusi script.
- **Password:** hashing **argon2id**.
- **Validasi input:** `go-playground/validator` memvalidasi email, phone, rating (1–5), quantity
  (>0), price/stock (≥0), discount value (>0), dll. Input invalid → 400 dengan pesan jelas per field.
- **Session / token:** JWT **HS256**, berlaku **24 jam** (`pkg/jwt.TokenTTL`). **Logout
  meng-invalidate token**: `jti` token dimasukkan ke denylist `revoked_tokens` dan dicek di setiap
  request (token logged-out → 401).
- **RBAC server-side:** otorisasi mengikuti **active role** di JWT (bukan klaim frontend).
  `RequireRole(...)` melindungi tiap endpoint privat; endpoint admin hanya untuk role `admin`.
- **Ownership:** user hanya bisa mengakses/mengubah miliknya — produk dicek vs toko seller, alamat &
  order & wallet difilter per user, job driver dicek `driver_id`. Mengakses resource milik orang
  lain → 403/404.

#### Cara uji keamanan singkat

- **XSS:** submit review `{"comment":"<script>alert(1)</script>", ...}` → tersimpan ter-escape,
  saat dirender muncul sebagai teks, tidak tereksekusi.
- **SQLi:** coba `?search=' OR 1=1 --` di `/products` atau email aneh di login → tidak
  mempengaruhi query (parameterized), hasil tetap aman.
- **Logout:** panggil `/auth/logout` lalu pakai token yang sama → 401 `token sudah tidak berlaku`.
- **RBAC:** login buyer, akses `/seller/products` atau `/admin/summary` → 403.

---

## Daftar Endpoint

Base URL: `http://localhost:5000/api/v1`. Endpoint privat butuh header
`Authorization: Bearer <token>`.

### Public (boleh guest)

| Method | Path                                  | Keterangan                                                      |
| ------ | ------------------------------------- | --------------------------------------------------------------- |
| GET    | `/ping`                             | Health check (di root, bukan`/api/v1`)                        |
| GET    | `/reviews`                          | List application reviews                                        |
| POST   | `/reviews`                          | Submit review (`reviewer_name`, `rating` 1–5, `comment`) |
| GET    | `/stores`                           | List toko                                                       |
| GET    | `/stores/:id`                       | Detail toko                                                     |
| GET    | `/products?search=&page=&per_page=` | Catalog produk (+ info toko)                                    |
| GET    | `/products/:id`                     | Detail produk (+ info toko)                                     |

### Auth

| Method | Path                      | Auth  | Keterangan                                                 |
| ------ | ------------------------- | ----- | ---------------------------------------------------------- |
| POST   | `/auth/register`        | —    | `email`, `password`, `confirm_password`, `roles[]` |
| POST   | `/auth/login`           | —    | `email`, `password`                                    |
| POST   | `/auth/logout`          | token | Stateless (client buang token)                             |
| POST   | `/auth/select-role`     | token | `{ "role": "seller" }` → token baru                     |
| GET    | `/auth/me`              | token | Profil + roles + active_role                               |
| GET    | `/auth/balance-summary` | token | Wallet balance (+ placeholder income/earnings)             |

### Seller (active role = `seller`)

| Method   | Path                           | Keterangan                                                 |
| -------- | ------------------------------ | ---------------------------------------------------------- |
| GET      | `/seller/store`              | Toko milik sendiri                                         |
| POST/PUT | `/seller/store`              | Buat/update toko (`name` unik, `description`)          |
| GET      | `/seller/products`           | Produk milik toko sendiri                                  |
| POST     | `/seller/products`           | Buat produk (`name`,`description`,`price`,`stock`) |
| PUT      | `/seller/products/:id`       | Update produk milik sendiri                                |
| DELETE   | `/seller/products/:id`       | Hapus produk milik sendiri                                 |
| GET      | `/seller/orders`             | Daftar order masuk untuk toko                              |
| GET      | `/seller/orders/:id`         | Detail order milik toko                                    |
| POST     | `/seller/orders/:id/process` | Proses order: Sedang Dikemas → Menunggu Pengirim          |
| GET      | `/seller/reports`            | Ringkasan pendapatan seller                                |

### Buyer (active role = `buyer`)

| Method | Path                             | Keterangan                                     |
| ------ | -------------------------------- | ---------------------------------------------- |
| GET    | `/buyer/wallet`                | Saldo wallet                                   |
| POST   | `/buyer/wallet/topup`          | Dummy top-up (`amount`)                      |
| GET    | `/buyer/wallet/transactions`   | Riwayat transaksi wallet                       |
| GET    | `/buyer/addresses`             | List alamat                                    |
| POST   | `/buyer/addresses`             | Tambah alamat                                  |
| PUT    | `/buyer/addresses/:id`         | Update alamat                                  |
| DELETE | `/buyer/addresses/:id`         | Hapus alamat                                   |
| GET    | `/buyer/cart`                  | Ringkasan cart                                 |
| POST   | `/buyer/cart/items`            | Tambah item (`product_id`, `quantity`)     |
| PUT    | `/buyer/cart/items/:productID` | Ubah qty (`quantity`)                        |
| DELETE | `/buyer/cart/items/:productID` | Hapus item                                     |
| DELETE | `/buyer/cart`                  | Kosongkan cart                                 |
| POST   | `/buyer/checkout/preview`      | Hitung ringkasan (`delivery_method`)         |
| POST   | `/buyer/checkout`              | Checkout (`address_id`, `delivery_method`) |
| GET    | `/buyer/orders`                | Riwayat order                                  |
| GET    | `/buyer/orders/:id`            | Detail order + timeline status                 |
| GET    | `/buyer/reports`               | Ringkasan pengeluaran buyer                    |

> Checkout & preview menerima `discount_code` opsional (voucher/promo).

### Driver (active role = `driver`)

| Method | Path                          | Keterangan                                     |
| ------ | ----------------------------- | ---------------------------------------------- |
| GET    | `/driver/jobs`              | Daftar job tersedia (status Menunggu Pengirim) |
| GET    | `/driver/jobs/:id`          | Detail job tersedia                            |
| POST   | `/driver/jobs/:id/take`     | Ambil job → Sedang Dikirim                    |
| POST   | `/driver/jobs/:id/complete` | Selesaikan job → Pesanan Selesai              |
| GET    | `/driver/dashboard`         | Job aktif, riwayat, total earning              |

### Admin (active role = `admin`)

| Method | Path                                                                          | Keterangan                                                                    |
| ------ | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| GET    | `/admin/summary`                                                            | Ringkasan monitoring (users/stores/products/orders/discount/delivery/overdue) |
| GET    | `/admin/users`                                                              | Monitoring users + roles                                                      |
| GET    | `/admin/stores`                                                             | Monitoring toko                                                               |
| GET    | `/admin/products`                                                           | Monitoring produk                                                             |
| GET    | `/admin/orders`                                                             | Monitoring order                                                              |
| GET    | `/admin/deliveries`                                                         | Monitoring pengiriman                                                         |
| GET    | `/admin/overdue-orders`                                                     | Order yang sedang overdue (virtual now)                                       |
| POST   | `/admin/vouchers` · GET `/admin/vouchers` · GET `/admin/vouchers/:id` | Generate / list / detail voucher                                              |
| POST   | `/admin/promos` · GET `/admin/promos` · GET `/admin/promos/:id`       | Generate / list / detail promo                                                |
| GET    | `/admin/simulate/now`                                                       | Lihat waktu virtual saat ini                                                  |
| POST   | `/admin/simulate/advance-day`                                               | Majukan waktu N hari + jalankan overdue (`{"days":1}`)                      |
| POST   | `/admin/overdue/run`                                                        | Jalankan overdue handling manual                                              |

> **Generate voucher** body: `{"code","discount_type":"percent|fixed","discount_value","max_discount?","min_spend?","expires_at":"RFC3339","usage_limit?"}`. Promo sama tanpa `usage_limit`.

---

## Contoh Alur Demo End-to-End

```bash
BASE=http://localhost:5000/api/v1

# 1. Guest lihat katalog
curl $BASE/products

# 2. Guest submit review
curl -X POST $BASE/reviews -H "Content-Type: application/json" \
  -d '{"reviewer_name":"Andi","rating":5,"comment":"Aplikasinya mantap!"}'

# 3. Login buyer (single role → langsung aktif)
TOKEN=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" \
  -d '{"email":"buyer1@seapedia.test","password":"Buyer123"}' | jq -r .data.token)

# 4. Top-up (opsional, buyer1 sudah punya saldo dari seed)
curl -X POST $BASE/buyer/wallet/topup -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"amount":500000}'

# 5. Tambah alamat
ADDR=$(curl -s -X POST $BASE/buyer/addresses -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"recipient_name":"Budi","phone":"08123456789","full_address":"Jl. Mawar 1","is_primary":true}' \
  | jq -r .data.id)

# 6. Tambah produk ke cart (ambil product_id dari GET /products)
curl -X POST $BASE/buyer/cart/items -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"product_id":"<PRODUCT_ID>","quantity":2}'

# 7. Preview & checkout
curl -X POST $BASE/buyer/checkout/preview -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"delivery_method":"regular"}'

curl -X POST $BASE/buyer/checkout -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"address_id\":\"$ADDR\",\"delivery_method\":\"regular\"}"

# 8. Multi-role: login multi1 → pilih role seller
TOKEN=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" \
  -d '{"email":"multi1@seapedia.test","password":"Multi123"}' | jq -r .data.token)   # need_role_selection:true
TOKEN=$(curl -s -X POST $BASE/auth/select-role -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"role":"seller"}' | jq -r .data.token)
curl $BASE/seller/orders -H "Authorization: Bearer $TOKEN"

# ── Level 4: checkout pakai voucher (sebagai buyer) ──
curl -X POST $BASE/buyer/checkout -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"address_id\":\"$ADDR\",\"delivery_method\":\"regular\",\"discount_code\":\"SEAPEDIA10\"}"

# ── Level 4: seller proses order (Sedang Dikemas → Menunggu Pengirim) ──
SELLER=$(curl -s -X POST $BASE/auth/login -d '{"email":"seller1@seapedia.test","password":"Seller123"}' | jq -r .data.token)
OID=$(curl -s $BASE/seller/orders -H "Authorization: Bearer $SELLER" | jq -r .list_data[0].id)
curl -X POST $BASE/seller/orders/$OID/process -H "Authorization: Bearer $SELLER"
curl $BASE/seller/reports -H "Authorization: Bearer $SELLER"

# ── Level 5: driver ambil & selesaikan job ──
DRIVER=$(curl -s -X POST $BASE/auth/login -d '{"email":"driver1@seapedia.test","password":"Driver123"}' | jq -r .data.token)
curl $BASE/driver/jobs -H "Authorization: Bearer $DRIVER"
curl -X POST $BASE/driver/jobs/$OID/take -H "Authorization: Bearer $DRIVER"
curl -X POST $BASE/driver/jobs/$OID/complete -H "Authorization: Bearer $DRIVER"
curl $BASE/driver/dashboard -H "Authorization: Bearer $DRIVER"

# ── Level 6: admin monitoring, generate voucher, simulasi waktu + overdue ──
ADMIN=$(curl -s -X POST $BASE/auth/login -d '{"email":"admin@seapedia.test","password":"Admin123"}' | jq -r .data.token)
curl $BASE/admin/summary -H "Authorization: Bearer $ADMIN"
curl -X POST $BASE/admin/vouchers -H "Authorization: Bearer $ADMIN" -H "Content-Type: application/json" \
  -d '{"code":"NEWYEAR","discount_type":"percent","discount_value":15,"max_discount":30000,"min_spend":100000,"expires_at":"2027-01-01T00:00:00Z","usage_limit":50}'
# Buat 1 order baru lalu majukan waktu 5 hari → order overdue otomatis di-refund:
curl -X POST $BASE/admin/simulate/advance-day -H "Authorization: Bearer $ADMIN" \
  -H "Content-Type: application/json" -d '{"days":5}'

# ── Level 7: logout meng-invalidate token ──
curl -X POST $BASE/auth/logout -H "Authorization: Bearer $DRIVER"
curl $BASE/driver/dashboard -H "Authorization: Bearer $DRIVER"   # → 401 token sudah tidak berlaku
```

> Koleksi Postman tersedia di `docs/SEAPEDIA.postman_collection.json`.

---

## Development Commands

```bash
make setup            # install migrate + air, go mod tidy
make run              # jalankan server
make air              # hot reload (dev)
make build            # build binary ke bin/
make seed                 # jalankan semua seeder
make seed-one name=user   # jalankan seeder satuan (data=user)
make migrate-up       # jalankan semua migration
make migrate-down     # rollback 1 step
make migrate-create name=add_something   # buat migration baru
make tidy             # go mod tidy
```
