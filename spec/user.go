package spec

import (
	"encoding/json"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Address             string          `gorm:"type:char(66);primary_key"`
	Point               decimal.Decimal `gorm:"type:decimal(24,4);not null;default:0.0000"`
	TaskPoint           decimal.Decimal `gorm:"type:decimal(24,4);not null;default:0.0000"`
	ReferralPoint       decimal.Decimal `gorm:"type:decimal(24,4);not null;default:0.0000"`
	SecondReferralPoint decimal.Decimal `gorm:"type:decimal(24,4);not null;default:0.0000"`
	TwitterID           string          `gorm:"type:varchar(255);not null;default:'';index:idx_twitter_id"`
	TwitterProfile      json.RawMessage `gorm:"type:jsonb;not null"`
	Blocked             bool            `gorm:"type:boolean;not null;default:false;index:idx_blocked"`
	Mining              bool            `gorm:"type:boolean;not null;default:false"`
	VerifiedBlocks      int32           `gorm:"type:integer;not null;default:0"`
	Campaign            string          `gorm:"type:varchar(64);not null;default:''"`
}
