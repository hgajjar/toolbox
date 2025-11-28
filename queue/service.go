package queue

import (
	"context"
	"time"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
)

type WorkerArgs struct {
	RabbitmqConnString string
	Queues             []string
	DaemonMode         bool
	CmdPrefix          []string
	CmdDir             string
	Cmd                []string
	connectionTimeout  time.Duration
}

func StartWorker(ctx context.Context, dic *container.Container, args WorkerArgs) {
	writer, stopFunc := dic.Writer()
	defer stopFunc()

	logger := dic.Logger()

	// Attach the Logger to the context.Context
	ctx = logger.WithContext(ctx)

	args.connectionTimeout = time.Hour
	conn, err := NewConnection(args).Setup(ctx, args)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to establish RabbitMQ connection")
	}
	defer conn.Close()

	worker := NewWorker(conn, args.Queues, args.DaemonMode, args.CmdPrefix, args.CmdDir, args.Cmd, writer, config.QueueDeclareRetryWait)
	worker.Execute(ctx)
}
