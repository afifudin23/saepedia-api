# Standard API Response — SEAPEDIA

Semua endpoint memakai envelope yang sama (`pkg/response`).

```json
{
  "status": true,
  "data": { },
  "list_data": [ ],
  "message": "...",
  "error": { },
  "error_code": "...",
  "pagination": { }
}
```

Field yang tidak relevan diomit (`omitempty`).

---

## Success

### Single object

```json
{ "status": true, "data": { "id": "uuid", "field": "value" }, "message": "ok" }
```

### List + pagination

```json
{
  "status": true,
  "list_data": [ { "id": "uuid" } ],
  "message": "",
  "pagination": { "page": 1, "per_page": 10, "total": 100, "total_pages": 10 }
}
```

> `pagination` berisi **integer**. Query list: `?page=&per_page=` (default 1 & 10, maks 100).

### Create

```json
{ "status": true, "data": { "id": "uuid" }, "message": "created" }
```

### Delete

```json
{ "status": true, "message": "deleted" }
```

---

## Error

Bentuk umum: `{ "status": false, "message": "...", "error_code": "..." }`.

| HTTP | error_code | Kondisi umum |
|------|-----------|--------------|
| 400 | `VALIDATION_ERROR` | Field gagal validasi (lihat `error` per field) |
| 400 | `BAD_REQUEST` | Request tidak valid |
| 401 | `UNAUTHORIZED` | Belum login / kredensial salah |
| 401 | `TOKEN_INVALID` | Token tidak valid / kadaluarsa / sudah logout |
| 403 | `FORBIDDEN` | Role aktif tidak diizinkan / bukan pemilik resource |
| 404 | `NOT_FOUND` | Resource tidak ada |
| 409 | `CONFLICT` | Bentrok unik (email/store name) / cart beda toko / job sudah diambil |
| 422 | `UNPROCESSABLE_ENTITY` | Cart kosong, stok kurang, saldo kurang, diskon invalid, transisi status invalid |
| 500 | `INTERNAL_SERVER_ERROR` | Kesalahan server (detail tidak dibocorkan) |

### Validation (400)

```json
{
  "status": false,
  "message": "Validation failed",
  "error": { "email": "email must be a valid email" },
  "error_code": "VALIDATION_ERROR"
}
```

### Token invalid (401)

```json
{ "status": false, "message": "token tidak valid atau kadaluarsa", "error_code": "TOKEN_INVALID" }
```

### Unprocessable (422)

```json
{ "status": false, "message": "insufficient wallet balance", "error_code": "UNPROCESSABLE_ENTITY" }
```

---

## Auth Header

Endpoint privat wajib header:

```
Authorization: Bearer <token>
```

**Endpoint publik (tanpa token):** `GET /ping`, `GET /` , `GET|POST /api/v1/reviews`,
`GET /api/v1/stores`, `GET /api/v1/stores/:id`, `GET /api/v1/products`, `GET /api/v1/products/:id`.

Token didapat dari `register`/`login` (`data.token`). Otorisasi mengikuti **active role** di token —
bila punya >1 role, pilih dulu via `POST /api/v1/auth/select-role`.
