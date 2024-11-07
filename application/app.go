package application

import (
	"fmt"
	"toolbox/application/cmd"
	"toolbox/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Toolbox struct {
	cmd *cobra.Command
}

func New() *Toolbox {
	rootCmd := &cobra.Command{
		Use: "toolbox",
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&config.CfgFile, "config", "c", "", "config file (default is config.yaml in current dir)")
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(cmd.NewSyncDataCmd().Cmd())
	rootCmd.AddCommand(cmd.NewQueueWorkerCmd().Cmd())
	// rootCmd.AddCommand(cmd.NewInstallCmd().Cmd())

	return &Toolbox{
		cmd: rootCmd,
	}
}

func (a *Toolbox) Execute() {
	cobra.CheckErr(a.cmd.Execute())
}

func initConfig() {
	if config.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(config.CfgFile)
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
