# PrintMart — Toko ATK & Jasa Fotocopy Online

Platform e-commerce untuk toko alat tulis kantor (ATK), perlengkapan sekolah, dan jasa print/fotocopy/jilid secara online. Dibangun dengan **Go** (backend) dan **Svelte** (frontend), didesain untuk siap production dengan payment gateway dalam mode **sandbox/testing**.

---

## 📋 Daftar Isi

- [PrintMart — Toko ATK \& Jasa Fotocopy Online](#printmart--toko-atk--jasa-fotocopy-online)
  - [📋 Daftar Isi](#-daftar-isi)
  - [Tentang Project](#tentang-project)
  - [Fitur Utama](#fitur-utama)
    - [🛒 Belanja Produk (ATK \& Perlengkapan)](#-belanja-produk-atk--perlengkapan)
    - [🖨️ Jasa Print / Fotocopy Online](#️-jasa-print--fotocopy-online)
    - [📎 Jasa Tambahan](#-jasa-tambahan)
    - [👤 Manajemen Pengguna](#-manajemen-pengguna)
    - [🛠️ Admin Dashboard](#️-admin-dashboard)
    - [🔔 Notifikasi](#-notifikasi)
  - [Tech Stack](#tech-stack)
  - [Kenapa Memilih Database Ini?](#kenapa-memilih-database-ini)
    - [PostgreSQL sebagai Database Utama](#postgresql-sebagai-database-utama)
    - [Redis sebagai Pelengkap (Cache \& Queue)](#redis-sebagai-pelengkap-cache--queue)
    - [Kenapa Bukan MySQL atau MongoDB?](#kenapa-bukan-mysql-atau-mongodb)
  - [Arsitektur Sistem](#arsitektur-sistem)
  - [Struktur Database](#struktur-database)
  - [Struktur Folder Project](#struktur-folder-project)
  - [Cara Instalasi \& Menjalankan](#cara-instalasi--menjalankan)
    - [Prasyarat](#prasyarat)
    - [Opsi 1: Menggunakan Docker Compose (disarankan)](#opsi-1-menggunakan-docker-compose-disarankan)
    - [Opsi 2: Manual](#opsi-2-manual)
  - [Environment Variables](#environment-variables)
  - [API Endpoints](#api-endpoints)
    - [Auth](#auth)
    - [Produk](#produk)
    - [Cart \& Order](#cart--order)
    - [Jasa Print](#jasa-print)
    - [Payment](#payment)
    - [WebSocket](#websocket)
  - [Payment Gateway (Mode Testing)](#payment-gateway-mode-testing)
  - [Roadmap](#roadmap)
  - [Kontribusi](#kontribusi)
  - [Lisensi](#lisensi)

---

## Tentang Project

**PrintMart** adalah platform digital yang menggabungkan dua kebutuhan sehari-hari masyarakat sekitar sekolah/perkantoran:

1. **Toko online** untuk alat tulis kantor, perlengkapan sekolah, kertas, tinta, dan consumables lainnya.
2. **Jasa print, fotocopy, jilid, laminating, dan scan dokumen** secara online — pelanggan cukup upload file, pilih spesifikasi, sistem otomatis menghitung harga, lalu pesanan bisa diambil di toko atau dikirim.

Target pengguna: pelajar, mahasiswa, guru, staf kantor, dan UMKM di sekitar area toko.

---

## Fitur Utama

### 🛒 Belanja Produk (ATK & Perlengkapan)
- Katalog produk dengan kategori (ATK, Kertas & Tinta, Perlengkapan Sekolah, Furniture Kantor)
- Pencarian & filter produk
- Keranjang belanja (cart)
- Checkout dengan pilihan ambil di toko atau dikirim
- Riwayat pesanan & fitur re-order cepat

### 🖨️ Jasa Print / Fotocopy Online
- Upload file (PDF, DOCX, JPG, PNG)
- Pilih spesifikasi: warna/hitam-putih, ukuran kertas (A4/F4/A3), jumlah rangkap, satu/dua sisi
- Kalkulasi harga otomatis berdasarkan jumlah halaman hasil parsing file
- Estimasi waktu selesai
- Tracking status pesanan real-time: `Menunggu` → `Diproses` → `Siap Diambil/Dikirim` → `Selesai`

### 📎 Jasa Tambahan
- Jilid (spiral, lakban, hardcover)
- Laminating (berbagai ukuran)
- Scan dokumen fisik ke PDF

### 👤 Manajemen Pengguna
- Registrasi & login (JWT-based auth)
- Profil & alamat pengiriman
- Riwayat transaksi

### 🛠️ Admin Dashboard
- CRUD produk & kategori
- Kelola pesanan (produk & jasa print)
- Update status pesanan jasa print
- Laporan penjualan sederhana
- Kelola voucher/promo

### 🔔 Notifikasi
- Notifikasi real-time via WebSocket saat status pesanan berubah (misal "Print Anda sudah siap diambil")

---

## Tech Stack

| Layer | Teknologi |
|---|---|
| Backend | Go (Golang), Gin/Fiber (HTTP framework) |
| Frontend | Svelte / SvelteKit |
| Database | PostgreSQL (dipilih karena data relational, ACID compliant, dan concurrency handling yang baik — lihat penjelasan lengkap di bawah) |
| Cache / Queue | Redis (untuk antrian proses print job & caching data yang sering diakses) |
| Auth | JWT |
| Real-time | WebSocket (status pesanan & notifikasi) |
| Payment | **Midtrans** / Xendit — **Sandbox Mode** |
| File Storage | Local storage (dev) / S3-compatible storage (production) |
| Containerization | Docker & Docker Compose |

---

## Kenapa Memilih Database Ini?

### PostgreSQL sebagai Database Utama

PostgreSQL dipilih sebagai database utama karena beberapa alasan yang relevan dengan kebutuhan project ini:

1. **Data bersifat relational** — entitas seperti `users`, `products`, `orders`, `order_items`, `print_jobs`, dan `payments` saling terhubung lewat foreign key. PostgreSQL didesain khusus untuk menangani relasi antar tabel seperti ini dengan efisien.

2. **ACID compliant** — penting untuk transaksi e-commerce. Contoh: saat checkout, stok produk harus berkurang **dan** order harus tercatat secara bersamaan. Jika salah satu proses gagal, PostgreSQL bisa melakukan rollback otomatis sehingga data tidak korup atau setengah tersimpan.

3. **Concurrency handling yang baik** — krusial untuk kasus seperti banyak user checkout produk yang sama secara bersamaan (stok terbatas), atau banyak print job masuk dalam waktu berdekatan. PostgreSQL menyediakan row-level locking yang membantu mencegah race condition.

4. **Dukungan JSON/JSONB** — berguna untuk menyimpan data semi-terstruktur seperti detail spesifikasi print job (kombinasi warna, ukuran kertas, opsi finishing) tanpa perlu database terpisah.

5. **Ekosistem matang di Go** — driver seperti `pgx` sudah stabil, banyak dipakai di production, dan terintegrasi baik dengan ORM/query builder populer (misalnya `sqlc`, `GORM`, atau `sqlx`).

### Redis sebagai Pelengkap (Cache & Queue)

Redis **bukan pengganti** PostgreSQL, melainkan pelengkap untuk dua kebutuhan spesifik:

- **Caching** — mempercepat response API untuk data yang sering diakses dan jarang berubah, misalnya daftar kategori atau produk populer.
- **Job Queue** — menampung antrian print job yang masuk, lalu diproses satu per satu oleh worker Go secara asynchronous, sehingga request upload file tidak perlu menunggu proses selesai (non-blocking).

### Kenapa Bukan MySQL atau MongoDB?

- **MySQL** — sebenarnya juga merupakan pilihan yang valid dan cukup mirip dengan PostgreSQL untuk kebutuhan project skala ini. Namun PostgreSQL unggul di fitur-fitur lanjutan seperti dukungan JSON yang lebih matang, full-text search bawaan, dan ekstensi tambahan (misalnya PostGIS jika suatu saat dibutuhkan fitur berbasis lokasi toko).

- **MongoDB (NoSQL)** — kurang cocok untuk kasus ini karena struktur data project sangat relational (order → order_items → products → categories). Memaksakan penggunaan NoSQL berisiko membuat query menjadi kompleks dan lebih rawan inkonsistensi data, terutama pada kasus race condition saat pengurangan stok.

---

## Arsitektur Sistem

```
┌─────────────┐        ┌─────────────────┐        ┌──────────────┐
│   Svelte    │ <----> │   Go REST API    │ <----> │  PostgreSQL  │
│  (Frontend) │        │    (Backend)     │        │  (Database)  │
└─────────────┘        └─────────────────┘        └──────────────┘
                               │
                 ┌─────────────┼─────────────┐
                 │             │             │
           ┌─────────┐   ┌──────────┐  ┌───────────┐
           │  Redis  │   │ WebSocket│  │  Payment   │
           │ (Queue) │   │  Server  │  │  Gateway   │
           │         │   │          │  │ (Sandbox)  │
           └─────────┘   └──────────┘  └───────────┘
```

Alur singkat jasa print:
1. User upload file → disimpan sementara → masuk **antrian (queue)** di Redis
2. Worker Go memproses file (hitung jumlah halaman → hitung estimasi harga)
3. User konfirmasi spesifikasi & harga → checkout
4. Status pesanan di-update dan dikirim ke frontend secara real-time via WebSocket

---

## Struktur Database

Skema inti (disederhanakan):

```
users            (id, name, email, password_hash, phone, role, created_at)
addresses        (id, user_id, label, full_address, city, postal_code)
categories       (id, name, type)              -- type: 'product' | 'service'
products         (id, category_id, name, description, price, stock, image_url)
print_services   (id, name, price_per_page, description)
finishing_options(id, name, type, price)        -- jilid, laminating, dll
orders           (id, user_id, order_type, status, total_price, delivery_method, created_at)
order_items      (id, order_id, product_id, quantity, subtotal)
print_jobs       (id, order_id, file_url, page_count, color_mode, paper_size,
                   copies, duplex, finishing_option_id, estimated_price, status)
payments         (id, order_id, provider, transaction_id, status, amount, paid_at)
vouchers         (id, code, discount_type, discount_value, valid_until, usage_limit)
notifications    (id, user_id, order_id, message, is_read, created_at)
```

---

## Struktur Folder Project

```
printmart/
├── backend/
│   ├── cmd/
│   │   └── server/main.go    # Entrypoint API + seed (-seed)
│   ├── internal/
│   │   ├── handler/         # HTTP handlers (controller)
│   │   ├── service/         # Business logic
│   │   ├── repository/      # Database layer (SQLite + seed)
│   │   ├── config/          # Env & konfigurasi
│   │   ├── middleware/      # Auth (JWT)
│   │   └── queue/           # Antrian print job (Redis/in-memory)
│   ├── pkg/
│   │   └── websocket/       # WebSocket handler
│   ├── scripts/go.mjs       # Bridge npm -> toolchain Go bundel
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── routes/          # SvelteKit pages
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   ├── stores/
│   │   │   └── api.js       # API client ke backend
│   │   └── app.html
│   ├── package.json
│   └── Dockerfile
│
├── .tools/                  # Toolchain Go bundel (lokal, tidak di-commit)
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Cara Instalasi & Menjalankan

### Prasyarat
- Go >= 1.22
- Node.js >= 18
- PostgreSQL >= 14
- Redis >= 6
- Docker & Docker Compose (opsional, untuk cara cepat)

### Opsi 1: Menggunakan Docker Compose (disarankan)

```bash
git clone https://github.com/username/printmart.git
cd printmart
cp .env.example .env
docker-compose up --build
```

Frontend akan berjalan di `http://localhost:5173`
Backend API akan berjalan di `http://localhost:8080`

### Opsi 2: Manual

**Backend:**
```bash
cd backend
cp .env.example .env
go mod download
go run ./cmd/server
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

> **Catatan:** Tidak perlu migrasi database manual — skema & seed (termasuk admin default) dibuat otomatis saat server pertama kali dijalankan.

---

## Environment Variables

Contoh isi `.env`:

```env
# Server
PORT=8080
APP_ENV=development

# Database
DATABASE_URL=postgres://user:password@localhost:5432/printmart?sslmode=disable

# Redis
REDIS_ADDR=localhost:6379

# JWT
JWT_SECRET=ganti_dengan_secret_yang_kuat
JWT_EXPIRES_IN=24h

# Payment Gateway (SANDBOX / TESTING MODE)
PAYMENT_PROVIDER=midtrans
MIDTRANS_SERVER_KEY=SB-Mid-server-xxxxxxxxxxxxxxxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxxxxxxxxxxxxxxx
MIDTRANS_IS_PRODUCTION=false

# File Storage
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=20
```

> ⚠️ **Catatan:** Payment gateway masih menggunakan **sandbox/test API key**. Sebelum production, ganti `MIDTRANS_IS_PRODUCTION=true` dan gunakan server key production dari akun Midtrans/Xendit resmi.

---

## API Endpoints

Ringkasan endpoint utama (dokumentasi lengkap ada di `/docs` atau Postman collection):

### Auth
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/api/auth/register` | Registrasi user baru |
| POST | `/api/auth/login` | Login, return JWT |

### Produk
| Method | Endpoint | Deskripsi |
|---|---|---|
| GET | `/api/products` | List produk (support query filter & search) |
| GET | `/api/products/:id` | Detail produk |
| POST | `/api/products` | Tambah produk (admin) |
| PUT | `/api/products/:id` | Update produk (admin) |
| DELETE | `/api/products/:id` | Hapus produk (admin) |

### Cart & Order
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/api/cart` | Tambah item ke cart |
| GET | `/api/cart` | Lihat isi cart |
| POST | `/api/orders` | Checkout / buat order |
| GET | `/api/orders/:id` | Detail order |
| GET | `/api/orders` | Riwayat order user |

### Jasa Print
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/api/print-jobs/upload` | Upload file untuk di-print |
| POST | `/api/print-jobs` | Buat print job dengan spesifikasi |
| GET | `/api/print-jobs/:id` | Detail & status print job |
| PATCH | `/api/print-jobs/:id/status` | Update status (admin) |

### Payment
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/api/payments/create` | Membuat transaksi pembayaran (sandbox) |
| POST | `/api/payments/webhook` | Callback/notifikasi dari payment gateway |

### WebSocket
| Endpoint | Deskripsi |
|---|---|
| `/ws/orders/:userId` | Real-time update status pesanan |

---

## Payment Gateway (Mode Testing)

Untuk tahap pengembangan, payment menggunakan **sandbox environment** dari Midtrans (bisa diganti Xendit sesuai preferensi):

- Semua transaksi **tidak menggunakan uang asli**
- Gunakan kartu kredit/VA test yang disediakan di dokumentasi sandbox masing-masing provider
- Status pembayaran (`pending`, `success`, `failed`) disimulasikan lewat dashboard sandbox atau webhook testing tool (contoh: Midtrans Simulator)

**Checklist sebelum go production:**
- [ ] Ganti API key sandbox → production
- [ ] Set `MIDTRANS_IS_PRODUCTION=true`
- [ ] Uji ulang seluruh flow checkout end-to-end
- [ ] Pastikan webhook URL sudah publicly accessible (HTTPS)
- [ ] Aktifkan logging & monitoring transaksi

---

## Roadmap

- [x] Autentikasi user (register/login)
- [x] Katalog produk & cart
- [x] Jasa print/fotocopy dengan kalkulasi harga otomatis
- [x] Payment integration (sandbox)
- [x] Real-time order tracking via WebSocket
- [ ] Voucher & sistem diskon
- [ ] Multi-alamat pengiriman
- [ ] Rekomendasi produk berdasarkan riwayat order
- [ ] Aplikasi mobile (opsional)

---

## Kontribusi

Pull request terbuka untuk siapa saja. Untuk perubahan besar, mohon buka issue terlebih dahulu untuk diskusi.

1. Fork repository
2. Buat branch baru (`git checkout -b fitur-baru`)
3. Commit perubahan (`git commit -m 'Tambah fitur X'`)
4. Push ke branch (`git push origin fitur-baru`)
5. Buka Pull Request

---

## Lisensi

Project ini menggunakan lisensi MIT. Lihat file `LICENSE` untuk detail lengkap.
