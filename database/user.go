package database

import (
	"github.com/fiamma-chain/fiamma-go-sdk/spec"
)

// User

func (db *Database) CreateUser(rcd *spec.User) error {
	return db.DB.Create(rcd).Error
}

func (db *Database) CreateAndGetUser(user *spec.User) (*spec.User, error) {
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	var usr spec.User
	if err := db.DB.Where("address = ?", user.Address).First(&usr).Error; err != nil {
		return nil, err
	}
	return &usr, nil
}

func (db *Database) GetUser(address string) (*spec.User, error) {
	var usr spec.User
	if err := db.DB.Where("address =  ?", address).First(&usr).Error; err != nil {
		return nil, err
	}
	return &usr, nil
}

func (db *Database) UpdateUser(usr *spec.User) error {
	return db.DB.Save(usr).Error
}
