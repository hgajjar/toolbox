package queue

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type connection struct {
	args WorkerArgs
}

func NewConnection(args WorkerArgs) *connection {
	return &connection{args: args}
}

func (c *connection) Setup(ctx context.Context, args WorkerArgs) (*amqp.Connection, error) {
	logger := zerolog.Ctx(ctx)
	conn, err := amqp.Dial(args.RabbitmqConnString)
	if err != nil {
		if !args.DaemonMode {
			return nil, errors.New("failed to connect to RabbitMQ")
		}

		// In daemon mode, we wait and retry connection with exponential backoff
		logger.Error().Err(err).Msg("failed to connect to RabbitMQ, retrying...")
		conn, err = c.waitAndRetry(ctx, args.RabbitmqConnString, logger)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (c *connection) waitAndRetry(ctx context.Context, connString string, logger *zerolog.Logger) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	backoff := time.Second

	for {
		conn, err = amqp.Dial(connString)
		if err == nil {
			logger.Info().Msg("successfully connected to RabbitMQ")
			return conn, nil
		}

		logger.Error().Err(err).Msgf("retrying connection in %v seconds...", backoff)

		select {
		case <-ctx.Done():
			return nil, errors.New("context cancelled, stopping retry attempts")
		case <-time.After(c.args.connectionTimeout):
			return nil, errors.New("connection timeout reached, stopping retry attempts")
		case <-time.After(backoff):
			backoff *= 2
		}
	}
}
