# Majoo — Revenue Reporting & Auth API (Laravel 13 + JWT + Swagger)

Aplikasi RESTful API berbasis Laravel 13 yang mengimplementasikan spesifikasi tes pemrograman **Majoo**:

- 🔐 **Autentikasi JWT & Registrasi** — Pendaftaran user + merchant atomik (`POST /api/auth/register`) & login JWT via `tymon/jwt-auth`.
- 📊 **Reporting Engine** — Laporan omzet harian per bulan (`SUM(bill_total)`) untuk **Merchant** (default November 2026) dan **Outlet** (default August 2026), dilengkapi paginasi dan otomatis melaporkan revenue `0` pada tanggal tanpa transaksi.
- 🛡️ **Multi-Tenancy & Data Security** — Pembatasan data ketat tingkat tenant: user hanya dapat mengakses data `merchant_id` / `outlet_id` milik akun mereka sendiri (kombinasi Guard JWT + `TenantScope` middleware + Global Scope pada Eloquent Model).
- 📄 **Dokumentasi Interaktif (Swagger / OpenAPI 3.0)** — Dokumentasi UI interaktif terintegrasi via `darkaonline/l5-swagger` di `/api/documentation`.
- ⚡ **Optimasi Database** — Dokumentasi DML dan strategi komposit indeks pada database MySQL/SQLite (lihat [`DATABASE.md`](DATABASE.md)).

---

## 🛠️ Persyaratan Sistem (Requirements)

- **PHP** >= 8.3 (dengan ekstensi `pdo_sqlite` untuk local dev / `pdo_mysql` untuk produksi)
- **Composer** >= 2.x
- **MySQL** (Opsional untuk dev lokal, SQLite pre-configured di `.env`)

---

## 🚀 Panduan Instalasi (Quickstart Setup)

1. **Clone repository & masuk ke direktori proyek:**
   ```bash
   git clone <repository-url>
   cd laravel-app
   ```

2. **Install dependensi Composer:**
   ```bash
   composer install
   ```

3. **Konfigurasi Environment File:**
   ```bash
   cp .env.example .env
   php artisan key:generate
   php artisan jwt:secret
   ```

4. **Migrasi Database & Seed Sample Data:**
   ```bash
   php artisan migrate --seed
   ```

5. **Generate Dokumentasi Swagger:**
   ```bash
   php artisan l5-swagger:generate
   ```

6. **Jalankan Development Server:**
   ```bash
   php artisan serve
   ```
   Aplikasi dapat diakses di `http://127.0.0.1:8000`.

---

## 📖 Dokumentasi Swagger UI (OpenAPI 3.0)

Aplikasi dilengkapi dengan dokumentasi interaktif **Swagger UI**.

- **URL Swagger UI**: `http://127.0.0.1:8000/api/documentation`
- **Re-generate Spec OpenAPI**:
  ```bash
  php artisan l5-swagger:generate
  ```

Dengan Swagger UI, Anda dapat langsung mencoba (*Try It Out*) seluruh endpoint, menguji autentikasi JWT Bearer token, serta melihat skema request & response secara visual.

---

## 📡 Daftar Endpoint API (API Endpoints)

### 1. Summary Endpoint

| Method | Endpoint | Auth | Deskripsi |
|--------|----------|------|-----------|
| `POST` | `/api/auth/register` | Public | Pendaftaran User + Merchant baru |
| `POST` | `/api/auth/login` | Public | Login akun & dapatkan token JWT |
| `POST` | `/api/auth/logout` | JWT | Invalidate token JWT saat ini |
| `POST` | `/api/auth/refresh` | JWT | Memperbarui token JWT (exchange expired token) |
| `GET` | `/api/auth/me` | JWT | Mengambil profil user & konteks merchant |
| `GET` | `/api/reports/merchant/daily` | JWT + Tenant | Laporan omzet harian merchant per bulan |
| `GET` | `/api/reports/outlet/daily` | JWT + Tenant | Laporan omzet harian outlet spesifik per bulan |

---

### 2. Detail Spesifikasi Endpoint

#### A. Registrasi Account & Merchant (`POST /api/auth/register`)
- **Request Body:**
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "password_confirmation": "password123",
    "merchant_name": "Toko Serba Ada"
  }
  ```
- **Response Success (201 Created):**
  ```json
  {
    "success": true,
    "message": "Registration successful. Please login to get your access token.",
    "data": {
      "id": 3,
      "name": "John Doe",
      "email": "john@example.com",
      "merchant": {
        "id": 3,
        "merchant_name": "Toko Serba Ada"
      }
    }
  }
  ```

#### B. Login (`POST /api/auth/login`)
- **Request Body:**
  ```json
  {
    "email": "merchant1@example.com",
    "password": "password"
  }
  ```
- **Response Success (200 OK):**
  ```json
  {
    "success": true,
    "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1...",
    "token_type": "bearer",
    "expires_in": 3600
  }
  ```

#### C. Laporan Omzet Merchant (`GET /api/reports/merchant/daily`)
- **Query Parameters:**
  - `month` (int 1-12, default: `11`)
  - `year` (int, default: `2026`)
  - `page` (int, default: `1`)
  - `per_page` (int 1-100, default: `10`)
- **Headers:** `Authorization: Bearer <JWT_TOKEN>`

#### D. Laporan Omzet Outlet (`GET /api/reports/outlet/daily`)
- **Query Parameters:**
  - `outlet_id` (int, **Required**)
  - `month` (int 1-12, default: `8`)
  - `year` (int, default: `2026`)
  - `page` (int, default: `1`)
  - `per_page` (int 1-100, default: `10`)
- **Headers:** `Authorization: Bearer <JWT_TOKEN>`

---

### 3. Kode Response HTTP (Status Codes)

| Status Code | Keterangan |
|-------------|------------|
| `200 OK` | Request berhasil. |
| `201 Created` | Resource berhasil dibuat (misal: Register). |
| `401 Unauthenticated` | Token JWT tidak ada, expired, atau salah kredensial. |
| `403 Forbidden` | Outlet tidak dimiliki oleh merchant autentikasi (Multi-tenancy violation). |
| `422 Unprocessable` | Validasi input gagal (misal: email duplikat / format salah). |
| `500 Server Error` | Unexpected error pada server. |

---

## 🔑 Akun Demo (Demo Accounts)

Tersedia akun bawaan hasil seeder (`php artisan db:seed`):

| Email | Password | Merchant | Outlets Milik Merchant |
|-------|----------|----------|------------------------|
| `merchant1@example.com` | `password` | Merchant 1 | Outlet 1 & Outlet 2 (ID: 1 & 3) |
| `merchant2@example.com` | `password` | Merchant 2 | Outlet 1 (ID: 2) |

---

## 🧪 Pengujian / Testing

Aplikasi ini dilengkapi dengan unit & feature tests menggunakan PHPUnit / Artisan Test.

```bash
php artisan test
```

Contoh pengujian mencakup:
- Pendaftaran akun & merchant baru (`RegisterTest`)
- Validasi email duplikat & konfirmasi password
- Alur login & pergantian token JWT
- Multi-tenancy isolation test

---

## 📚 Dokumentasi Tambahan

- Silakan baca [`DATABASE.md`](DATABASE.md) untuk dokumentasi lengkap sintaks SQL DML yang dihasilkan Eloquent serta penjelasan optimasi B-Tree **Composite Indexes** pada kolom `(merchant_id, created_at)` dan `(outlet_id, created_at)`.

