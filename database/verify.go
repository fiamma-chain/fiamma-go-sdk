package database

import (
	"github.com/fiamma-chain/fiamma-go-sdk/spec"
)

// Task record

func (db *Database) CreateTaskRecord(rcd *spec.TaskRecord) error {
	return db.DB.Create(rcd).Error
}

func (db *Database) CreateAndGetTaskRecord(rcd *spec.TaskRecord) (*spec.TaskRecord, error) {
	if err := db.DB.Create(rcd).Error; err != nil {
		return nil, err
	}
	var record spec.TaskRecord
	if err := db.DB.Where("tx_hash = ?", rcd.TxHash).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *Database) GetTaskRecord(txHash string) (*spec.TaskRecord, error) {
	var rcd spec.TaskRecord
	if err := db.DB.Where("tx_hash =  ?", txHash).First(&rcd).Error; err != nil {
		return nil, err
	}
	return &rcd, nil
}

func (db *Database) UpdateTaskRecord(rcd *spec.TaskRecord) error {
	return db.DB.Save(rcd).Error
}

func (db *Database) ListPendingTaskRecord(number int) (*[]spec.TaskRecord, error) {
	var rcds []spec.TaskRecord
	if err := db.DB.Where("status = ?", spec.TaskPending).Order("created_at asc").Limit(number).Find(&rcds).Error; err != nil {
		return nil, err
	}
	return &rcds, nil
}
