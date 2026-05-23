package main

import (
	"log"

	"dpp-backend/internal/config"
	"dpp-backend/internal/db"
	"dpp-backend/internal/handlers"
	"dpp-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ayarları yükle
	cfg := config.Load()

	// Veritabanı bağlantısı ve migration'ı
	database, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Varsayılan admin oluşturulur (sadece ilk çalıştırmada)
	if err := db.SeedAdmin(database, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Printf("Warning: failed to seed admin: %v", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Handler'lar başlatılır
	authHandler := handlers.NewAuthHandler(database, cfg)
	productHandler := handlers.NewProductHandler(database)
	userHandler := handlers.NewUserHandler(database)

	// Public rotalar
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.SecretKey), authHandler.Me)
		}
	}

	// Korumalı rotalar — JWT gerektirir
	protected := api.Group("/", middleware.AuthMiddleware(cfg.SecretKey))
	{
		products := protected.Group("/products")
		{
			products.GET("/", productHandler.List)
			products.GET("/stats", productHandler.Stats)
			products.GET("/categories", productHandler.Categories)
			products.GET("/:id", productHandler.Get)

			// Sadece admin erişebilir
			adminProducts := products.Group("/", middleware.AdminMiddleware())
			{
				adminProducts.POST("/", productHandler.Create)
				adminProducts.PUT("/:id", productHandler.Update)
				adminProducts.DELETE("/:id", productHandler.Delete)
			}
		}

		users := protected.Group("/users")
		{
			users.PATCH("/me/profile", userHandler.UpdateProfile)
			users.PATCH("/me/password", userHandler.ChangePassword)

			// Sadece admin erişebilir
			adminUsers := users.Group("/", middleware.AdminMiddleware())
			{
				adminUsers.GET("/", userHandler.List)
				adminUsers.PATCH("/:id/role", userHandler.UpdateRole)
				adminUsers.DELETE("/:id", userHandler.Delete)
			}
		}
	}

	log.Printf("Server running on :%s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}