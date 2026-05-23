package handlers

import (
	"net/http"

	"dpp-backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type RoleUpdateRequest struct {
	Role models.UserRole `json:"role" binding:"required"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	h.db.Order("created_at DESC").Find(&users)
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = userToResponse(u)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	currentUserID, _ := c.Get("userID")
	targetID := c.Param("id")

	// Admin kendi rolünü değiştiremez
	if targetID == currentUserID.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot change your own role"})
		return
	}

	var req RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
		return
	}

	user.Role = req.Role
	h.db.Save(&user)
	c.JSON(http.StatusOK, userToResponse(user))
}

func (h *UserHandler) Delete(c *gin.Context) {
	currentUserID, _ := c.Get("userID")
	targetID := c.Param("id")

	if targetID == currentUserID.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot delete your own account"})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "User not found"})
		return
	}

	h.db.Delete(&user)
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	var user models.User
	h.db.First(&user, "id = ?", userID)

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" && req.Email != user.Email {
		var existing models.User
		if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Email already in use"})
			return
		}
		user.Email = req.Email
	}

	h.db.Save(&user)
	c.JSON(http.StatusOK, userToResponse(user))
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	var user models.User
	h.db.First(&user, "id = ?", userID)

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Current password is incorrect"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.HashedPassword = string(hash)
	h.db.Save(&user)
	c.Status(http.StatusNoContent)
}
