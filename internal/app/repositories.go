package app

import (
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/chat-backend/internal/repository"
	"github.com/chat-backend/internal/repository/cassandra"
	"github.com/chat-backend/internal/repository/postgres"
	redisrepo "github.com/chat-backend/internal/repository/redis"
)

type repositories struct {
	userRepo    repository.UserRepository
	groupRepo   repository.GroupRepository
	messageRepo repository.MessageRepository
	statusRepo  repository.StatusRepository
}

func initRepositories(db *gorm.DB, cassandraSession *gocql.Session, redisClient *redis.Client) *repositories {
	return &repositories{
		userRepo:    postgres.NewUserRepository(db),
		groupRepo:   postgres.NewGroupRepository(db),
		messageRepo: cassandra.NewMessageRepository(cassandraSession),
		statusRepo:  redisrepo.NewStatusRepository(redisClient),
	}
}
