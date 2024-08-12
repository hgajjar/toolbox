package cmd

import (
	"context"
	"queue-worker/queue"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	queueNamesKey = "worker.queues"

	argDaemonMode      = "daemon-mode"
	argDaemonModeShort = "d"
	argDaemonModeUsage = `Keep queue workers running in daemon mode.`
)

var (
	daemonModeOpt bool
)

type QueueWorkerCmd struct {
	cmd *cobra.Command
}

func (s *QueueWorkerCmd) Cmd() *cobra.Command {
	return s.cmd
}

func NewQueueWorkerCmd() *QueueWorkerCmd {
	queueWorkerCmd.PersistentFlags().BoolVarP(&daemonModeOpt, argDaemonMode, argDaemonModeShort, false, argDaemonModeUsage)

	return &QueueWorkerCmd{
		cmd: queueWorkerCmd,
	}
}

var queueWorkerCmd = &cobra.Command{
	Use: "queue:worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := amqp.Dial(viper.GetString(argRabbitmqConnString))
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		queues := viper.GetStringSlice(queueNamesKey)

		worker := queue.NewWorker(conn, queues, daemonModeOpt)
		worker.Execute(ctx)
	},
}
