# SIAKAD Microservices

## Sistem Informasi Akademik — Arsitektur Microservices & Clean Architecture (Minggu 3)

### 1. Deskripsi Singkat

SIAKAD Minggu 2 dikembangkan sebagai REST API monolitik — seluruh fungsionalitas (CRUD Mahasiswa, Input Nilai, Transkrip, Rekap Jurusan, Top IPK) berjalan dalam satu proses. Pada Minggu 3, sistem dimigrasikan ke arsitektur **Microservices**:

- **`akademik-service`** — Mencatat data mahasiswa dan nilai akademiknya secara persisten (memiliki database).
- **`rekap-service`** — Meringkas keadaan akademik mahasiswa menjadi laporan statistik tanpa menyimpan data (stateless, database-less).

Kedua service diorkestrasi dengan **Docker Compose** dan berkomunikasi melalui **HTTP/REST sinkron** via DNS internal.

---

### 2. Arsitektur Sistem

#### 2.1 Diagram Komunikasi

```
+-----------------------------------------------------------+
|                         KLIEN                             |
|                     (Postman / Web)                       |
+-----------------------------------------------------------+
       |                                    |
       | Port 8081                          | Port 8082
       v                                    v
+----------------------+         +-------------------------+
|  AKADEMIK SERVICE    |         |    REKAP SERVICE        |
|  CRUD Mahasiswa,     | <------ |    Rekap Jurusan,       |
|  Nilai, Transkrip,   | HTTP    |    Top IPK, Ringkasan   |
|  Ringkasan           | REST    |    [No Database]        |
+----------------------+         +-------------------------+
       |
       | TCP 5432
       v
+----------------------+
|     AKADEMIK-DB      |
|  (PostgreSQL 18)     |
+----------------------+
```

#### 2.2 Clean Architecture per Service

| Layer | Akademik Service | Rekap Service |
|:---|:---|:---|
| **Infrastructure** | Gin Engine, pgx Driver, Config, Env | Gin Engine, HTTP Client, Config, Env |
| **Interface Adapter** | HTTP Handlers, DB Repository (SQL) | HTTP Handlers, HTTP Client Implementation |
| **Usecase / Service** | Logic Akademik, Hitung IPK, Validasi | Logic Rekap (PerJurusan, TopIPK, Sorting) |
| **Domain** | Entities (Mahasiswa, Nilai), Repo Interfaces, Sentinel Errors | Entities (Ringkasan), Client Interface, Errors |

#### 2.3 Evaluasi Arsitektur

1. **Deploy Independen?** Ya — tiap service dapat dideploy tanpa mengubah yang lain. Rekap Service hanya bergantung pada kontrak API v1 Akademik Service.
2. **Kepemilikan Data?** Ya — `akademik-service` adalah pemilik tunggal `akademik-db`. `rekap-service` murni mengkonsumsi data via HTTP.
3. **Tanggung Jawab Jelas?**
   - **Akademik Service:** Mencatat data mahasiswa dan nilai akademiknya secara persisten.
   - **Rekap Service:** Meringkas keadaan akademik mahasiswa menjadi laporan statistik tanpa menyimpan data.

---

### 3. Pembagian Domain & Tanggung Jawab

| Aspek | Akademik Service | Rekap Service |
|:---|:---|:---|
| **Tanggung Jawab Utama** | Mencatat mahasiswa dan nilainya | Meringkas keadaan akademik menjadi laporan |
| **Data yang Dimiliki** | Tabel `mahasiswa` & `nilai` di `akademik-db` | Tidak Ada (Data dipinjam via HTTP) |
| **Operasi Penulisan** | Ya (POST / PUT / DELETE) | Tidak (Read-Only) |
| **Lapisan Data** | Handler → Service → Repository → PostgreSQL | Handler → Service → Client → HTTP REST |
| **Port Publikasi** | `8081:8080` | `8082:8080` |
| **Dependensi Eksternal** | PostgreSQL (`akademik-db:5432`) | Akademik Service (`akademik-service:8080`) |

#### Migration Mapping (Minggu 2 → Minggu 3)

| Kapabilitas Minggu 2 | Status Minggu 3 | Lokasi Baru |
|:---|:---|:---|
| CRUD Mahasiswa | Tetap | `akademik-service` |
| Input Nilai & Transkrip | Tetap | `akademik-service` |
| Perhitungan IPK | Tetap | `akademik-service` |
| `GET /api/v1/rekap/jurusan` | **PINDAH** | `rekap-service` |
| `GET /api/v1/rekap/top-ipk` | **PINDAH** | `rekap-service` |
| `GET /health` | **BARU** | Kedua Service |
| `GET /api/v1/mahasiswa/:nim/ringkasan` | **BARU (Kontrak Internal)** | `akademik-service` |
| `GET /api/v1/rekap/mahasiswa/:nim` | **BARU** | `rekap-service` |

---

### 4. Spesifikasi API

#### 4.1 Akademik Service (Port 8081)

Base URL: `http://localhost:8081`

