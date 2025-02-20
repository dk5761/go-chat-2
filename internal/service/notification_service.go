package service

import (
	"context"
	"encoding/json"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sourcegraph/conc"

	"github.com/chat-backend/internal/models"
)

type NotificationService struct {
	fcmClient    *messaging.Client
	rabbitmqChan *amqp.Channel
	queueName    string
	exchangeName string
	routingKey   string
}

type PushNotification struct {
	UserID      string            `json:"user_id"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data"`
	Priority    string            `json:"priority"`
	DeviceToken string            `json:"device_token"`
}

func NewNotificationService(app *firebase.App, rabbitmqChan *amqp.Channel, queueName, exchangeName, routingKey string) (*NotificationService, error) {
	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Messaging client: %v", err)
	}

	// Declare RabbitMQ exchange and queue
	err = rabbitmqChan.ExchangeDeclare(
		exchangeName,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	_, err = rabbitmqChan.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	err = rabbitmqChan.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %v", err)
	}

	return &NotificationService{
		fcmClient:    fcmClient,
		rabbitmqChan: rabbitmqChan,
		queueName:    queueName,
		exchangeName: exchangeName,
		routingKey:   routingKey,
	}, nil
}

func (s *NotificationService) QueueNotification(ctx context.Context, notification *PushNotification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %v", err)
	}

	return s.rabbitmqChan.PublishWithContext(
		ctx,
		s.exchangeName,
		s.routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (s *NotificationService) StartConsumer(ctx context.Context) error {
	msgs, err := s.rabbitmqChan.Consume(
		s.queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	var wg conc.WaitGroup
	for msg := range msgs {
		msg := msg // Create new variable for goroutine
		wg.Go(func() {
			var notification PushNotification
			if err := json.Unmarshal(msg.Body, &notification); err != nil {
				// TODO: Add proper error logging
				return
			}

			if err := s.sendPushNotification(ctx, &notification); err != nil {
				// TODO: Add proper error logging and retry mechanism
				return
			}
		})
	}

	return nil
}

func (s *NotificationService) sendPushNotification(ctx context.Context, notification *PushNotification) error {
	message := &messaging.Message{
		Token: notification.DeviceToken,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Data: notification.Data,
	}

	if notification.Priority == "high" {
		message.Android = &messaging.AndroidConfig{
			Priority: "high",
		}
	}

	_, err := s.fcmClient.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

func (s *NotificationService) NotifyNewMessage(ctx context.Context, message *models.Message, recipientToken string) error {
	notification := &PushNotification{
		UserID:      message.RecipientID.String(),
		Title:       "New Message",
		Body:        fmt.Sprintf("You have a new message from %s", message.SenderID.String()),
		DeviceToken: recipientToken,
		Priority:    "high",
		Data: map[string]string{
			"message_id": message.ID.String(),
			"sender_id":  message.SenderID.String(),
			"type":       "new_message",
		},
	}

	return s.QueueNotification(ctx, notification)
}

func (s *NotificationService) NotifyGroupMessage(ctx context.Context, message *models.Message, groupName string, recipientTokens []string) error {
	var wg conc.WaitGroup
	errors := make(chan error, len(recipientTokens))

	for _, token := range recipientTokens {
		token := token // Create new variable for goroutine
		wg.Go(func() {
			notification := &PushNotification{
				UserID:      message.GroupID.String(),
				Title:       fmt.Sprintf("New message in %s", groupName),
				Body:        fmt.Sprintf("New message from %s", message.SenderID.String()),
				DeviceToken: token,
				Priority:    "high",
				Data: map[string]string{
					"message_id": message.ID.String(),
					"sender_id":  message.SenderID.String(),
					"group_id":   message.GroupID.String(),
					"type":       "group_message",
				},
			}

			if err := s.QueueNotification(ctx, notification); err != nil {
				errors <- err
			}
		})
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}
