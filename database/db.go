package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	zlog "github.com/fiamma-chain/fiamma-go-sdk/log"
)

type Database struct {
	DB  *gorm.DB
	log *zlog.Logger
}

func NewDB(url string) (*Database, error) {
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			IgnoreRecordNotFoundError: true,
		},
	)
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}
	return &Database{
		DB:  db,
		log: zlog.L().With(zlog.Any("service", "db")),
	}, nil
}
