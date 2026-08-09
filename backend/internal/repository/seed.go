package repository

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"

	"printmart/backend/internal/config"
)

func Seed(db *sql.DB, cfg *config.Config) error {
	log.Printf("[seed] mulai inisialisasi data...")
	return seedAll(db, cfg)
}

func seedAll(db *sql.DB, cfg *config.Config) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return seedAdmin(db, cfg)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	catInsert := `INSERT INTO categories (id, name, type, created_at) VALUES (?, ?, ?, ?)`
	cat := func(name, typ string) string {
		id := UUID()
		tx.Exec(catInsert, id, name, typ, Now())
		return id
	}
	prod := func(catID, name string, price, stock int64, desc string) {
		tx.Exec(`INSERT INTO products (id, category_id, name, description, price, stock, image_url, is_active, created_at)
		         VALUES (?, ?, ?, ?, ?, ?, '', 1, ?)`,
			UUID(), catID, name, desc, price, stock, Now())
	}

	atk := cat("Alat Tulis Kantor (ATK)", "product")
	kertas := cat("Kertas & Tinta", "product")
	sekolah := cat("Perlengkapan Sekolah", "product")
	furniture := cat("Furniture Kantor", "product")

	prod(atk, "Pulpen Standard AE7 Hitam", 3000, 120, "Pulpen ballpoint standar dengan tinta hitam, nyaman untuk menulis sehari-hari.")
	prod(atk, "Pensil 2B Faber-Castell", 5000, 80, "Pensil kayu 2B premium, mudah dihapus dan tidak mudah patah.")
	prod(atk, "Buku Tulis SiDU 38 Lembar", 3500, 200, "Buku tulis isi 38 lembar, kertas halus dan tebal.")
	prod(atk, "Penghapus Joyko", 3000, 90, "Penghapus karet putih bersih, tidak meninggalkan noda.")
	prod(atk, "Penggaris 30cm Kenko", 6000, 40, "Penggaris plastik bening 30cm dengan skala presisi.")
	prod(atk, "Stabilo Boss Highlighter", 12000, 50, "Highlighter warna-warni dengan tinta tahan lama.")
	prod(atk, "Lakban Bening 2 inch", 8500, 60, "Lakban bening serbaguna untuk packing dan keseharian.")
	prod(atk, "Binder Clip No.3", 1500, 200, "Penjepit kertas ukuran sedang, isi 12 pcs.")
	prod(kertas, "Kertas HVS A4 70gsm (1 rim)", 48000, 30, "Kertas HVS putih 70gsm ukuran A4, 500 lembar per rim.")
	prod(kertas, "Kertas HVS F4 70gsm (1 rim)", 55000, 25, "Kertas HVS putih 70gsm ukuran F4 folio, 500 lembar per rim.")
	prod(kertas, "Kertas HVS A3 80gsm (1 rim)", 95000, 15, "Kertas HVS putih 80gsm ukuran A3, 500 lembar per rim.")
	prod(kertas, "Tinta Epson 003 Black 65ml", 85000, 20, "Botol tinta original Epson 003 warna hitam, isi 65ml.")
	prod(kertas, "Tinta Epson 664 CMYK 70ml", 95000, 20, "Botol tinta original Epson 664, 1 set 4 botol 70ml.")
	prod(sekolah, "Tempat Pensil Jepit", 15000, 35, "Tempat pensil kain dengan penutup jepit, motif simple.")
	prod(sekolah, "Buku Gambar A4", 9000, 45, "Buku gambar ukuran A4, isi 30 halaman kertas gambar.")
	prod(sekolah, "Map Folder Tali", 12000, 40, "Map plastik tali bening ukuran folio.")
	prod(furniture, "Rak Buku Mini 3 Tingkat", 85000, 8, "Rak buku mini kayu 3 tingkat, cocok untuk meja kerja.")
	prod(furniture, "Kursi Lipat Portable", 180000, 6, "Kursi lipat portable kokoh, ringan dan mudah dibawa.")

	svc := func(name string, price int64, desc string) {
		tx.Exec(`INSERT INTO print_services (id, name, price_per_page, description, created_at) VALUES (?, ?, ?, ?, ?)`,
			UUID(), name, price, desc, Now())
	}
	svc("Print Hitam Putih", 500, "Print dokumen hitam putih per halaman. Ukuran A4/F4/A3 tersedia.")
	svc("Print Warna", 1500, "Print dokumen warna per halaman.")
	svc("Fotocopy", 300, "Fotocopy dokumen hitam putih per halaman.")

	fin := func(name, typ string, price int64) {
		tx.Exec(`INSERT INTO finishing_options (id, name, type, price, created_at) VALUES (?, ?, ?, ?, ?)`,
			UUID(), name, typ, price, Now())
	}
	fin("Jilid Spiral", "binding", 5000)
	fin("Jilid Lakban", "binding", 2500)
	fin("Jilid Hardcover", "binding", 25000)
	fin("Laminating A4", "lamination", 5000)
	fin("Laminating A3", "lamination", 8000)
	fin("Laminating Kartu", "lamination", 3000)
	fin("Scan Dokumen per Halaman", "scan", 1000)

	vou := func(code, desc, dtype string, value, minSpend, usageLimit int64) {
		tx.Exec(`INSERT INTO vouchers (id, code, description, discount_type, discount_value, min_spend, valid_until, usage_limit, used_count, is_active, created_at)
		         VALUES (?, ?, ?, ?, ?, ?, NULL, ?, 0, 1, ?)`,
			UUID(), code, desc, dtype, value, minSpend, usageLimit, Now())
	}
	vou("HEMAT10", "Diskon 10% untuk semua pesanan", "percent", 10, 0, 0)
	vou("POTONG5K", "Potongan Rp 5.000 min. belanja Rp 25.000", "fixed", 5000, 25000, 0)
	vou("NGANTOR20", "Diskon 20% min. belanja Rp 100.000 (kuota 50)", "percent", 20, 100000, 50)

	if err := tx.Commit(); err != nil {
		return err
	}
	return seedAdmin(db, cfg)
}

func seedAdmin(db *sql.DB, cfg *config.Config) error {
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, cfg.AdminEmail).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 10)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO users (id, name, email, password_hash, phone, role, created_at) VALUES (?, ?, ?, ?, ?, 'admin', ?)`,
		UUID(), "Admin PrintMart", cfg.AdminEmail, string(hash), "0812-0000-0000", Now())
	log.Printf("[seed] Admin default dibuat: %s / %s", cfg.AdminEmail, cfg.AdminPassword)
	return err
}