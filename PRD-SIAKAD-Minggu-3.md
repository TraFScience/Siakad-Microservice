# Product Requirement Document (PRD)
## Sistem Informasi Akademik (SIAKAD) Microservices
### Migrasi Arsitektur Microservices & Clean Architecture (Minggu 3)

---

## 1. Ringkasan Eksekutif & Konteks

### 1.1 Latar Belakang
Pada Minggu 2, SIAKAD telah dikembangkan menjadi REST API berlapis dengan media penyimpanan permanen di PostgreSQL. Meskipun rapi di dalam, sistem Minggu 2 masih berarsitektur **Monolitik** (*single binary, single deployment, single point of failure*). Seluruh fungsionalitas—mulai dari CRUD Mahasiswa, Input Nilai, Transkrip, hingga Rekap Jurusan dan Top IPK—dideploy sebagai satu proses yang sama.

Pada Minggu 3, SIAKAD dimigrasikan menjadi arsitektur **Microservices**. Fungsionalitas Rekap memisahkan diri dari monolit menjadi **Rekap Service** independen yang berinteraksi dengan **Akademik Service** melalui protokol HTTP/REST sinkron, dengan orkestrasisasi kontainer menggunakan **Docker Compose**.

### 1.2 Tujuan Proyek
* **Decomposisi Monolit:** Memecah monolit Minggu 2 menjadi dua microservices yang berjalan mandiri: `akademik-service` dan `rekap-service`.
* **Prinsip Kepemilikan Data (*Database per Service*):** Memastikan `akademik-db` hanya dimiliki dan diakses oleh `akademik-service`. `rekap-service` bersifat *stateless* dan *database-less*.
* **Penerapan Clean Architecture pada Microservices:** Menjaga isolasi lapisan (*Domain, Usecase/Service, Interface Adapter/Client/Repository, Infrastructure*) pada masing-masing modul service.
* **Resiliensi & Isolasi Kegagalan:** Mengimplementasikan timeout berlapis, propagasi *context*, penanganan kegagalan jaringan, dan pemetaan HTTP status code khusus (`503 Service Unavailable` dan `504 Gateway Timeout`).
* **Containerization & Deployment:** Menggunakan multi-stage `Dockerfile` (menghasilkan image kecil < 30 MB, pengguna non-root) dan `docker-compose.yml` dengan *healthcheck* serta *DNS internal*.
* **Observabilitas:** Mengimplementasikan *structured logging* (`log/slog` JSON) dan penelusuran *distributed tracing* menggunakan header `X-Request-ID`.

---

## 2. Arsitektur Microservices & Clean Architecture

### 2.1 Alur Sistem & Diagram Komunikasi

```
+-----------------------------------------------------------------------------------+
|                                    KLIEN                                          |
|                                (Postman / Web)                                    |
+-----------------------------------------------------------------------------------+
       |                                                    |
       | Port 8081 (Akademik API)                           | Port 8082 (Rekap API)
       v                                                    v
+-----------------------------+                      +------------------------------+
|     AKADEMIK SERVICE        |                      |        REKAP SERVICE         |
|  (CRUD Mahasiswa, Nilai,    |                      |   (Rekap Jurusan, Top IPK,   |
|     Transkrip, Ringkasan)   | <--- HTTP REST ----- |    Ringkasan Mahasiswa)      |
|                             |   (DNS Internal)     |   [Stateless, No Database]   |
+-----------------------------+                      +------------------------------+
       |
       | TCP 5432
       v
+-----------------------------+
|        AKADEMIK-DB          |
|    (PostgreSQL 18 - DB)     |
+-----------------------------+
```

### 2.2 Struktur Clean Architecture per Service

Kedua service menerapkan Clean Architecture dengan batasan tanggung jawab yang ketat:

