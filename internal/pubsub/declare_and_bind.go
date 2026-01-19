package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := queueType == SimpleQueueDurable
	autoDelete := queueType == SimpleQueueTransient
	exclusive := queueType == SimpleQueueTransient

	q, err := ch.QueueDeclare(
		queueName,
		durable,    // durable
		autoDelete, // autoDelete
		exclusive,  // exclusive
		false,      // noWait
		nil,        // args
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	if err := ch.QueueBind(
		q.Name,   // queue name
		key,      // routing key
		exchange, // exchange
		false,    // noWait
		nil,      // args
	); err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
