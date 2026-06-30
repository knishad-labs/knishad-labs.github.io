package model

import (
	"time"
)

// DatabaseConnection represents configuration for target databases to optimize
type DatabaseConnection struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"type:varchar(100);not null" json:"name"`
	DbType            string    `gorm:"type:varchar(20);not null" json:"db_type"` // 'postgres', 'mssql', 'mariadb', 'mysql'
	Host              string    `gorm:"type:varchar(255);not null" json:"host"`
	Port              int       `gorm:"not null" json:"port"`
	Username          string    `gorm:"type:varchar(100);not null" json:"username"`
	PasswordEncrypted string    `gorm:"type:text;not null" json:"password_encrypted,omitempty"` // stored encrypted
	DatabaseName      string    `gorm:"type:varchar(100);not null" json:"database_name"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AIProvider represents configurations for custom or standard AI engines (Gemini, OpenAI, Ollama, etc.)
type AIProvider struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name"`
	ApiEndpoint     string    `gorm:"type:text;not null" json:"api_endpoint"`
	ApiKeyEncrypted string    `gorm:"type:text;not null" json:"api_key_encrypted,omitempty"` // stored encrypted
	ModelName       string    `gorm:"type:varchar(100);not null" json:"model_name"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}
