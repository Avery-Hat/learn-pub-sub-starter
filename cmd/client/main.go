package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	// Connect to RabbitMQ
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Successfully connected to RabbitMQ")

	// Create a channel for publishing (PublishJSON needs *amqp.Channel)
	pubCh, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open RabbitMQ channel:", err)
		os.Exit(1)
	}
	defer pubCh.Close()

	// Prompt for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Failed to get username:", err)
		os.Exit(1)
	}

	// Create a new game state
	gamestate := gamelogic.NewGameState(username)

	// ---- Subscribe to pause/resume messages (direct exchange) ----
	pauseQueueName := routing.PauseKey + "." + username
	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gamestate), // must return pubsub.AckType now
	); err != nil {
		fmt.Println("Failed to subscribe to pause messages:", err)
		os.Exit(1)
	}

	// ---- Subscribe to army move messages (topic exchange) ----
	armyMovesSlug := "army_moves"
	moveQueueName := armyMovesSlug + "." + username
	moveBindingKey := armyMovesSlug + ".*"

	if err := pubsub.SubscribeJSON[gamelogic.ArmyMove](
		conn,
		routing.ExchangePerilTopic,
		moveQueueName,
		moveBindingKey,
		pubsub.SimpleQueueTransient,
		handlerMove(gamestate, pubCh), // <-- CHANGED: pass pubCh, returns AckType
	); err != nil {
		fmt.Println("Failed to subscribe to move messages:", err)
		os.Exit(1)
	}

	// ---- NEW: Subscribe to war recognitions (topic exchange) ----
	// durable shared queue named "war"
	warBindingKey := routing.WarRecognitionsPrefix + ".*"
	if err := pubsub.SubscribeJSON[gamelogic.RecognitionOfWar](
		conn,
		routing.ExchangePerilTopic,
		"war",
		warBindingKey,
		pubsub.SimpleQueueDurable,
		handlerWar(gamestate, pubCh),
	); err != nil {
		fmt.Println("Failed to subscribe to war messages:", err)
		os.Exit(1)
	}

	// Print available client commands
	gamelogic.PrintClientHelp()

	// Client REPL loop
	for {
		words := gamelogic.GetInput()
		if words == nil {
			gamelogic.PrintQuit()
			return
		}
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			if err := gamestate.CommandSpawn(words); err != nil {
				fmt.Println("Error:", err)
			}

		case "move":
			// CommandMove updates local state and returns the ArmyMove event payload
			mv, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			// Publish move to army_moves.<username> on the topic exchange
			moveRoutingKey := armyMovesSlug + "." + username
			if err := pubsub.PublishJSON(
				pubCh,
				routing.ExchangePerilTopic,
				moveRoutingKey,
				mv,
			); err != nil {
				fmt.Println("Failed to publish move:", err)
				continue
			}

			fmt.Println("Move published successfully!")

		case "status":
			gamestate.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) < 2 {
				fmt.Println("usage: spam <n>")
				break
			}

			n, err := strconv.Atoi(words[1])
			if err != nil || n < 1 {
				fmt.Println("usage: spam <n> (n must be a positive integer)")
				break
			}

			key := routing.GameLogSlug + "." + username

			for i := 0; i < n; i++ {
				msg := gamelogic.GetMaliciousLog()

				gl := routing.GameLog{
					CurrentTime: time.Now(),
					Message:     msg,
					Username:    username,
				}

				if err := pubsub.PublishGob(pubCh, routing.ExchangePerilTopic, key, gl); err != nil {
					fmt.Println("Failed to publish spam log:", err)
					break
				}
			}

			fmt.Printf("Published %d log(s)\n", n)

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("I don't understand that command.")
		}
	}
}
