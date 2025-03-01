package app

import (
	"context"

	firebase "firebase.google.com/go/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

func initRabbitMQ() (*amqp.Connection, error) {
	uri := viper.GetString("rabbitmq.uri")
	if uri == "" {
		uri = "amqp://guest:guest@rabbitmq:5672"
	}
	return amqp.Dial(uri)
}

func initFirebase() (*firebase.App, error) {
	opt := option.WithCredentialsFile(viper.GetString("fcm.credentials_file"))
	config := &firebase.Config{ProjectID: viper.GetString("fcm.project_id")}
	return firebase.NewApp(context.Background(), config, opt)
}
