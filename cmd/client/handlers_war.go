package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func handlerWar(gs *gamelogic.GameState, pubCh *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(w gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(w)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			if err := publishGameLog(pubCh, w.Attacker.Username, msg); err != nil {
				fmt.Println("Failed to publish game log:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			if err := publishGameLog(pubCh, w.Attacker.Username, msg); err != nil {
				fmt.Println("Failed to publish game log:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		default:
			fmt.Println("Unknown war outcome:", outcome)
			return pubsub.NackDiscard
		}
	}
}
