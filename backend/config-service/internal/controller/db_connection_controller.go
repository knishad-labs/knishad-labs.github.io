package controller

import (
	"net/http"
	"strconv"
	"time"

	"config-service/internal/model"
	"config-service/internal/repository"
	"config-service/internal/utils"
	"github.com/gin-gonic/gin"
)

type DatabaseConnectionInput struct {
	Name         string `json:"name" binding:"required"`
	DbType       string `json:"db_type" binding:"required"`
	Host         string `json:"host" binding:"required"`
	Port         int    `json:"port" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	DatabaseName string `json:"database_name" binding:"required"`
}

type DatabaseConnectionUpdateInput struct {
	Name         string `json:"name"`
	DbType       string `json:"db_type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"database_name"`
}

// GetDatabaseConnections retrieves all configured target database connections
func GetDatabaseConnections(c *gin.Context) {
	var connections []model.DatabaseConnection
	if err := repository.DB.Find(&connections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch connections"})
		return
	}

	// Remove encrypted passwords from response for security
	for i := range connections {
		connections[i].PasswordEncrypted = ""
	}

	c.JSON(http.StatusOK, connections)
}

// GetDatabaseConnection retrieves a specific target database connection detail by ID
func GetDatabaseConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var connection model.DatabaseConnection
	if err := repository.DB.First(&connection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	// Remove encrypted password from response
	connection.PasswordEncrypted = ""

	c.JSON(http.StatusOK, connection)
}

// CreateDatabaseConnection registers a new target database connection config
func CreateDatabaseConnection(c *gin.Context) {
	var input DatabaseConnectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedPassword, err := utils.Encrypt(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
		return
	}

	connection := model.DatabaseConnection{
		Name:              input.Name,
		DbType:            input.DbType,
		Host:              input.Host,
		Port:              input.Port,
		Username:          input.Username,
		PasswordEncrypted: encryptedPassword,
		DatabaseName:      input.DatabaseName,
	}

	if err := repository.DB.Create(&connection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database connection"})
		return
	}

	connection.PasswordEncrypted = ""
	c.JSON(http.StatusCreated, connection)
}

// UpdateDatabaseConnection updates a database connection configuration
func UpdateDatabaseConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var connection model.DatabaseConnection
	if err := repository.DB.First(&connection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	var input DatabaseConnectionUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		connection.Name = input.Name
	}
	if input.DbType != "" {
		connection.DbType = input.DbType
	}
	if input.Host != "" {
		connection.Host = input.Host
	}
	if input.Port != 0 {
		connection.Port = input.Port
	}
	if input.Username != "" {
		connection.Username = input.Username
	}
	if input.DatabaseName != "" {
		connection.DatabaseName = input.DatabaseName
	}
	if input.Password != "" {
		encryptedPassword, err := utils.Encrypt(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
			return
		}
		connection.PasswordEncrypted = encryptedPassword
	}

	connection.UpdatedAt = time.Now()

	if err := repository.DB.Save(&connection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connection"})
		return
	}

	connection.PasswordEncrypted = ""
	c.JSON(http.StatusOK, connection)
}

// DeleteDatabaseConnection deletes a database connection configuration
func DeleteDatabaseConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var connection model.DatabaseConnection
	if err := repository.DB.First(&connection, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	if err := repository.DB.Delete(&connection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database connection deleted successfully"})
}
