# Standard API Response

## Format Umum

Semua response menggunakan struktur yang sama:

```json
{
  "status": true | false,
  "data": { ... } | null,
  "list_data": [ ... ] | null,
  "message": "...",
  "error": { ... } | null,
  "error_code": "..." | null,
  "pagination": { ... } | null
}
```

---

## Success Responses

### GET — Single Object
```json
{
  "status": true,
  "data": {
    "id": "uuid",
    "field": "value"
  },
  "message": ""
}
```

### GET — List
```json
{
  "status": true,
  "list_data": [
    { "id": "uuid", "field": "value" },
    { "id": "uuid", "field": "value" }
  ],
  "message": "",
  "pagination": {
    "page": "1",
    "per_page": "10",
    "total": "100",
    "total_pages": "10"
  }
}
```

### POST / PUT / PATCH — Return ID
```json
{
  "status": true,
  "data": { "id": "uuid" },
  "message": ""
}
```

### DELETE
```json
{
  "status": true,
  "data": { "deleted_count": 2 },
  "message": ""
}
```

---

## Error Responses

### Validation Error (400)
```json
{
  "status": false,
  "data": null,
  "message": "Validation failed",
  "error": {
    "field_name": "error message"
  },
  "error_code": "VALIDATION_ERROR"
}
```

### Bad Request (400)
```json
{
  "status": false,
  "data": null,
  "message": "...",
  "error": null,
  "error_code": "BAD_REQUEST"
}
```

### Unauthorized (401)
```json
{
  "status": false,
  "data": null,
  "message": "...",
  "error": null,
  "error_code": "UNAUTHORIZED"
}
```

### Forbidden (403)
```json
{
  "status": false,
  "data": null,
  "message": "...",
  "error": null,
  "error_code": "FORBIDDEN"
}
```

### Not Found (404)
```json
{
  "status": false,
  "data": null,
  "message": "...",
  "error": null,
  "error_code": "NOT_FOUND"
}
```

### Internal Server Error (500)
```json
{
  "status": false,
  "data": null,
  "message": "Internal server error",
  "error": null,
  "error_code": "INTERNAL_SERVER_ERROR"
}
```

---

## Auth Header

Semua endpoint kecuali `/auth/login`, `/auth/register`, `/auth/forgot-password`, `/auth/reset-password` wajib menyertakan:

```
Authorization: Bearer <access_token>
```
