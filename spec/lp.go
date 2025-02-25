package spec

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	LpTxStateCreated    = "created"
	LpTxStatePending    = "pending"
	LpTxStateProcessing = "processing"
	LpTxStateInvalid    = "invalid"
	LpTxStateDone       = "done"
)

type LpTxInfo struct {
	gorm.Model
	TxHash     string          `gorm:"not null;uniqueIndex"`
	BtcAddress string          `gorm:"not null;type:citext;index"`
	EvmAddress string          `gorm:"not null;type:citext;index"`
	Amount     decimal.Decimal `gorm:"type:decimal(24,6);default:0"`
	State      string          `gorm:"not null;default:'created'"`
}
