# PrintKu — Toko ATK & Jasa Fotocopy Online

Platform e-commerce untuk toko alat tulis kantor (ATK), perlengkapan sekolah, dan jasa print/fotocopy/jilid secara online. Dibangun dengan **Go** (backend) dan **SvelteKit** (frontend), payment gateway dalam mode **sandbox/testing**.

---

## Daftar Isi

- [Tentang Project](#tentang-project)
- [Fitur Utama](#fitur-utama)
- [Tech Stack](#tech-stack)
- [Arsitektur Sistem](#arsitektur-sistem)
- [Struktur Database](#struktur-database)
- [Struktur Folder Project](#struktur-folder-project)
- [Cara Instalasi & Menjalankan](#cara-instalasi--menjalankan)
- [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
- [Security](#security)
- [Roadmap](#roadmap)
- [Kontribusi](#kontribusi)
- [Lisensi](#lisensi)

---

## Tentang Project

**PrintKu** adalah platform digital yang menggabungkan dua kebutuhan sehari-hari masyarakat sekitar sekolah/perkantoran:

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
- Upload file (PDF, DOCX, JPG, PNG) — format .doc **tidak didukung** (menggunakan format ZIP-based DOCX)
- Pilih spesifikasi: warna/hitam-putih, ukuran kertas (A4/F4/A3), jumlah rangkap, satu/dua sisi
- Kalkulasi harga otomatis berdasarkan jumlah halaman hasil parsing server-side
- Estimasi waktu selesai
- Tracking status pesanan real-time: `Menunggu` → `Diproses` → `Siap Diambil/Dikirim` → `Selesai`

### 📎 Jasa Tambahan
- Jilid (spiral, lakban, hardcover)
- Laminating (berbagai ukuran)
- Scan dokumen fisik ke PDF

### 👤 Manajemen Pengguna
- Registrasi & login (JWT-based auth, HS256)
- Profil & alamat pengiriman
- Riwayat transaksi

### 🛠️ Admin Dashboard
- CRUD produk & kategori
- Kelola pesanan (produk & jasa print)
- Update status pesanan jasa print
- Laporan penjualan sederhana
- Kelola voucher/promo

### 🔔 Notifikasi
- Notifikasi real-time via WebSocket saat status pesanan berubah
- WebSocket membutuhkan autentikasi JWT

### 🎫 Voucher & Diskon
- Voucher persen atau nominal
- Minimum belanja
- Kuota penggunaan
- Validasi waktu (expiry) menggunakan time.Time
- Diskon hanya dihitung saat checkout, used_count diincrement saat payment sukses

---

## Tech Stack

| Layer | Teknologi |
|---|---|
| Backend | Go (Golang), Gin (HTTP framework) |
| Frontend | SvelteKit (Svelte 5), adapter-node |
| Database | SQLite (modernc.org/sqlite, pure Go, WAL mode) |
| Cache / Queue | Redis (opsional — fallback in-memory untuk development) |
| Auth | JWT (HS256, typed claims) |
| Real-time | WebSocket (Gorilla, dengan autentikasi) |
| Payment | Midtrans — **Sandbox Mode** (simulasi lokal) |
| File Storage | Local storage (UUID filenames, auth-protected download) |
| Containerization | Docker & Docker Compose |

---

## Arsitektur Sistem

```
┌─────────────┐        ┌─────────────────┐        ┌──────────────┐
│   SvelteKit │ <----> │   Go REST API    │ <----> │    SQLite    │
│  (Frontend) │        │    (Backend)     │        │  (Database)  │
└─────────────┘        └─────────────────┘        └──────────────┘
                               │
                 ┌─────────────┼─────────────┐
                 │             │             │
           ┌─────────┐   ┌──────────┐  ┌───────────┐
           │  Redis  │   │ WebSocket│  │  Payment   │
           │ (Queue) │   │  Server  │  │  Gateway   │
           │         │   │  (auth)  │  │ (Sandbox)  │
           └─────────┘   └──────────┘  └───────────┘
```

---

## Struktur Database

```
users            (id, name, email, password_hash, phone, role, created_at)
addresses        (id, user_id, label, full_address, city, postal_code, created_at)
categories       (id, name, type, created_at)
products         (id, category_id, name, description, price, stock, image_url, is_active, created_at)
print_services   (id, name, price_per_page, description, created_at)
finishing_options(id, name, type, price, created_at)
cart_items       (id, user_id, product_id, quantity, created_at, UNIQUE(user_id, product_id))
orders           (id, user_id, order_type, status, total_price, delivery_method, delivery_address, notes, customer_name, subtotal, discount, voucher_code, created_at)
order_items      (id, order_id, product_id, product_name, quantity, unit_price, subtotal)
print_jobs       (id, order_id, file_url, file_name, page_count, color_mode, paper_size, copies, duplex, finishing_option_id, estimated_price, status, created_at)
payments         (id, order_id, provider, transaction_id, status, amount, va_number, paid_at, created_at)
notifications    (id, user_id, order_id, message, is_read, created_at)
vouchers         (id, code, description, discount_type, discount_value, min_spend, valid_until, usage_limit, used_count, is_active, created_at)
```

---

## Struktur Folder Project

```
printku/
├── backend/
│   ├── cmd/
│   │   └── server/main.go          # Entrypoint + graceful shutdown
│   ├── internal/
│   │   ├── config/                 # Env & konfigurasi
│   │   ├── handler/                # HTTP handlers (controller)
│   │   ├── middleware/              # Auth (JWT HS256), rate limit
│   │   ├── model/                  # Domain models
│   │   ├── queue/                  # Redis/in-memory job queue
│   │   ├── repository/             # Database layer (SQLite + seed)
│   │   └── service/                # Business logic
│   ├── pkg/
│   │   └── websocket/              # WebSocket hub (auth, heartbeat)
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── routes/                 # SvelteKit pages
│   │   ├── lib/
│   │   │   ├── components/         # UI components
│   │   │   ├── api.js              # API client
│   │   │   ├── stores.js           # State management
│   │   │   └── format.js           # Formatting helpers
│   │   └── app.css                 # Design system
│   ├── package.json
│   └── Dockerfile
│
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Cara Instalasi & Menjalankan

### Prasyarat
- Go >= 1.22
- Node.js >= 18
- Redis >= 6 (opsional — fallback in-memory jika tidak tersedia)
- Docker & Docker Compose (opsional)

### Opsi 1: Docker Compose (disarankan)

```bash
git clone https://github.com/zharrDev/PrintKu.git
cd PrintKu
cp .env.example .env
docker compose up --build
```

Frontend: `http://localhost:5173`
Backend API: `http://localhost:8080`

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
npm run build
node build
```

> **Catatan:** Tidak perlu migrasi database manual — skema & seed (termasuk admin default) dibuat otomatis saat server pertama kali dijalankan.

---

## Environment Variables

```env
# Server
PORT=8080
APP_ENV=development

# Database (SQLite pure-Go)
DATABASE_PATH=./data/printku.db

# Redis (opsional — kosong = fallback memory)
REDIS_ADDR=localhost:6379

# JWT (minimal 32 karakter untuk production)
JWT_SECRET=ganti_dengan_secret_yang_kuat_printku_2024
JWT_EXPIRES_IN=24h

# Payment (sandbox)
PAYMENT_PROVIDER=midtrans-sandbox
PAYMENT_IS_PRODUCTION=false

# File upload
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=20

# CORS
CORS_ORIGIN=http://localhost:5173

# Admin default
ADMIN_EMAIL=admin@printku.local
ADMIN_PASSWORD=admin123
```

---

## API Endpoints

### Auth
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| POST | `/api/auth/register` | No | Registrasi user baru |
| POST | `/api/auth/login` | No | Login, return JWT |
| GET | `/api/auth/me` | Yes | Profil user saat ini |
| GET | `/api/auth/addresses` | Yes | Daftar alamat |
| POST | `/api/auth/addresses` | Yes | Tambah alamat |
| DELETE | `/api/auth/addresses/:id` | Yes | Hapus alamat |

### Produk
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/products` | Optional | List produk (search, filter, sort) |
| GET | `/api/products/:id` | Optional | Detail produk |
| POST | `/api/products` | Admin | Tambah produk |
| PUT | `/api/products/:id` | Admin | Update produk |
| DELETE | `/api/products/:id` | Admin | Hapus produk |

### Kategori
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/categories` | No | List kategori (filter by type) |

### Cart
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/cart` | Yes | Lihat isi cart |
| POST | `/api/cart` | Yes | Tambah item |
| PATCH | `/api/cart/:cartItemId` | Yes | Ubah jumlah |
| DELETE | `/api/cart/:cartItemId` | Yes | Hapus item |
| DELETE | `/api/cart` | Yes | Kosongkan cart |

### Order
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| POST | `/api/orders` | Yes | Checkout / buat order |
| GET | `/api/orders` | Yes | Riwayat order |
| GET | `/api/orders/:id` | Yes | Detail order |
| PATCH | `/api/orders/:id/status` | Admin | Update status |

### Print Jobs
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/print-jobs/options` | No | Opsi print & finishing |
| POST | `/api/print-jobs/upload` | Yes | Upload file (validasi MIME) |
| POST | `/api/print-jobs` | Yes | Buat print job |
| GET | `/api/print-jobs` | Yes | List print job saya |
| GET | `/api/print-jobs/:id` | Yes | Detail print job (ownership check) |
| PATCH | `/api/print-jobs/:id` | Yes | Update spesifikasi |
| PATCH | `/api/print-jobs/:id/status` | Admin | Update status |

### Payment
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| POST | `/api/payments/create` | Yes | Buat transaksi (sandbox) |
| POST | `/api/payments/webhook` | **No** | Callback dari payment gateway |
| GET | `/api/payments/:orderId` | Yes | Status pembayaran |

### File
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/files/download/:filename` | Yes | Download file (ownership check) |

### Notification
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/notifications` | Yes | Daftar notifikasi |
| POST | `/api/notifications/read-all` | Yes | Tandai semua sudah dibaca |
| POST | `/api/notifications/:id/read` | Yes | Tandai sudah dibaca |

### Voucher
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| POST | `/api/vouchers/validate` | Yes | Validasi voucher |

### Admin
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/admin/stats` | Admin | Dashboard stats |
| GET | `/api/admin/orders` | Admin | Semua pesanan |
| GET | `/api/admin/print-jobs` | Admin | Semua print jobs |
| GET | `/api/admin/vouchers` | Admin | Semua voucher |
| POST | `/api/admin/vouchers` | Admin | Buat voucher |
| PATCH | `/api/admin/vouchers/:id` | Admin | Update voucher |
| DELETE | `/api/admin/vouchers/:id` | Admin | Hapus voucher |

### WebSocket
| Endpoint | Auth | Deskripsi |
|---|---|---|
| `/ws/orders/:userId?token=<jwt>` | JWT | Real-time update status pesanan |

> WebSocket membutuhkan token JWT valid. User hanya bisa subscribe ke event milik sendiri.

### Health
| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| GET | `/api/health` | No | Server health check |

---

## Security

### Autentikasi & Otorisasi
- JWT dengan signing method **HS256** — metode lain ditolak
- Token divalidasi terhadap database (user harus ada dan aktif)
- Role diambil dari database, bukan dari token client
- Server **gagal start** jika JWT_SECRET terlalu lemah di production

### WebSocket
- Autentikasi via JWT token (query param atau Authorization header)
- Origin allowlist dari konfigurasi
- Ping/pong heartbeat (60 detik timeout)
- Cleanup koneksi mati secara periodik

### File Upload
- Request body dibatasi sebelum diproses
- Validasi ekstensi, MIME type, dan magic bytes
- Nama file menggunakan UUID random (tidak dari client)
- Download membutuhkan authorization & ownership check
- Format `.doc` **tidak didukung** (hanya `.docx`)

### Payment
- Webhook payment gateway **tidak membutuhkan JWT user**
- Sandbox mode hanya untuk testing
- Payment idempotent — callback success yang sama tidak menggandakan efek
- Voucher used_count diincrement hanya saat payment sukses

### Input Validation
- Enum validation untuk colorMode, paperSize, status order
- Validasi negatif ditolak
- Copies dibatasi max 100

---

## Payment Gateway (Mode Testing)

Semua transaksi menggunakan **sandbox environment** — tidak ada uang asli.

**Checklist sebelum go production:**
- [ ] Ganti API key sandbox → production
- [ ] Set `PAYMENT_IS_PRODUCTION=true`
- [ ] Implementasikan signature verification webhook
- [ ] Uji ulang seluruh flow checkout end-to-end
- [ ] Pastikan webhook URL sudah publicly accessible (HTTPS)

---

## Roadmap

### Selesai
- [x] Autentikasi user (register/login) dengan JWT HS256
- [x] Katalog produk & cart
- [x] Jasa print/fotocopy dengan kalkulasi harga otomatis
- [x] Payment integration (sandbox)
- [x] Real-time order tracking via WebSocket (authenticated)
- [x] Voucher & sistem diskon
- [x] Order state machine dengan validasi transisi
- [x] File upload dengan validasi MIME & magic bytes
- [x] Auth-protected file download
- [x] Redis queue dengan retry & dead-letter

### Dalam Pengembangan
- [ ] Multi-alamat pengiriman
- [ ] Pagination untuk produk, order, print job
- [ ] Structured logging & metrics
- [ ] OpenAPI documentation
- [ ] Rate limiting untuk login & register
- [ ] Migrasi ke TypeScript untuk frontend
- [ ] Backup strategy untuk database

### Future
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

Project ini menggunakan lisensi MIT.
