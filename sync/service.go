package sync

import (
	"context"
	"database/sql"
	"io"
	"strconv"
	"strings"

	"github.com/Adaendra/uilive"
	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
	syncData "github.com/hgajjar/toolbox/data/sync"
	"github.com/hgajjar/toolbox/queue"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type SyncDataArgs struct {
	RunQueueWorkerOpt  bool
	Queues             []string
	CmdPrefix          []string
	CmdDir             string
	Cmd                []string
	SyncDataEntities   []config.SyncEntity
	RabbitmqConnString string
	PostgresConnString string
	ResourceFilter     string
	IDsOpt             string
}

func RunSyncData(ctx context.Context, dic *container.Container, args SyncDataArgs) {
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	logger := dic.Logger(writer.Bypass())

	// Attach the Logger to the context.Context
	ctx = logger.WithContext(ctx)

	conn, err := amqp.Dial(args.RabbitmqConnString)
	if err != nil {
		logger.Panic().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	dbconn, err := sql.Open("postgres", args.PostgresConnString)
	if err != nil {
		logger.Panic().Err(err).Msg("Failed to connect to Postgres")
	}
	defer dbconn.Close()

	var workerDoneCh <-chan any
	var queueWorker *queue.Worker
	if args.RunQueueWorkerOpt {
		workerDoneCh, queueWorker = startQueueWorker(ctx, args.Queues, conn, true, args.CmdPrefix, args.CmdDir, args.Cmd, writer)
	}

	exporter := NewExporter(conn, getSyncDataPlugins(dbconn, args.ResourceFilter, args.SyncDataEntities))
	err = exporter.Export(ctx, getIDs(ctx, args.IDsOpt))
	if err != nil {
		logger.Panic().Err(err).Msg("Failed to export data to rabbitmq")
	}

	if args.RunQueueWorkerOpt {
		queueWorker.SetDaemonMode(false)
		<-workerDoneCh
	}
}

func startQueueWorker(ctx context.Context, queues []string, conn *amqp.Connection, daemonMode bool, cmdPrefix []string, cmdDir string, cmd []string, writer io.Writer) (<-chan any, *queue.Worker) {
	done := make(chan any)
	worker := queue.NewWorker(conn, queues, daemonMode, cmdPrefix, cmdDir, cmd, writer)

	go func(ctx context.Context, worker *queue.Worker) {
		worker.Execute(ctx)
		done <- struct{}{}
	}(ctx, worker)

	return done, worker
}

func getSyncDataPlugins(dbconn *sql.DB, resourceFilter string, syncConfigEntities []config.SyncEntity) []SyncDataPluginInterface {
	allPlugins := buildSyncDataPlugins(dbconn, syncConfigEntities)

	if resourceFilter == "" {
		return allPlugins
	}

	var filteredPlugins []SyncDataPluginInterface
	for _, resourceName := range strings.Split(resourceFilter, ",") {
		for _, plugin := range allPlugins {
			if plugin.GetResourceName() == resourceName {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		}
	}

	return filteredPlugins
}

func buildSyncDataPlugins(dbconn *sql.DB, syncConfigEntities []config.SyncEntity) []SyncDataPluginInterface {
	var plugins []SyncDataPluginInterface

	for _, syncConfigEntity := range syncConfigEntities {
		plugins = append(plugins, NewPlugin(syncData.NewRepository(dbconn, &syncConfigEntity), &syncConfigEntity))
	}

	return plugins
}

func getIDs(ctx context.Context, idsFilter string) []int {
	var IDs []int
	if idsFilter == "" {
		return IDs
	}
	for _, id := range strings.Split(idsFilter, ",") {
		intId, err := strconv.Atoi(id)
		if err != nil {
			zerolog.Ctx(ctx).Panic().Stack().Err(err).Msg("Failed to parse 'ids' filter, it must be an integer number")
		}

		IDs = append(IDs, intId)
	}

	return IDs
}
