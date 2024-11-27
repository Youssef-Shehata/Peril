package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)


func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(wr gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
        wo,_,_ :=gs.HandleWar(wr)
        switch wo{

	case gamelogic.WarOutcomeNoUnits:
        return pubsub.NackDiscard

	case gamelogic.WarOutcomeNotInvolved:
        return pubsub.NackRequeue

    case gamelogic.WarOutcomeDraw:
        return pubsub.Ack

    case gamelogic.WarOutcomeYouWon:
        return pubsub.Ack

    case gamelogic.WarOutcomeOpponentWon:
        return pubsub.Ack

	default:
        fmt.Println("unknown war outcome")
        return pubsub.NackDiscard
	}
	}
}
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}
func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(mv gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
            return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
            if err := pubsub.PublishJSON(ch,
            routing.ExchangePerilTopic ,
            fmt.Sprintf("%s.%s",routing.WarRecognitionsPrefix,gs.Player.Username) ,
            gamelogic.RecognitionOfWar{
            	Attacker: mv.Player,
            	Defender: gs.GetPlayerSnap(),
            });err != nil{

                fmt.Printf("declaring war failed, messanger was decapitated by enemy forces on sight. %e",err)
                return pubsub.NackRequeue

            }

            return pubsub.Ack

		default:
            return pubsub.NackDiscard
		}

	}
}
func main() {
	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("failed to dial rabbitmq")
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("failed to create a channel")
	}

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("hello %v\n", userName)

	state := gamelogic.NewGameState(userName)

	//sub to pausing
	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		1,
		handlerPause(state))
	if err != nil {
		log.Fatal(err)
	}

	//sub to moves
	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		1,
		handlerMove(state , ch))
	if err != nil {
		log.Fatal(err)
	}

	//sub to war
	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
        routing.WarRecognitionsPrefix,
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		0,
		handlerWar(state))
	if err != nil {
		log.Fatal(err)
	}


    //game
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := state.CommandSpawn(words)
			if err != nil {
				fmt.Printf("%v\n", err)
			}

		case "move":
			mv, err := state.CommandMove(words)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName), mv)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}
			fmt.Println("Unit moved successfully")

		case "status":
			state.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed")

		case "quit":
			fmt.Println("Closing connection")
			return

		default:
			fmt.Println("wtf was that ")
		}
	}
}
