package queue

import (
	"context"
	"os"
	"toolbox/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

func StartWorker(ctx context.Context, rabbitmqConnString string, queues []string, daemonMode bool, cmdPrefix []string, cmdDir string, cmd []string) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if config.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// Attach the Logger to the context.Context
	ctx = logger.WithContext(ctx)

	conn, err := amqp.Dial(rabbitmqConnString)
	if err != nil {
		logger.Panic().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	worker := NewWorker(conn, queues, daemonMode, cmdPrefix, cmdDir, cmd)
	worker.Execute(ctx)
}
