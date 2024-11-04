package database

import (
	"github.com/fiamma-chain/fiamma-go-sdk/spec"
)

// Verify record

func (db *Database) CreateVerifyRecord(rcd *spec.VerificationRecord) error {
	return db.DB.Create(rcd).Error
}

func (db *Database) CreateAndGetVerifyRecord(rcd *spec.VerificationRecord) (*spec.VerificationRecord, error) {
	if err := db.DB.Create(rcd).Error; err != nil {
		return nil, err
	}
	var record spec.VerificationRecord
	if err := db.DB.Where("tx_hash = ?", rcd.TxHash).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *Database) GetVerifyRecord(txHash string) (*spec.VerificationRecord, error) {
	var rcd spec.VerificationRecord
	if err := db.DB.Where("tx_hash =  ?", txHash).First(&rcd).Error; err != nil {
		return nil, err
	}
	return &rcd, nil
}

func (db *Database) UpdateVerifyRecord(rcd *spec.VerificationRecord) error {
	return db.DB.Save(rcd).Error
}

func (db *Database) ListPendingVerifyRecord(number int) (*[]spec.VerificationRecord, error) {
	var rcds []spec.VerificationRecord
	if err := db.DB.Where("status = ?", spec.VerifyTaskPending).Order("created_at asc").Limit(number).Find(&rcds).Error; err != nil {
		return nil, err
	}
	return &rcds, nil
}
