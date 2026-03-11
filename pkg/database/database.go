package database

import (
	"event-tracking-service/config"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Connections struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewConnections(cfg *config.Config) (*Connections, error) {
	postgresDB, err := NewPostgresConnection(&cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	redisClient, err := NewRedisClient(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Connections{
		DB:    postgresDB,
		Redis: redisClient,
	}, nil
}

func (c *Connections) Close() {
	_ = DisconnectPostgres(c.DB)
	_ = DisconnectRedis(c.Redis)
}
