package cmd

import (
	"log"
	"strings"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/sync"

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
		rabbitmqConnStr := viper.GetString(argRabbitmqConnString)
		postgresConnStr := viper.GetString(argPostgresConnString)
		queues := viper.GetStringSlice(queueNamesKey)

		cmdPrefix := strings.Split(viper.GetString(config.ConsoleCmdPrefixKey), " ")
		cmdDir := viper.GetString(config.ConsoleCmdDirKey)
		consoleCmd := strings.Split(viper.GetString(config.ConsoleCmdKey), " ")

		resourceFilter := viper.GetString(argResource)

		var syncConfigEntities []config.SyncEntity

		err := viper.UnmarshalKey(syncDataEntitiesKey, &syncConfigEntities)
		if err != nil {
			log.Panicf("Failed to parse sync-data.entities config: %s", err)
		}

		sync.RunSyncData(cmd.Context(), runQueueWorkerOpt, queues, cmdPrefix, cmdDir, consoleCmd, syncConfigEntities, rabbitmqConnStr, postgresConnStr, resourceFilter, idsOpt)
	},
}
