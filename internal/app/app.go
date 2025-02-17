package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"

	"github.com/chat-backend/internal/migrations"
)

type App struct {
	server   *Server
	logger   *logrus.Logger
	services *services
	handlers *handlers
	wg       conc.WaitGroup
}

func New() (*App, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	if err := loadConfig(); err != nil {
		return nil, err
	}

	// Initialize databases
	db, err := initPostgres()
	if err != nil {
		return nil, err
	}

	mongoDB, err := initMongoDB(context.Background())
	if err != nil {
		return nil, err
	}

	redisClient := initRedis()

	// Initialize messaging
	rabbitmqConn, err := initRabbitMQ()
	if err != nil {
		return nil, err
	}

	rabbitmqChan, err := rabbitmqConn.Channel()
	if err != nil {
		return nil, err
	}

	firebaseApp, err := initFirebase()
	if err != nil {
		return nil, err
	}

	// Run migrations
	migrator := migrations.NewMigrator(logger, db, mongoDB)
	if err := migrator.RunMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repositories
	repos := initRepositories(db, mongoDB, redisClient)

	// Initialize services
	services, err := initServices(repos, logger, rabbitmqChan, firebaseApp)
	if err != nil {
		return nil, err
	}

	// Initialize handlers
	handlers := initHandlers(services)

	// Initialize server
	server := NewServer(
		logger,
		handlers.userHandler,
		handlers.groupHandler,
		handlers.messageHandler,
		handlers.wsHandler,
		handlers.authMiddleware,
		handlers.healthHandler,
	)

	return &App{
		server:   server,
		logger:   logger,
		services: services,
		handlers: handlers,
		wg:       conc.WaitGroup{},
	}, nil
}

func (a *App) Start() error {
	ctx := context.Background()

	// Start WebSocket manager
	a.wg.Go(func() {
		a.logger.Info("Starting WebSocket manager")
		a.services.wsManager.Start(ctx)
	})

	// Start notification consumer
	a.wg.Go(func() {
		a.logger.Info("Starting notification consumer")
		if err := a.services.notificationService.StartConsumer(ctx); err != nil {
			a.logger.Error("Notification consumer error: ", err)
		}
	})

	// Start HTTP server
	a.wg.Go(func() {
		a.logger.Info("Starting HTTP server")
		if err := a.server.Start(); err != nil {
			a.logger.Error("Server error: ", err)
		}
	})

	return nil
}

func (a *App) WaitForShutdown() {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down...")

	// Create shutdown timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown: ", err)
	}

	// Wait for all goroutines to complete
	a.wg.Wait()

	a.logger.Info("Server exited properly")
}
