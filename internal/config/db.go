package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		Log.Errorf("failed to connect to database: %v", err)
		return err
	}
	Log.Info("database connection established")
	return nil
}
