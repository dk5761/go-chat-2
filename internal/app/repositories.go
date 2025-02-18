package app

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/chat-backend/internal/repository"
	"github.com/chat-backend/internal/repository/postgres"
	redisrepo "github.com/chat-backend/internal/repository/redis"
)

type repositories struct {
	userRepo    repository.UserRepository
	groupRepo   repository.GroupRepository
	messageRepo repository.MessageRepository
	statusRepo  repository.StatusRepository
}

func initRepositories(db *gorm.DB, redisClient *redis.Client) *repositories {
	return &repositories{
		userRepo:    postgres.NewUserRepository(db),
		groupRepo:   postgres.NewGroupRepository(db),
		messageRepo: postgres.NewMessageRepository(db),
		statusRepo:  redisrepo.NewStatusRepository(redisClient),
	}
}
