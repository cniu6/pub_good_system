package models

import (
	"time"
)

// BaseModel defines the common fields for all models
type BaseModel struct {
	ID        uint64     `db:"id" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"` // Soft delete field
}

// IsDeleted checks if the record is soft-deleted
func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt != nil
}
