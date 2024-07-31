package application

import (
	"fmt"
	"queue-worker/application/cmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type Toolbox struct {
	cmd *cobra.Command
}

func New() *Toolbox {
	rootCmd := &cobra.Command{
		Use: "toolbox",
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yaml in current dir)")

	rootCmd.AddCommand(cmd.NewSyncDataCmd().Cmd())
	rootCmd.AddCommand(cmd.QueueWorkerCmd)

	return &Toolbox{
		cmd: rootCmd,
	}
}

func (a *Toolbox) Execute() {
	cobra.CheckErr(a.cmd.Execute())
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "toolbox.yaml".
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("toolbox.yml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())
}
