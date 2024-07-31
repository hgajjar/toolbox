package cmd

import (
	"queue-worker/queue"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	queueNamesKey = "worker.queues"
)

var QueueWorkerCmd = &cobra.Command{
	Use: "queue:worker",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := amqp.Dial(viper.GetString(argRabbitmqConnString))
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		failOnError(err, "Failed to open a rabbitmq channel")
		defer ch.Close()

		queues := viper.GetStringSlice(queueNamesKey)

		worker := queue.NewWorker(ch, queues)
		worker.Execute()
	},
}
