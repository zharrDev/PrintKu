package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/repository"
	"printmart/backend/internal/service"
)

// allowedExt — ekstensi file yang didukung.
// CATATAN: .doc TIDAK didukung karena parser DOCX (ZIP-based)
// tidak bisa membaca format .doc (binary lama). Jika diperlukan,
// gunakan LibreOffice headless untuk konversi.
var allowedExt = map[string]bool{
	".pdf": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true,
}

// allowedMimeTypes — MIME types yang diizinkan per ekstensi.
// Validasi berlapis: extension + MIME type.
var allowedMimeTypes = map[string][]string{
	".pdf":  {"application/pdf"},
	".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/zip"},
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".png":  {"image/png"},
}

// maxBodySize — batas ukuran request body (25MB untuk upload header + file).
const maxBodySize = 25 << 20 // 25MB

func (s *Server) printOptions(c *gin.Context) {
	services, _ := allRows(s.DB, `SELECT * FROM print_services ORDER BY price_per_page`)
	finishing, _ := allRows(s.DB, `SELECT * FROM finishing_options ORDER BY price`)
	c.JSON(http.StatusOK, gin.H{
		"printServices": services,
		"finishingOptions": finishing,
		"paperSizes": []gin.H{
			{"id": "a4", "name": "A4"},
			{"id": "f4", "name": "F4 (Folio)"},
			{"id": "a3", "name": "A3"},
		},
		"colorModes": []gin.H{
			{"id": "bw", "name": "Hitam Putih"},
			{"id": "color", "name": "Warna"},
		},
	})
}

func (s *Server) uploadFile(c *gin.Context) {
	// Batasi request body sebelum file diproses
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)

	file, err := c.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("File terlalu besar. Maksimal %dMB", s.Cfg.MaxUploadSizeMB)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "File wajib diupload"})
		return
	}

	// Validasi ukuran file
	if file.Size > s.Cfg.MaxUploadSizeMB*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("File terlalu besar. Maksimal %dMB", s.Cfg.MaxUploadSizeMB)})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ekstensi tidak didukung. Gunakan: .pdf, .docx, .jpg, .jpeg, .png"})
		return
	}

	// Validasi MIME type dari header Content-Type
	mimeType := file.Header.Get("Content-Type")
	if mimeType != "" {
		// Ambil MIME type utama (tanpa parameter)
		mainMime := strings.Split(mimeType, ";")[0]
		mainMime = strings.TrimSpace(mainMime)
		validMimes := allowedMimeTypes[ext]
		mimeValid := false
		for _, vm := range validMimes {
			if mainMime == vm {
				mimeValid = true
				break
			}
		}
		if !mimeValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MIME type tidak sesuai dengan ekstensi file"})
			return
		}
	}

	// Validasi magic bytes (cek header file)
	if !validateMagicBytes(file, ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File corrupt atau Magic bytes tidak sesuai"})
		return
	}	// Simpan file dengan UUID random, bukan nama file client
	storedName := fmt.Sprintf("%d-%s%s", repository.NowUnixMilli(), newID(), ext)
	dst := filepath.Join(s.Cfg.UploadDir, storedName)
	originalName := file.Filename

	// Simpan file, lalu hitung page count dari data yang sudah terimpan
	if err := c.SaveUploadedFile(file, dst); err != nil {
		s.jsonErr(c, err)
		return
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		s.jsonErr(c, err)
		return
	}

	extNoDot := strings.TrimPrefix(ext, ".")
	pageCount := service.CountPages(extNoDot, data)
	fileURL := "/api/files/download/" + storedName

	status := "uploaded"
	c.JSON(http.StatusCreated, gin.H{
		"upload": gin.H{
			"fileUrl":    fileURL,
			"fileName":   originalName,
			"pageCount":  pageCount,
			"size":       file.Size,
			"status":     status,
			"storedName": storedName, // simpan untuk ownership check
		},
	})
}