#### A. Akademik Service (`akademik-service/`)
```
+-------------------------------------------------------------------------+
| Infrastructure: Gin Engine, PostgreSQL Driver (pgx/sql), Config, Env     |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Interface Adapter: HTTP Handlers (Gin Controllers), DB Repository (SQL)  |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Usecase / Service: Logic Akademik, Hitung IPK, Validasi Domain          |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Domain: Entities (Mahasiswa, Nilai), Repo Interfaces, Sentinel Errors   |
+-------------------------------------------------------------------------+
```

#### B. Rekap Service (`rekap-service/`)
```
+-------------------------------------------------------------------------+
| Infrastructure: Gin Engine, HTTP Client (net/http), Config, Env         |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Interface Adapter: HTTP Handlers, Akademik HTTP Client Implementation    |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Usecase / Service: Logic Rekap (PerJurusan, TopIPK, Sorting)            |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Domain: Entities (Ringkasan, Mahasiswa), Client Interface, Errors       |
+-------------------------------------------------------------------------+
```

### 2.3 Evaluasi Tiga Pertanyaan Dasar Arsitektur
1. **Bisakah tiap service di-deploy sendiri tanpa mengubah yang lain?**
   Ya. `akademik-service` dapat diperbarui atau dideploy ulang secara independen. `rekap-service` hanya bergantung pada kontrak API v1 HTTP `akademik-service`.
2. **Apakah tiap service benar-benar memiliki datanya sendiri?**
   Ya. `akademik-service` adalah satu-satunya pemilik `akademik-db`. `rekap-service` tidak memiliki database, tidak menyimpan DSN database di `.env`, dan murni mengkonsumsi data lewat HTTP.
3. **Bisakah tanggung jawab tiap service dijelaskan dalam satu kalimat?**
   * **Akademik Service:** Mencatat data mahasiswa dan nilai akademiknya secara persisten.
   * **Rekap Service:** Meringkas keadaan akademik mahasiswa menjadi laporan statistik tanpa menyimpan data.

---

## 3. Pembagian Domain & Batas Fungsionalitas

### 3.1 Matriks Kepemilikan & Responsibilitas

| Aspek | Akademik Service | Rekap Service |
| :--- | :--- | :--- |
| **Tanggung Jawab Utama** | Mencatat mahasiswa dan nilainya | Meringkas keadaan akademik menjadi laporan |
| **Data yang Dimiliki** | Tabel `mahasiswa` & `nilai` di `akademik-db` | Tidak Ada (Data dipinjam via HTTP) |
| **Operasi Penulisan (Write)** | Ya (POST / PUT / DELETE) | Tidak (Read-Only) |
| **Lapisan Data** | Handler -> Service -> Repository -> PostgreSQL | Handler -> Service -> Client -> HTTP REST |
| **Port Publikasi (Host:Container)** | `8081:8080` | `8082:8080` |
| **Dependensi Eksternal** | PostgreSQL Container (`akademik-db:5432`) | Akademik Service (`akademik-service:8080`) |

### 3.2 Pemindahan Kapabilitas (Migration Mapping)

| Kapabilitas Minggu 2 | Status Minggu 3 | Lokasi Baru |
| :--- | :--- | :--- |
| CRUD Mahasiswa | Tetap | `akademik-service` |
| Input Nilai & Transkrip | Tetap | `akademik-service` |
| Perhitungan IPK (`HitungIPK`) | Tetap | `akademik-service` |
| `GET /api/v1/rekap/jurusan` | **PINDAH** | `rekap-service` (Dihapus dari Akademik) |
| `GET /api/v1/rekap/top-ipk` | **PINDAH** | `rekap-service` (Dihapus dari Akademik) |
| `GET /health` | **BARU** | Kedua Service |
| `GET /api/v1/mahasiswa/:nim/ringkasan` | **BARU (Kontrak Internal)** | `akademik-service` |
| `GET /api/v1/rekap/mahasiswa/:nim` | **BARU** | `rekap-service` |

---

## 4. Kontrak & Antarmuka API (Wajib Dipenuhi)

### 4.1 Spesifikasi Endpoint Akademik Service (Port Publikasi 8081)

Base URL Internal: `http://akademik-service:8080`
Base URL Publik: `http://localhost:8081`

