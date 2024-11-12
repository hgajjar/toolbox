//go:build integration
// +build integration

package rabbitmq

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config parameters for queue service
type Config struct {
	URL        string
	RoutingKey string
	BindingKey string
}

// Rabbitmq implements the Queue interface
type Rabbitmq struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  Config
}

// New instance of Queue
func New(config Config) (*Rabbitmq, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Rabbitmq{
		conn:    conn,
		channel: ch,
		config:  config,
	}, nil
}

func (r *Rabbitmq) Connection() *amqp.Connection {
	return r.conn
}

func (r *Rabbitmq) Queue(name string) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		name,    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return ch.QueueBind(
		q.Name,              // queue name
		r.config.BindingKey, // routing key
		name,                // exchange
		false,
		nil,
	)
}

// Publish message to rabbitmq
func (r *Rabbitmq) Publish(message []byte, exchange string) error {
	return r.channel.Publish(
		exchange,            // exchange
		r.config.RoutingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			Body: message,
		},
	)
}

// Consume message from rabbitmq
func (r *Rabbitmq) Consume(queue string, size int) (<-chan []byte, error) {
	err := r.channel.Qos(size, 0, false)
	if err != nil {
		return nil, err
	}

	msgs, err := r.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	deliveries := make(chan []byte)
	go func() {
		var i int

		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					close(deliveries)
					return
				}
				deliveries <- msg.Body
				msg.Ack(true)
				i++

				if i == size {
					close(deliveries)
					return
				}
			case <-time.After(2 * time.Second):
				close(deliveries)
				return
			}
		}
	}()
	return (<-chan []byte)(deliveries), nil
}

func (r *Rabbitmq) GetMessageCount(queue string) (int, error) {
	q, err := r.channel.QueueDeclarePassive(queue, false, false, false, false, nil)
	if err != nil {
		return 0, err
	}
	return q.Messages, nil
}

// close rabbitmq connection
func (r *Rabbitmq) Close() error {
	if err := r.conn.Close(); err != nil {
		return err
	}
	if err := r.channel.Close(); err != nil {
		return err
	}
	return nil
}
