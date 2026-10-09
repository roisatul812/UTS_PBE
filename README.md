# UTS-PBE_SIAKAD — SIAKAD Mini RESTful API

RESTful API backend untuk **SIAKAD Mini**, layanan akademik sederhana yang mengelola data pengguna (user), mahasiswa (student), mata kuliah (course), dan Kartu Rencana Studi (KRS/enrollment).

Project ini dikembangkan untuk memenuhi Ujian Tengah Semester (UTS) mata kuliah **Pemrograman Backend Lanjut**, Program Studi D4 Teknik Informatika.

---

## 1. Tech Stack

- **Bahasa Pemrograman**: Go (Golang) v1.26.5
- **Web Framework**: Go Fiber v2 (`github.com/gofiber/fiber/v2`)
- **Database**: PostgreSQL 18
- **Database Driver**: `github.com/lib/pq` (Standard Go `database/sql` Library, Non-ORM)
- **Password Hashing**: `golang.org/x/crypto/bcrypt`
- **Autentikasi Token**: JWT v5 (`github.com/golang-jwt/jwt/v5`)
- **Konfigurasi Environment**: `github.com/joho/godotenv`

---

## 2. Struktur Database

Database terdiri dari 4 tabel utama:

### 1. `users`
| Kolom | Tipe Data | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik user |
| `email` | VARCHAR(255) UNIQUE NOT NULL | Email login user |
| `password` | VARCHAR(255) NOT NULL | Hash password (bcrypt) |
| `role` | VARCHAR(20) NOT NULL | Peran: `'admin'` atau `'mahasiswa'` |
| `created_at` | TIMESTAMPTZ | Waktu dibuat |
| `updated_at` | TIMESTAMPTZ | Waktu diperbarui |

### 2. `students`
| Kolom | Tipe Data | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik mahasiswa |
| `user_id` | INT UNIQUE NOT NULL | Foreign key ke `users(id)` ON DELETE CASCADE |
| `nim` | VARCHAR(20) UNIQUE NOT NULL | NIM 12 digit mahasiswa |
| `nama` | VARCHAR(255) NOT NULL | Nama lengkap mahasiswa |
| `prodi` | VARCHAR(100) NOT NULL | Program studi |
| `angkatan` | INT NOT NULL | Tahun angkatan (4 digit, $\le$ tahun berjalan) |
| `ipk_terakhir` | NUMERIC(3, 2) DEFAULT 0.00 | IPK terakhir (0.00 - 4.00) |
| `created_at` | TIMESTAMPTZ | Waktu dibuat |
| `updated_at` | TIMESTAMPTZ | Waktu diperbarui |
| `deleted_at` | TIMESTAMPTZ NULL | Timestamp soft delete |

### 3. `courses`
| Kolom | Tipe Data | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik mata kuliah |
| `kode_mk` | VARCHAR(20) UNIQUE NOT NULL | Kode mata kuliah (unik) |
| `nama_mk` | VARCHAR(255) NOT NULL | Nama mata kuliah |
| `sks` | INT NOT NULL CHECK (sks > 0) | Beban SKS mata kuliah |
| `semester` | INT NOT NULL CHECK (semester > 0) | Semester kurikulum |
| `kuota` | INT NOT NULL CHECK (kuota $\ge$ 0) | Kapasitas maksimal kuota mahasiswa |
| `created_at` | TIMESTAMPTZ | Waktu dibuat |
| `updated_at` | TIMESTAMPTZ | Waktu diperbarui |

### 4. `enrollments`
| Kolom | Tipe Data | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik enrollment / KRS |
| `student_id` | INT NOT NULL | Foreign key ke `students(id)` ON DELETE CASCADE |
| `course_id` | INT NOT NULL | Foreign key ke `courses(id)` ON DELETE CASCADE |
| `tahun_akademik` | VARCHAR(30) NOT NULL | Tahun akademik (contoh: `2026/2027-Ganjil`) |
| `created_at` | TIMESTAMPTZ | Waktu pendaftaran |

**Constraint & Index Unik**:
- `UNIQUE (student_id, course_id, tahun_akademik)`

---

## 3. Struktur Project

