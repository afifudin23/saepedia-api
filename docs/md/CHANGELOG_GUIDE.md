# Changelog - Putra Sunda Trans API

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial project setup placeholder

---

## [0.1.0] - 2025-01-08

### Detail Versi 0.1.0

#### 🚀 Initial Setup & Configuration

- **Deskripsi:**
  - **FastAPI Project Setup:** Inisialisasi project FastAPI dengan struktur folder modular
  - **Database Configuration:** Setup PostgreSQL/MySQL dengan SQLAlchemy ORM
  - **API Documentation:** Auto-generated OpenAPI (Swagger) documentation di `/docs`
  - **Environment Configuration:** Setup `.env` untuk database credentials dan configuration
  - **CORS Configuration:** Setup CORS middleware untuk frontend integration
  - **Authentication Base:** Setup JWT token authentication system

#### 🛠️ Technical Setup

- **Deskripsi:**
  - **Alembic Migration:** Setup database migration tool untuk version control
  - **Pydantic Models:** Setup request/response validation schemas
  - **Router Structure:** Modular router setup untuk maintainability
  - **Error Handling:** Global exception handler dan custom error responses
  - **Logging:** Setup structured logging dengan Loguru

#### 📊 Database Schema

- **Deskripsi:**
  - **Contacts Table:** Schema untuk menyimpan data kontak
    - Fields: id, name, email, phone, company, created_at, updated_at
  - **Users Table:** Schema untuk authentication
    - Fields: id, email, hashed_password, is_active, created_at

---

## Template untuk Update Selanjutnya

Gunakan template berikut saat menambahkan perubahan baru:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Detail Versi X.Y.Z

#### 📦 Kategori Perubahan (pilih yang sesuai)

- **Deskripsi:**
  - **Fitur/Fix Name:** Penjelasan detail tentang perubahan
  - **Impact:** Dampak perubahan terhadap sistem atau API consumers
  - **Technical Notes:** Catatan teknis jika diperlukan
  - **Migration Required:** Ya/Tidak (jika ada database migration)
```

### Kategori yang Tersedia

- `✨ Fitur Baru` - Penambahan endpoint atau fitur baru
- `🐛 Bug Fix` - Perbaikan bug
- `🚀 Peningkatan Performa` - Optimasi performa API
- `🔒 Security Fix` - Perbaikan keamanan
- `📝 Documentation` - Update dokumentasi API
- `♻️ Refactor` - Refactoring code tanpa mengubah fungsionalitas
- `🗃️ Database` - Perubahan skema atau migrasi database
- `🔧 Configuration` - Perubahan konfigurasi
- `🧪 Testing` - Penambahan atau update test
- `🔨 Breaking Changes` - Perubahan yang break backward compatibility

---

## Contoh Entry Changelog

### Contoh 1: Penambahan Fitur Baru (MINOR Version)

```markdown
## [1.2.0] - 2025-01-15

### Detail Versi 1.2.0

#### ✨ Bulk Contact Import API

- **Deskripsi:**
  - **New Endpoint:** `POST /api/v1/contacts/bulk-import`
  - **Request Body:**
    ```json
    {
      "contacts": [
        {"name": "John Doe", "email": "john@example.com", "phone": "+1234567890"},
        {"name": "Jane Smith", "email": "jane@example.com", "phone": "+0987654321"}
      ]
    }
    ```
  - **Response:** Returns job_id untuk tracking progress import
  - **Validation:** Email format, unique constraint, phone number format
  - **Performance:** Batch insert dengan chunk 1000 records
  - **Impact:** Memungkinkan import ribuan kontak sekaligus dengan efficient processing

#### 🚀 Peningkatan Performa

- **Deskripsi:**
  - **Database Indexing:** Menambahkan index pada kolom `email` dan `phone` untuk faster search
  - **Query Optimization:** Refactor query `/contacts` dengan pagination optimization
  - **Caching:** Implementasi Redis caching untuk frequently accessed data
  - **Impact:** Response time berkurang dari 500ms menjadi 50ms untuk list contacts

#### 🗃️ Database Migration

- **Migration Required:** ✅ Yes
- **Migration Command:**
  ```bash
  alembic upgrade head
```

- **Changes:**
  - Added index on `contacts.email`
  - Added index on `contacts.phone`

```

### Contoh 2: Bug Fix (PATCH Version)

```markdown
## [1.1.1] - 2025-01-10

### Detail Versi 1.1.1

#### 🐛 Bug Fix Contact Creation

