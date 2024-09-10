package main

import (
	"go-echo/app/handlers"
	"go-echo/app/rabbitmq"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Start the RabbitMQ consumer
	go rabbitmq.StartConsumer()

	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/orders", handlers.OrderHandler)

	// Start server
	log.Println("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