```
UTS-PBE_SIAKAD/
├── app/
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── student_handler.go
│   │   ├── course_handler.go
│   │   └── enrollment_handler.go
│   ├── model/
│   │   ├── user.go
│   │   ├── student.go
│   │   ├── course.go
│   │   ├── enrollment.go
│   │   └── response.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── student_repository.go
│   │   ├── course_repository.go
│   │   └── enrollment_repository.go
│   └── service/
│       ├── auth_service.go
│       ├── student_service.go
│       ├── course_service.go
│       └── enrollment_service.go
├── config/
│   └── config.go
├── database/
│   ├── postgres.go
│   ├── migration.go
│   ├── seeder.go
│   └── database_test.go
├── helper/
│   ├── response.go
│   └── jwt.go
├── middleware/
│   ├── auth.go
│   └── rate_limiter.go
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_students.sql
│   ├── 003_create_courses.sql
│   └── 004_create_enrollments.sql
├── route/
│   └── route.go
├── test/
│   ├── setup_test.go
│   ├── auth_test.go
│   ├── student_test.go
│   ├── course_test.go
│   └── enrollment_test.go
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── test_e2e.ps1
└── README.md
```

---

## 4. Konfigurasi Environment (`.env`)

Salin file `.env.example` menjadi `.env`:

```bash
cp .env.example .env
```

Sesuaikan konfigurasi koneksi database PostgreSQL dan JWT:

```env
PORT=3000
DATABASE_URL=postgres://postgres:password@localhost:5432/siakad_mini?sslmode=disable
JWT_SECRET=super-secret-key-siakad-mini-2026
JWT_EXPIRES_IN=86400
RATE_LIMIT_LOGIN_MAX=5
RATE_LIMIT_LOGIN_EXP=60
```

> **Catatan Keamanan**: File `.env` telah dimasukkan ke dalam `.gitignore` sehingga tidak akan terunggah ke repositori Git.

---

## 5. Cara Menjalankan Project

### A. Persiapan Database
Buat database `siakad_mini` di PostgreSQL:
```sql
CREATE DATABASE siakad_mini;
```

### B. Menjalankan Server & Migration Otomatis
Jalankan aplikasi Go:
```bash
go run main.go
```
*Aplikasi secara otomatis membaca folder `migrations/` dan mengeksekusi semua migration SQL yang belum dijalankan, serta melakukan seeding awal jika tabel `users` masih kosong.*

### C. Menjalankan Seeder Secara Eksplisit
Untuk menjalankan seeder ulang secara manual:
```bash
go run main.go -seed
```

**Data Seeder Bawaan**:
- **1 Akun Admin**:
  - Email: `admin@example.com`
  - Password: `admin1234`
- **20 Akun Mahasiswa**:
  - Email: `rina.putri@student.siakad.ac.id`, `budi.santoso@student.siakad.ac.id`, dst.
  - Password Awal: Sesuai NIM masing-masing (misal: `187221000001`, `187221000002`, dst.) dalam bentuk bcrypt hash.
- **10 Mata Kuliah**:
  - `IF101` (Pemrograman Dasar - 3 SKS - Kuota 30)
  - `IF102` (Struktur Data dan Algoritma - 3 SKS - Kuota 30)
  - `IF201` (Basis Data Lanjut - 3 SKS - Kuota 25)
  - `IF202` (Pemrograman Web - 3 SKS - Kuota 25)
  - `IF301` (Pemrograman Backend Lanjut - 4 SKS - Kuota 20)
  - `IF302` (Rekayasa Perangkat Lunak - 3 SKS - Kuota 20)
  - `IF303` (Jaringan Komputer - 3 SKS - Kuota 25)
  - `IF401` (Kecerdasan Buatan - 3 SKS - Kuota 20)
  - `IF402` (Keamanan Informasi - 3 SKS - Kuota 2)
  - `IF403` (Cloud Computing - 2 SKS - Kuota 1)

---

## 6. Daftar 10 Endpoint Resmi SIAKAD Mini

| No | Method | Endpoint | Akses / Role | Fungsi | Status Sukses |
|---|---|---|---|---|---|
| 1 | `POST` | `/api/v1/auth/login` | Publik | Autentikasi user & mengembalikan access token | `200 OK` |
| 2 | `GET` | `/api/v1/auth/me` | Semua Role | Menampilkan profil user yang sedang login | `200 OK` |
| 3 | `GET` | `/api/v1/students` | Admin | Daftar mahasiswa (pagination, search, filter, sort) | `200 OK` |
| 4 | `POST` | `/api/v1/students` | Admin | Menambah mahasiswa sekaligus user account | `201 Created` |
| 5 | `GET` | `/api/v1/students/:id` | Admin, Mahasiswa (diri sendiri) | Detail mahasiswa beserta total SKS & daftar MK | `200 OK` |
| 6 | `PUT` | `/api/v1/students/:id` | Admin | Memperbarui data mahasiswa (NIM tidak boleh diubah) | `200 OK` |
| 7 | `DELETE` | `/api/v1/students/:id` | Admin | Soft delete mahasiswa (set `deleted_at`) | `204 No Content` |
| 8 | `GET` | `/api/v1/courses` | Semua Role | Daftar mata kuliah beserta sisa kuota (`terisi`, `sisa_kuota`) | `200 OK` |
| 9 | `POST` | `/api/v1/enrollments` | Mahasiswa | Mengambil mata kuliah / menambah KRS (dengan transaction & locking) | `201 Created` |
| 10 | `DELETE` | `/api/v1/enrollments/:id` | Mahasiswa (milik sendiri) | Membatalkan mata kuliah dari KRS | `204 No Content` |