- **Deskripsi:**
  - **Duplicate Email Validation:** Fix issue dimana duplicate email masih bisa tersimpan
  - **Phone Number Validation:** Perbaikan regex untuk international phone format
  - **Error Response:** Mengembalikan proper 409 Conflict untuk duplicate entries
  - **Transaction Rollback:** Memastikan rollback saat validation error
  - **Impact:** Mencegah data inconsistency di database

#### 🔒 Security Fix

- **Deskripsi:**
  - **SQL Injection:** Fix potential SQL injection di search query
  - **Password Hashing:** Update bcrypt rounds dari 10 ke 12 untuk stronger hashing
  - **Rate Limiting:** Implementasi rate limiting 100 requests per minute per IP
  - **Impact:** Meningkatkan keamanan API dari common vulnerabilities

#### 📝 Documentation

- **Deskripsi:**
  - **Swagger Examples:** Menambahkan request/response examples di Swagger UI
  - **Error Codes:** Dokumentasi lengkap untuk semua error codes
  - **Authentication Guide:** Step-by-step guide untuk JWT authentication
```

### Contoh 3: Breaking Changes (MAJOR Version)

```markdown
## [2.0.0] - 2025-02-01

### Detail Versi 2.0.0

#### 🔨 Breaking Changes

- **Deskripsi:**
  - **API Versioning:**
    - **BREAKING:** Base path berubah dari `/api/contacts` ke `/api/v2/contacts`
    - Old endpoints tetap available di `/api/v1/` sampai Q2 2025
  
  - **Response Structure:**
    - **BREAKING:** Unified response format untuk consistency
    - Old:
      ```json
      {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
      }
      ```
    - New:
      ```json
      {
        "success": true,
        "data": {
          "id": 1,
          "name": "John Doe",
          "email": "john@example.com"
        },
        "meta": {
          "timestamp": "2025-02-01T10:00:00Z"
        }
      }
      ```
  
  - **Contact Model:**
    - **BREAKING:** Field `phone` split menjadi `phone_primary` dan `phone_secondary`
    - Field `company` sekarang required (NOT NULL)
    - Field `tags` ditambahkan (JSON array)

#### ✨ Fitur Baru

- **Deskripsi:**
  - **Contact Groups API:** CRUD endpoints untuk contact grouping
    - `POST /api/v2/groups` - Create group
    - `GET /api/v2/groups` - List groups
    - `POST /api/v2/groups/{id}/contacts` - Add contacts to group
  
  - **Advanced Search:** Full-text search dengan filters
    - `GET /api/v2/contacts/search?q=john&company=acme&tags=vip`
  
  - **Export API:** Export contacts ke CSV/Excel
    - `GET /api/v2/contacts/export?format=csv`

#### 🗃️ Database Migration

- **Migration Required:** ✅ Yes
- **Migration Command:**
  ```bash
  # Backup database first!
  pg_dump supercontact > backup_$(date +%Y%m%d).sql
  
  # Run migration
  alembic upgrade head
```

- **Changes:**

  ```sql
  -- Split phone field
  ALTER TABLE contacts ADD COLUMN phone_primary VARCHAR(20);
  ALTER TABLE contacts ADD COLUMN phone_secondary VARCHAR(20);
  UPDATE contacts SET phone_primary = phone;
  ALTER TABLE contacts DROP COLUMN phone;

  -- Make company required
  UPDATE contacts SET company = 'N/A' WHERE company IS NULL;
  ALTER TABLE contacts ALTER COLUMN company SET NOT NULL;

  -- Add tags field
  ALTER TABLE contacts ADD COLUMN tags JSONB DEFAULT '[]';

  -- Create groups table
  CREATE TABLE contact_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
  );
  ```

#### 📝 Migration Guide

**For API Consumers:**

1. Update base URL:

   ```python
   # Old
   BASE_URL = "https://supercontact-api.com/api/contacts"

   # New
   BASE_URL = "https://supercontact-api.com/api/v2/contacts"
   ```
2. Update response parsing:

   ```python
   # Old
   response = requests.get(f"{BASE_URL}/1")
   contact = response.json()  # Direct access
   name = contact["name"]

   # New
   response = requests.get(f"{BASE_URL}/1")
   result = response.json()
   contact = result["data"]  # Access via 'data' key
   name = contact["name"]
   ```
