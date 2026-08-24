package service

import (
	"strings"
	"time"
)

// Voucher merepresentasikan baris voucher yang relevan untuk perhitungan diskon.
type Voucher struct {
	Code          string
	DiscountType  string
	DiscountValue int64
	MinSpend      int64
	ValidUntil    string
	UsageLimit    int64
	UsedCount     int64
	IsActive      bool
}

// CalcDiscount memvalidasi voucher terhadap subtotal & waktu sekarang,
// mengembalikan nominal diskon dan pesan error (kosong bila valid).
//
// PENTING:
// - Expiry divalidasi menggunakan time.Time, bukan string comparison.
// - Discount percent tidak boleh lebih dari 100.
// - used_count TIDAK diincrement di sini (hanya saat payment sukses).
func CalcDiscount(v Voucher, subtotal int64, nowRFC3339 string) (int64, string) {
	if !v.IsActive {
		return 0, "Voucher tidak aktif"
	}

	// Validasi expiry menggunakan time.Time, bukan string comparison
	if v.ValidUntil != "" {
		validUntil, err := time.Parse(time.RFC3339, v.ValidUntil)
		if err == nil {
			now, err := time.Parse(time.RFC3339, nowRFC3339)
			if err == nil && now.After(validUntil) {
				return 0, "Voucher sudah kedaluwarsa"
			}
		}
	}

	if v.UsageLimit > 0 && v.UsedCount >= v.UsageLimit {
		return 0, "Kuota voucher sudah habis"
	}
	if subtotal < v.MinSpend {
		return 0, "Belanja minimal " + FormatRupiah(v.MinSpend) + " untuk memakai voucher ini"
	}

	var discount int64
	switch strings.ToLower(v.DiscountType) {
	case "percent", "percentage":
		// Validasi: discount percent tidak boleh lebih dari 100
		if v.DiscountValue > 100 {
			return 0, "Persentase diskon tidak boleh lebih dari 100%"
		}
		discount = subtotal * v.DiscountValue / 100
	case "fixed", "nominal":
		discount = v.DiscountValue
	default:
		return 0, "Tipe voucher tidak dikenal"
	}

	if discount > subtotal {
		discount = subtotal
	}
	if discount < 0 {
		discount = 0
	}
	return discount, ""
}