// validateMagicBytes memvalidasi header file berdasarkan magic bytes.
func validateMagicBytes(file *multipart.FileHeader, ext string) bool {
	f, err := file.Open()
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil || n < 4 {
		return false
	}

	switch ext {
	case ".pdf":
		// PDF: %PDF
		return len(header) >= 4 && header[0] == '%' && header[1] == 'P' && header[2] == 'D' && header[3] == 'F'
	case ".docx":
		// DOCX: PK (ZIP header) — DOCX adalah file ZIP
		return len(header) >= 2 && header[0] == 'P' && header[1] == 'K'
	case ".jpg", ".jpeg":
		// JPEG: FF D8 FF
		return len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF
	case ".png":
		// PNG: 89 50 4E 47
		return len(header) >= 4 && header[0] == 0x89 && header[1] == 'P' && header[2] == 'N' && header[3] == 'G'
	}
	return false
}

func (s *Server) enrichJob(c *gin.Context, row map[string]any) gin.H {
	finishing := gin.H(nil)
	if v := toStr(row["finishing_option_id"]); v != "" {
		if f, _ := s.q1(c, `SELECT * FROM finishing_options WHERE id = ?`, v); f != nil {
			finishing = f
		}
	}
	return gin.H{
		"id": row["id"], "order_id": row["order_id"], "file_url": row["file_url"],
		"file_name": row["file_name"], "page_count": row["page_count"],
		"color_mode": row["color_mode"], "paper_size": row["paper_size"],
		"copies": row["copies"], "duplex": row["duplex"],
		"finishing_option_id": row["finishing_option_id"],
		"estimated_price": row["estimated_price"], "status": row["status"],
		"created_at": row["created_at"],
		"finishing": finishing,
		"totalPages": toInt(row["page_count"]) * toInt(row["copies"]),
	}
}

func (s *Server) serviceFor(colorMode string) (map[string]any, error) {
	pat := "%Warna%"
	if colorMode == "bw" {
		pat = "%Hitam Putih%"
	}
	return getRow(s.DB, `SELECT * FROM print_services WHERE name LIKE ? ORDER BY price_per_page LIMIT 1`, pat)
}

