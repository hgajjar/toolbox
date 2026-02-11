package cmd

import (
	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
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
		cfg := config.New()

		syncDataArgs := sync.SyncDataArgs{
			IDsOpt:            idsOpt,
			RunQueueWorkerOpt: runQueueWorkerOpt,
		}

		dic := container.New()
		defer dic.Close()

		sync.RunSyncData(
			cmd.Context(),
			dic,
			cfg,
			syncDataArgs,
		)
	},
}