| Method | Path | Deskripsi | Status Code |
|:---|:---|:---|:---|
| `GET` | `/health` | Liveness check service & koneksi DB | `200 OK` |
| `POST` | `/api/v1/mahasiswa` | Mendaftarkan mahasiswa baru | `201 Created` / `400`, `409` |
| `GET` | `/api/v1/mahasiswa` | Seluruh daftar mahasiswa | `200 OK` |
| `GET` | `/api/v1/mahasiswa/:nim` | Detail mahasiswa + IPK | `200 OK` / `404` |
| `PUT` | `/api/v1/mahasiswa/:nim` | Mengubah data mahasiswa | `200 OK` / `400`, `404` |
| `DELETE` | `/api/v1/mahasiswa/:nim` | Menghapus mahasiswa | `200 OK` / `404` |
| `POST` | `/api/v1/mahasiswa/:nim/nilai` | Input nilai mata kuliah | `201 Created` / `400`, `404` |
| `GET` | `/api/v1/mahasiswa/:nim/transkrip` | Transkrip nilai + IPK | `200 OK` / `404` |
| `GET` | `/api/v1/mahasiswa/:nim/ringkasan` | **Internal:** Ringkasan 1 mahasiswa | `200 OK` / `404` |
| `GET` | `/api/v1/rekap/*` | Route lama (telah dipindah) | `404 Not Found` |

#### 4.2 Rekap Service (Port 8082)

Base URL: `http://localhost:8082`

| Method | Path | Parameter | Deskripsi | Status Code |
|:---|:---|:---|:---|:---|
| `GET` | `/health` | - | Liveness check service | `200 OK` |
| `GET` | `/api/v1/rekap/jurusan` | - | Rekap mahasiswa per jurusan | `200 OK` / `503`, `504` |
| `GET` | `/api/v1/rekap/top-ipk` | `?n=3` | `n` mahasiswa IPK tertinggi | `200 OK` / `400`, `503`, `504` |
| `GET` | `/api/v1/rekap/mahasiswa/:nim` | - | Ringkasan 1 mahasiswa | `200 OK` / `404`, `503`, `504` |

> **Aturan `?n` pada Top IPK:** `n` bukan angka atau `n <= 0` → `400 Bad Request`. `n` > total mahasiswa → kembalikan seluruh mahasiswa.

#### 4.3 Kontrak Internal: Ringkasan Mahasiswa

**Response 200 OK:**
```json
{
  "sukses": true,
  "data": {
    "nim": "23010001",
    "nama": "Bunga",
    "jurusan": "Teknik Informatika",
    "status": "Aktif",
    "total_sks": 5,
    "ipk": 3.60
  }
}
```

**Response 404 Not Found:**
```json
{
  "sukses": false,
  "error": "mahasiswa tidak ditemukan"
}
```

---

### 5. Manajemen Error & Resiliensi

#### 5.1 Sentinel Error → HTTP Status Code

| Error Sentinel | Package | HTTP | Pemicu |
|:---|:---|:---|:---|
| `ErrNIMTidakValid` | siakad / rekap | `400` | NIM bukan 8 digit angka |
| `ErrMahasiswaSudahAda` | siakad | `409` | NIM sudah terdaftar |
| `ErrMahasiswaTidakAda` | siakad / rekap | `404` | NIM tidak ditemukan |
| `ErrNilaiTidakValid` | siakad | `400` | Mutu di luar 0.0–4.0 |
| `ErrSKSTidakValid` | siakad | `400` | SKS <= 0 |
| `ErrAkademikTimeout` | rekap | `504` | HTTP timeout > 2 detik |
| `ErrAkademikTidakTersedia` | rekap | `503` | Akademik mati / 5xx / connection refused |

#### 5.2 Layered Timeout Budget

```
Klien / Postman           : 10 detik
    └── Rekap Gin Handler :  5 detik  (context.WithTimeout)
          └── HTTP Client :  2 detik  (http.Client{Timeout: 2s})
                └── DB Query :  1 detik
```

#### 5.3 Aturan Resiliensi

- `http.DefaultClient` dilarang — client dikonfigurasi eksplisit dengan `Timeout: 2s`
- Context dari Gin Handler diteruskan hingga `http.NewRequestWithContext(ctx, ...)`
- `defer resp.Body.Close()` pada setiap HTTP outbound
- Error internal Rekap tidak meneruskan detail Akademik — dipetakan ke `503` atau `504`

---

### 6. Struktur Proyek

