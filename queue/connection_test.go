package queue

import (
	"context"
	"testing"
	"time"
)

func TestConnection_New(t *testing.T) {
	args := WorkerArgs{
		RabbitmqConnString: "amqp://guest:guest@localhost:5672/",
	}
	conn := NewConnection(args)
	if conn == nil {
		t.Fatal("Expected NewConnection to return a non-nil connection")
	}

	if conn.args.RabbitmqConnString != args.RabbitmqConnString {
		t.Fatalf("Expected RabbitmqConnString to be %s, got %s", args.RabbitmqConnString, conn.args.RabbitmqConnString)
	}
}

func TestConnection_Setup(t *testing.T) {
	t.Run("It should return an error when RabbitMQ is not running", func(t *testing.T) {
		t.Parallel()

		args := WorkerArgs{
			RabbitmqConnString: "amqp://guest:guest@invalid-host:5672/",
			DaemonMode:         false,
		}
		conn := NewConnection(args)
		ctx := context.Background()

		_, err := conn.Setup(ctx, args)
		if err == nil {
			t.Fatal("Expected Setup to return an error when RabbitMQ is not running")
		}
	})

	t.Run("It should retry connection in DaemonMode", func(t *testing.T) {
		t.Parallel()

		args := WorkerArgs{
			RabbitmqConnString: "amqp://guest:guest@invalid-host:5672/",
			DaemonMode:         true,
			connectionTimeout:  time.Second * 3,
		}
		conn := NewConnection(args)
		ctx := context.Background()
		start := time.Now()

		_, err := conn.Setup(ctx, args)
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
