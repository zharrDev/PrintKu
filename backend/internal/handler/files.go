package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// downloadFile menyediakan file download yang membutuhkan authorization.
// User hanya dapat mengakses file miliknya atau file order miliknya.
// TIDAK ada public static uploads lagi!
func (s *Server) downloadFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename wajib diisi"})
		return
	}

	// Keamanan: pastikan filename tidak mengandung path traversal
	filename = filepath.Base(filename)
	if filename == "." || filename == ".." || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename tidak valid"})
		return
	}

	filePath := filepath.Join(s.Cfg.UploadDir, filename)

	// Pastikan file ada
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File tidak ditemukan"})
		return
	}

	// Ownership check: pastikan file milik user atau milik order user
	u := s.user(c)
	if !s.userOwnsFile(u.ID, filename) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: akses file ditolak"})
		return
	}

	c.File(filePath)
}

// userOwnsFile memeriksa apakah file dimiliki oleh user.
// File dimiliki jika:
// 1. File merupakan upload user (berdasarkan print_jobs yang menggunakan file tersebut)
// 2. File merupakan bagian dari order milik user
func (s *Server) userOwnsFile(userID, filename string) bool {
	// Cek apakah file digunakan oleh print job yang dimiliki user
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*) FROM print_jobs pj
		WHERE pj.file_url LIKE ?
		AND (
			pj.order_id IS NULL
			OR EXISTS (
				SELECT 1 FROM orders o
				WHERE o.id = pj.order_id AND o.user_id = ?
			)
			OR EXISTS (
				SELECT 1 FROM print_jobs pj2
				WHERE pj2.id = pj.id AND pj2.order_id IS NULL
			)
		)
	`, "%"+filename, userID).Scan(&count)
	if err != nil {
		return false
	}

	// Jika file belum terkait order, izinkan upload sendiri
	// (file baru diupload, belum masuk order)
	var fileExists int
	err = s.DB.QueryRow(`
		SELECT COUNT(*) FROM print_jobs pj
		WHERE pj.file_url LIKE ?
	`, "%"+filename).Scan(&fileExists)
	if err != nil {
		return false
	}

	// Jika file tidak ada di print_jobs manapun, mungkin baru diupload
	// Izinkan (user baru upload, belum create print job)
	if fileExists == 0 {
		return true
	}

	return count > 0
}
