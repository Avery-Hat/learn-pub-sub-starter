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
	fmt.Println("Starting Peril server...")

	// Show available REPL commands
	gamelogic.PrintServerHelp()

	// Connect to RabbitMQ
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Successfully connected to RabbitMQ")

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open RabbitMQ channel:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// REPL loop
	for {
		words := gamelogic.GetInput()

		// GetInput returns nil when input fails (EOF, Ctrl+D, etc.)
		if words == nil {
			fmt.Println("Input closed, exiting...")
			return
		}

		// User just hit enter
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message...")
			state := routing.PlayingState{IsPaused: true}
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, state); err != nil {
				fmt.Println("Failed to publish pause message:", err)
			}

		case "resume":
			fmt.Println("Sending resume message...")
			state := routing.PlayingState{IsPaused: false}
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, state); err != nil {
				fmt.Println("Failed to publish resume message:", err)
			}

		case "quit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("I don't understand that command.")
		}
	}
}
