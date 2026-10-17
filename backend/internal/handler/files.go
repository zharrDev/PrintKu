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
//
// Aturan:
//  1. File yang diunggah user tercatat di tabel file_uploads — inilah
//     pemilik yang sebenarnya, terverifikasi saat upload.
//  2. File tetap boleh diakses pemilik order yang memakai file tersebut
//     (menutup data lama yang sudah terikat pesanan).
//
// File tanpa catatan upload dan tanpa order tidak dapat diverifikasi
// pemiliknya, sehingga aksesnya ditolak (bukan diloloskan).
func (s *Server) userOwnsFile(userID, filename string) bool {
	var count int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM file_uploads WHERE stored_name = ? AND user_id = ?`,
		filename, userID).Scan(&count)
	if err == nil && count > 0 {
		return true
	}

	fileURL := "/api/files/download/" + filename
	err = s.DB.QueryRow(`
		SELECT COUNT(*) FROM print_jobs pj
		WHERE pj.file_url = ?
		AND pj.order_id IS NOT NULL
		AND EXISTS (
			SELECT 1 FROM orders o
			WHERE o.id = pj.order_id AND o.user_id = ?
		)`, fileURL, userID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
