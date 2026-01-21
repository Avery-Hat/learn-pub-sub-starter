package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	// Limit unacked messages per consumer (prefetch)
	if err := ch.Qos(10, 0, false); err != nil {
		_ = ch.Close()
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
			// Optional but helpful: verify content type
			if msg.ContentType != "" && msg.ContentType != "application/gob" {
				fmt.Println("[pubsub] Wrong content type -> NackDiscard:", msg.ContentType)
				_ = msg.Nack(false, false)
				continue
			}

			var val T
			dec := gob.NewDecoder(bytes.NewReader(msg.Body))
			if err := dec.Decode(&val); err != nil {
				// Poison message: discard so it doesn't loop forever
				fmt.Println("[pubsub] Gob decode failed -> Ack (discarding bad message):", err)
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
				fmt.Println("[pubsub] Unknown AckType -> NackDiscard")
				_ = msg.Nack(false, false)
			}
		}

		_ = ch.Close()
	}()

	return nil
}
