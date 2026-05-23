package handlers

import (
	"net/http"
	"time"

	"dpp-backend/internal/config"
	"dpp-backend/internal/models"
	"dpp-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        string           `json:"id"`
	Email     string           `json:"email"`
	FullName  string           `json:"full_name"`
	Role      models.UserRole  `json:"role"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
}

func userToResponse(u models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	var existing models.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to process password"})
		return
	}

	// Register'dan oluşturulan kullanıcılar varsayılan olarak auditor rolüyle açılır
	user := models.User{
		ID:             uuid.New().String(),
		Email:          req.Email,
		FullName:       req.FullName,
		HashedPassword: string(hash),
		Role:           models.RoleAuditor,
		IsActive:       true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, userToResponse(user))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		// Güvenlik: kullanıcı bulunamadı ile yanlış şifre aynı mesajı döner (enumeration önleme)
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user.ID, string(user.Role), h.cfg.SecretKey, h.cfg.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "bearer",
		"user":         userToResponse(user),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "User not found"})
		return
	}
	c.JSON(http.StatusOK, userToResponse(user))
}
