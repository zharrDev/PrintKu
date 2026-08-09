package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/repository"
	"printmart/backend/internal/service"
)

var allowedExt = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true,
}

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
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File wajib diupload"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ekstensi tidak didukung. Gunakan: .pdf, .doc, .docx, .jpg, .jpeg, .png"})
		return
	}
	if file.Size > s.Cfg.MaxUploadSizeMB*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("File terlalu besar. Maksimal %dMB", s.Cfg.MaxUploadSizeMB)})
		return
	}

	storedName := fmt.Sprintf("%d-%s%s", repository.NowUnixMilli(), newID(), ext)
	dst := filepath.Join(s.Cfg.UploadDir, storedName)
	originalName := file.Filename
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
	fileURL := "/uploads/" + storedName

	status := "uploaded"
	c.JSON(http.StatusCreated, gin.H{
		"upload": gin.H{
			"fileUrl": fileURL, "fileName": originalName, "pageCount": pageCount,
			"size": file.Size, "status": status,
		},
	})
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
		FileURL            string `json:"fileUrl"`
		FileName           string `json:"fileName"`
		PageCount          any     `json:"pageCount"`
		ColorMode          string `json:"colorMode"`
		PaperSize          string `json:"paperSize"`
		Copies             any     `json:"copies"`
		Duplex             any     `json:"duplex"`
		FinishingOptionID  string `json:"finishingOptionId"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.FileURL == "" || body.FileName == "" || body.PageCount == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileUrl, fileName, dan pageCount wajib diisi"})
		return
	}
	if body.ColorMode == "" {
		body.ColorMode = "bw"
	}
	if body.PaperSize == "" {
		body.PaperSize = "a4"
	}
	storedPath := filepath.Join(s.Cfg.UploadDir, filepath.Base(body.FileURL))
	if _, err := os.Stat(storedPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File tidak ditemukan di server"})
		return
	}

	serviceRow, err := s.serviceFor(body.ColorMode)
	if err != nil || serviceRow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan print tidak ditemukan"})
		return
	}
	var finishing map[string]any
	finishingPrice := int64(0)
	if body.FinishingOptionID != "" {
		if f, _ := getRow(s.DB, `SELECT * FROM finishing_options WHERE id = ?`, body.FinishingOptionID); f != nil {
			finishing = f
			finishingPrice = toInt(f["price"])
		}
	}

	pageCount := num(body.PageCount)
	copies := num(body.Copies)
	if copies < 1 {
		copies = 1
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
	var body struct {
		ColorMode         string `json:"colorMode"`
		PaperSize         string `json:"paperSize"`
		Copies            any    `json:"copies"`
		Duplex            any    `json:"duplex"`
		FinishingOptionID any    `json:"finishingOptionId"`
	}
	_ = c.ShouldBindJSON(&body)

	colorMode := body.ColorMode
	if colorMode == "" {
		colorMode = toStr(job["color_mode"])
	}
	paperSize := body.PaperSize
	if paperSize == "" {
		paperSize = toStr(job["paper_size"])
	}
	copies := num(body.Copies)
	if body.Copies == nil {
		copies = toInt(job["copies"])
	}
	if copies < 1 {
		copies = 1
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
	c.JSON(http.StatusOK, gin.H{"printJob": s.enrichJob(c, job)})
}

func (s *Server) patchPrintJobStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)
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