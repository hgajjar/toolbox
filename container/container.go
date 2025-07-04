package container

import (
	"io"
	"log"

	"github.com/hgajjar/toolbox/config"
	"github.com/rs/zerolog"
)

type Container struct {
	logger *zerolog.Logger
}

func New() *Container {
	return &Container{}
}

func (c *Container) Logger(bypassWriter io.Writer) *zerolog.Logger {
	if c.logger == nil {
		level := zerolog.ErrorLevel

		switch config.Verbose {
		case 1:
			level = zerolog.ErrorLevel
		case 2:
			level = zerolog.InfoLevel
		case 3:
			level = zerolog.DebugLevel
		default:
			// Handle unexpected verbosity levels
			log.Fatal("Invalid verbosity level.")
		}

		logger := zerolog.New(zerolog.ConsoleWriter{Out: bypassWriter}).Level(level).With().Timestamp().Logger()
		c.logger = &logger
	}

	return c.logger
}