3. Update contact model:

   ```python
   # Old
   contact = {
       "phone": "+1234567890",
       "company": None  # Optional
   }

   # New
   contact = {
       "phone_primary": "+1234567890",
       "phone_secondary": None,  # Optional
       "company": "Acme Corp"  # Required!
   }
   ```

**For Database Admins:**

1. Schedule maintenance window (estimated 30 minutes for 1M records)
2. Create backup before migration
3. Run migration script
4. Verify data integrity:

   ```sql
   SELECT COUNT(*) FROM contacts WHERE phone_primary IS NULL;
   SELECT COUNT(*) FROM contacts WHERE company IS NULL;
   ```

#### ⚠️ Deprecation Notice

- `/api/v1/*` endpoints will be deprecated on **June 1, 2025**
- Start migrating to `/api/v2/*` as soon as possible
- V1 will return deprecation warning header: `X-API-Deprecated: true`

```

---

## Versioning Guidelines

### Kapan Increment Version?

**MAJOR (X.0.0):**
- Breaking changes di API contract (endpoint path, response structure)
- Database schema changes yang tidak backward-compatible
- Removal of deprecated endpoints
- Changes requiring migration atau action dari API consumers
- Perubahan authentication/authorization mechanism

**MINOR (0.X.0):**
- Penambahan endpoint baru (backward-compatible)
- Penambahan optional fields di request/response
- New features yang tidak break existing functionality
- Database schema additions (new tables, optional columns)
- Performance improvements yang significant

**PATCH (0.0.X):**
- Bug fixes
- Security patches
- Performance improvements tanpa API changes
- Documentation updates
- Internal refactoring tanpa API changes
- Logging dan monitoring improvements

---

## Database Migration Checklist

Untuk setiap perubahan yang memerlukan migration:

- [ ] Alembic migration script sudah dibuat
- [ ] Migration tested di development environment
- [ ] Rollback script sudah ditest
- [ ] Backup strategy sudah documented
- [ ] Estimated downtime sudah dihitung
- [ ] Migration guide untuk API consumers sudah dibuat
- [ ] Team sudah di-notify minimal 1 minggu sebelumnya

---

## Changelog Maintenance

### Best Practices

1. **Update setiap deploy ke production** - Document semua changes
2. **Include database changes** - Selalu dokumentasikan schema changes
3. **Provide migration guides** - Untuk breaking changes, sertakan step-by-step
4. **Version API endpoints** - Gunakan versioning untuk backward compatibility
5. **Deprecation warnings** - Kasih minimum 3 bulan notice sebelum remove endpoints
6. **Document breaking changes clearly** - Gunakan badge 🔨 dan BREAKING prefix
7. **Include performance impacts** - Dokumentasikan perubahan response time

### Bad Examples ❌

```markdown
## [1.2.0] - 2025-01-15
- Added search
- Fixed bugs
- Updated database
```

### Good Examples ✅

```markdown
## [1.2.0] - 2025-01-15

### Detail Versi 1.2.0

#### ✨ Advanced Search API

- **Deskripsi:**
  - **New Endpoint:** `GET /api/v1/contacts/search`
  - **Query Parameters:**
    - `q`: Full-text search (name, email, company)
    - `company`: Filter by company name
    - `tags`: Filter by tags (comma-separated)
    - `limit`: Results per page (default 20, max 100)
    - `offset`: Pagination offset
  - **Response Time:** Average 80ms for 100K records
  - **Impact:** Memungkinkan user mencari kontak dengan multiple criteria
  - **Example Request:**
    ```bash
    GET /api/v1/contacts/search?q=john&company=acme&limit=10
    ```

#### 🗃️ Database Migration

- **Migration Required:** ✅ Yes
- **Command:** `alembic upgrade head`
- **Changes:** Added GIN index on contacts for full-text search
- **Downtime:** ~5 minutes for index creation
```

---

## Version History Reference

| Version | Date       | Type    | Description                               |
| ------- | ---------- | ------- | ----------------------------------------- |
| 0.1.0   | 2025-01-08 | Initial | Project setup, basic CRUD, authentication |

---

## API Deprecation Policy

When deprecating endpoints:

1. **Announce** deprecation in changelog with target removal date (minimum 3 months)
2. **Add header** `X-API-Deprecated: true` to deprecated endpoints
3. **Update docs** with deprecation notice and migration path
4. **Monitor usage** of deprecated endpoints
5. **Remove** only after grace period and usage drops to near-zero

---

**Note:** Changelog ini akan terus diupdate seiring development. Untuk breaking changes, selalu provide migration guide dan minimum 3 bulan deprecation period.
