package app

import (
	firebase "firebase.google.com/go/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/chat-backend/internal/service"
	"github.com/chat-backend/internal/websocket"
)

type services struct {
	userService         *service.UserService
	groupService        *service.GroupService
	messageService      *service.MessageService
	notificationService *service.NotificationService
	wsManager           *websocket.Manager
}

func initServices(repos *repositories, logger *logrus.Logger, rabbitmqChan *amqp.Channel, firebaseApp *firebase.App, redisClient *redis.Client) (*services, error) {
	wsManager := websocket.NewManager(logger, redisClient)

	userService := service.NewUserService(repos.userRepo, repos.statusRepo, viper.GetString("jwt.secret"))
	groupService := service.NewGroupService(repos.groupRepo, repos.userRepo)
	messageService := service.NewMessageService(repos.messageRepo, repos.userRepo, repos.groupRepo, wsManager)

	notificationService, err := service.NewNotificationService(
		firebaseApp,
		rabbitmqChan,
		"notifications",
		"chat_exchange",
		"notifications",
	)
	if err != nil {
		return nil, err
	}

	return &services{
		userService:         userService,
		groupService:        groupService,
		messageService:      messageService,
		notificationService: notificationService,
		wsManager:           wsManager,
	}, nil
}
