package app

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chat-backend/internal/repository/mongodb"
)

func initPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("postgres.host"),
		viper.GetInt("postgres.port"),
		viper.GetString("postgres.user"),
		viper.GetString("postgres.password"),
		viper.GetString("postgres.database"),
	)

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(viper.GetInt("postgres.max_open_conns"))
	sqlDB.SetMaxIdleConns(viper.GetInt("postgres.max_idle_conns"))
	sqlDB.SetConnMaxLifetime(viper.GetDuration("postgres.conn_max_lifetime"))

	return db, nil
}

func initMongoDB(ctx context.Context) (*mongodb.DB, error) {
	return mongodb.NewDB(
		ctx,
		viper.GetString("mongodb.uri"),
		viper.GetString("mongodb.database"),
	)
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	})
}
