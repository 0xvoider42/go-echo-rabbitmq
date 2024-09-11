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

	// Route to handle GET all orders
	e.GET("/orders", handlers.GetAllOrdersHandler)

	// // Route to handle GET a specific order by ID
	e.GET("/orders/:id", handlers.GetOrderHandler)

	// // Route to handle PUT to update a specific order by ID
	e.PUT("/orders/:id", handlers.UpdateOrderHandler)

	// // Route to handle DELETE a specific order by ID
	e.DELETE("/orders/:id", handlers.DeleteOrderHandler)

	// Start server
	log.Println("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
