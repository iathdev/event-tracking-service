package database

import (
	"event-tracking-service/config"
	"event-tracking-service/pkg/utils"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 100
	defaultConnMaxLifetime = 60 // minutes
	defaultConnMaxIdleTime = 10 // minutes
)

func NewPostgresConnection(cfg *config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	maxIdle := cfg.MaxIdle
	maxOpen := cfg.MaxOpen
	maxLife := cfg.MaxLife
	maxIdleTime := cfg.MaxIdleTime
	sqlDB.SetMaxIdleConns(utils.MaxInt(maxIdle, defaultMaxIdleConns))
	sqlDB.SetMaxOpenConns(utils.MaxInt(maxOpen, defaultMaxOpenConns))
	sqlDB.SetConnMaxLifetime(time.Duration(utils.MaxInt(maxLife, defaultConnMaxLifetime)) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(utils.MaxInt(maxIdleTime, defaultConnMaxIdleTime)) * time.Minute)

	return db, nil
}

func DisconnectPostgres(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB from gorm: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close Postgres connection: %v", err)
	}

	return nil
}
