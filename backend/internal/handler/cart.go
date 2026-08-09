package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) runTx(c *gin.Context, fn func() (gin.H, int)) {
	tx, err := s.DB.Begin()
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	// Simpan pointer tx di context api untuk dipakai query dalam transaksi.
	s.setTx(c, tx)
	data, status := fn()
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		s.jsonErr(c, err)
		return
	}
	s.setTx(c, nil)
	if status >= 400 {
		tx.Rollback()
	}
	c.JSON(status, data)
}

// enrichCart meniru fungsi Node dengan output: { items, total }
func (s *Server) enrichCart(c *gin.Context, userID string) (gin.H, error) {
	rows, err := s.q(c, `SELECT ci.id AS cart_item_id, ci.quantity, p.*
		FROM cart_items ci JOIN products p ON p.id = ci.product_id
		WHERE ci.user_id = ? ORDER BY ci.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	var total int64
	items := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		sub := num(r["price"]) * num(r["quantity"])
		total += sub
		items = append(items, r)
	}
	return gin.H{"items": items, "total": total}, nil
}

func (s *Server) getCart(c *gin.Context) {
	u := s.user(c)
	data, err := s.enrichCart(c, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (s *Server) addToCart(c *gin.Context) {
	var body struct {
		ProductID string `json:"product_id"`
		Quantity  any     `json:"quantity"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.ProductID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id wajib diisi"})
		return
	}
	row, err := s.q1(c, `SELECT * FROM products WHERE id = ?`, body.ProductID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	qty := num(body.Quantity)
	if qty < 1 {
		qty = 1
	}
	u := s.user(c)
	existing, _ := s.q1(c, `SELECT * FROM cart_items WHERE user_id = ? AND product_id = ?`, u.ID, body.ProductID)
	if existing != nil {
		if _, err := s.DB.Exec(`UPDATE cart_items SET quantity = ? WHERE id = ?`, num(existing["quantity"])+qty, existing["id"]); err != nil {
			s.jsonErr(c, err)
			return
		}
	} else {
		if _, err := s.DB.Exec(`INSERT INTO cart_items (id, user_id, product_id, quantity, created_at) VALUES (?, ?, ?, ?, ?)`,
			newID(), u.ID, body.ProductID, qty, s.now()); err != nil {
			s.jsonErr(c, err)
			return
		}
	}
	data, err := s.enrichCart(c, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, data)
}

func (s *Server) patchCartItem(c *gin.Context) {
	var body struct {
		Quantity any `json:"quantity"`
	}
	_ = c.ShouldBindJSON(&body)
	u := s.user(c)
	row, err := s.q1(c, `SELECT * FROM cart_items WHERE id = ? AND user_id = ?`, c.Param("cartItemId"), u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item cart tidak ditemukan"})
		return
	}
	qty := num(body.Quantity)
	if qty < 1 {
		qty = 1
	}
	if _, err := s.DB.Exec(`UPDATE cart_items SET quantity = ? WHERE id = ?`, qty, row["id"]); err != nil {
		s.jsonErr(c, err)
		return
	}
	data, err := s.enrichCart(c, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (s *Server) deleteCartItem(c *gin.Context) {
	u := s.user(c)
	s.DB.Exec(`DELETE FROM cart_items WHERE id = ? AND user_id = ?`, c.Param("cartItemId"), u.ID)
	data, err := s.enrichCart(c, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (s *Server) clearCart(c *gin.Context) {
	u := s.user(c)
	s.DB.Exec(`DELETE FROM cart_items WHERE user_id = ?`, u.ID)
	data, err := s.enrichCart(c, u.ID)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}