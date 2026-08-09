package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"printmart/backend/internal/middleware"
	"printmart/backend/internal/repository"
)

func (s *Server) listProducts(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")
	sort := c.Query("sort")

	var where []string
	var params []any
	isAdmin := false
	if u, ok := c.Get("user"); ok {
		if au, ok := u.(middleware.AuthUser); ok && au.Role == "admin" {
			isAdmin = true
		}
	}
	if isAdmin {
		where = append(where, "(is_active = 1 OR is_active = 0)")
	} else {
		where = append(where, "is_active = 1")
	}
	if category != "" {
		where = append(where, "category_id = ?")
		params = append(params, category)
	}
	if search != "" {
		where = append(where, "(p.name LIKE ? OR p.description LIKE ?)")
		params = append(params, "%"+search+"%", "%"+search+"%")
	}
	orderBy := "name ASC"
	switch sort {
	case "price_asc":
		orderBy = "price ASC"
	case "price_desc":
		orderBy = "price DESC"
	case "newest":
		orderBy = "created_at DESC"
	}

	rows, err := allRows(s.DB,
		fmt.Sprintf(`SELECT p.*, c.name AS category_name FROM products p
			LEFT JOIN categories c ON c.id = p.category_id
			WHERE %s ORDER BY %s`, strings.Join(where, " AND "), orderBy),
		params...)
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": rows, "count": len(rows)})
}

func (s *Server) getProduct(c *gin.Context) {
	row, err := getRow(s.DB, `SELECT p.*, c.name AS category_name FROM products p
		LEFT JOIN categories c ON c.id = p.category_id WHERE p.id = ?`, c.Param("id"))
	if err != nil {
		s.jsonErr(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": row})
}

func (s *Server) createProduct(c *gin.Context) {
	var body struct {
		CategoryID string `json:"category_id"`
		Name       string `json:"name"`
		Description string `json:"description"`
		Price      any     `json:"price"`
		Stock      any     `json:"stock"`
		ImageURL   string `json:"image_url"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.CategoryID == "" || strings.TrimSpace(body.Name) == "" || body.Price == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_id, name, dan price wajib diisi"})
		return
	}
	id := newID()
	price := num(body.Price)
	stock := num(body.Stock)
	if _, err := s.DB.Exec(
		`INSERT INTO products (id, category_id, name, description, price, stock, image_url, is_active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?)`,
		id, body.CategoryID, body.Name, body.Description, price, stock, body.ImageURL, repository.Now(),
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	row, _ := getRow(s.DB, `SELECT * FROM products WHERE id = ?`, id)
	c.JSON(http.StatusCreated, gin.H{"product": row})
}

func (s *Server) updateProduct(c *gin.Context) {
	existing, err := getRow(s.DB, `SELECT * FROM products WHERE id = ?`, c.Param("id"))
	if err != nil || existing == nil {
		if err != nil {
			s.jsonErr(c, err)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	var body struct {
		CategoryID  *string `json:"category_id"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Price       *float64 `json:"price"`
		Stock       *float64 `json:"stock"`
		ImageURL    *string `json:"image_url"`
		IsActive    *float64 `json:"is_active"`
	}
	_ = c.ShouldBindJSON(&body)

	categoryID := body.CategoryID
	if categoryID == nil {
		v := toStr(existing["category_id"])
		categoryID = &v
	}
	name := body.Name
	if name == nil {
		v := toStr(existing["name"])
		name = &v
	}
	description := body.Description
	if description == nil {
		v := toStr(existing["description"])
		description = &v
	}
	price := body.Price
	if price == nil {
		v := float64(toInt(existing["price"]))
		price = &v
	}
	stock := body.Stock
	if stock == nil {
		v := float64(toInt(existing["stock"]))
		stock = &v
	}
	imageURL := body.ImageURL
	if imageURL == nil {
		v := toStr(existing["image_url"])
		imageURL = &v
	}
	isActive := body.IsActive
	if isActive == nil {
		v := float64(toInt(existing["is_active"]))
		isActive = &v
	}

	if _, err := s.DB.Exec(
		`UPDATE products SET category_id = ?, name = ?, description = ?, price = ?, stock = ?, image_url = ?, is_active = ? WHERE id = ?`,
		*categoryID, *name, *description, int64(*price), int64(*stock), *imageURL, int64(*isActive), c.Param("id"),
	); err != nil {
		s.jsonErr(c, err)
		return
	}
	row, _ := getRow(s.DB, `SELECT * FROM products WHERE id = ?`, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"product": row})
}

func (s *Server) deleteProduct(c *gin.Context) {
	s.DB.Exec(`DELETE FROM products WHERE id = ?`, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) decrementStock(tx *sql.Tx, productID string, qty int64) error {
	_, err := tx.Exec(`UPDATE products SET stock = stock - ? WHERE id = ?`, qty, productID)
	return err
}