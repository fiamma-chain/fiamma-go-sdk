package database

import (
	"github.com/fiamma-chain/fiamma-go-sdk/spec"
)

// Devnet Pioneer

func (db *Database) CreatePioneer(pioneer *spec.DevnetPioneer) error {
	return db.DB.Create(pioneer).Error
}

func (db *Database) CreateAndGetPioneer(pioneer *spec.DevnetPioneer) (*spec.DevnetPioneer, error) {
	if err := db.DB.Create(pioneer).Error; err != nil {
		return nil, err
	}
	var res spec.DevnetPioneer
	if err := db.DB.Where("id = ?", pioneer.ID).First(&res).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

func (db *Database) GetPioneer(email string) (*spec.DevnetPioneer, error) {
	var pioneer spec.DevnetPioneer
	if err := db.DB.Where("email =  ?", email).First(&pioneer).Error; err != nil {
		return nil, err
	}
	return &pioneer, nil
}

func (db *Database) ListPioneers() (*[]spec.DevnetPioneer, error) {
	var res []spec.DevnetPioneer
	if err := db.DB.Order("created_at asc").Find(&res).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

func (db *Database) UpdatePioneer(usr *spec.DevnetPioneer) error {
	return db.DB.Save(usr).Error
}
