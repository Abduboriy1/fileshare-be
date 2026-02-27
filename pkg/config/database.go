package config

import (
	"fileshare-be/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *Config) (*gorm.DB, error) {
	logLevel := logger.Error
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	if err := db.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.AuditLog{},
		&models.DocumentView{},
		&models.Group{},
		&models.GroupMember{},
		&models.DocumentShare{},
		&models.DocumentGroupShare{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
