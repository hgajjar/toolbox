package queue

import (
	"context"

	"github.com/hgajjar/toolbox/container"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WorkerArgs struct {
	RabbitmqConnString string
	Queues             []string
	DaemonMode         bool
	CmdPrefix          []string
	CmdDir             string
	Cmd                []string
}

func StartWorker(ctx context.Context, dic *container.Container, args WorkerArgs) {
	writer, stopFunc := dic.Writer()
	defer stopFunc()

	logger := dic.Logger()

	// Attach the Logger to the context.Context
	ctx = logger.WithContext(ctx)

	conn, err := amqp.Dial(args.RabbitmqConnString)
	if err != nil {
		logger.Panic().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	worker := NewWorker(conn, args.Queues, args.DaemonMode, args.CmdPrefix, args.CmdDir, args.Cmd, writer)
	worker.Execute(ctx)
}
