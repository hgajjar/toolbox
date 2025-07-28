package container

import (
	"io"
	"log"

	"github.com/Adaendra/uilive"
	"github.com/Adaendra/uilive/pkg/writer"
	"github.com/hgajjar/toolbox/config"
	"github.com/rs/zerolog"
)

type Container struct {
	writer *writer.Writer
	logger *zerolog.Logger
}

func New() *Container {
	return &Container{}
}

func (c *Container) Writer() (io.Writer, func()) {
	if c.writer == nil {
		c.writer = uilive.New()
		c.writer.Start()
	}

	return c.writer, func() { c.writer.Stop() }
}

func (c *Container) Logger() *zerolog.Logger {
	if c.logger == nil {
		level := zerolog.ErrorLevel

		switch config.Verbose {
		case 0:
			level = zerolog.Disabled
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

		if c.writer == nil {
			log.Fatal("Writer must be initialized before Logger.")
		}

		logger := zerolog.New(zerolog.ConsoleWriter{Out: c.writer.Bypass()}).Level(level).With().Timestamp().Logger()
		c.logger = &logger
	}

	return c.logger
}
