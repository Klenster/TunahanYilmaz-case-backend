package db

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"dpp-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"
)

// Connect veritabanı bağlantısını açar ve tabloları oluşturur.
func Connect(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Model değişikliklerini otomatik olarak veritabanına yansıtır
	if err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.Material{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// SeedAdmin ilk çalıştırmada varsayılan admin kullanıcısı oluşturur.
func SeedAdmin(db *gorm.DB, email, password string) error {
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)
	if count > 0 {
		return nil
	}

	if password == "" {
		var err error
		password, err = generateSecurePassword(20)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin := models.User{
		ID:             uuid.New().String(),
		Email:          email,
		FullName:       "Default Admin",
		HashedPassword: string(hash),
		Role:           models.RoleAdmin,
		IsActive:       true,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	fmt.Println("\n" + "============================================================")
	fmt.Println("DEFAULT ADMIN CREDENTIALS (save these now!)")
	fmt.Printf("  Email:    %s\n", email)
	fmt.Printf("  Password: %s\n", password)
	fmt.Println("These will NOT be shown again.")
	fmt.Println("============================================================\n")

	return nil
}

// generateSecurePassword crypto/rand kullanarak karışık bir şifre üretir.
func generateSecurePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[n.Int64()]
	}
	password[0] = 'A'
	password[1] = '1'
	password[2] = '!'
	log.Println("Secure admin password generated")
	return string(password), nil
}