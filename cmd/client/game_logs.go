package main

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func publishGameLog(ch *amqp.Channel, initiatorUsername, msg string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     msg,
		Username:    initiatorUsername,
	}

	key := routing.GameLogSlug + "." + initiatorUsername
	return pubsub.PublishGob(ch, routing.ExchangePerilTopic, key, log)
}
