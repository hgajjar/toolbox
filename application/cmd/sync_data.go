package cmd

import (
	"context"
	"database/sql"
	"log"
	"queue-worker/config"
	syncData "queue-worker/data/sync"
	"queue-worker/sync"
	"queue-worker/sync/plugin"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
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

	syncDataEntitiesKey = "sync-data.entities"
)

type SyncDataCmd struct {
	cmd *cobra.Command
}

func NewSyncDataCmd() *SyncDataCmd {
	syncDataCmd.PersistentFlags().StringP(argResource, argResourceShort, "", argResourceUsage)
	viper.BindPFlag(argResource, syncDataCmd.PersistentFlags().Lookup(argResource))

	syncDataCmd.PersistentFlags().StringP(argIds, argIdsShort, "", argIdsUsage)
	viper.BindPFlag(argIds, syncDataCmd.PersistentFlags().Lookup(argIds))

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
		ctx := context.Background()

		conn, err := amqp.Dial(viper.GetString(argRabbitmqConnString))
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		dbconn, err := sql.Open("postgres", viper.GetString(argPostgresConnString))
		failOnError(err, "Failed to connect to Postgres")
		defer dbconn.Close()

		resourceFilter := viper.GetString(argResource)
		idsFilter := viper.GetString(argIds)

		exporter := sync.NewExporter(conn, getSyncDataPlugins(dbconn, resourceFilter))
		err = exporter.Export(ctx, getIDs(idsFilter))
		failOnError(err, "Failed to export data to rabbitmq")
	},
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
