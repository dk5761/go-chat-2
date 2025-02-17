package main

import (
	"log"

	"github.com/chat-backend/internal/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatal("Failed to initialize application: ", err)
	}

	if err := app.Start(); err != nil {
		log.Fatal("Failed to start application: ", err)
	}

	app.WaitForShutdown()
}
