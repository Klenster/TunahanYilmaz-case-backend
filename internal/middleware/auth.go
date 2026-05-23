package middleware

import (
	"net/http"
	"strings"

	"dpp-backend/internal/models"
	"dpp-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware her korumalı rotada çalışır.
// Authorization header'dan Bearer token alır, doğrular ve context'e kullanıcı bilgisi ekler.
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Authorization header required"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateToken(token, secretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid or expired token"})
			return
		}

		
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// AdminMiddleware yalnızca admin rolündeki kullanıcılara izin verir.
// Rol kontrolü BACKEND'de yapılır — frontend'de gizlenmek yerine erişim kesilir
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		if role != string(models.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "Admin access required"})
			return
		}
		c.Next()
	}
}
