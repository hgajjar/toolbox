package application

import (
	"bytes"
	"fmt"

	"github.com/hgajjar/toolbox/application/cmd"
	"github.com/hgajjar/toolbox/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "embed"
)

var defaultConfig []byte

type Toolbox struct {
	cmd *cobra.Command
}

func New(defaultConf []byte) *Toolbox {
	rootCmd := &cobra.Command{
		Use: "toolbox",
	}
	defaultConfig = defaultConf

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&config.CfgFile, "config", "c", "", "config file (default is toolbox.yml in current dir)")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity")

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
		if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
			cobra.CheckErr(err)
		}
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())
}
