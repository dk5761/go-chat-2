package app

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/chat-backend/internal/repository"
	"github.com/chat-backend/internal/repository/mongodb"
	"github.com/chat-backend/internal/repository/postgres"
	redisrepo "github.com/chat-backend/internal/repository/redis"
)

type repositories struct {
	userRepo    repository.UserRepository
	groupRepo   repository.GroupRepository
	messageRepo repository.MessageRepository
	statusRepo  repository.StatusRepository
}

func initRepositories(db *gorm.DB, mongoDB *mongodb.DB, redisClient *redis.Client) *repositories {
	return &repositories{
		userRepo:    postgres.NewUserRepository(db),
		groupRepo:   postgres.NewGroupRepository(db),
		messageRepo: mongodb.NewMessageRepository(mongoDB.GetDatabase()),
		statusRepo:  redisrepo.NewStatusRepository(redisClient),
	}
}