| Method | Path | Fungsi / Deskripsi | HTTP Status Code |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Liveness check service & koneksi DB | `200 OK` |
| `POST` | `/api/v1/mahasiswa` | Mendaftarkan mahasiswa baru | `201 Created` / `400`, `409` |
| `GET` | `/api/v1/mahasiswa` | Mendapatkan seluruh daftar mahasiswa | `200 OK` |
| `GET` | `/api/v1/mahasiswa/:nim` | Detail mahasiswa + kalkulasi IPK | `200 OK` / `404` |
| `PUT` | `/api/v1/mahasiswa/:nim` | Mengubah data mahasiswa | `200 OK` / `400`, `404` |
| `DELETE` | `/api/v1/mahasiswa/:nim` | Menghapus data mahasiswa | `200 OK` / `404` |
| `POST` | `/api/v1/mahasiswa/:nim/nilai` | Input nilai mata kuliah | `201 Created` / `400`, `404` |
| `GET` | `/api/v1/mahasiswa/:nim/transkrip` | Daftar nilai + IPK mahasiswa | `200 OK` / `404` |
| `GET` | `/api/v1/mahasiswa/:nim/ringkasan` | **Kontrak Internal:** Ringkasan 1 mahasiswa | `200 OK` / `404` |
| `GET` | `/api/v1/rekap/*` | Route lama yang telah dipindah | `404 Not Found` |

### 4.2 Spesifikasi Endpoint Rekap Service (Port Publikasi 8082)

Base URL Publik: `http://localhost:8082`

| Method | Path | Parameter | Fungsi / Deskripsi | HTTP Status Code |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/health` | - | Liveness check service | `200 OK` |
| `GET` | `/api/v1/rekap/jurusan` | - | Rekap mahasiswa per jurusan + rata-rata IPK | `200 OK` / `503`, `504` |
| `GET` | `/api/v1/rekap/top-ipk` | `?n=3` (Optional, Default: 3) | `n` mahasiswa IPK tertinggi, urut descending | `200 OK` / `400`, `503`, `504` |
| `GET` | `/api/v1/rekap/mahasiswa/:nim` | - | Ringkasan 1 mahasiswa (Uji propagasi 404) | `200 OK` / `404`, `503`, `504` |

*Aturan Query Param `n` pada `top-ipk`:*
* Jika `n` bukan angka atau `n <= 0` -> Kembalikan HTTP `400 Bad Request`.
* Jika `n` lebih besar dari total mahasiswa -> Kembalikan seluruh mahasiswa (bukan error).

### 4.3 Kontrak Internal: Endpoint Ringkasan Mahasiswa
Endpoint ini digunakan khusus untuk komunikasi antar-service (`akademik-service` -> `rekap-service`).

**Request:**
`GET /api/v1/mahasiswa/23010001/ringkasan`

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

### 4.4 Kontrak Interface & Model Code

#### A. Interface Repository pada Akademik Service
```go
package siakad

import "context"

type MahasiswaRepository interface {
    Tambah(ctx context.Context, m Mahasiswa) error
    Cari(ctx context.Context, nim string) (Mahasiswa, error)
    Semua(ctx context.Context) ([]Mahasiswa, error)
    Update(ctx context.Context, m Mahasiswa) error
    Hapus(ctx context.Context, nim string) error
}

type NilaiRepository interface {
    Tambah(ctx context.Context, n Nilai) error
    PerMahasiswa(ctx context.Context, nim string) ([]Nilai, error)
}
```

#### B. Interface Client pada Rekap Service (`internal/rekap/client.go`)
```go
package rekap

import "context"

type AkademikClient interface {
    DaftarMahasiswa(ctx context.Context) ([]Mahasiswa, error)
    Ringkasan(ctx context.Context, nim string) (Ringkasan, error)
}
```

#### C. Service Layer Methods pada Rekap Service (`internal/rekap/service.go`)
```go
package rekap

import "context"

