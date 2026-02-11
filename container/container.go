package container

import (
	"database/sql"
	"io"
	"log"

	_ "github.com/lib/pq"

	"github.com/Adaendra/uilive"
	"github.com/Adaendra/uilive/pkg/writer"
	"github.com/hgajjar/toolbox/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type Container struct {
	writer *writer.Writer
	logger *zerolog.Logger
	db     *sql.DB
	queue  *amqp.Connection
}

func New() *Container {
	return &Container{}
}

func (c *Container) Writer() io.Writer {
	if c.writer == nil {
		c.writer = uilive.New()
		c.writer.Start()
	}

	return c.writer
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

func (c *Container) DB(connStr string) *sql.DB {
	if c.db == nil {
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			c.Logger().Panic().Err(err).Msg("Failed to connect to Postgres")
		}
		c.db = db
	}

	return c.db
}

func (c *Container) Queue(connStr string) *amqp.Connection {
	if c.queue == nil {
		conn, err := amqp.Dial(connStr)
		if err != nil {
			c.Logger().Panic().Err(err).Msg("Failed to connect to RabbitMQ")
		}
		c.queue = conn
	}

	return c.queue
}

func (c *Container) Close() {
	if c.db != nil {
		err := c.db.Close()
		if err != nil {
			c.Logger().Error().Err(err).Msg("Failed to close Postgres connection")
		}
	}
	if c.queue != nil {
		err := c.queue.Close()
		if err != nil {
			c.Logger().Error().Err(err).Msg("Failed to close RabbitMQ connection")
		}
	}
	if c.writer != nil {
		c.writer.Stop()
	}
}
