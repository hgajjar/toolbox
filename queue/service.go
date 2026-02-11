package queue

import (
	"context"
	"time"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
)

type WorkerArgs struct {
	DaemonMode        bool
	connectionTimeout time.Duration
}

func StartWorker(ctx context.Context, dic *container.Container, cfg *config.Config, args WorkerArgs) {
	writer := dic.Writer()
	logger := dic.Logger()

	// Attach the Logger to the context.Context
	ctx = logger.WithContext(ctx)

	args.connectionTimeout = time.Hour
	conn, err := NewConnection(args).Setup(ctx, cfg, args)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to establish RabbitMQ connection")
	}
	defer conn.Close()

	worker := NewWorker(conn, cfg.QueueNames, args.DaemonMode, cfg.ConsoleCmdPrefix, cfg.ConsoleCmdDir, cfg.ConsoleCmd, writer, config.QueueDeclareRetryWait)
	worker.Execute(ctx)
}
