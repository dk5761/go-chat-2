package migrations

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/chat-backend/internal/repository/mongodb"
)

type Migrator struct {
	logger     *logrus.Logger
	postgresDB *gorm.DB
	mongoDB    *mongodb.DB
}

func NewMigrator(logger *logrus.Logger, postgresDB *gorm.DB, mongoDB *mongodb.DB) *Migrator {
	return &Migrator{
		logger:     logger,
		postgresDB: postgresDB,
		mongoDB:    mongoDB,
	}
}

func (m *Migrator) RunMigrations(ctx context.Context) error {
	if err := m.runPostgresMigrations(); err != nil {
		return fmt.Errorf("postgres migrations failed: %w", err)
	}

	if err := m.runMongoMigrations(ctx); err != nil {
		return fmt.Errorf("mongodb migrations failed: %w", err)
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

func (m *Migrator) runMongoMigrations(ctx context.Context) error {
	m.logger.Info("Running MongoDB migrations")

	// Read and execute migration files
	files := []string{
		"scripts/init-mongo.js",
	}

	for _, file := range files {
		m.logger.Infof("Applying MongoDB migration: %s", file)
		if err := m.executeJSFile(ctx, file); err != nil {
			return err
		}
	}

	m.logger.Info("MongoDB migrations completed")
	return nil
}

func (m *Migrator) executeJSFile(ctx context.Context, filePath string) error {
	// Read the JS file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
	}

	// Split into individual statements by semicolon
	statements := strings.Split(string(content), ";")

	// Execute each non-empty statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := m.mongoDB.GetDatabase().RunCommand(ctx, stmt).Err(); err != nil {
			return fmt.Errorf("failed to execute statement from %s: %w", filePath, err)
		}
	}

	return nil
}
