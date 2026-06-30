# SEAPEDIA API — Daftar Endpoint

Ringkasan seluruh endpoint REST API SEAPEDIA (Level 1–7).

- **Base URL:** `http://localhost:5000/api/v1` (kecuali `/ping` yang ada di root `/`)
- **Autentikasi:** kirim header `Authorization: Bearer <token>` untuk endpoint privat.
- **Otorisasi:** mengikuti **active role** pada JWT (bukan sekadar daftar role yang dimiliki).
  Pilih role aktif via `POST /auth/select-role` setelah login bila punya >1 role.
- **Format response:** lihat [README — Standard API Response](../README.md#standard-api-response).
- **Pagination:** endpoint list menerima query `?page=` & `?per_page=` (default `1` & `10`, maks `100`).

Daftar isi:

1. [Publik (Guest)](#1-publik-guest)
2. [Auth](#2-auth)
3. [Buyer](#3-buyer-active-role--buyer)
4. [Seller](#4-seller-active-role--seller)
5. [Driver](#5-driver-active-role--driver)
6. [Admin](#6-admin-active-role--admin)

---

## 1. Publik (Guest)

Bisa diakses tanpa login.

| Method | Path              | Keterangan                             |
| ------ | ----------------- | -------------------------------------- |
| GET    | `/ping`         | Health check (root, bukan`/api/v1`)  |
| GET    | `/reviews`      | List review aplikasi                   |
| POST   | `/reviews`      | Kirim review aplikasi                  |
| GET    | `/stores`       | List toko                              |
| GET    | `/stores/:id`   | Detail toko                            |
| GET    | `/products`     | Katalog produk (mendukung`?search=`) |
| GET    | `/products/:id` | Detail produk + info toko              |

**POST `/reviews`**

```json
{ "reviewer_name": "Andi", "rating": 5, "comment": "Aplikasinya keren!" }
```

---

## 2. Auth

| Method | Path                      | Auth  | Keterangan                                     |
| ------ | ------------------------- | ----- | ---------------------------------------------- |
| POST   | `/auth/register`        | —    | Registrasi user                                |
| POST   | `/auth/login`           | —    | Login, dapat token + daftar role + active role |
| POST   | `/auth/select-role`     | token | Pilih role aktif → token baru                 |
| GET    | `/auth/me`              | token | Profil + roles + active role                   |
| GET    | `/auth/balance-summary` | token | Ringkasan saldo lintas role                    |
| POST   | `/auth/logout`          | token | Invalidasi token (denylist)                    |

**POST `/auth/register`** — `roles` opsional (default `["buyer"]`)

```json
{
  "email": "newuser@mail.com",
  "password": "Password123",
  "confirm_password": "Password123",
  "roles": ["buyer", "seller"]
}
```

**POST `/auth/login`** — login memakai email

```json
{ "email": "buyer1@seapedia.test", "password": "Buyer123" }
```

Respons memuat `token`, `user.roles`, `active_role`, dan `need_role_selection`
(`true` bila user punya >1 role non-admin → wajib `select-role` dulu).

**POST `/auth/select-role`**

```json
{ "role": "seller" }
```

---

## 3. Buyer (active role = `buyer`)

### Wallet

| Method | Path                           | Keterangan                              |
| ------ | ------------------------------ | --------------------------------------- |
| GET    | `/buyer/wallet`              | Saldo wallet                            |
| POST   | `/buyer/wallet/topup`        | Top-up dummy (`{ "amount": 500000 }`) |
| GET    | `/buyer/wallet/transactions` | Riwayat transaksi wallet                |

### Alamat Pengiriman

| Method | Path                     | Keterangan    |
| ------ | ------------------------ | ------------- |
| GET    | `/buyer/addresses`     | List alamat   |
| POST   | `/buyer/addresses`     | Tambah alamat |
| PUT    | `/buyer/addresses/:id` | Ubah alamat   |
| DELETE | `/buyer/addresses/:id` | Hapus alamat  |

```json
// POST/PUT /buyer/addresses
{
  "recipient_name": "Budi",
  "phone": "08123456789",
  "full_address": "Jl. Mawar No. 1",
  "is_primary": true
}
```

### Cart (single-store rule)

| Method | Path                             | Keterangan                                                 |
| ------ | -------------------------------- | ---------------------------------------------------------- |
| GET    | `/buyer/cart`                  | Lihat ringkasan cart                                       |
| POST   | `/buyer/cart/items`            | Tambah produk (`{ "product_id": "...", "quantity": 2 }`) |
| PUT    | `/buyer/cart/items/:productID` | Ubah qty (`{ "quantity": 3 }`)                           |
| DELETE | `/buyer/cart/items/:productID` | Hapus 1 item                                               |
| DELETE | `/buyer/cart`                  | Kosongkan cart                                             |

> Satu cart hanya boleh berisi produk dari **satu toko**. Menambah produk toko lain → `409 CONFLICT`.

### Checkout & Order

| Method | Path                        | Keterangan                              |
| ------ | --------------------------- | --------------------------------------- |
| POST   | `/buyer/checkout/preview` | Hitung ringkasan biaya sebelum checkout |
| POST   | `/buyer/checkout`         | Buat order dari cart                    |
| GET    | `/buyer/orders`           | Riwayat order                           |
| GET    | `/buyer/orders/:id`       | Detail order + timeline status          |
| GET    | `/buyer/reports`          | Ringkasan pengeluaran                   |

```json
// POST /buyer/checkout/preview
{ "delivery_method": "regular", "discount_code": "SEAPEDIA10" }

// POST /buyer/checkout  (discount_code opsional)
{ "address_id": "<uuid>", "delivery_method": "regular", "discount_code": "SEAPEDIA10" }
```

`delivery_method`: `instant` (Rp20.000) · `next_day` (Rp10.000) · `regular` (Rp5.000).
PPN 12% dihitung dari `(subtotal − discount)`.

---

## 4. Seller (active role = `seller`)

### Toko

| Method | Path              | Keterangan         |
| ------ | ----------------- | ------------------ |
| GET    | `/seller/store` | Toko milik sendiri |
| POST   | `/seller/store` | Buat toko          |
| PUT    | `/seller/store` | Ubah toko          |

```json
// POST/PUT /seller/store  (nama toko WAJIB unik)
{ "name": "Toko Maju Jaya", "description": "Toko serba ada" }
```

### Produk

| Method | Path                     | Keterangan                 |
| ------ | ------------------------ | -------------------------- |
| GET    | `/seller/products`     | Produk milik toko sendiri  |
| POST   | `/seller/products`     | Tambah produk              |
| PUT    | `/seller/products/:id` | Ubah produk milik sendiri  |
| DELETE | `/seller/products/:id` | Hapus produk milik sendiri |

```json
// POST/PUT /seller/products  (image_url opsional, harus URL valid)
{ "name": "Sepatu Lari", "description": "Ringan & nyaman", "price": 350000, "stock": 10, "image_url": "https://images.unsplash.com/photo-..." }
```

### Order & Laporan

| Method | Path                           | Keterangan                                                 |
| ------ | ------------------------------ | ---------------------------------------------------------- |
| GET    | `/seller/orders`             | Daftar order masuk                                         |
| GET    | `/seller/orders/:id`         | Detail order                                               |
| POST   | `/seller/orders/:id/process` | Proses order:**Sedang Dikemas → Menunggu Pengirim** |
| GET    | `/seller/reports`            | Ringkasan pendapatan                                       |

---

## 5. Driver (active role = `driver`)

| Method | Path                          | Keterangan                                       |
| ------ | ----------------------------- | ------------------------------------------------ |
| GET    | `/driver/jobs`              | Job tersedia (status**Menunggu Pengirim**) |
| GET    | `/driver/jobs/:id`          | Detail job                                       |
| POST   | `/driver/jobs/:id/take`     | Ambil job →**Sedang Dikirim**             |
| POST   | `/driver/jobs/:id/complete` | Selesaikan job →**Pesanan Selesai**       |
| GET    | `/driver/dashboard`         | Job aktif, riwayat, total earning                |

> Pendapatan driver = **80% dari ongkir**, dikunci saat job selesai.

---

## 6. Admin (active role = `admin`)

### Monitoring

| Method | Path                      | Keterangan                                                                                |
| ------ | ------------------------- | ----------------------------------------------------------------------------------------- |
| GET    | `/admin/summary`        | Ringkasan: users, stores, products, orders (per status), voucher/promo, delivery, overdue |
| GET    | `/admin/users`          | Monitoring users + roles                                                                  |
| GET    | `/admin/stores`         | Monitoring toko                                                                           |
| GET    | `/admin/products`       | Monitoring produk                                                                         |
| GET    | `/admin/orders`         | Monitoring order                                                                          |
| GET    | `/admin/deliveries`     | Monitoring pengiriman                                                                     |
| GET    | `/admin/overdue-orders` | Order yang sedang overdue (waktu virtual)                                                 |

### Voucher & Promo

| Method | Path                    | Keterangan       |
| ------ | ----------------------- | ---------------- |
| POST   | `/admin/vouchers`     | Generate voucher |
| GET    | `/admin/vouchers`     | List voucher     |
| GET    | `/admin/vouchers/:id` | Detail voucher   |
| POST   | `/admin/promos`       | Generate promo   |
| GET    | `/admin/promos`       | List promo       |
| GET    | `/admin/promos/:id`   | Detail promo     |

```json
// POST /admin/vouchers  (promo sama, tanpa "usage_limit")
{
  "code": "NEWYEAR",
  "discount_type": "percent",      // "percent" atau "fixed"
  "discount_value": 15,
  "max_discount": 30000,            // cap untuk percent (opsional)
  "min_spend": 100000,              // minimal belanja (opsional)
  "expires_at": "2027-01-01T00:00:00Z",
  "usage_limit": 50                 // hanya voucher
}
```

### Simulasi Waktu & Overdue

| Method | Path                            | Keterangan                                                  |
| ------ | ------------------------------- | ----------------------------------------------------------- |
| GET    | `/admin/simulate/now`         | Lihat waktu virtual saat ini                                |
| POST   | `/admin/simulate/advance-day` | Majukan waktu N hari + jalankan overdue (`{ "days": 1 }`) |
| POST   | `/admin/overdue/run`          | Jalankan overdue handling secara manual                     |

> SLA overdue: Instant **1 hari** · Next Day **2 hari** · Regular **3 hari**. Order yang
> melewati SLA dan belum selesai akan **auto-refund** ke wallet buyer + stok dipulihkan +
> status jadi **Dikembalikan**.

---

## Akun Demo (setelah `make seed`)

Login memakai **email**. Password mengikuti pola `Role123`:

| Email                     | Password     | Role                                 |
| ------------------------- | ------------ | ------------------------------------ |
| `admin@seapedia.test`   | `Admin123`  | Admin                                |
| `seller1@seapedia.test` | `Seller123` | Seller (Toko Sumber Rejeki)          |
| `seller2@seapedia.test` | `Seller123` | Seller (Elektronik Jaya)             |
| `seller3@seapedia.test` | `Seller123` | Seller (Dapur Sehat)                 |
| `buyer1@seapedia.test`  | `Buyer123`  | Buyer (saldo Rp1.000.000)            |
| `buyer2@seapedia.test`  | `Buyer123`  | Buyer (saldo Rp500.000)              |
| `buyer3@seapedia.test`  | `Buyer123`  | Buyer (saldo Rp2.000.000)            |
| `driver1@seapedia.test` | `Driver123` | Driver                               |
| `driver2@seapedia.test` | `Driver123` | Driver                               |
| `multi1@seapedia.test`  | `Multi123`  | Buyer + Seller + Driver (multi-role) |

Diskon demo (10): voucher `SEAPEDIA10` `GROCERY15` `NEWUSER20` `FLASH25` `HEMAT50K` ·
promo `HEMAT5K` `POTONG10K` `POTONG20K` `PROMO10` `GAJIAN`.

> Koleksi siap-impor: [`SEAPEDIA.postman_collection.json`](SEAPEDIA.postman_collection.json).

