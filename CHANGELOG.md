# Changelog - SEAPEDIA API

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.1] - 2026-06-30

### Detail Versi 0.1.1

Perbaikan & penyempurnaan setelah 0.1.0 (tanpa perubahan kontrak API).

#### 🐛 Bug Fix

- **Deskripsi:**
    - **Release manager:** regex bump versi di-anchor ke `const Version` agar tidak ikut menimpa `const APIVersion`
    - **Swagger base URL:** hapus `@host`/`@schemes` → URL relatif, "Try it out" jalan via HTTP (bukan paksa HTTPS)
    - **Logger:** auto-init saat belum di-`Init()` agar `database.Connect()` dari script seed tidak panic (nil logger)
    - **.env:** nilai tanpa tanda kutip agar perintah `make migrate-*` (PowerShell) terbentuk benar

#### ♻️ Refactor

- **Deskripsi:**
    - **Versi terpisah:** pindah ke `config/version.go` (`Version` semver + `APIVersion` path + `APIBasePath()`)
    - **Swagger version:** `info.version` diisi dari `config.Version` saat runtime (bukan hardcode anotasi)

#### ✨ Penyempurnaan

- **Deskripsi:**
    - **Swagger UI:** disajikan di `/docs/v1` (version-aware) + redirect `/docs`, `/docs/v1`, `/docs/v1/` ke `index.html`
    - **Root & Health:** endpoint `/` dan `/ping` mengikuti standard response envelope

#### 📝 Documentation

- **Deskripsi:**
    - **docs/md disamakan dengan project:** API_SUMMARY, API_RESPONSE, DEVELOPER_GUIDE, DEVELOPER_GUIDE_DETAIL, CHANGELOG_GUIDE (sebelumnya masih template project lain)

---

## [0.1.0] - 2026-06-29

### Detail Versi 0.1.0

Backend API marketplace multi-role (Admin, Seller, Buyer, Driver) — implementasi
**Level 1 sampai Level 7** Technical Challenge COMPFEST 18 SEAPEDIA.

#### 🚀 Initial Setup & Configuration

- **Deskripsi:**
    - **Go Project Setup:** Inisialisasi project Go dengan Clean Architecture per modul (domain, repository, usecase, handler, routes)
    - **Database Configuration:** PostgreSQL + GORM ORM, koneksi dengan connection pooling
    - **Environment Configuration:** `.env` untuk kredensial database, port, dan `ACCESS_KEY` (JWT)
    - **CORS Configuration:** Middleware CORS via Gin agar bisa diintegrasi frontend web/mobile

#### 🛠️ Technical Setup

- **Deskripsi:**
    - **Golang-Migrate:** Migrasi database via CLI dengan perintah Makefile (`migrate-up`, `migrate-down`, dll)
    - **Request Validation:** Validasi request memakai go-playground/validator
    - **Router Structure:** Router modular per modul, dirakit di `internal/router`
    - **Error Handling:** Response terpusat dengan typed `ErrorCode` (`pkg/response`)
    - **Logging:** Structured logging dengan zap (`pkg/logger`)
    - **Hot Reload:** Air untuk auto-reload saat development
    - **Transaction Manager:** `pkg/tx` (transaksi berbasis context) — business logic di usecase, repository hanya operasi data
    - **API Documentation:** Swagger/OpenAPI via swaggo + gin-swagger (UI di `/docs/v1`, `make swag`) + Postman collection
    - **Versi & Path:** `config/version.go` — `Version` (semver rilis, dipakai `/ping` & Swagger) dan `APIVersion` (segmen path `/api/v1`)
    - **Seeder:** Data demo (±10 baris/tabel) di `scripts/seed` — `make seed` (semua) / `make seed-one name=<user|discount|review|address|order>` (satuan)
    - **Release Manager:** Script interaktif `scripts/release` untuk bump version, commit, tag, dan push

#### 🔐 Level 1 — Marketplace Publik, Autentikasi, Multi-Role & Review

