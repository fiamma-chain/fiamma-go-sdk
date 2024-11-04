package spec

import (
	"gorm.io/gorm"
)

const (
	VerifyTaskPending = "pending"
	VerifyTaskDone    = "done"
	VerifyTaskFailed  = "failed"
)

type VerificationRecord struct {
	gorm.Model
	RequestID string `gorm:"unique;not null;"`
	TxHash    string `gorm:"unique;not null;"`
	Address   string `gorm:"not null;"`
	Status    string `gorm:"not null;default:'pending'"`
}