type Service interface {
    PerJurusan(ctx context.Context) (map[string][]Ringkasan, error)
    TopIPK(ctx context.Context, n int) ([]Ringkasan, error)
    RingkasanMahasiswa(ctx context.Context, nim string) (Ringkasan, error)
}
```

---

## 5. Manajemen Error, Timeout & Resiliensi

### 5.1 Matriks Sentinel Error & Status Code HTTP

| Error Sentinel Go | Package | HTTP Status Code | Pemicu / Kondisi |
| :--- | :--- | :--- | :--- |
| `ErrNIMTidakValid` | `siakad` / `rekap` | `400 Bad Request` | NIM bukan 8 digit angka |
| `ErrMahasiswaSudahAda` | `siakad` | `409 Conflict` | NIM sudah terdaftar pada database |
| `ErrMahasiswaTidakAda` | `siakad` / `rekap` | `404 Not Found` | NIM tidak ditemukan (termasuk 404 dari Akademik Service) |
| `ErrNilaiTidakValid` | `siakad` | `400 Bad Request` | Mutu di luar skala 0.0 - 4.0 |
| `ErrSKSTidakValid` | `siakad` | `400 Bad Request` | SKS <= 0 |
| `ErrAkademikTimeout` **(BARU)** | `rekap` | `504 Gateway Timeout` | Panggilan HTTP melebihi batas waktu (2 detik) / `context.DeadlineExceeded` |
| `ErrAkademikTidakTersedia` **(BARU)** | `rekap` | `503 Service Unavailable` | Akademik Service mati, *connection refused*, DNS gagal, atau HTTP Status 5xx |

### 5.2 Anggaran Timeout Berlapis (*Layered Timeout Budget*)
Untuk mencegah *cascading latency* dan memastikan error diterjemahkan secara presisi, batas timeout dikonfigurasi secara hierarkis:

```
+-------------------------------------------------------------------+
| Klien / Postman Timeout                     : 10 detik             |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
| Server Timeout (Rekap Service Gin Middleware): 5 detik            |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
| Rekap HTTP Client Timeout ke Akademik       : 2 detik             |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
| Query Database PostgreSQL di Akademik       : 1 detik             |
+-------------------------------------------------------------------+
```

### 5.3 Aturan Resiliensi Komunikasi Network
1. **Dilarang Menggunakan `http.DefaultClient`:** Client wajib diinisialisasi secara eksplisit dengan konfigurasi `http.Client{ Timeout: 2 * time.Second }`.
2. **Propagasi Context:** Context dari `c.Request.Context()` di Gin Handler wajib diteruskan sampai ke `http.NewRequestWithContext(ctx, ...)` pada client.
3. **Penyelamatan Resource:** Setiap panggilan HTTP outbound wajib mengeksekusi `defer resp.Body.Close()`.
4. **Isolasi Error 500:** Error internal pada Rekap Service tidak boleh meneruskan detail stacktrace internal Akademik Service. Semua kegagalan tetangga wajib dipetakan ke 503 atau 504.

---

## 6. Struktur Direktori Proyek

Seluruh source code disimpan dalam satu repository `siakad-microservices/` dengan dua modul Go yang terisolasi (*no cross-module imports*):

```
siakad-microservices/
├── docker-compose.yml                  # Orchestration seluruh service & DB
├── README.md                           # Dokumentasi, diagram, & instruksi eksekusi
├── postman/
│   └── minggu3.postman_collection.json # Collection pengujian API Minggu 3
│
├── akademik-service/                   # MODUL 1: AKADEMIK SERVICE
│   ├── Dockerfile                      # Multi-stage Dockerfile
│   ├── .dockerignore
│   ├── go.mod                          # module akademik-service
│   ├── main.go                         # Inisialisasi DB, Router, & Server
│   ├── migrations/
│   │   └── 001_init.sql                # Skema tabel mahasiswa & nilai
│   └── internal/
│       ├── config/                     # Configuration loader (.env)
│       ├── response/                   # JSON response envelope formatter
│       └── siakad/
│           ├── model.go                # Struct Mahasiswa, Nilai, DTO
│           ├── errors.go               # 5 Error Sentinel Utama
│           ├── repository.go           # Interface Mahasiswa & Nilai Repository
│           ├── postgres.go             # Implementasi repository Postgres (sqlx/database/sql)
│           ├── service.go              # Business logic (HitungIPK, InputNilai)
│           ├── service_test.go        # Table-driven test HitungIPK
│           └── handler.go              # Gin HTTP handlers
│
└── rekap-service/                      # MODUL 2: REKAP SERVICE
    ├── Dockerfile                      # Multi-stage Dockerfile
    ├── .dockerignore
    ├── go.mod                          # module rekap-service
    ├── main.go                         # Inisialisasi HTTP Client, Router, & Server
    └── internal/
        ├── config/                     # Configuration loader (.env)
        ├── response/                   # JSON response envelope formatter
        └── rekap/
            ├── model.go                # Struct Ringkasan, Mahasiswa (rekap domain)
            ├── errors.go               # Error sentinel (termasuk 503 & 504)
            ├── client.go               # Interface AkademikClient + Implementasi HTTP
            ├── client_test.go          # Testing client menggunakan httptest.NewServer
            ├── service.go              # Business logic (PerJurusan, TopIPK)
            ├── service_test.go         # Testing service menggunakan mock client
            └── handler.go              # Gin HTTP handlers + Error mapping 503/504