```
├── docker-compose.yml
├── postman/
│   └── minggu3.postman_collection.json
├── akademik-service/
│   ├── Dockerfile              (multi-stage, image < 30 MB, user appuser)
│   ├── go.mod                  (module akademik-service)
│   ├── main.go                 (slog JSON, X-Request-ID middleware)
│   ├── migrations/
│   │   └── 001_init.sql
│   └── internal/
│       ├── config/config.go
│       ├── response/response.go
│       └── siakad/
│           ├── model.go        (Mahasiswa, Nilai, RingkasanDTO, DTOs)
│           ├── errors.go       (5 sentinel error)
│           ├── repository.go   (interface: MahasiswaRepo, NilaiRepo)
│           ├── postgres.go     (implementasi PostgreSQL)
│           ├── service.go      (business logic: HitungIPK, CRUD, Ringkasan)
│           ├── service_test.go (table-driven HitungIPK + validasi)
│           └── handler.go      (Gin HTTP handlers, 8 endpoint)
└── rekap-service/
    ├── Dockerfile              (multi-stage, image < 30 MB, user appuser)
    ├── go.mod                  (module rekap-service)
    ├── main.go                 (slog JSON, X-Request-ID middleware + propagate)
    └── internal/
        ├── config/config.go    (PORT + AKADEMIK_BASE_URL dari env)
        ├── response/response.go
        └── rekap/
            ├── model.go        (Mahasiswa, Ringkasan domain rekap)
            ├── errors.go       (7 sentinel: +503, +504)
            ├── client.go       (AkademikClient interface + HTTP impl, timeout 2s)
            ├── client_test.go  (5 skenario: httptest.NewServer)
            ├── service.go      (PerJurusan, TopIPK, RingkasanMahasiswa)
            ├── service_test.go (4 skenario: mock client)
            └── handler.go      (Gin handlers, 5s timeout + 503/504 mapping)
```

---

### 7. Menjalankan

#### Prasyarat

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Git](https://git-scm.com/)

#### Quick Start

```bash
# Clone repository
git clone https://github.com/TraFScience/Siakad-Microservice.git
cd Siakad-Microservice

# Jalankan seluruh sistem (3 kontainer)
docker compose up --build
```

Setelah semua kontainer healthy:

| Service | URL |
|:---|:---|
| Akademik Service | `http://localhost:8081` |
| Rekap Service | `http://localhost:8082` |
| PostgreSQL | `localhost:5432` |

#### Stopping

```bash
docker compose down          # Hentikan, hapus container
docker compose down -v       # Hentikan + hapus volume (data reset)
```

#### Development Lokal (tanpa Docker untuk 1 service)

```bash
# Terminal 1 — Jalankan PostgreSQL via Docker saja
docker compose up akademik-db

# Terminal 2 — Jalankan akademik-service
cd akademik-service
cp .env.example .env
go run main.go

# Terminal 3 — Jalankan rekap-service
cd rekap-service
cp .env.example .env
go run main.go
```

#### Environment Variables

**akademik-service/.env.example:**
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=siakad
DB_SSLMODE=disable
```

**rekap-service/.env.example:**
```env
PORT=8080
AKADEMIK_BASE_URL=http://akademik-service:8080
```

#### Menjalankan Unit Tests

```bash
# Akademik Service
cd akademik-service
go test ./...

# Rekap Service
cd rekap-service
go test ./...
```

---

### 8. Acceptance Criteria (14 Skenario)

| No | Tindakan | Hasil yang Diharapkan |
|:---|:---|:---|
| 1 | `docker compose up --build` | 3 kontainer (DB, Akademik, Rekap) running & healthy |
| 2 | GET `:8081/health` & `:8082/health` | Keduanya `200 OK` |
| 3 | POST Mahasiswa NIM `23010001` "Bunga" (8081) | `201 Created` |
| 4 | POST Mahasiswa `23010001` lagi | `409 Conflict` |
| 5 | POST Nilai: Go (3 SKS, 4.0), DB (2 SKS, 3.0) | `201 Created`, IPK 3.60 |
| 6 | POST 2 Mahasiswa lain beda jurusan + nilai | `201 Created` seluruhnya |
| 7 | GET `:8082/api/v1/rekap/jurusan` | `200 OK`, data per jurusan rapi |
| 8 | GET `:8082/api/v1/rekap/top-ipk?n=2` | `200 OK`, 2 mahasiswa IPK tertinggi |
| 9 | GET `:8082/api/v1/rekap/mahasiswa/99999999` | `404 Not Found` |
| 10 | `docker compose stop akademik-service` lalu #7 | `503 Service Unavailable` |
| 11 | `docker compose start akademik-service` lalu #7 | `200 OK` (tanpa restart rekap) |
| 12 | Simulasi sleep 5s di akademik, panggil #7 | `504 Gateway Timeout` (≈2 detik) |
| 13 | `docker compose down` lalu `up -d`, panggil #7 | `200 OK`, data utuh (persistensi) |
| 14 | `docker compose logs -f rekap-service` | Log JSON dengan `service_name` & `X-Request-ID` |

---

### 9. Repository & Tools

- **GitHub:** [https://github.com/TraFScience/Siakad-Microservice](https://github.com/TraFScience/Siakad-Microservice)
- **Postman Collection:** `postman/minggu3.postman_collection.json` — 7 test suite mencakup seluruh 14 acceptance criteria
- **Tech Stack:** Go 1.26, Gin (HTTP framework), pgx/v5 (PostgreSQL driver), PostgreSQL 18, Docker Compose, `log/slog` (structured logging), `X-Request-ID` (distributed tracing)
