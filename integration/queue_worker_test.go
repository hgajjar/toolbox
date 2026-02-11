//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/queue"
)

const (
	queueChunkSize = "100"
)

func TestQueueWorker(t *testing.T) {
	ctx := context.Background()

	deferFn, hostPort, err := setupRabbitMqService(ctx)
	defer deferFn()
	if err != nil {
		t.Fatal(err)
	}

	config.Verbose = 1

	rmq, err := setupRabbitMqConnection(hostPort)
	if err != nil {
		t.Fatal(err)
	}
	defer rmq.Close()

	queues := []string{"test.product", "test.category", "test.user"}
	if err := defineQueues(queues, rmq); err != nil {
		t.Fatal(err)
	}

	logger := io.Discard

	t.Run("It consumes all messages from the queue and exits", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			rmq.Publish([]byte(fmt.Sprintf("product-%d", i)), "test.product")
			rmq.Publish([]byte(fmt.Sprintf("category-%d", i)), "test.category")
			rmq.Publish([]byte(fmt.Sprintf("user-%d", i)), "test.user")
		}

		worker := queue.NewWorker(rmq.Connection(), queues, false, []string{}, "rabbitmq/consumer", []string{"./consumer", hostPort, queueChunkSize}, logger, config.QueueDeclareRetryWait)
		worker.Execute(ctx)

		for _, queue := range queues {
			count, err := rmq.GetMessageCount(queue)
			if err != nil {
				t.Fatal(err)
			}

			if count != 0 {
				t.Fatalf("expected 0 messages in %s queue, got %d", queue, count)
			}
		}
	})

	t.Run("It waits for queues to be available and can process messages that came later in daemon mode", func(t *testing.T) {
		newQueues := []string{"test.product.new", "test.category.new", "test.user.new"}
		worker := queue.NewWorker(rmq.Connection(), newQueues, true, []string{}, "rabbitmq/consumer", []string{"./consumer", hostPort, queueChunkSize}, logger, 2*time.Second)
		go worker.Execute(ctx)

		time.Sleep(3 * time.Second)

		for _, queue := range newQueues {
			if err := rmq.Queue(queue); err != nil {
				t.Fatal(err)
			}
		}

		for i := 0; i < 10; i++ {
			rmq.Publish([]byte(fmt.Sprintf("product-%d", i)), "test.product.new")
			rmq.Publish([]byte(fmt.Sprintf("category-%d", i)), "test.category.new")
			rmq.Publish([]byte(fmt.Sprintf("user-%d", i)), "test.user.new")
		}

		for _, queue := range newQueues {
		loop:
			for {
				select {
				case <-time.After(5 * time.Second):
					t.Fatalf("timed-out while waiting for queue %s to get empty", queue)
				default:
					count, err := rmq.GetMessageCount(queue)
					if err != nil {
						t.Fatal(err)
					}
					if count == 0 {
						break loop
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	})

	t.Run("It consumes all existing and new messages from the queue and keeps running in daemon mode", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			rmq.Publish([]byte(fmt.Sprintf("product-%d", i)), "test.product")
			rmq.Publish([]byte(fmt.Sprintf("category-%d", i)), "test.category")
			rmq.Publish([]byte(fmt.Sprintf("user-%d", i)), "test.user")
		}

		worker := queue.NewWorker(rmq.Connection(), queues, true, []string{}, "rabbitmq/consumer", []string{"./consumer", hostPort, queueChunkSize}, logger, config.QueueDeclareRetryWait)
		go worker.Execute(ctx)

		time.Sleep(3 * time.Second)

		for i := 0; i < 1000; i++ {
			rmq.Publish([]byte(fmt.Sprintf("product-%d", i)), "test.product")
			rmq.Publish([]byte(fmt.Sprintf("category-%d", i)), "test.category")
			rmq.Publish([]byte(fmt.Sprintf("user-%d", i)), "test.user")
		}

		for _, queue := range queues {
		loop:
			for {
				select {
				case <-time.After(5 * time.Second):
					t.Fatalf("timed-out while waiting for queue %s to get empty", queue)
				default:
					count, err := rmq.GetMessageCount(queue)
					if err != nil {
						t.Fatal(err)
					}
					if count == 0 {
						break loop
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}
	})
}
