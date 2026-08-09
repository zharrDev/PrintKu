package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"printmart/backend/internal/middleware"
	"printmart/backend/internal/repository"
)

func (s *Server) register(c *gin.Context) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, dan password wajib diisi"})
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.TrimSpace(body.Email)
	if body.Name == "" || body.Email == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, dan password wajib diisi"})
		return
	}
	if len(body.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password minimal 6 karakter"})
		return
	}
	email := strings.ToLower(body.Email)

	var exists string
	if err := s.DB.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&exists); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	id := newID()
	if _, err := s.DB.Exec(
		`INSERT INTO users (id, name, email, password_hash, phone, role, created_at) VALUES (?, ?, ?, ?, ?, 'user', ?)`,
		id, body.Name, email, string(hash), body.Phone, repository.Now(),
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user, _ := getRow(s.DB, `SELECT id, name, email, phone, role, created_at FROM users WHERE id = ?`, id)
	token, err := middleware.SignToken(s.Cfg, middleware.AuthUser{
		ID: id, Name: body.Name, Email: email, Role: "user",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func (s *Server) login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email dan password wajib diisi"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))

	var user middleware.AuthUser
	var passwordHash string
	var phone string
	var createdAt string
	err := s.DB.QueryRow(`SELECT id, name, email, phone, password_hash, role, created_at FROM users WHERE email = ?`, email).
		Scan(&user.ID, &user.Name, &user.Email, &phone, &passwordHash, &user.Role, &createdAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}
	token, err := middleware.SignToken(s.Cfg, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	safe := gin.H{
		"id": user.ID, "name": user.Name, "email": user.Email, "phone": phone, "role": user.Role, "created_at": createdAt,
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": safe})
}

func (s *Server) me(c *gin.Context) {
	u := s.user(c)
	user, err := getRow(s.DB, `SELECT id, name, email, phone, role, created_at FROM users WHERE id = ?`, u.ID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *Server) listAddresses(c *gin.Context) {
	u := s.user(c)
	rows, err := allRows(s.DB, `SELECT * FROM addresses WHERE user_id = ? ORDER BY created_at DESC`, u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"addresses": rows})
}

func (s *Server) createAddress(c *gin.Context) {
	var body struct {
		Label       string `json:"label"`
		FullAddress string `json:"full_address"`
		City        string `json:"city"`
		PostalCode  string `json:"postal_code"`
	}
	_ = c.ShouldBindJSON(&body)
	if strings.TrimSpace(body.FullAddress) == "" || strings.TrimSpace(body.City) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "full_address dan city wajib diisi"})
		return
	}
	id := newID()
	if _, err := s.DB.Exec(
		`INSERT INTO addresses (id, user_id, label, full_address, city, postal_code, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, s.user(c).ID, body.Label, body.FullAddress, body.City, body.PostalCode, repository.Now(),
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addr, _ := getRow(s.DB, `SELECT * FROM addresses WHERE id = ?`, id)
	c.JSON(http.StatusCreated, gin.H{"address": addr})
}

func (s *Server) deleteAddress(c *gin.Context) {
	s.DB.Exec(`DELETE FROM addresses WHERE id = ? AND user_id = ?`, c.Param("id"), s.user(c).ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}