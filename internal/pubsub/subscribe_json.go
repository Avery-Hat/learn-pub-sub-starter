package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType, // <-- changed
) error {
	// Ensure queue exists and is bound
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveries, err := ch.Consume(
		queueName,
		"",    // consumer name auto-generated
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		_ = ch.Close()
		return err
	}

	go func() {
		for msg := range deliveries {
			var val T
			if err := json.Unmarshal(msg.Body, &val); err != nil {
				fmt.Println("[pubsub] Unmarshal failed -> Ack (discarding bad message):", err)
				_ = msg.Ack(false)
				continue
			}

			action := handler(val)

			switch action {
			case Ack:
				fmt.Println("[pubsub] Ack")
				_ = msg.Ack(false)
			case NackRequeue:
				fmt.Println("[pubsub] NackRequeue")
				_ = msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("[pubsub] NackDiscard")
				_ = msg.Nack(false, false)
			default:
				// Safe default for unexpected return values: discard
				fmt.Println("[pubsub] Unknown AckType -> NackDiscard")
				_ = msg.Nack(false, false)
			}
		}

		_ = ch.Close()
	}()

	return nil
}