func (s *Server) createPrintJob(c *gin.Context) {
	var body struct {
		FileURL           string `json:"fileUrl"`
		FileName          string `json:"fileName"`
		ColorMode         string `json:"colorMode"`
		PaperSize         string `json:"paperSize"`
		Copies            any    `json:"copies"`
		Duplex            any    `json:"duplex"`
		FinishingOptionID string `json:"finishingOptionId"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.FileURL == "" || body.FileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileUrl dan fileName wajib diisi"})
		return
	}
	if body.ColorMode == "" {
		body.ColorMode = "bw"
	}
	if body.PaperSize == "" {
		body.PaperSize = "a4"
	}

	// Validasi enum color_mode
	if body.ColorMode != "bw" && body.ColorMode != "color" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "colorMode harus 'bw' atau 'color'"})
		return
	}

	// Validasi enum paper_size
	validPaperSizes := map[string]bool{"a4": true, "f4": true, "a3": true}
	if !validPaperSizes[body.PaperSize] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paperSize harus 'a4', 'f4', atau 'a3'"})
		return
	}

	// Pastikan file ada di server
	storedPath := filepath.Join(s.Cfg.UploadDir, filepath.Base(body.FileURL))
	if _, err := os.Stat(storedPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File tidak ditemukan di server"})
		return
	}

	// Ownership check: pastikan user memiliki file ini
	u := s.user(c)
	if !s.userOwnsFile(u.ID, filepath.Base(body.FileURL)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: file tidak dimiliki"})
		return
	}

	// Ambil page count dari metadata file yang sudah diverifikasi server
	// JANGAN percaya pageCount dari client!
	data, err := os.ReadFile(storedPath)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	extNoDot := strings.TrimPrefix(strings.ToLower(filepath.Ext(body.FileURL)), ".")
	pageCount := service.CountPages(extNoDot, data)

	serviceRow, err := s.serviceFor(body.ColorMode)
	if err != nil || serviceRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan print tidak ditemukan"})
		return
	}
	var finishing map[string]any
	finishingPrice := int64(0)
	if body.FinishingOptionID != "" {
		// Validasi finishing option ID
		if f, _ := getRow(s.DB, `SELECT * FROM finishing_options WHERE id = ?`, body.FinishingOptionID); f != nil {
			finishing = f
			finishingPrice = toInt(f["price"])
		}
	}

	copies := num(body.Copies)
	if copies < 1 {
		copies = 1
	}
	// Validasi copies max
	if copies > 100 {
		copies = 100
	}
	duplex := toBoolInt(body.Duplex)
	pricing := service.CalculatePrintPrice(toInt(serviceRow["price_per_page"]), pageCount, copies, finishingPrice, body.PaperSize)
	minutes := service.EstimateMinutes(pageCount, copies)

	id := newID()
	if _, err := s.DB.Exec(
		`INSERT INTO print_jobs (id, order_id, file_url, file_name, page_count, color_mode, paper_size, copies, duplex, finishing_option_id, estimated_price, status, created_at)
		 VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'menunggu', ?)`,
		id, body.FileURL, body.FileName, pageCount, body.ColorMode, body.PaperSize, copies, duplex,
		nilableID(finishing), pricing.Total, repository.Now(),
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	row, _ := getRow(s.DB, `SELECT * FROM print_jobs WHERE id = ?`, id)
	c.JSON(http.StatusCreated, gin.H{
		"printJob": s.enrichJob(c, row),
		"pricing":  gin.H{"perPage": pricing.PerPage, "subtotal": pricing.Subtotal, "finishingPrice": pricing.FinishingPrice, "total": pricing.Total, "formatted": pricing.Formatted},
		"estimateMinutes": minutes,
	})
}

func nilableID(m map[string]any) any {
	if m == nil {
		return nil
	}
	return m["id"]
}

func (s *Server) patchPrintJob(c *gin.Context) {
	job, err := getRow(s.DB, `SELECT * FROM print_jobs WHERE id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Print job tidak ditemukan"})
		return
	}

	// Ownership check
	u := s.user(c)
	if !s.userOwnsPrintJob(c, u.ID, toStr(job["id"])) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var body struct {
		ColorMode        string `json:"colorMode"`
		PaperSize        string `json:"paperSize"`
		Copies           any    `json:"copies"`
		Duplex           any    `json:"duplex"`
		FinishingOptionID any    `json:"finishingOptionId"`
	}
	_ = c.ShouldBindJSON(&body)

	colorMode := body.ColorMode
	if colorMode == "" {
		colorMode = toStr(job["color_mode"])
	}
	// Validasi enum
	if colorMode != "bw" && colorMode != "color" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "colorMode harus 'bw' atau 'color'"})
		return
	}

	paperSize := body.PaperSize
	if paperSize == "" {
		paperSize = toStr(job["paper_size"])
	}
	validPaperSizes := map[string]bool{"a4": true, "f4": true, "a3": true}
	if !validPaperSizes[paperSize] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paperSize harus 'a4', 'f4', atau 'a3'"})
		return
	}

	copies := num(body.Copies)
	if body.Copies == nil {
		copies = toInt(job["copies"])
	}
	if copies < 1 {
		copies = 1
	}
	if copies > 100 {
		copies = 100
	}

	var duplex int64
	if body.Duplex != nil {
		duplex = toBoolInt(body.Duplex)
	} else {
		duplex = toInt(job["duplex"])
	}
	finishingOptionID := toStr(job["finishing_option_id"])
	if body.FinishingOptionID != nil {
		finishingOptionID = toStr(body.FinishingOptionID)
	}

	serviceRow, err := s.serviceFor(colorMode)
	if err != nil || serviceRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan print tidak ditemukan"})
		return
	}
	var finishing map[string]any
	finishingPrice := int64(0)
	if finishingOptionID != "" {
		if f, _ := getRow(s.DB, `SELECT * FROM finishing_options WHERE id = ?`, finishingOptionID); f != nil {
			finishing = f
			finishingPrice = toInt(f["price"])
		}
	}
	pageCount := toInt(job["page_count"])
	pricing := service.CalculatePrintPrice(toInt(serviceRow["price_per_page"]), pageCount, copies, finishingPrice, paperSize)

	if _, err := s.DB.Exec(
		`UPDATE print_jobs SET color_mode = ?, paper_size = ?, copies = ?, duplex = ?, finishing_option_id = ?, estimated_price = ? WHERE id = ?`,
		colorMode, paperSize, copies, duplex, nilableID(finishing), pricing.Total, job["id"],
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	row, _ := getRow(s.DB, `SELECT * FROM print_jobs WHERE id = ?`, job["id"])
	c.JSON(http.StatusOK, gin.H{
		"printJob": s.enrichJob(c, row),
		"pricing":  gin.H{"perPage": pricing.PerPage, "subtotal": pricing.Subtotal, "finishingPrice": pricing.FinishingPrice, "total": pricing.Total, "formatted": pricing.Formatted},
		"estimateMinutes": service.EstimateMinutes(toInt(row["page_count"]), copies),
	})
}

