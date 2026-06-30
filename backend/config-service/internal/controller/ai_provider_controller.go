package controller

import (
	"net/http"
	"strconv"

	"config-service/internal/model"
	"config-service/internal/repository"
	"config-service/internal/utils"
	"github.com/gin-gonic/gin"
)

type AIProviderInput struct {
	Name        string `json:"name" binding:"required"`
	ApiEndpoint string `json:"api_endpoint" binding:"required"`
	ApiKey      string `json:"api_key" binding:"required"`
	ModelName   string `json:"model_name" binding:"required"`
	IsActive    *bool  `json:"is_active" binding:"required"`
}

type AIProviderUpdateInput struct {
	Name        string `json:"name"`
	ApiEndpoint string `json:"api_endpoint"`
	ApiKey      string `json:"api_key"`
	ModelName   string `json:"model_name"`
	IsActive    *bool  `json:"is_active"`
}

// GetAIProviders retrieves all configured AI endpoints
func GetAIProviders(c *gin.Context) {
	var providers []model.AIProvider
	if err := repository.DB.Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch AI providers"})
		return
	}

	// Remove encrypted API keys from response for security
	for i := range providers {
		providers[i].ApiKeyEncrypted = ""
	}

	c.JSON(http.StatusOK, providers)
}

// GetAIProvider retrieves a specific AI provider config by ID
func GetAIProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider model.AIProvider
	if err := repository.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI provider not found"})
		return
	}

	// Remove encrypted API key from response
	provider.ApiKeyEncrypted = ""

	c.JSON(http.StatusOK, provider)
}

// CreateAIProvider registers a new AI endpoint config
func CreateAIProvider(c *gin.Context) {
	var input AIProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedApiKey, err := utils.Encrypt(input.ApiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure API key"})
		return
	}

	provider := model.AIProvider{
		Name:            input.Name,
		ApiEndpoint:     input.ApiEndpoint,
		ApiKeyEncrypted: encryptedApiKey,
		ModelName:       input.ModelName,
		IsActive:        *input.IsActive,
	}

	if err := repository.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create AI provider"})
		return
	}

	provider.ApiKeyEncrypted = ""
	c.JSON(http.StatusCreated, provider)
}

// UpdateAIProvider updates an AI provider configuration
func UpdateAIProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider model.AIProvider
	if err := repository.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI provider not found"})
		return
	}

	var input AIProviderUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		provider.Name = input.Name
	}
	if input.ApiEndpoint != "" {
		provider.ApiEndpoint = input.ApiEndpoint
	}
	if input.ModelName != "" {
		provider.ModelName = input.ModelName
	}
	if input.IsActive != nil {
		provider.IsActive = *input.IsActive
	}
	if input.ApiKey != "" {
		encryptedApiKey, err := utils.Encrypt(input.ApiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure API key"})
			return
		}
		provider.ApiKeyEncrypted = encryptedApiKey
	}

	if err := repository.DB.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update AI provider"})
		return
	}

	provider.ApiKeyEncrypted = ""
	c.JSON(http.StatusOK, provider)
}

// DeleteAIProvider deletes an AI provider configuration
func DeleteAIProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider model.AIProvider
	if err := repository.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AI provider not found"})
		return
	}

	if err := repository.DB.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete AI provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI provider deleted successfully"})
}