- **Deskripsi:**
    - **Autentikasi:** Register, login (pakai **email**), logout dengan hashing password **argon2id** + **JWT HS256**
    - **Multi-Role & Active Role:** Satu akun bisa punya banyak role (buyer/seller/driver); otorisasi mengikuti **role aktif** di JWT, dipilih via `/auth/select-role`
    - **Profil & Balance Summary:** Endpoint `/auth/me` dan ringkasan saldo lintas role
    - **Public Review:** Review aplikasi (rating 1–5 + komentar) bisa diisi guest tanpa transaksi
    - **Katalog Publik:** Endpoint produk & toko dapat diakses tanpa login

#### 🏪 Level 2 — Seller: Toko & Produk

- **Deskripsi:**
    - **Store Management:** Seller membuat/mengubah toko (nama **unik**), hanya bisa mengelola toko sendiri
    - **Product CRUD:** Buat/ubah/hapus produk milik sendiri (ownership di-enforce), termasuk `image_url`
    - **Public Catalog:** Katalog produk + detail dari data backend lengkap dengan info toko & gambar

#### 🛒 Level 3 — Buyer: Wallet, Cart & Checkout

- **Deskripsi:**
    - **Wallet:** Saldo, dummy top-up, dan riwayat transaksi
    - **Delivery Address:** CRUD alamat pengiriman
    - **Cart (single-store rule):** Satu cart hanya berisi produk dari satu toko
    - **Checkout & Order:** Hitung subtotal, ongkir per metode, **PPN 12%**, total; order dibuat dalam **satu transaksi atomik** (kurangi stok aman tanpa minus, potong wallet, riwayat status), status awal **Sedang Dikemas**

#### 🎟️ Level 4 — Diskon & Pemrosesan Order Seller

- **Deskripsi:**
    - **Voucher & Promo:** Voucher (punya kuota) & promo (tanpa kuota), tipe percent/fixed, dengan expiry & min belanja; admin generate/list/detail
    - **Diskon di Checkout:** Satu kode per checkout, diskon sebelum PPN, kuota voucher dikurangi atomik
    - **Seller Process Order:** Sedang Dikemas → Menunggu Pengirim
    - **Laporan:** Ringkasan pengeluaran buyer & pendapatan seller

#### 🚚 Level 5 — Pengiriman & Workflow Driver

- **Deskripsi:**
    - **Delivery Job:** Driver mencari job (status Menunggu Pengirim), ambil job (atomik, anti dua driver), dan konfirmasi selesai
    - **Status Lifecycle:** Menunggu Pengirim → Sedang Dikirim → Pesanan Selesai (dengan timestamp)
    - **Driver Earning:** 80% dari ongkir; dashboard job aktif, riwayat, dan total pendapatan

#### 🛡️ Level 6 — Admin Monitoring & Overdue

- **Deskripsi:**
    - **Admin Dashboard:** Monitoring users, stores, products, orders, voucher/promo, delivery, overdue
    - **Simulasi Waktu:** Offset waktu virtual (`pkg/clock` + tabel `app_settings`), maju N hari via admin
    - **Overdue Auto-Refund:** SLA per metode (Instant 1/Next Day 2/Regular 3 hari); order lewat SLA → refund ke wallet + restore stok + status Dikembalikan, idempotent (anti double refund)

#### 🔒 Level 7 — Security Hardening & Finalisasi

- **Deskripsi:**
    - **SQL Injection:** Seluruh query parameterized via GORM
    - **XSS:** Konten user-generated di-escape saat disimpan
    - **Logout Invalidation:** Token denylist (`revoked_tokens` + `jti`) — token logged-out ditolak
    - **RBAC Server-side:** Otorisasi per active role; ownership resource di-enforce
    - **Dokumentasi:** README, Swagger UI (`/docs/v1`), `docs/md/` (API_ENDPOINTS, API_SUMMARY, API_RESPONSE, Developer Guide), dan Postman collection

#### 📊 Database Schema

- **Deskripsi:**
    - `users`, `user_roles` — akun (identitas email) + role
    - `app_reviews` — review aplikasi publik
    - `stores`, `products` — toko & produk seller (produk punya `image_url`)
    - `wallets`, `wallet_transactions`, `addresses` — buyer
    - `carts`, `cart_items` — keranjang (single-store)
    - `orders`, `order_items`, `order_status_histories` — order + riwayat status + field driver
    - `discounts` — voucher & promo
    - `app_settings` — offset simulasi waktu
    - `revoked_tokens` — denylist logout

---
