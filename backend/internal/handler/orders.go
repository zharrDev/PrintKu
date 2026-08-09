package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/repository"
	"printmart/backend/internal/service"
)

func (s *Server) enrichOrder(c *gin.Context, id string) (gin.H, error) {
	order, err := s.q1(c, `SELECT * FROM orders WHERE id = ?`, id)
	if err != nil || order == nil {
		return nil, err
	}
	items, err := s.q(c, `SELECT * FROM order_items WHERE order_id = ?`, id)
	if err != nil {
		return nil, err
	}
	jobs, err := s.q(c, `SELECT pj.*, f.name AS finishing_name, f.price AS finishing_price
		FROM print_jobs pj LEFT JOIN finishing_options f ON f.id = pj.finishing_option_id
		WHERE pj.order_id = ?`, id)
	if err != nil {
		return nil, err
	}
	payment, _ := s.q1(c, `SELECT * FROM payments WHERE order_id = ?`, id)
	data := gin.H{
		"id": order["id"], "user_id": order["user_id"], "order_type": order["order_type"],
		"status": order["status"], "total_price": order["total_price"],
		"subtotal": order["subtotal"], "discount": order["discount"], "voucher_code": order["voucher_code"],
		"delivery_method": order["delivery_method"], "delivery_address": order["delivery_address"],
		"notes": order["notes"], "customer_name": order["customer_name"], "created_at": order["created_at"],
		"items": items, "printJobs": jobs, "payment": nil,
	}
	if payment != nil {
		data["payment"] = payment
	}
	return data, nil
}

