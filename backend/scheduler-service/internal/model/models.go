package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

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