```

---

## 7. Spesifikasi Kontainarisasi & Docker Compose

### 7.1 Standar Dockerfile Multi-Stage
Kedua service (`akademik-service` dan `rekap-service`) wajib menggunakan pola multi-stage Dockerfile yang identik untuk mengoptimalkan ukuran image dan keamanan:

```dockerfile
# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service main.go

# Stage 2: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /app/service .
USER appuser
EXPOSE 8080
ENTRYPOINT ["./service"]
```

### 7.2 Spesifikasi `docker-compose.yml`
File `docker-compose.yml` di root repository mengorkestrasi tiga kontainer utama:

```yaml
version: '3.8'

services:
  akademik-db:
    image: postgres:18-alpine
    container_name: akademik-db
    environment:
      POSTGRES_USER: siakad_user
      POSTGRES_PASSWORD: siakad_password
      POSTGRES_DB: siakad_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./akademik-service/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U siakad_user -d siakad_db"]
      interval: 3s
      timeout: 3s
      retries: 5

  akademik-service:
    build:
      context: ./akademik-service
      dockerfile: Dockerfile
    container_name: akademik-service
    ports:
      - "8081:8080"
    environment:
      PORT: "8080"
      DB_HOST: akademik-db
      DB_PORT: "5432"
      DB_USER: siakad_user
      DB_PASSWORD: siakad_password
      DB_NAME: siakad_db
    depends_on:
      akademik-db:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 5s
      timeout: 3s
      retries: 3

  rekap-service:
    build:
      context: ./rekap-service
      dockerfile: Dockerfile
    container_name: rekap-service
    ports:
      - "8082:8080"
    environment:
      PORT: "8080"
      AKADEMIK_BASE_URL: "http://akademik-service:8080"
    depends_on:
      akademik-service:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 5s
      timeout: 3s
      retries: 3

volumes:
  postgres_data:
