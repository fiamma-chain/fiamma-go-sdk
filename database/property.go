package database

import (
	"github.com/fiamma-chain/fiamma-go-sdk/spec"
)

func (db *Database) GetProperty(name string) (string, error) {
	var prop spec.Property
	if err := db.DB.Where("name = ?", name).First(&prop).Error; err != nil {
		return "", err
	}
	return prop.Value, nil
}

func (db *Database) UpdateProperty(prop *spec.Property) error {
	return db.DB.Model(&spec.Property{}).Where("name = ?", prop.Name).
		Update("value", prop.Value).Error
}

func (db *Database) UpdateAndGetProperty(prop *spec.Property) (*spec.Property, error) {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&spec.Property{}).Where("name = ?", prop.Name).
		Update("value", prop.Value).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var newProp spec.Property
	if err := tx.Where("name = ?", prop.Name).First(&newProp).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &newProp, nil
}
