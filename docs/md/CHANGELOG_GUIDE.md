# Panduan Changelog — SEAPEDIA API

Panduan cara menulis entri di [`CHANGELOG.md`](../../CHANGELOG.md) (root project).
Format mengikuti [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) +
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> Stack project: **Go · Gin · GORM · PostgreSQL · golang-migrate**. Migrasi dijalankan dengan
> `make migrate-up` (bukan tool lain). Versi rilis ada di `config/version.go` (`const Version`) dan
> di-bump otomatis lewat `make release`.

---

## Format Entri

Tiap rilis ditulis seperti ini di `CHANGELOG.md`:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Detail Versi X.Y.Z

#### <emoji> Kategori

- **Deskripsi:**
  - **Nama fitur/fix:** penjelasan singkat
  - **Impact:** dampak ke API consumer / sistem
  - **Migration Required:** Ya/Tidak (`make migrate-up` bila ada migrasi baru)

---
```

Diakhiri `---` agar mudah diparse (release manager membaca blok antar-versi).

### Kategori

`✨ Fitur Baru` · `🐛 Bug Fix` · `🚀 Performa` · `🔒 Security` · `📝 Documentation` ·
`♻️ Refactor` · `🗃️ Database` · `🔧 Configuration` · `🧪 Testing` · `🔨 Breaking Changes`

---

## Contoh

```markdown
## [0.2.0] - 2026-07-01

### Detail Versi 0.2.0

#### ✨ Wishlist Buyer

- **Deskripsi:**
  - **Endpoint baru:** `POST /api/v1/buyer/wishlist`, `GET /api/v1/buyer/wishlist`
  - **Impact:** buyer bisa menyimpan produk favorit
  - **Migration Required:** Ya → tabel `wishlists` (`make migrate-up`)

#### 🐛 Bug Fix Checkout

- **Deskripsi:**
  - **Stok race:** perketat `UPDATE ... WHERE stock >= qty` agar stok tak pernah negatif
  - **Impact:** mencegah oversell saat checkout bersamaan
```

---

## Kapan Naik Versi (SemVer)

- **MAJOR (X.0.0)** — breaking: ubah kontrak API, hapus endpoint, ganti mekanisme auth, perubahan
  skema tak backward-compatible.
- **MINOR (0.X.0)** — fitur/endpoint baru yang backward-compatible, tambah kolom opsional.
- **PATCH (0.0.X)** — bug fix, security patch, dokumentasi, refactor tanpa ubah API.

Cara bump: `make release` (memperbarui `const Version` di `config/version.go`, commit, tag, push).

---

## Checklist Saat Rilis

- [ ] Entri `## [X.Y.Z]` sudah ditambahkan di `CHANGELOG.md` (release manager mensyaratkan ini).
- [ ] Migrasi baru (jika ada) sudah ada pasangan up/down & teruji (`make migrate-up` / `migrate-down`).
- [ ] `go build ./...` & `go vet ./...` hijau.
- [ ] `make swag` dijalankan bila ada perubahan endpoint/anotasi.
- [ ] Commit bertahap (hindari satu commit raksasa).

---

## Riwayat Versi

| Version | Tanggal | Keterangan |
| ------- | ------- | ---------- |
| 0.1.0 | 2026-06-29 | Backend Level 1–7: auth/multi-role, store/produk, wallet/cart/checkout, diskon, driver, admin/overdue, security + Swagger |

Detail lengkap tiap versi: lihat [`CHANGELOG.md`](../../CHANGELOG.md).
