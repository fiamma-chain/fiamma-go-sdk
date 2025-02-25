package database

import (
	"github.com/fiamma-chain/fiamma-go-sdk/spec"
)

func (db *Database) GetLpTransaction(txHash, btcAddress string) (*spec.LpTxInfo, error) {
	var txInfo spec.LpTxInfo
	if err := db.DB.Where("tx_hash = ? AND btc_address = ?", txHash, btcAddress).First(&txInfo).Error; err != nil {
		return nil, err
	}
	return &txInfo, nil
}

func (db *Database) CreateLpTransaction(tx *spec.LpTxInfo) error {
	return db.DB.Create(tx).Error
}

func (db *Database) CreateAndGetLpTransaction(tx *spec.LpTxInfo) (*spec.LpTxInfo, error) {
	if err := db.DB.Create(tx).Error; err != nil {
		return nil, err
	}
	var txInfo spec.LpTxInfo
	if err := db.DB.Where("tx_hash = ? AND btc_address = ?", tx.TxHash, tx.BtcAddress).First(&txInfo).Error; err != nil {
		return nil, err
	}
	return &txInfo, nil
}

func (db *Database) ListLpTransactionByAddress(btcAddress string) (*[]spec.LpTxInfo, error) {
	var txInfos []spec.LpTxInfo
	if err := db.DB.Where("btc_address = ?", btcAddress).Find(&txInfos).Error; err != nil {
		return nil, err
	}
	return &txInfos, nil
}

func (db *Database) UpdateLpTransaction(tx *spec.LpTxInfo) error {
	return db.DB.Save(tx).Error
}

func (db *Database) DeleteLpTransaction(txHash, btcAddress string) error {
	var txInfo spec.LpTxInfo
	if err := db.DB.Where("tx_hash = ? AND btc_address = ?", txHash, btcAddress).Delete(&txInfo).Error; err != nil {
		return err
	}
	return nil
}
