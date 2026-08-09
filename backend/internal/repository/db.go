// Package repository — database layer (schema, seed, helpers).
package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"printmart/backend/internal/config"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  phone TEXT,
  role TEXT NOT NULL DEFAULT 'user',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS addresses (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  label TEXT NOT NULL DEFAULT 'Rumah',
  full_address TEXT NOT NULL,
  city TEXT NOT NULL,
  postal_code TEXT,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS categories (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL DEFAULT 'product',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS products (
  id TEXT PRIMARY KEY,
  category_id TEXT REFERENCES categories(id),
  name TEXT NOT NULL,
  description TEXT DEFAULT '',
  price INTEGER NOT NULL DEFAULT 0,
  stock INTEGER NOT NULL DEFAULT 0,
  image_url TEXT DEFAULT '',
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS print_services (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  price_per_page INTEGER NOT NULL DEFAULT 0,
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS finishing_options (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL DEFAULT 'binding',
  price INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS cart_items (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  UNIQUE(user_id, product_id)
);
CREATE TABLE IF NOT EXISTS orders (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  order_type TEXT NOT NULL DEFAULT 'product',
  status TEXT NOT NULL DEFAULT 'menunggu_pembayaran',
  total_price INTEGER NOT NULL DEFAULT 0,
  delivery_method TEXT NOT NULL DEFAULT 'ambil',
  delivery_address TEXT DEFAULT '',
  notes TEXT DEFAULT '',
  customer_name TEXT DEFAULT '',
  subtotal INTEGER NOT NULL DEFAULT 0,
  discount INTEGER NOT NULL DEFAULT 0,
  voucher_code TEXT DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS order_items (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id TEXT REFERENCES products(id),
  product_name TEXT NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 1,
  unit_price INTEGER NOT NULL DEFAULT 0,
  subtotal INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS print_jobs (
  id TEXT PRIMARY KEY,
  order_id TEXT REFERENCES orders(id) ON DELETE SET NULL,
  file_url TEXT NOT NULL,
  file_name TEXT DEFAULT '',
  page_count INTEGER NOT NULL DEFAULT 1,
  color_mode TEXT NOT NULL DEFAULT 'bw',
  paper_size TEXT NOT NULL DEFAULT 'a4',
  copies INTEGER NOT NULL DEFAULT 1,
  duplex INTEGER NOT NULL DEFAULT 0,
  finishing_option_id TEXT REFERENCES finishing_options(id),
  estimated_price INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'menunggu',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS payments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  provider TEXT NOT NULL DEFAULT 'midtrans-sandbox',
  transaction_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  amount INTEGER NOT NULL DEFAULT 0,
  va_number TEXT,
  paid_at TEXT,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS notifications (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  order_id TEXT REFERENCES orders(id),
  message TEXT NOT NULL,
  is_read INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS vouchers (
  id TEXT PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  description TEXT DEFAULT '',
  discount_type TEXT NOT NULL DEFAULT 'percent',
  discount_value INTEGER NOT NULL DEFAULT 0,
  min_spend INTEGER NOT NULL DEFAULT 0,
  valid_until TEXT,
  usage_limit INTEGER NOT NULL DEFAULT 0,
  used_count INTEGER NOT NULL DEFAULT 0,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_print_jobs_order ON print_jobs(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers(code);
`

// migrate menambah kolom baru pada DB lama secara idempoten (SQLite tidak
// punya "ADD COLUMN IF NOT EXISTS", jadi error "duplicate column" diabaikan).
func migrate(db *sql.DB) {
	adds := []string{
		`ALTER TABLE orders ADD COLUMN subtotal INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE orders ADD COLUMN discount INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE orders ADD COLUMN voucher_code TEXT DEFAULT ''`,
	}
	for _, stmt := range adds {
		db.Exec(stmt) // abaikan error bila kolom sudah ada
	}
}

func Open(cfg *config.Config) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", cfg.DatabasePath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // sqlite: single writer
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	migrate(db)
	return db, nil
}

func Now() string { return time.Now().UTC().Format(time.RFC3339) }

func NowUnixMilli() int64 { return time.Now().UnixMilli() }

func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}

func Notify(db *sql.DB, userID, orderID, message string) {
	db.Exec(`INSERT INTO notifications (id, user_id, order_id, message, is_read, created_at) VALUES (?, ?, ?, ?, 0, ?)`,
		UUID(), userID, orderID, message, Now())
}