func (s *Server) createOrder(c *gin.Context) {
	var body struct {
		OrderType       string   `json:"orderType"`
		DeliveryMethod  string   `json:"deliveryMethod"`
		DeliveryAddress string   `json:"delivery_address"`
		Notes           string   `json:"notes"`
		CustomerName    string   `json:"customer_name"`
		PrintJobIDs     []string `json:"printJobIds"`
		VoucherCode     string   `json:"voucher_code"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.OrderType == "" {
		body.OrderType = "product"
	}
	if body.DeliveryMethod == "" {
		body.DeliveryMethod = "ambil"
	}
	u := s.user(c)

	tx, err := s.DB.Begin()
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	defer tx.Rollback()
	s.setTx(c, tx)

	id := newID()
	var total int64

	if body.OrderType == "print" {
		if len(body.PrintJobIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "printJobIds wajib diisi"})
			return
		}
		for _, jobID := range body.PrintJobIDs {
			job, err := s.q1(c, `SELECT * FROM print_jobs WHERE id = ?`, jobID)
			if err != nil {
				s.jsonErr(c, err)
				return
			}
			if job == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Print job " + jobID + " tidak ditemukan"})
				return
			}
			if v := toStr(job["order_id"]); v != "" {
				c.JSON(http.StatusConflict, gin.H{"error": "Print job " + jobID + " sudah masuk order lain"})
				return
			}
			total += toInt(job["estimated_price"])
		}
	} else {
		cartItems, err := s.q(c, `SELECT * FROM cart_items WHERE user_id = ?`, u.ID)
		if err != nil {
			s.jsonErr(c, err)
			return
		}
		if len(cartItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Keranjang kosong"})
			return
		}
		for _, ci := range cartItems {
			product, _ := s.q1(c, `SELECT * FROM products WHERE id = ?`, ci["product_id"])
			if product == nil || toInt(product["is_active"]) != 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Produk " + toStr(ci["product_id"]) + " tidak tersedia"})
				return
			}
			if toInt(product["stock"]) < toInt(ci["quantity"]) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Stok \"" + toStr(product["name"]) + "\" tidak mencukupi (sisa " + toStr(product["stock"]) + ")"})
				return
			}
			total += toInt(product["price"]) * toInt(ci["quantity"])
		}
	}

	customerName := body.CustomerName
	if customerName == "" {
		customerName = u.Name
	}

	// Voucher / diskon (opsional). Validasi dilakukan server-side terhadap
	// subtotal yang baru dihitung; nominal diskon tidak dipercayakan ke client.
	subtotal := total
	var discount int64
	var voucherCode, voucherID string
	if code := strings.ToUpper(strings.TrimSpace(body.VoucherCode)); code != "" {
		v, err := s.q1(c, `SELECT * FROM vouchers WHERE UPPER(code) = ?`, code)
		if err != nil {
			s.jsonErr(c, err)
			return
		}
		if v == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Voucher \"" + code + "\" tidak ditemukan"})
			return
		}
		d, msg := service.CalcDiscount(voucherFromRow(v), subtotal, repository.Now())
		if msg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		discount = d
		voucherCode = toStr(v["code"])
		voucherID = toStr(v["id"])
	}
	total = subtotal - discount

	if _, err := tx.Exec(
		`INSERT INTO orders (id, user_id, order_type, status, total_price, delivery_method, delivery_address, notes, customer_name, subtotal, discount, voucher_code, created_at)
		 VALUES (?, ?, ?, 'menunggu_pembayaran', ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, u.ID, body.OrderType, total, body.DeliveryMethod, body.DeliveryAddress, body.Notes, customerName, subtotal, discount, voucherCode, repository.Now(),
	); err != nil {
		s.jsonErr(c, err)
		return
	}

	if voucherID != "" {
		if _, err := tx.Exec(`UPDATE vouchers SET used_count = used_count + 1 WHERE id = ?`, voucherID); err != nil {
			s.jsonErr(c, err)
			return
		}
	}

	if body.OrderType == "print" {
		for _, jobID := range body.PrintJobIDs {
			if _, err := tx.Exec(`UPDATE print_jobs SET order_id = ? WHERE id = ?`, id, jobID); err != nil {
				s.jsonErr(c, err)
				return
			}
		}
	} else {
		cartItems, _ := s.q(c, `SELECT * FROM cart_items WHERE user_id = ?`, u.ID)
		for _, ci := range cartItems {
			product, _ := s.q1(c, `SELECT * FROM products WHERE id = ?`, ci["product_id"])
			qty := toInt(ci["quantity"])
			subtotal := toInt(product["price"]) * qty
			if _, err := tx.Exec(`UPDATE products SET stock = stock - ? WHERE id = ?`, qty, ci["product_id"]); err != nil {
				s.jsonErr(c, err)
				return
			}
			if _, err := tx.Exec(
				`INSERT INTO order_items (id, order_id, product_id, product_name, quantity, unit_price, subtotal) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				newID(), id, ci["product_id"], product["name"], qty, product["price"], subtotal,
			); err != nil {
				s.jsonErr(c, err)
				return
			}
		}
		if _, err := tx.Exec(`DELETE FROM cart_items WHERE user_id = ?`, u.ID); err != nil {
			s.jsonErr(c, err)
			return
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO payments (id, order_id, provider, status, amount, created_at) VALUES (?, ?, 'midtrans-sandbox', 'pending', ?, ?)`,
		newID(), id, total, repository.Now(),
	); err != nil {
		s.jsonErr(c, err)
		return
	}

	if err := tx.Commit(); err != nil {
		s.jsonErr(c, err)
		return
	}
	s.setTx(c, nil)

	msg := "Pesanan #" + truncateOrderId(id) + " dibuat. Total " + service.FormatRupiah(total)
	if discount > 0 {
		msg += " (hemat " + service.FormatRupiah(discount) + " dengan voucher " + voucherCode + ")"
	}
	repository.Notify(s.DB, u.ID, id, msg)

	row, _ := s.enrichOrder(c, id)
	c.JSON(http.StatusCreated, gin.H{"order": row})
}

// voucherFromRow mengubah baris DB voucher menjadi struct service.Voucher.
func voucherFromRow(v map[string]any) service.Voucher {
	return service.Voucher{
		Code:          toStr(v["code"]),
		DiscountType:  toStr(v["discount_type"]),
		DiscountValue: toInt(v["discount_value"]),
		MinSpend:      toInt(v["min_spend"]),
		ValidUntil:    toStr(v["valid_until"]),
		UsageLimit:    toInt(v["usage_limit"]),
		UsedCount:     toInt(v["used_count"]),
		IsActive:      toInt(v["is_active"]) == 1,
	}
}

func (s *Server) listOrders(c *gin.Context) {
	u := s.user(c)
	rows, err := allRows(s.DB, `SELECT * FROM orders WHERE user_id = ? ORDER BY created_at DESC`, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		e, err := s.enrichOrder(c, toStr(r["id"]))
		if err != nil {
			continue
		}
		out = append(out, e)
	}
	c.JSON(http.StatusOK, gin.H{"orders": out})
}

func (s *Server) getOrder(c *gin.Context) {
	u := s.user(c)
	order, err := s.q1(c, `SELECT * FROM orders WHERE id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order tidak ditemukan"})
		return
	}
	if toStr(order["user_id"]) != u.ID && u.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	e, err := s.enrichOrder(c, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": e})
}

func (s *Server) updateOrderStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)
	allowed := []string{"menunggu_pembayaran", "diproses", "siap", "selesai", "dibatalkan"}
	valid := false
	for _, s := range allowed {
		if s == body.Status {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status harus salah satu dari: " + strings.Join(allowed, ", ")})
		return
	}
	order, err := s.q1(c, `SELECT * FROM orders WHERE id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order tidak ditemukan"})
		return
	}
	updated := service.SetOrderStatus(s.DB, c.Param("id"), body.Status, s.broadcast)
	if updated == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order tidak ditemukan"})
		return
	}
	e, _ := s.enrichOrder(c, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"order": e})
}