package migrations

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gocql/gocql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Migrator struct {
	logger           *logrus.Logger
	postgresDB       *gorm.DB
	cassandraSession *gocql.Session
}

func NewMigrator(logger *logrus.Logger, postgresDB *gorm.DB, cassandraSession *gocql.Session) *Migrator {
	return &Migrator{
		logger:           logger,
		postgresDB:       postgresDB,
		cassandraSession: cassandraSession,
	}
}

func (m *Migrator) RunMigrations(ctx context.Context) error {
	if err := m.runPostgresMigrations(); err != nil {
		return fmt.Errorf("postgres migrations failed: %w", err)
	}

	if err := m.runCassandraMigrations(ctx); err != nil {
		return fmt.Errorf("cassandra migrations failed: %w", err)
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

func (m *Migrator) runCassandraMigrations(ctx context.Context) error {
	m.logger.Info("Running Cassandra migrations")

	// Read and execute migration files
	files := []string{
		"migrations/cassandra/000001_create_keyspace.up.cql",
		"migrations/cassandra/000002_create_messages.up.cql",
	}

	for _, file := range files {
		m.logger.Infof("Applying Cassandra migration: %s", file)
		if err := m.executeCQLFile(ctx, file); err != nil {
			return err
		}
	}

	m.logger.Info("Cassandra migrations completed")
	return nil
}

func (m *Migrator) executeCQLFile(ctx context.Context, filePath string) error {
	// In a real implementation, you would:
	// 1. Read the CQL file
	// 2. Split into individual statements
	// 3. Execute each statement
	// For now, we'll use the hardcoded schema since we know it

	// Read the CQL file content
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

		if err := m.cassandraSession.Query(stmt).WithContext(ctx).Exec(); err != nil {
			return fmt.Errorf("failed to execute statement from %s: %w", filePath, err)
		}
	}

	return nil
}
