// Package models holds the GORM entities for the iam module.
package models

import (
	"time"

	"github.com/google/uuid"
)

// User is a panel user belonging to a company.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID    uuid.UUID `gorm:"type:uuid"`
	Email        string
	PasswordHash string
	Name         string
	Role         string // admin | operator | knowledge_manager
	Status       string // active | invited | disabled
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string { return "users" }
