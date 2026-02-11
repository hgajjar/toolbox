package cmd

import (
	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
	"github.com/hgajjar/toolbox/queue"

	"github.com/spf13/cobra"
)

const (
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
		cfg := config.New()

		workerArgs := queue.WorkerArgs{
			DaemonMode: daemonModeOpt,
		}

		dic := container.New()
		defer dic.Close()

		queue.StartWorker(cmd.Context(), dic, cfg, workerArgs)
	},
}
