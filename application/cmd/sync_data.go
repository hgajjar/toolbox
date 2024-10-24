package cmd

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"
	"toolbox/config"
	syncData "toolbox/data/sync"
	"toolbox/queue"
	"toolbox/sync"
	"toolbox/sync/plugin"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	argResource      = "resource"
	argResourceShort = "r"
	argResourceUsage = `Defines which resource(s) should be exported, if there is more than one, use comma to separate them.
If not, full export will be executed.
	`

	argIds      = "ids"
	argIdsShort = "i"
	argIdsUsage = `Defines ids for entities which should be exported, if there is more than one, use comma to separate them.
If not, full export will be executed.`

	argRunQueueWorker      = "run-queue-worker"
	argRunQueueWorkerShort = "q"
	argRunQueueWorkerUsage = `Run queue workers in the background.`

	syncDataEntitiesKey = "sync-data.entities"
)

var (
	idsOpt            string
	runQueueWorkerOpt bool
)

type SyncDataCmd struct {
	cmd *cobra.Command
}

func NewSyncDataCmd() *SyncDataCmd {
	syncDataCmd.PersistentFlags().StringP(argResource, argResourceShort, "", argResourceUsage)
	viper.BindPFlag(argResource, syncDataCmd.PersistentFlags().Lookup(argResource))

	syncDataCmd.PersistentFlags().StringVarP(&idsOpt, argIds, argIdsShort, "", argIdsUsage)
	syncDataCmd.PersistentFlags().BoolVarP(&runQueueWorkerOpt, argRunQueueWorker, argRunQueueWorkerShort, false, argRunQueueWorkerUsage)

	return &SyncDataCmd{
		cmd: syncDataCmd,
	}
}

func (s *SyncDataCmd) Cmd() *cobra.Command {
	return s.cmd
}

var syncDataCmd = &cobra.Command{
	Use: "sync:data",
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if config.Verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

		// Attach the Logger to the context.Context
		ctx := logger.WithContext(cmd.Context())

		conn, err := amqp.Dial(viper.GetString(argRabbitmqConnString))
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		dbconn, err := sql.Open("postgres", viper.GetString(argPostgresConnString))
		failOnError(err, "Failed to connect to Postgres")
		defer dbconn.Close()

		var workerDoneCh <-chan any
		var queueWorker *queue.Worker
		if runQueueWorkerOpt {
			workerDoneCh, queueWorker = startQueueWorker(ctx, conn, true)
		}

		resourceFilter := viper.GetString(argResource)

		exporter := sync.NewExporter(conn, getSyncDataPlugins(dbconn, resourceFilter))
		err = exporter.Export(ctx, getIDs(idsOpt))
		failOnError(err, "Failed to export data to rabbitmq")

		if runQueueWorkerOpt {
			queueWorker.SetDaemonMode(false)
			<-workerDoneCh
		}
	},
}

func startQueueWorker(ctx context.Context, conn *amqp.Connection, daemonMode bool) (<-chan any, *queue.Worker) {
	done := make(chan any)
	queues := viper.GetStringSlice(queueNamesKey)
	worker := queue.NewWorker(conn, queues, daemonMode)

	go func(ctx context.Context, worker *queue.Worker) {
		worker.Execute(ctx)
		done <- struct{}{}
	}(ctx, worker)

	return done, worker
}

func getSyncDataPlugins(dbconn *sql.DB, resourceFilter string) []sync.SyncDataPluginInterface {
	allPlugins := generateSyncDataPlugins(dbconn)

	if resourceFilter == "" {
		return allPlugins
	}

	var filteredPlugins []sync.SyncDataPluginInterface
	for _, resourceName := range strings.Split(resourceFilter, ",") {
		for _, plugin := range allPlugins {
			if plugin.GetResourceName() == resourceName {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		}
	}

	return filteredPlugins
}

func generateSyncDataPlugins(dbconn *sql.DB) []sync.SyncDataPluginInterface {
	var syncConfigEntities []config.SyncEntity

	err := viper.UnmarshalKey(syncDataEntitiesKey, &syncConfigEntities)
	failOnError(err, "Failed to parse sync-data.entities config")

	var plugins []sync.SyncDataPluginInterface
	for _, syncConfigEntity := range syncConfigEntities {
		plugins = append(plugins, plugin.New(syncData.NewRepository(dbconn, &syncConfigEntity), &syncConfigEntity))
	}

	return plugins
}

func getIDs(idsFilter string) []int {
	var IDs []int
	if idsFilter == "" {
		return IDs
	}
	for _, id := range strings.Split(idsFilter, ",") {
		intId, err := strconv.Atoi(id)
		failOnError(err, "Failed to parse 'ids' filter, it must be an integer number")

		IDs = append(IDs, intId)
	}

	return IDs
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