```

### 7.3 Logging & Trace ID Cross-Service
1. **Structured Logging:** Menggunakan standard library `log/slog` dengan output JSON ke stdout.
2. **Propagasi Header `X-Request-ID`:** 
   - Middleware pada `rekap-service` mengekstrak atau generate `X-Request-ID`.
   - `HTTPAkademikClient` menyertakan `X-Request-ID` dalam HTTP Request Header saat memanggil `akademik-service`.
   - `akademik-service` mencatat `X-Request-ID` tersebut pada log internalnya.

---

## 8. Strategi Pengujian (Quality Assurance & Unit Test)

### 8.1 Pengujian Unit Akademik Service (`akademik-service/internal/siakad/service_test.go`)
* **Pengujian Function `HitungIPK`:** Menggunakan paradigma *table-driven test* dengan komparasi epsilon (`math.Abs(got - want) < 1e-9`). Minimal 4 skenario:
  1. Slice nilai kosong -> `0.0`.
  2. 1 Mata Kuliah (3 SKS, Mutu 4.0) -> `4.0`.
  3. 2 Mata Kuliah (3 SKS Mutu 4.0 & 2 SKS Mutu 3.0) -> `3.6`.
  4. Skenario kustom variasi mutu & SKS.

### 8.2 Pengujian HTTP Client Rekap Service (`rekap-service/internal/rekap/client_test.go`)
Menggunakan `httptest.NewServer` untuk menyimulasikan Akademik Service tanpa koneksi jaringan sungguhan/Docker:

| Skenario Mock Server | Respon Server | Hasil yang Diharapkan pada Client |
| :--- | :--- | :--- |
| **Respon Valid** | Status `200 OK` + Body JSON valid | `struct` terisi benar, `err == nil` |
| **Data Tidak Ada** | Status `404 Not Found` | `errors.Is(err, ErrMahasiswaTidakAda)` bernilai `true` |
| **Server Error** | Status `500 Internal Server Error` | `errors.Is(err, ErrAkademikTidakTersedia)` bernilai `true` |
| **Latensi / Timeout** | Delay 3 detik sebelum merespon | `errors.Is(err, context.DeadlineExceeded)` / `ErrAkademikTimeout` |
| **Bad JSON Body** | Status `200 OK` + Body JSON rusak/malformed | Mengembalikan error unmarshal yang dibungkus `%w` |

### 8.3 Pengujian Service Layer Rekap (`rekap-service/internal/rekap/service_test.go`)
Menggunakan Mock Client (`MockAkademikClient`) untuk menguji logika `TopIPK` dan `PerJurusan`:
1. `TopIPK(ctx, 2)` pada 4 mahasiswa -> Mengembalikan 2 mahasiswa IPK tertinggi, urut descending.
2. `TopIPK(ctx, 10)` pada 4 mahasiswa -> Mengembalikan 4 mahasiswa tanpa error.
3. `PerJurusan(ctx)` -> Memetakan 3 mahasiswa di 2 jurusan menjadi map dengan 2 key secara tepat.
4. Client mengembalikan `ErrAkademikTidakTersedia` -> Service meneruskan error tersebut, tidak menelannya menjadi data kosong.

---

## 9. Kriteria Penerimaan & Skenario Demo (Acceptance Criteria)

Sistem dinyatakan LULUS uji apabila berhasil melewati 14 alur pengujian berikut secara berurutan:

```
+----+---------------------------------------------------+---------------------------------------------------------+
| No | Tindakan / Eksekusi                               | Hasil yang Diharapkan                                   |
+----+---------------------------------------------------+---------------------------------------------------------+
| 1  | `docker compose up --build`                       | 3 kontainer (DB, Akademik, Rekap) running & healthy.    |
| 2  | GET `http://localhost:8081/health` & `:8082/health`| Both return 200 OK.                                     |
| 3  | POST Mahasiswa NIM "23010001" "Bunga" (Port 8081)  | Status 201 Created; IPK Awal 0.00.                      |
| 4  | POST Mahasiswa NIM "23010001" lagi                | Status 409 Conflict (NIM sudah terdaftar).              |
| 5  | POST Nilai NIM "23010001": Go (3 SKS, 4.0), DB(2,3)| Status 201 Created; IPK terhitung menjadi 3.60.         |
| 6  | POST 2 Mahasiswa lain beda jurusan + nilai         | Status 201 Created pada seluruh input.                  |
| 7  | GET `http://localhost:8082/api/v1/rekap/jurusan`  | Status 200 OK; Data terkelompok per jurusan, IPK tepat. |
| 8  | GET `http://localhost:8082/api/v1/rekap/top-ipk?n=2`| Status 200 OK; 2 Mahasiswa IPK tertinggi (descending).  |
| 9  | GET `http://localhost:8082/api/v1/rekap/mahasiswa/99`| Status 404 Not Found ("mahasiswa tidak ditemukan").      |
| 10 | `docker compose stop akademik-service` lalu #7    | Status 503 Service Unavailable (Cepat, tanpa panic).    |
| 11 | `docker compose start akademik-service` lalu #7   | Status 200 OK kembali normal tanpa restart Rekap.       |
| 12 | Simulasikan Sleep 5s di Akademik, panggil #7      | Status 504 Gateway Timeout tepat setelah ±2 detik.      |
| 13 | `docker compose down` lalu `up -d`, panggil #7    | Status 200 OK; Data tetap utuh (Persistensi DB terbukti)|
| 14 | Cek `docker compose logs -f rekap-service`        | Log JSON rapi dengan field service_name & X-Request-ID. |
+----+---------------------------------------------------+---------------------------------------------------------+
```

---

## 10. Checklist Pengumpulan & Definition of Done (DoD)

- [ ] Kode Minggu 2 telah dipindahkan penuh ke folder `akademik-service/` dan tetap berfungsi.
- [ ] Endpoint `/health` tersedia di kedua service dan mengembalikan HTTP 200 OK.
- [ ] Endpoint internal `/api/v1/mahasiswa/:nim/ringkasan` sesuai spesifikasi kontrak JSON.
- [ ] Seluruh route `/api/v1/rekap/*` telah dihapus dari `akademik-service`.
- [ ] `rekap-service` murni *database-less* (tidak ada SQL driver, DSN, atau import database).
- [ ] Interface `AkademikClient` pada Rekap Service diimplementasikan secara implisit oleh HTTP Client.
- [ ] HTTP Client menggunakan instance kustom dengan Timeout eksplisit (bukan `http.DefaultClient`).
- [ ] Context dari Gin Handler diteruskan hingga ke `http.NewRequestWithContext` dan dipasangi `defer cancel()`.
- [ ] Pernyataan `defer resp.Body.Close()` dipasang di seluruh panggilan HTTP outbound.
- [ ] Pemetaan 7 Error Sentinel terpetakan secara presisi ke HTTP status: 400, 404, 409, 503, dan 504.
- [ ] Multi-stage Dockerfile dibuat untuk kedua service (image < 30 MB, user non-root `appuser`).
- [ ] File `docker-compose.yml` di root dapat menyalakan seluruh sistem dengan satu perintah (`docker compose up --build`).
- [ ] Variable `AKADEMIK_BASE_URL` dibaca dari environment variable, bukan hardcoded.
- [ ] Structured logging JSON (`log/slog`) aktif dan meneruskan `X-Request-ID`.
- [ ] Seluruh pengujian unit (`go test ./...`) HIJAU di kedua modul (`akademik-service` & `rekap-service`).
- [ ] Postman Collection Minggu 3 tersedia pada direktori `postman/`.

---

## 11. Fitur Tantangan Bonus (Opsional Scope)

* **Panggilan Paralel (Errgroup):** Rekap Service mengambil ringkasan mahasiswa secara paralel menggunakan `golang.org/x/sync/errgroup` dengan *worker pool* terbatas untuk mengatasi masalah $N+1$ query HTTP.
* **Endpoint Batch di Akademik Service:** Menyediakan `GET /api/v1/mahasiswa/ringkasan` pada Akademik Service untuk mengembalikan seluruh ringkasan dalam 1 HTTP request.
* **Retry + Exponential Backoff:** Menambahkan mekanisme retry 2-3 kali dengan jeda menaik hanya untuk operasi idempoten (`GET`).
* **In-Memory Caching (In-Memory TTL):** Menyimpan rekap jurusan selama 5-10 detik pada memori `rekap-service` untuk mengurangi beban panggilan ke Akademik Service.
* **Circuit Breaker Pattern:** Menghentikan panggilan ke Akademik Service sementara jika terjadi kegagalan berturut-turut (*fast fail*).
* **API Gateway (Nginx / Traefik):** Menyediakan single entry point pada port 80 untuk merouting `/api/v1/rekap/*` ke Rekap Service dan sisanya ke Akademik Service.
