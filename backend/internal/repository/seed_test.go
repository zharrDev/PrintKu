package repository

import (
	"path/filepath"
	"testing"
	"time"

	"printmart/backend/internal/config"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		AppEnv:        "test",
		DatabasePath:  filepath.Join(dir, "seed.db"),
		JWTSecret:     "seed-test-secret",
		JWTExpiresIn:  time.Hour,
		AdminEmail:    "admin@seed.test",
		AdminPassword: "rahasia123",
	}
}

// Seed pertama harus mengisi foto untuk seluruh produk awal.
func TestSeedFillsProductImages(t *testing.T) {
	cfg := testConfig(t)
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := Seed(db, cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rows, err := db.Query(`SELECT name, COALESCE(image_url, '') FROM products`)
	if err != nil {
		t.Fatalf("query products: %v", err)
	}
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var name, img string
		if err := rows.Scan(&name, &img); err != nil {
			t.Fatalf("scan: %v", err)
		}
		seen[name] = true
		want, ok := seedProductImages[name]
		if !ok {
			t.Errorf("produk %q tidak ada di seedProductImages", name)
			continue
		}
		if img != want {
			t.Errorf("image_url %q = %q, mau %q", name, img, want)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterasi: %v", err)
	}
	for name := range seedProductImages {
		if !seen[name] {
			t.Errorf("seedProductImages punya %q tapi produk tidak di-seed", name)
		}
	}
}

// DB lama (produk sudah ada, image_url kosong) harus dilengkapi tanpa
// menimpa foto yang sudah diubah admin, dan tanpa menduplikasi data.
func TestBackfillProductImagesOnExistingDB(t *testing.T) {
	cfg := testConfig(t)
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := Seed(db, cfg); err != nil {
		t.Fatalf("seed awal: %v", err)
	}

	// Tiru DB lama: semua foto kosong, satu produk sudah punya foto sendiri.
	if _, err := db.Exec(`UPDATE products SET image_url = ''`); err != nil {
		t.Fatalf("kosongkan image_url: %v", err)
	}
	if _, err := db.Exec(`UPDATE products SET image_url = '/uploads/foto-toko.jpg' WHERE name = ?`, "Binder Clip No.3"); err != nil {
		t.Fatalf("set foto admin: %v", err)
	}

	if err := Seed(db, cfg); err != nil {
		t.Fatalf("seed ulang: %v", err)
	}
	// Backfill harus idempoten.
	if err := Seed(db, cfg); err != nil {
		t.Fatalf("seed ketiga: %v", err)
	}

	var empty int
	if err := db.QueryRow(`SELECT COUNT(*) FROM products WHERE COALESCE(image_url, '') = ''`).Scan(&empty); err != nil {
		t.Fatalf("hitung kosong: %v", err)
	}
	if empty != 0 {
		t.Errorf("masih ada %d produk tanpa foto setelah backfill", empty)
	}

	var custom string
	if err := db.QueryRow(`SELECT image_url FROM products WHERE name = ?`, "Binder Clip No.3").Scan(&custom); err != nil {
		t.Fatalf("baca foto admin: %v", err)
	}
	if custom != "/uploads/foto-toko.jpg" {
		t.Errorf("foto pilihan admin tertimpa: %q", custom)
	}

	var products, admins int
	if err := db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&products); err != nil {
		t.Fatalf("hitung produk: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&admins); err != nil {
		t.Fatalf("hitung admin: %v", err)
	}
	if products != len(seedProductImages) {
		t.Errorf("jumlah produk berubah jadi %d", products)
	}
	if admins != 1 {
		t.Errorf("admin terduplikasi: %d", admins)
	}
}
