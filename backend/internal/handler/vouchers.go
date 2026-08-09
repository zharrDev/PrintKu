package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/repository"
	"printmart/backend/internal/service"
)

// validateVoucher dipakai halaman checkout untuk melihat pratinjau diskon
// terhadap subtotal tertentu tanpa membuat order.
func (s *Server) validateVoucher(c *gin.Context) {
	var body struct {
		Code     string `json:"code"`
		Subtotal any    `json:"subtotal"`
	}
	_ = c.ShouldBindJSON(&body)
	code := strings.ToUpper(strings.TrimSpace(body.Code))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode voucher wajib diisi"})
		return
	}
	subtotal := num(body.Subtotal)
	v, err := getRow(s.DB, `SELECT * FROM vouchers WHERE UPPER(code) = ?`, code)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Voucher tidak ditemukan"})
		return
	}
	discount, msg := service.CalcDiscount(voucherFromRow(v), subtotal, repository.Now())
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg, "valid": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"voucher":  v,
		"discount": discount,
		"subtotal": subtotal,
		"total":    subtotal - discount,
	})
}

// ===== Admin CRUD voucher =====

func (s *Server) adminListVouchers(c *gin.Context) {
	rows, err := allRows(s.DB, `SELECT * FROM vouchers ORDER BY created_at DESC`)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"vouchers": rows})
}

func (s *Server) adminCreateVoucher(c *gin.Context) {
	var body struct {
		Code          string `json:"code"`
		Description   string `json:"description"`
		DiscountType  string `json:"discount_type"`
		DiscountValue any    `json:"discount_value"`
		MinSpend      any    `json:"min_spend"`
		ValidUntil    string `json:"valid_until"`
		UsageLimit    any    `json:"usage_limit"`
		IsActive      any    `json:"is_active"`
	}
	_ = c.ShouldBindJSON(&body)
	code := strings.ToUpper(strings.TrimSpace(body.Code))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode voucher wajib diisi"})
		return
	}
	dtype := strings.ToLower(strings.TrimSpace(body.DiscountType))
	if dtype != "percent" && dtype != "fixed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount_type harus 'percent' atau 'fixed'"})
		return
	}
	if num(body.DiscountValue) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount_value harus lebih dari 0"})
		return
	}
	var dup string
	if err := s.DB.QueryRow(`SELECT id FROM vouchers WHERE UPPER(code) = ?`, code).Scan(&dup); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Kode voucher sudah ada"})
		return
	}
	isActive := int64(1)
	if body.IsActive != nil {
		isActive = toBoolInt(body.IsActive)
	}
	id := newID()
	if _, err := s.DB.Exec(
		`INSERT INTO vouchers (id, code, description, discount_type, discount_value, min_spend, valid_until, usage_limit, used_count, is_active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		id, code, body.Description, dtype, num(body.DiscountValue), num(body.MinSpend),
		nullIfEmpty(body.ValidUntil), num(body.UsageLimit), isActive, repository.Now(),
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	row, _ := getRow(s.DB, `SELECT * FROM vouchers WHERE id = ?`, id)
	c.JSON(http.StatusCreated, gin.H{"voucher": row})
}

func (s *Server) adminUpdateVoucher(c *gin.Context) {
	existing, err := getRow(s.DB, `SELECT * FROM vouchers WHERE id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Voucher tidak ditemukan"})
		return
	}
	var body struct {
		Description   *string `json:"description"`
		DiscountType  *string `json:"discount_type"`
		DiscountValue *int64  `json:"discount_value"`
		MinSpend      *int64  `json:"min_spend"`
		ValidUntil    *string `json:"valid_until"`
		UsageLimit    *int64  `json:"usage_limit"`
		IsActive      *int64  `json:"is_active"`
	}
	_ = c.ShouldBindJSON(&body)

	description := toStr(existing["description"])
	if body.Description != nil {
		description = *body.Description
	}
	dtype := toStr(existing["discount_type"])
	if body.DiscountType != nil {
		dtype = strings.ToLower(*body.DiscountType)
	}
	dvalue := toInt(existing["discount_value"])
	if body.DiscountValue != nil {
		dvalue = *body.DiscountValue
	}
	minSpend := toInt(existing["min_spend"])
	if body.MinSpend != nil {
		minSpend = *body.MinSpend
	}
	validUntil := existing["valid_until"]
	if body.ValidUntil != nil {
		validUntil = nullIfEmpty(*body.ValidUntil)
	}
	usageLimit := toInt(existing["usage_limit"])
	if body.UsageLimit != nil {
		usageLimit = *body.UsageLimit
	}
	isActive := toInt(existing["is_active"])
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	if _, err := s.DB.Exec(
		`UPDATE vouchers SET description = ?, discount_type = ?, discount_value = ?, min_spend = ?, valid_until = ?, usage_limit = ?, is_active = ? WHERE id = ?`,
		description, dtype, dvalue, minSpend, validUntil, usageLimit, isActive, c.Param("id"),
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	row, _ := getRow(s.DB, `SELECT * FROM vouchers WHERE id = ?`, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"voucher": row})
}

func (s *Server) adminDeleteVoucher(c *gin.Context) {
	s.DB.Exec(`DELETE FROM vouchers WHERE id = ?`, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
