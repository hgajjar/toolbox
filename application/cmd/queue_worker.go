package cmd

import (
	"strings"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
	"github.com/hgajjar/toolbox/queue"

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
		queues := viper.GetStringSlice(queueNamesKey)

		cmdPrefix := strings.Split(viper.GetString(config.ConsoleCmdPrefixKey), " ")
		cmdDir := viper.GetString(config.ConsoleCmdDirKey)
		consoleCmd := strings.Split(viper.GetString(config.ConsoleCmdKey), " ")

		workerArgs := queue.WorkerArgs{
			RabbitmqConnString: config.GetRabbitMQConnectionString(),
			Queues:             queues,
			DaemonMode:         daemonModeOpt,
			CmdPrefix:          cmdPrefix,
			CmdDir:             cmdDir,
			Cmd:                consoleCmd,
		}

		dic := container.New()

		queue.StartWorker(cmd.Context(), dic, workerArgs)
	},
}
