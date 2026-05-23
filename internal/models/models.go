package models

import (
	"time"
)

// UserRole kullanıcı rollerini tanımlar.
// String tipi olması JSON serializasyonunu kolaylaştırır.
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleAuditor UserRole = "auditor"
)

// User veritabanı kullanıcı modelidir.
// Şifre asla düz metin olarak saklanmaz — bcrypt hash kullanılır.
type User struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	Email          string    `gorm:"uniqueIndex;not null" json:"email"`
	FullName       string    `gorm:"not null" json:"full_name"`
	HashedPassword string    `gorm:"not null" json:"-"` // json:"-" ile API yanıtına dahil edilmez
	Role           UserRole  `gorm:"not null;default:'auditor'" json:"role"`
	IsActive       bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type WashTemperature string

const (
	Wash30   WashTemperature = "30"
	Wash40   WashTemperature = "40"
	Wash60   WashTemperature = "60"
	WashNone WashTemperature = "none"
)

// Product bir tekstil ürününün dijital pasaportunu temsil eder.
type Product struct {
	ID                  string          `gorm:"primaryKey" json:"id"`
	Name                string          `gorm:"not null" json:"name"`
	Brand               string          `gorm:"not null" json:"brand"`
	Category            string          `gorm:"not null" json:"category"`
	CountryOfProduction string          `gorm:"not null" json:"country_of_production"`
	ProductionDate      string          `gorm:"not null" json:"production_date"`
	WashTemperature     WashTemperature `gorm:"not null;default:'30'" json:"wash_temperature"`
	DryCleaning         bool            `gorm:"not null;default:false" json:"dry_cleaning"`
	AdditionalNotes     string          `json:"additional_notes"`
	CreatorID           string          `json:"creator_id"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	// Preload ile ilişkili materyaller yüklenir
	Materials []Material `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"materials"`
}

// Material bir ürünün materyal kompozisyonundaki tek bir kalemi temsil eder.
// Toplam yüzde validasyonu servis katmanında yapılır.
type Material struct {
	ID         uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID  string  `gorm:"not null" json:"product_id"`
	Name       string  `gorm:"not null" json:"name"`
	Percentage float64 `gorm:"not null" json:"percentage"`
	IsRecycled bool    `gorm:"not null;default:false" json:"is_recycled"`
}

// DashboardStats dashboard için istatistik verilerini tutar.
type DashboardStats struct {
	TotalProducts             int                `json:"total_products"`
	TotalCategories           int                `json:"total_categories"`
	TotalUsers                int                `json:"total_users"`
	ProductsByCategory        map[string]int     `json:"products_by_category"`
	ProductsByCountry         map[string]int     `json:"products_by_country"`
	RecycledMaterialPercentage float64           `json:"recycled_material_percentage"`
}
