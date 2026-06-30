# Ringkasan API — SEAPEDIA

Ringkasan singkat + contoh **request/response** untuk endpoint kunci.
Daftar endpoint **lengkap** ada di [API_ENDPOINTS.md](API_ENDPOINTS.md); format envelope di
[API_RESPONSE.md](API_RESPONSE.md).

## Info Umum

| Item | Nilai |
|------|-------|
| Base URL | `http://localhost:{APP_PORT}` (default `:5000`) |
| Base Path | `/api/v1` |
| Format | JSON (`Content-Type: application/json`) |
| Auth | JWT via header `Authorization: Bearer <token>` |
| Identitas akun | **email** (bukan username) |
| Otorisasi | mengikuti **active role** pada JWT |
| Swagger UI | `/docs/v1/index.html` |

Semua response memakai envelope standar:

```json
{ "status": true, "data": {}, "list_data": [], "message": "", "error": {}, "error_code": "", "pagination": {} }
```

---

## 1. Health Check

`GET /ping`

```json
{ "status": true, "data": { "app": "SEAPEDIA API", "version": "0.1.0" }, "message": "pong" }
```

---

## 2. Register

`POST /api/v1/auth/register`

| Field | Tipe | Aturan |
|-------|------|--------|
| `email` | string | wajib, format email |
| `password` | string | wajib, min 6 |
| `confirm_password` | string | wajib, sama dengan `password` |
| `roles` | array | opsional, isi `buyer`/`seller`/`driver` (default `["buyer"]`) |

```json
{
  "email": "user@mail.com",
  "password": "Password123",
  "confirm_password": "Password123",
  "roles": ["buyer", "seller"]
}
```

**Response — 201 Created**

```json
{
  "status": true,
  "message": "user registered",
  "data": {
    "user": { "id": "uuid", "email": "user@mail.com", "is_admin": false, "roles": ["buyer","seller"], "created_at": "..." },
    "token": "jwt",
    "active_role": "",
    "need_role_selection": true
  }
}
```

> `need_role_selection: true` bila punya >1 role → wajib `POST /auth/select-role` dulu.

**Error:** `400 VALIDATION_ERROR`, `409 CONFLICT` (email sudah dipakai).

---

## 3. Login

`POST /api/v1/auth/login`

```json
{ "email": "buyer1@seapedia.test", "password": "Buyer123" }
```

**Response — 200 OK** (sama bentuk dengan register: `user`, `token`, `active_role`, `need_role_selection`).
Bila user punya 1 role, `active_role` langsung terisi; admin → `active_role: "admin"`.

**Error:** `401 UNAUTHORIZED` (email/password salah).

---

## 4. Pilih Role Aktif

`POST /api/v1/auth/select-role` — header `Authorization: Bearer <token>`

```json
{ "role": "seller" }
```

Mengembalikan **token baru** dengan `active_role` terisi. Error `403 FORBIDDEN` bila role tidak dimiliki.

---

## 5. Checkout (contoh alur transaksi)

`POST /api/v1/buyer/checkout` (active role buyer)

```json
{ "address_id": "uuid", "delivery_method": "regular", "discount_code": "SEAPEDIA10" }
```

Response berisi order: `subtotal`, `discount`, `delivery_fee`, `tax` (PPN 12%), `total`, `status` (awal `Sedang Dikemas`), `items`, `status_history`.

**Error:** `422 UNPROCESSABLE_ENTITY` (cart kosong / stok kurang / saldo kurang / diskon invalid), `404 NOT_FOUND` (alamat).

---

## Akun Demo (setelah `make seed`)

Login pakai **email**, password pola `Role123`:

| Email | Password | Role |
|-------|----------|------|
| `admin@seapedia.test` | `Admin123` | Admin |
| `seller1@seapedia.test` | `Seller123` | Seller |
| `buyer1@seapedia.test` | `Buyer123` | Buyer |
| `driver1@seapedia.test` | `Driver123` | Driver |
| `multi1@seapedia.test` | `Multi123` | Buyer+Seller+Driver |

---

## Referensi Kode

| Bagian | File |
|--------|------|
| Router | [internal/router/router.go](../../internal/router/router.go) |
| Auth | [internal/auth/](../../internal/auth/) |
| Envelope response | [pkg/response/](../../pkg/response/) |
| JWT (uid + active_role + jti) | [pkg/jwt/jwt.go](../../pkg/jwt/jwt.go) |
