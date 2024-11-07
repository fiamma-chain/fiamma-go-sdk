package spec

import (
	"gorm.io/gorm"
)

const (
	TaskPending = "pending"
	TaskDone    = "done"
	TaskFailed  = "failed"
)

type TaskRecord struct {
	gorm.Model
	TaskType  string  `gorm:"not null;default:'verify'"`
	RequestID string  `gorm:"uniqueIndex;not null;"`
	TxHash    *string `gorm:"uniqueIndex"`
	Address   string  `gorm:"not null;"`
	Status    string  `gorm:"not null;default:'pending'"`
}
