package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	PasswordEncrypted string    `gorm:"type:text;not null" json:"password_encrypted,omitempty"`
	DatabaseName      string    `gorm:"type:varchar(100);not null" json:"database_name"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AIProvider represents configurations for custom or standard AI engines (Gemini, OpenAI, Ollama, etc.)
type AIProvider struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name"`
	ApiEndpoint     string    `gorm:"type:text;not null" json:"api_endpoint"`
	ApiKeyEncrypted string    `gorm:"type:text;not null" json:"api_key_encrypted,omitempty"`
	ModelName       string    `gorm:"type:varchar(100);not null" json:"model_name"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

// StringArray handles string slices serialized as JSON in PostgreSQL
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return errors.New("failed to scan StringArray")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, a)
}

// ScheduledTask defines standard tasks that run periodically to optimize queries using rules
type ScheduledTask struct {
	ID                   uint        `gorm:"primaryKey" json:"id"`
	Name                 string      `gorm:"type:varchar(100);not null" json:"name"`
	CronExpression       string      `gorm:"type:varchar(50);not null" json:"cron_expression"`
	DatabaseConnectionID uint        `json:"database_connection_id"`
	AIProviderID         uint        `json:"ai_provider_id"`
	TargetQueries        StringArray `gorm:"type:text;not null" json:"target_queries"`
	SkillRules           string      `gorm:"type:text" json:"skill_rules"`
	IsActive             bool        `gorm:"default:true" json:"is_active"`
	CreatedAt            time.Time   `json:"created_at"`
}

// OptimizationReport stores query execution plans, recommendations, and fix status
type OptimizationReport struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	TaskID               *uint      `json:"task_id,omitempty"`
	DatabaseConnectionID uint       `json:"database_connection_id"`
	QueryText            string     `gorm:"type:text;not null" json:"query_text"`
	ExecutionPlan        string     `gorm:"type:text" json:"execution_plan"`
	AnalysisResult       string     `gorm:"type:jsonb;not null" json:"analysis_result"`
	Status               string     `gorm:"type:varchar(50);default:'Pending Review'" json:"status"` // 'Pending Review', 'Applied', 'Skipped'
	AppliedFix           string     `gorm:"type:text" json:"applied_fix,omitempty"`
	AppliedAt            *time.Time `json:"applied_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}
