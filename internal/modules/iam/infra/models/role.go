package models

import (
	"time"

	"github.com/google/uuid"
)

// SystemRole is a global application role (admin, operator, knowledge_manager).
type SystemRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string
	Description string
	CreatedAt   time.Time
}

func (SystemRole) TableName() string { return "roles" }
