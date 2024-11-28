package spec

import (
	"gorm.io/gorm"
)

type Property struct {
	gorm.Model
	Name  string `gorm:"unique;not null"`
	Value string `gorm:"not null"`
}
