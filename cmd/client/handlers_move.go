package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerMove(gs *gamelogic.GameState, pubCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			defender := gs.GetPlayerSnap()

			warMsg := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: defender,
			}

			routingKey := routing.WarRecognitionsPrefix + "." + defender.Username
			if err := pubsub.PublishJSON(pubCh, routing.ExchangePerilTopic, routingKey, warMsg); err != nil {
				fmt.Println("Failed to publish war recognition:", err)
				// fix: publish failure -> requeue
				return pubsub.NackRequeue
			}

			fmt.Println("Published war recognition:", routingKey)
			// fix: publish success -> ack (no more requeue hell)
			return pubsub.Ack

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard

		default:
			// Anything unexpected: discard
			fmt.Println("Unexpected move outcome:", outcome)
			return pubsub.NackDiscard
		}
	}
}
