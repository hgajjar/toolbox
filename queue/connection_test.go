package queue

import (
	"context"
	"testing"
	"time"

	"github.com/hgajjar/toolbox/config"
)

func TestConnection_New(t *testing.T) {
	args := WorkerArgs{}
	conn := NewConnection(args)
	if conn == nil {
		t.Fatal("Expected NewConnection to return a non-nil connection")
	}
}

func TestConnection_Setup(t *testing.T) {
	t.Run("It should return an error when RabbitMQ is not running", func(t *testing.T) {
		t.Parallel()

		args := WorkerArgs{
			DaemonMode: false,
		}
		cfg := &config.Config{
			RabbitMq: config.RabbitMQ{
				ConnectionString: "amqp://guest:guest@invalid-host:5672/",
			},
		}
		conn := NewConnection(args)
		ctx := context.Background()

		_, err := conn.Setup(ctx, cfg, args)
		if err == nil {
			t.Fatal("Expected Setup to return an error when RabbitMQ is not running")
		}
	})

	t.Run("It should retry connection in DaemonMode", func(t *testing.T) {
		t.Parallel()

		args := WorkerArgs{
			DaemonMode:        true,
			connectionTimeout: time.Second * 3,
		}
		cfg := &config.Config{
			RabbitMq: config.RabbitMQ{
				ConnectionString: "amqp://guest:guest@invalid-host:5672/",
			},
		}
		conn := NewConnection(args)
		ctx := context.Background()
		start := time.Now()

		_, err := conn.Setup(ctx, cfg, args)
		if err == nil {
			t.Fatal("Expected Setup to return an error after retrying in DaemonMode when RabbitMQ is not running")
		}

		// check if it waited at least the timeout duration
		elapsed := time.Since(start)
		if elapsed < args.connectionTimeout {
			t.Fatalf("Expected Setup to wait at least %v, but it waited %v", args.connectionTimeout, elapsed)
		}
	})
}
