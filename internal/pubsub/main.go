package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack         AckType = iota // 0
	NackDiscard                // 1
	NackRequeue                // 2
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jValue, err := json.Marshal(val)

	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: jValue})

	return err
}
func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, errors.New("failed to connect to rabbitmq")
	}

	var (
		durable    bool
		autoDelete bool
		exclusive  bool
	)

	noWait := false

	if simpleQueueType == 0 {
		durable = true
		autoDelete = true
		exclusive = false
	} else {
		durable = false
		autoDelete = true
		exclusive = true
	}

    queue, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, amqp.Table{"x-dead-letter-exchange":routing.ExchangePerilDlx})
	if err != nil {
		return nil, amqp.Queue{}, errors.New("failed to declare rbmq queue")
	}

	err = ch.QueueBind(queueName, key, exchange, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, errors.New("failed to bind queue to exchange")
	}

	return ch, queue, nil
}
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) AckType,
) error {

	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	delivery, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range delivery {
			var data T
			err = json.Unmarshal(d.Body, &data)
			if err != nil {
				fmt.Printf("failed to unmarshal json : %v", err)
				return
			}
			ack := handler(data)
			switch ack {

			case Ack:
                err = d.Ack(false)
                if err != nil {
                    fmt.Printf("failed to Ack message: %v", err)
                    return
                }
			case NackRequeue:
                err = d.Nack(false,true)
                if err != nil {
                    fmt.Printf("failed to Nack message: %v", err)
                    return
                }
			case NackDiscard:
                err = d.Nack(false,false)
                if err != nil {
                    fmt.Printf("failed to Discard message: %v", err)
                    return
                }
			default:
				panic(fmt.Sprintf("unexpected pubsub.AckType: %#v", ack))
			}
		}
	}()

	return nil
}
