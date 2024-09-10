package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// ConnectRabbitMQ establishes a connection to RabbitMQ, sets up necessary queues and exchanges, and returns a channel and connection.
// This function performs the following steps:
// 1. Connects to RabbitMQ server using default credentials.
// 2. Opens a channel on the connection.
// 3. Declares a dead-letter queue (DLQ) named "orders_dlq".
// 4. Deletes any existing "orders" queue to ensure a fresh setup.
// 5. Declares a new "orders" queue with a dead-letter exchange argument pointing to "orders_dlq".
// 6. Declares a topic exchange named "order_topic".
// 7. Binds the "orders" queue to the "order_topic" exchange with a routing key pattern "order.*".
func ConnectRabbitMQ() (*amqp.Channel, *amqp.Connection, error) {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, nil, fmt.Errorf("connectRabbitMQ failed to connect to RabbitMQ: %v", err)
	}

	// Open a channel on the connection
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	// Declare the dead-letter queue (DLQ)
	_, err = ch.QueueDeclare(
		"orders_dlq", // DLQ name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Printf("Error declaring DLQ: %v", err)
		return nil, nil, fmt.Errorf("failed to declare the DLQ: %v", err)
	}

	// Check if the orders queue exists
	_, err = ch.QueueInspect("orders")
	if err != nil {
		// If the queue does not exist, declare it with the DLQ argument
		_, err = ch.QueueDeclare(
			"orders", // queue name
			true,     // durable
			false,    // delete when unused
			false,    // exclusive
			false,    // no-wait
			amqp.Table{
				"x-dead-letter-exchange":    "",           // use the default exchange for DLQ
				"x-dead-letter-routing-key": "orders_dlq", // route to DLQ if message is rejected
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to declare the orders queue: %v", err)
		}
	} else {
		// If the queue exists, ensure it has the same arguments
		_, err = ch.QueueInspect("orders_dlq")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to inspect the DLQ: %v", err)
		}
	}

	// Declare a topic exchange
	err = ch.ExchangeDeclare(
		"order_topic", // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare an exchange: %v", err)
	}

	// Bind the queue to the exchange with the "order" routing key
	err = ch.QueueBind(
		"orders",      // queue name
		"order.*",     // routing key
		"order_topic", // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to bind the queue: %v", err)
	}

	return ch, conn, nil
}

// PublishMessage sends a message to the RabbitMQ exchange with a specific routing key.
// This function performs the following steps:
// 1. Constructs the routing key by appending the messageType to "order.".
// 2. Publishes the message to the "order_topic" exchange with the constructed routing key.
// 3. Logs the message and routing key if the publish is successful.
func PublishMessage(ch *amqp.Channel, message string, messageType string) error {
	// Construct the routing key
	routingKey := "order." + messageType

	// Publish the message to the exchange
	err := ch.Publish(
		"order_topic", // exchange
		routingKey,    // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}
	log.Printf("Message sent: %s with routing key: %s", message, routingKey)
	return nil
}
