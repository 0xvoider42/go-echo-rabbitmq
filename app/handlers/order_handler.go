package handlers

import (
	"log"
	"net/http"
	"sync"

	"go-echo/app/rabbitmq"

	"github.com/labstack/echo/v4"
)

// Order represents the structure of an order with fields for ID, item, price, and message type.
type Order struct {
	ID          string `json:"id"`
	Item        string `json:"item"`
	Price       int    `json:"price"`
	MessageType string `json:"message_type"`
}

var orders = make(map[string]*Order)
var mu sync.RWMutex

// OrderHandler receives HTTP requests and sends messages to RabbitMQ
// This function handles incoming HTTP requests, binds the request body to an Order struct,
// connects to RabbitMQ, publishes a message, and returns an appropriate HTTP response.
func OrderHandler(c echo.Context) error {
	// Create a new Order instance
	order := new(Order)

	// Bind the incoming request body to the Order struct
	if err := c.Bind(order); err != nil {
		// Log an error if binding fails and return a 400 Bad Request response
		log.Printf("Error binding request body: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body",
		})
	}

	// Save order to the global map
	mu.Lock()
	orders[order.ID] = order
	mu.Unlock()

	// Connect to RabbitMQ
	// This function returns a channel and a connection to RabbitMQ
	ch, conn, err := rabbitmq.ConnectRabbitMQ()
	if err != nil {
		// Log an error if connection fails and return a 500 Internal Server Error response
		log.Printf("Error connecting to RabbitMQ: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "OrderHandler failed to connect to RabbitMQ",
		})
	}
	// Ensure the connection and channel are closed when the function exits
	defer conn.Close()
	defer ch.Close()

	// Publish the message to RabbitMQ
	// The message type (e.g., order.created, order.updated) determines the routing of the message
	err = rabbitmq.PublishMessage(ch, order.ID, order.MessageType)
	if err != nil {
		// Log an error if publishing fails and return a 500 Internal Server Error response
		log.Printf("Error publishing message to RabbitMQ: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to send message to RabbitMQ",
		})
	}

	// Log the successful receipt and queuing of the order
	log.Printf("Order received and queued: ID=%s, MessageType=%s", order.ID, order.MessageType)

	// Return a 200 OK response with the order details
	return c.JSON(http.StatusOK, echo.Map{
		"message":     "Order received and queued",
		"orderID":     order.ID,
		"messageType": order.MessageType,
	})
}

func GetAllOrdersHandler(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	return c.JSON(http.StatusOK, orders)
}

func GetOrderHandler(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	id := c.Param("id")
	order, ok := orders[id]
	if !ok {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Order not found",
		})
	}

	return c.JSON(http.StatusOK, order)
}

func UpdateOrderHandler(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	id := c.Param("id")
	order, ok := orders[id]
	if !ok {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Order not found",
		})
	}

	newOrder := new(Order)
	if err := c.Bind(newOrder); err != nil {
		log.Printf("Error binding request body: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request body",
		})
	}

	order.Item = newOrder.Item
	order.Price = newOrder.Price

	return c.JSON(http.StatusOK, order)
}

func DeleteOrderHandler(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	id := c.Param("id")
	_, ok := orders[id]
	if !ok {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Order not found",
		})
	}

	delete(orders, id)

	return c.NoContent(http.StatusNoContent)
}
