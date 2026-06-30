# Ringkasan API — Putra Sunda Trans

Ringkasan endpoint beserta contoh **request** dan **response**.
Untuk detail format envelope response, lihat [API_RESPONSE.md](API_RESPONSE.md).

## Info Umum

| Item | Nilai |
|------|-------|
| Base URL | `http://localhost:{APP_PORT}` |
| Base Path | `/api/v1` |
| Format | JSON (`Content-Type: application/json`) |
| Auth | JWT via header `Authorization: Bearer <token>` |
| Versi | `0.1.0` |

Semua response sukses/gagal memakai envelope standar:

```json
{
  "status": true,
  "data": {},
  "message": "",
  "error": {},
  "error_code": ""
}
```

---

## Daftar Endpoint

| Method | Path | Auth | Keterangan |
|--------|------|------|------------|
| `GET`  | `/ping` | — | Health check |
| `POST` | `/api/v1/auth/register` | — | Registrasi user baru |
| `POST` | `/api/v1/auth/login` | — | Login & ambil token |

---

## 1. Health Check

`GET /ping`

**Response — 200 OK**

```json
{
  "message": "pong"
}
```

> Catatan: endpoint ini tidak memakai envelope standar.

---

## 2. Register

`POST /api/v1/auth/register`

**Request Body**

| Field | Tipe | Aturan |
|-------|------|--------|
| `email` | string | wajib, format email |
| `password` | string | wajib, min 3 karakter |
| `confirm_password` | string | wajib, harus sama dengan `password` |

```json
{
  "email": "user@example.com",
  "password": "secret",
  "confirm_password": "secret"
}
```

**Response — 201 Created**

```json
{
  "status": true,
  "message": "user registered",
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "created_at": "2026-06-29T10:00:00Z"
    },
    "token": "jwt-token"
  }
}
```

**Kemungkinan Error**

| Status | error_code | Kondisi |
|--------|-----------|---------|
| 400 | `VALIDATION_ERROR` | Field gagal validasi |
| 409 | `CONFLICT` | Email sudah terdaftar |
| 500 | `INTERNAL_SERVER_ERROR` | Kesalahan server |

Contoh **409 Conflict**:

```json
{
  "status": false,
  "message": "email already exists",
  "error_code": "CONFLICT"
}
```

Contoh **400 Validation**:

```json
{
  "status": false,
  "message": "Validation failed",
  "error": {
    "email": "email must be a valid email",
    "confirm_password": "confirm_password must match Password"
  },
  "error_code": "VALIDATION_ERROR"
}
```

---

## 3. Login

`POST /api/v1/auth/login`

**Request Body**

| Field | Tipe | Aturan |
|-------|------|--------|
| `email` | string | wajib, format email |
| `password` | string | wajib, min 3 karakter |

```json
{
  "email": "user@example.com",
  "password": "secret"
}
```

**Response — 200 OK**

```json
{
  "status": true,
  "message": "login success",
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "created_at": "2026-06-29T10:00:00Z"
    },
    "token": "jwt-token"
  }
}
```

**Kemungkinan Error**

| Status | error_code | Kondisi |
|--------|-----------|---------|
| 400 | `VALIDATION_ERROR` | Field gagal validasi |
| 401 | `UNAUTHORIZED` | Email/password salah |
| 500 | `INTERNAL_SERVER_ERROR` | Kesalahan server |

Contoh **401 Unauthorized**:

```json
{
  "status": false,
  "message": "invalid credentials",
  "error_code": "UNAUTHORIZED"
}
```

---

## Autentikasi

Endpoint yang dilindungi membutuhkan header:

```
Authorization: Bearer <token>
```

Token didapat dari response `register` atau `login` (field `data.token`).

Jika token tidak ada / tidak valid:

```json
{ "message": "token tidak ada" }
```
```json
{ "message": "token tidak valid" }
```

---

## Referensi Kode

| Bagian | File |
|--------|------|
| Routing root | [internal/router/router.go](../../internal/router/router.go) |
| Route auth | [internal/auth/routes.go](../../internal/auth/routes.go) |
| Handler auth | [internal/auth/handler/handler.go](../../internal/auth/handler/handler.go) |
| Request DTO | [internal/auth/dto/request.go](../../internal/auth/dto/request.go) |
| Response DTO | [internal/auth/dto/response.go](../../internal/auth/dto/response.go) |
| Envelope response | [pkg/response/response.go](../../pkg/response/response.go) |
