package migrations

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Migrator struct {
	logger     *logrus.Logger
	postgresDB *gorm.DB
}

func NewMigrator(logger *logrus.Logger, postgresDB *gorm.DB) *Migrator {
	return &Migrator{
		logger:     logger,
		postgresDB: postgresDB,
	}
}

func (m *Migrator) RunMigrations(ctx context.Context) error {
	if err := m.runPostgresMigrations(); err != nil {
		return fmt.Errorf("postgres migrations failed: %w", err)
	}

	return nil
}

func (m *Migrator) runPostgresMigrations() error {
	m.logger.Info("Running PostgreSQL migrations")

	sqlDB, err := m.postgresDB.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations/postgres",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	m.logger.Info("PostgreSQL migrations completed")
	return nil
}
