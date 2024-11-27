package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
        log.Fatal(fmt.Errorf("failed to dial rabbitmq : %w",err))
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
        log.Fatal(fmt.Errorf("failed to connect to rMQ : %w",err))
	}

	fmt.Println("Starting Peril server...")
	fmt.Println("successfully connected to rabbitMQ...")
	gamelogic.PrintServerHelp()

    //_,_,err = pubsub.DeclareAndBind(conn ,routing.ExchangePerilTopic,routing.GameLogSlug,"game_logs.*" , 0)
    //if err != nil {
    //    log.Fatal(fmt.Errorf("failed to create q : %w",err))
    //}



	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			fmt.Println("please provide an option")
			continue
		}

		switch words[0] {
		case "pause":

			fmt.Println("publishing a pause message")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Println(fmt.Errorf("failed to publish last message %w", err))
			}

		case "resume":

			fmt.Println("publishing a resume message")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Println(fmt.Errorf("failed to publish last message %w", err))
			}

		case "help":
			gamelogic.PrintServerHelp()

		case "quit":

			fmt.Println("Closing server...")
			return

        default:
            fmt.Println("wtf is that")
		}

	}

}
