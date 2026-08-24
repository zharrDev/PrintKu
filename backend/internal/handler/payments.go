package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/repository"
	"printmart/backend/internal/service"
)

func genVaNumber() string {
	return fmt.Sprintf("4811%d", 10000000+rand.Intn(89999999))
}

func transactionID() string {
	return strings.ToUpper(fmt.Sprintf("SANDBOX-%d-%s", repository.NowUnixMilli(), newID()[:6]))
}

func (s *Server) createPayment(c *gin.Context) {
	var body struct {
		OrderID string `json:"orderId"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId wajib diisi"})
		return
	}
	order, err := getRow(s.DB, `SELECT * FROM orders WHERE id = ?`, body.OrderID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order tidak ditemukan"})
		return
	}
	u := s.user(c)
	if toStr(order["user_id"]) != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	existing, _ := getRow(s.DB, `SELECT * FROM payments WHERE order_id = ?`, body.OrderID)
	if existing != nil {
		if toStr(existing["va_number"]) == "" {
			if _, err := s.DB.Exec(
				`UPDATE payments SET va_number = ?, transaction_id = ? WHERE id = ?`,
				genVaNumber(), transactionID(), existing["id"],
			); err != nil {
				s.jsonErr(c, err)
				return
			}
		}
		payment, _ := getRow(s.DB, `SELECT * FROM payments WHERE order_id = ?`, body.OrderID)
		c.JSON(http.StatusOK, gin.H{"payment": payment})
		return
	}

	id := newID()
	payment := gin.H{
		"id": id, "order_id": body.OrderID, "provider": s.Cfg.PaymentProvider,
		"transaction_id": transactionID(), "status": "pending",
		"amount": order["total_price"], "va_number": genVaNumber(), "created_at": repository.Now(),
	}
	if _, err := s.DB.Exec(
		`INSERT INTO payments (id, order_id, provider, transaction_id, status, amount, va_number, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, body.OrderID, s.Cfg.PaymentProvider, payment["transaction_id"], "pending",
		order["total_price"], payment["va_number"], repository.Now(),
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"payment": payment})
}

func (s *Server) paymentWebhook(c *gin.Context) {
	var body struct {
		OrderID string `json:"orderId"`
		Status  string `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.OrderID == "" || body.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId dan status wajib diisi"})
		return
	}
	payment, err := getRow(s.DB, `SELECT * FROM payments WHERE order_id = ?`, body.OrderID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if payment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment tidak ditemukan"})
		return
	}
	if toStr(payment["status"]) == "success" {
		c.JSON(http.StatusOK, gin.H{"payment": payment, "message": "Transaksi sudah sukses sebelumnya"})
		return
	}

	// Validasi status enum
	validStatuses := map[string]bool{"success": true, "failed": true, "pending": true}
	newStatus := "pending"
	if validStatuses[body.Status] {
		newStatus = body.Status
	}
	var paidAt any
	if newStatus == "success" {
		paidAt = repository.Now()
	}
	if _, err := s.DB.Exec(`UPDATE payments SET status = ?, paid_at = ? WHERE id = ?`, newStatus, paidAt, payment["id"]); err != nil {
		s.jsonErr(c, err)
		return
	}

	var updatedOrder gin.H
	if newStatus == "success" {
		// Increment voucher used_count hanya saat payment sukses
		if vcode := orderVoucherCode(s.DB, body.OrderID); vcode != "" {
			s.DB.Exec(`UPDATE vouchers SET used_count = used_count + 1 WHERE UPPER(code) = ?`, vcode)
		}

		order := service.SetOrderStatus(s.DB, body.OrderID, "diproses", s.broadcast)
		if order != nil {
			repository.Notify(s.DB, order.UserID, body.OrderID,
				"Pembayaran diterima ("+s.Cfg.PaymentProvider+"). Pesanan sedang diproses.")
			s.broadcast(order.UserID, gin.H{"type": "payment", "orderId": body.OrderID, "status": "success"})
			updatedOrder = gin.H{
				"id": order.ID, "user_id": order.UserID, "order_type": order.OrderType,
				"status": order.Status, "total_price": order.TotalPrice,
				"delivery_method": order.DeliveryMethod, "delivery_address": order.DeliveryAddress,
				"notes": order.Notes, "customer_name": order.CustomerName, "created_at": order.CreatedAt,
			}
		}
	} else if newStatus == "failed" {
		service.SetOrderStatus(s.DB, body.OrderID, "dibatalkan", s.broadcast)
	}
	payment, _ = getRow(s.DB, `SELECT * FROM payments WHERE order_id = ?`, body.OrderID)
	c.JSON(http.StatusOK, gin.H{"payment": payment, "order": updatedOrder})
}

func (s *Server) getPayment(c *gin.Context) {
	order, err := getRow(s.DB, `SELECT * FROM orders WHERE id = ?`, c.Param("orderId"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order tidak ditemukan"})
		return
	}
	u := s.user(c)
	if toStr(order["user_id"]) != u.ID && u.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	payment, _ := getRow(s.DB, `SELECT * FROM payments WHERE order_id = ?`, c.Param("orderId"))
	if payment == nil {
		c.JSON(http.StatusOK, gin.H{"payment": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment": payment})
}