package spec

import (
	"gorm.io/gorm"
)

type DevnetPioneer struct {
	gorm.Model
	Name      string `gorm:"not null"`
	Email     string `gorm:"not null;unique"`
	TwitterID string
	Project   string
	BuildPlan string `gorm:"not null"`
}
