package main

import (
	"fmt"
	"os"

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

	// Prompt for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Failed to get username:", err)
		os.Exit(1)
	}

	// Declare + bind a transient pause queue: pause.<username>
	queueName := routing.PauseKey + "." + username
	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		fmt.Println("Failed to declare/bind queue:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// Create a new game state
	gamestate := gamelogic.NewGameState(username)

	// Print available client commands
	gamelogic.PrintClientHelp()

	// Client REPL loop
	for {
		words := gamelogic.GetInput()
		if words == nil {
			// Input failed (EOF / stdin closed)
			gamelogic.PrintQuit()
			return
		}
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			// spawn <location> <unit_type>
			// e.g. spawn europe infantry
			if err := gamestate.CommandSpawn(words); err != nil {
				fmt.Println("Error:", err)
			}
			// CommandSpawn should print the new unit ID itself (per assignment expectation)

		case "move":
			// move <destination> <unit_id>
			// e.g. move europe 1
			_, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Println("Move successful!")

		case "status":
			gamestate.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("I don't understand that command.")
		}
	}
}