---

## 7. Business Rules KRS

1. **Batas SKS Berdasarkan IPK Terakhir**:
   - $\text{IPK} \ge 3.00 \longrightarrow$ Maksimal **24 SKS**
   - $\text{IPK } 2.50 - 2.99 \longrightarrow$ Maksimal **21 SKS**
   - $\text{IPK} < 2.50 \longrightarrow$ Maksimal **18 SKS**
   *Jika penambahan mata kuliah menyebabkan total SKS melebihi batas, API mengembalikan HTTP 422 dengan pesan spesifik yang menyebutkan sisa SKS yang masih tersedia.*

2. **Duplikasi Mata Kuliah**:
   - Mahasiswa tidak dapat mengambil mata kuliah yang sama dua kali pada tahun akademik yang sama.
   - Pelanggaran menghasilkan status **HTTP 409 Conflict**.

3. **Validasi Kuota**:
   - Mata kuliah yang kuotanya sudah penuh (`terisi >= kuota`) tidak dapat diambil.
   - Menghasilkan status **HTTP 422 Unprocessable Entity**.

4. **Ownership KRS**:
   - Mahasiswa hanya dapat melihat data dirinya sendiri dan hanya dapat membatalkan enrollment miliknya sendiri.
   - Percobaan mengakses atau membatalkan KRS mahasiswa lain menghasilkan **HTTP 403 Forbidden**.

5. **Soft Delete**:
   - Mahasiswa yang di-soft-delete tidak muncul pada daftar `GET /api/v1/students` dan tidak dapat login ke sistem.

6. **Database Transaction & Row Locking**:
   - `POST /api/v1/enrollments` dijalankan dalam 1 database transaction dengan `SELECT ... FOR UPDATE` pada tabel `courses` untuk mencegah race condition / over-enrollment pada konkurensi tinggi.

---

## 8. Format Response JSON

### Response Sukses
```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": { ... }
}
```

### Response List dengan Pagination
```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [ ... ],
  "meta": {
    "current_page": 1,
    "per_page": 10,
    "total": 20,
    "last_page": 2
  }
}
```

### Response Error Validasi (HTTP 422)
```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "nim": ["NIM sudah terdaftar"],
    "email": ["Format email tidak valid"]
  }
}
```

### Response Rate Limiting (HTTP 429)
```json
{
  "success": false,
  "message": "Terlalu banyak percobaan login gagal, silakan coba lagi dalam 1 menit"
}
```

---

## 9. Contoh Request Penting

### A. Login Admin
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin1234"
  }'
```

### B. Login Mahasiswa
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rina.putri@student.siakad.ac.id",
    "password": "187221000001"
  }'
```

### C. Tambah Mahasiswa (Admin)
```bash
curl -X POST http://localhost:3000/api/v1/students \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "187221000099",
    "nama": "Alya Maharani",
    "email": "alya.m@student.siakad.ac.id",
    "prodi": "Teknik Informatika",
    "angkatan": 2024,
    "ipk_terakhir": 3.75
  }'
```

### D. Ambil KRS (Mahasiswa)
```bash
curl -X POST http://localhost:3000/api/v1/enrollments \
  -H "Authorization: Bearer <MAHASISWA_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "course_id": 1,
    "tahun_akademik": "2026/2027-Ganjil"
  }'
```

### E. Batalkan KRS (Mahasiswa)
```bash
curl -X DELETE http://localhost:3000/api/v1/enrollments/1 \
  -H "Authorization: Bearer <MAHASISWA_TOKEN>"
```

---

## 10. Pengujian (Testing)

Project dilengkapi dengan automated testing lengkap yang mencakup unit test, integration test, dan live E2E script.

### Menjalankan Unit & Integration Test
```bash
# Format kode
gofmt -w .

# Analisis statis
go vet ./...

# Menjalankan seluruh test suite tanpa cache
go test -v -count=1 ./...

# Build binary
go build ./...
```

### Menjalankan Live E2E Verification
Tersedia skrip pengujian langsung pada server yang sedang berjalan:
```powershell
powershell -ExecutionPolicy Bypass -File .\test_e2e.ps1
```
