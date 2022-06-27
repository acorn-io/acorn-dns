package db

import (
	"time"

	"gorm.io/gorm"
)

type Domain struct {
	gorm.Model
	UniqueSlug  string `gorm:"uniqueIndex"`
	Domain      string `gorm:"uniqueIndex"`
	TokenHash   string
	LastCheckIn time.Time
}

type Record struct {
	ID          uint   `gorm:"primarykey"`
	FQDN        string `gorm:"uniqueIndex:idx_record,priority:1"`
	Type        string `gorm:"uniqueIndex:idx_record,priority:2"`
	DomainID    uint
	Domain      Domain `gorm:"constraint:OnDelete:SET NULL;"`
	Values      string `gorm:"type:text"` // Intentionally denormalized because we don't want to create a values table
	CreatedAt   time.Time
	LastCheckIn time.Time
}
