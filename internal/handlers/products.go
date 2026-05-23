package handlers

import (
	"math"
	"net/http"

	"dpp-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

type MaterialInput struct {
	Name       string  `json:"name" binding:"required"`
	Percentage float64 `json:"percentage" binding:"required,min=0,max=100"`
	IsRecycled bool    `json:"is_recycled"`
}

type ProductInput struct {
	Name                string                 `json:"name" binding:"required"`
	Brand               string                 `json:"brand" binding:"required"`
	Category            string                 `json:"category" binding:"required"`
	CountryOfProduction string                 `json:"country_of_production" binding:"required"`
	ProductionDate      string                 `json:"production_date" binding:"required"`
	WashTemperature     models.WashTemperature `json:"wash_temperature"`
	DryCleaning         bool                   `json:"dry_cleaning"`
	AdditionalNotes     string                 `json:"additional_notes"`
	Materials           []MaterialInput        `json:"materials"`
}

func validateMaterials(materials []MaterialInput) bool {
	if len(materials) == 0 {
		return true
	}
	total := 0.0
	for _, m := range materials {
		total += m.Percentage
	}
	// Ondalık hassasiyeti için 0.01 tolerans
	return math.Abs(total-100.0) < 0.01
}

func (h *ProductHandler) List(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")

	var products []models.Product
	q := h.db.Preload("Materials")

	if search != "" {
		q = q.Where("name LIKE ? OR brand LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}

	q.Order("created_at DESC").Find(&products)
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) Get(c *gin.Context) {
	var product models.Product
	if err := h.db.Preload("Materials").First(&product, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var input ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	if !validateMaterials(input.Materials) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Material percentages must sum to 100%"})
		return
	}

	creatorID, _ := c.Get("userID")
	product := models.Product{
		ID:                  uuid.New().String(),
		Name:                input.Name,
		Brand:               input.Brand,
		Category:            input.Category,
		CountryOfProduction: input.CountryOfProduction,
		ProductionDate:      input.ProductionDate,
		WashTemperature:     input.WashTemperature,
		DryCleaning:         input.DryCleaning,
		AdditionalNotes:     input.AdditionalNotes,
		CreatorID:           creatorID.(string),
	}

	if err := h.db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create product"})
		return
	}

	for _, m := range input.Materials {
		h.db.Create(&models.Material{
			ProductID:  product.ID,
			Name:       m.Name,
			Percentage: m.Percentage,
			IsRecycled: m.IsRecycled,
		})
	}

	h.db.Preload("Materials").First(&product, "id = ?", product.ID)
	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var product models.Product
	if err := h.db.First(&product, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Product not found"})
		return
	}

	var input ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	if !validateMaterials(input.Materials) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Material percentages must sum to 100%"})
		return
	}

	product.Name = input.Name
	product.Brand = input.Brand
	product.Category = input.Category
	product.CountryOfProduction = input.CountryOfProduction
	product.ProductionDate = input.ProductionDate
	product.WashTemperature = input.WashTemperature
	product.DryCleaning = input.DryCleaning
	product.AdditionalNotes = input.AdditionalNotes

	h.db.Save(&product)

	// Materyalleri silme ve yeniden ekleme
	h.db.Where("product_id = ?", product.ID).Delete(&models.Material{})
	for _, m := range input.Materials {
		h.db.Create(&models.Material{
			ProductID:  product.ID,
			Name:       m.Name,
			Percentage: m.Percentage,
			IsRecycled: m.IsRecycled,
		})
	}

	h.db.Preload("Materials").First(&product, "id = ?", product.ID)
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	var product models.Product
	if err := h.db.First(&product, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Product not found"})
		return
	}
	h.db.Delete(&product)
	c.Status(http.StatusNoContent)
}

func (h *ProductHandler) Categories(c *gin.Context) {
	var categories []string
	h.db.Model(&models.Product{}).Distinct("category").Pluck("category", &categories)
	c.JSON(http.StatusOK, categories)
}

func (h *ProductHandler) Stats(c *gin.Context) {
	var products []models.Product
	h.db.Preload("Materials").Find(&products)

	var users int64
	h.db.Model(&models.User{}).Count(&users)

	byCategory := map[string]int{}
	byCountry := map[string]int{}
	categories := map[string]bool{}

	for _, p := range products {
		byCategory[p.Category]++
		byCountry[p.CountryOfProduction]++
		categories[p.Category] = true
	}

	var allMaterials []models.Material
	h.db.Find(&allMaterials)
	recycled := 0
	for _, m := range allMaterials {
		if m.IsRecycled {
			recycled++
		}
	}
	recycledPct := 0.0
	if len(allMaterials) > 0 {
		recycledPct = math.Round(float64(recycled)/float64(len(allMaterials))*1000) / 10
	}

	c.JSON(http.StatusOK, models.DashboardStats{
		TotalProducts:              len(products),
		TotalCategories:            len(categories),
		TotalUsers:                 int(users),
		ProductsByCategory:         byCategory,
		ProductsByCountry:          byCountry,
		RecycledMaterialPercentage: recycledPct,
	})
}
