package app

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func initCassandra() (*gocql.Session, error) {
	cluster := gocql.NewCluster(viper.GetStringSlice("cassandra.hosts")...)
	cluster.Keyspace = viper.GetString("cassandra.keyspace")
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = viper.GetDuration("cassandra.timeout")
	cluster.ConnectTimeout = viper.GetDuration("cassandra.connect_timeout")

	return cluster.CreateSession()
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	})
}