func (s *Server) listMyPrintJobs(c *gin.Context) {
	u := s.user(c)
	rows, err := allRows(s.DB, `SELECT pj.* FROM print_jobs pj
		INNER JOIN orders o ON o.id = pj.order_id
		WHERE o.user_id = ? ORDER BY pj.created_at DESC`, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.enrichJob(c, r))
	}
	c.JSON(http.StatusOK, gin.H{"printJobs": out})
}

func (s *Server) getPrintJob(c *gin.Context) {
	job, err := getRow(s.DB, `SELECT * FROM print_jobs WHERE id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Print job tidak ditemukan"})
		return
	}

	// Ownership check — user hanya bisa melihat print job miliknya
	u := s.user(c)
	if !s.userOwnsPrintJob(c, u.ID, toStr(job["id"])) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"printJob": s.enrichJob(c, job)})
}

func (s *Server) patchPrintJobStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)

	// Validasi enum status
	allowed := []string{"menunggu", "diproses", "siap", "selesai", "dibatalkan"}
	valid := false
	for _, st := range allowed {
		if st == body.Status {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status harus salah satu dari: " + strings.Join(allowed, ", ")})
		return
	}

	job, err := getRow(s.DB, `SELECT * FROM print_jobs WHERE id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Print job tidak ditemukan"})
		return
	}
	if _, err := s.DB.Exec(`UPDATE print_jobs SET status = ? WHERE id = ?`, body.Status, job["id"]); err != nil {
		s.jsonErr(c, err)
		return
	}
	updated, _ := getRow(s.DB, `SELECT * FROM print_jobs WHERE id = ?`, job["id"])

	if v := toStr(job["order_id"]); v != "" {
		order, _ := getRow(s.DB, `SELECT * FROM orders WHERE id = ?`, v)
		if order != nil {
			repository.Notify(s.DB, toStr(order["user_id"]), v,
				"Status print \""+toStr(job["file_name"])+"\" menjadi \""+body.Status+"\"")
			s.broadcast(toStr(order["user_id"]), gin.H{
				"type": "order_status", "orderId": v,
				"message": "Status print \"" + toStr(job["file_name"]) + "\" menjadi \"" + body.Status + "\"",
			})
		}
		service.SyncOrderStatusFromJobs(s.DB, v, s.broadcast)
	}
	c.JSON(http.StatusOK, gin.H{"printJob": s.enrichJob(c, updated)})
}

// userOwnsPrintJob memeriksa apakah print job dimiliki oleh user.
func (s *Server) userOwnsPrintJob(c *gin.Context, userID, jobID string) bool {
	job, err := s.q1(c, `SELECT * FROM print_jobs WHERE id = ?`, jobID)
	if err != nil || job == nil {
		return false
	}
	// Jika print job belum terkait order, izinkan (user baru upload)
	if v := toStr(job["order_id"]); v == "" {
		return true
	}
	// Cek ownership via order
	order, err := getRow(s.DB, `SELECT user_id FROM orders WHERE id = ?`, toStr(job["order_id"]))
	if err != nil || order == nil {
		return false
	}
	return toStr(order["user_id"]) == userID
}
