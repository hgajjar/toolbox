package cmd

import (
	"os"

	"github.com/hgajjar/toolbox/recipe"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	recipeFile string
)

type InstallCmd struct {
	cmd *cobra.Command
}

func (s *InstallCmd) Cmd() *cobra.Command {
	return s.cmd
}

func NewInstallCmd() *InstallCmd {
	installCmd.PersistentFlags().StringVarP(&recipeFile, "recipe", "r", "", `Name of the recipe you want to use for install. [default: "development"]`)

	return &InstallCmd{
		cmd: installCmd,
	}
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run install for a specified environment.",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

		// Attach the Logger to the context.Context
		ctx := logger.WithContext(cmd.Context())

		installer := recipe.NewInstaller(recipeFile)
		installer.Install(ctx)
	},
}
