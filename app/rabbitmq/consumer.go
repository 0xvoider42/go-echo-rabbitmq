package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

// StartConsumer listens for messages on the orders queue
// This function establishes a connection to RabbitMQ, declares a queue, and consumes messages from it.
// It processes each message by calling the processOrder function.
func StartConsumer() {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		// Log an error and exit if the connection fails
		log.Fatalf("StartConsumer failed to connect to RabbitMQ: %v", err)
	}
	// Ensure the connection is closed when the function exits
	defer conn.Close()

	// Open a channel to communicate with RabbitMQ
	ch, err := conn.Channel()
	if err != nil {
		// Log an error and exit if opening the channel fails
		log.Fatalf("Failed to open a channel: %v", err)
	}
	// Ensure the channel is closed when the function exits
	defer ch.Close()

	// Declare the orders queue
	// This ensures the queue exists before we try to consume messages from it
	q, err := ch.QueueDeclare(
		"orders", // name of the queue
		true,     // durable (the queue will survive a broker restart)
		false,    // delete when unused
		false,    // exclusive (used by only one connection and the queue will be deleted when that connection closes)
		false,    // no-wait (do not wait for a server response)
		nil,      // arguments (optional additional arguments)
	)
	if err != nil {
		// Log an error and exit if declaring the queue fails
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // name of the queue
		"",     // consumer tag (can be empty)
		true,   // auto-ack (automatically acknowledge message receipt)
		false,  // exclusive (used by only this consumer)
		false,  // no-local (not supported by RabbitMQ)
		false,  // no-wait (do not wait for a server response)
		nil,    // arguments (optional additional arguments)
	)
	if err != nil {
		// Log an error and exit if registering the consumer fails
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// Process the messages in a separate goroutine
	go func() {
		for msg := range msgs {
			// Log the received message
			log.Printf("Received a message: %s", msg.Body)

			// Process the message
			// This function can be extended to integrate with databases or other services
			processOrder(string(msg.Body))
		}
	}()

	// Log that the consumer has started and is waiting for messages
	log.Println("Consumer started. Waiting for messages...")
	// Block forever to keep the consumer running
	select {}
}

// processOrder processes the message (order data)
// This function simulates order processing, such as saving to a database or calling an external API.
func processOrder(orderID string) {
	// Log the order being processed
	log.Printf("Processing order: %s", orderID)
	// Simulate order processing (e.g., saving to DB, external API call)
}
