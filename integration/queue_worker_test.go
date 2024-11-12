//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"
	"toolbox/config"
	"toolbox/integration/rabbitmq"
	"toolbox/queue"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	rmqDockerImage = "docker.io/library/rabbitmq:3.9-management-alpine"
	queueChunkSize = "100"
)

func TestQueueWorker(t *testing.T) {
	ctx := context.Background()

	deferFn, hostPort, err := setupRabbitMqService(ctx)
	defer deferFn()
	if err != nil {
		t.Fatal(err)
	}

	config.Verbose = true

	rmq, err := rabbitmq.New(rabbitmq.Config{
		URL: fmt.Sprintf("amqp://guest:guest@127.0.0.1:%s/", hostPort),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer rmq.Close()

	queues := []string{"test.product", "test.category", "test.user"}
	for _, queue := range queues {
		if err := rmq.Queue(queue); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("It consumes all messages from the queue and exits", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			rmq.Publish([]byte(fmt.Sprintf("product-%d", i)), "test.product")
			rmq.Publish([]byte(fmt.Sprintf("category-%d", i)), "test.category")
			rmq.Publish([]byte(fmt.Sprintf("user-%d", i)), "test.user")
		}

		worker := queue.NewWorker(rmq.Connection(), queues, false, []string{}, "rabbitmq/consumer", []string{"./consumer", hostPort, queueChunkSize})
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

	t.Run("It consumes all existing and new messages from the queue and keeps running in daemon mode", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			rmq.Publish([]byte(fmt.Sprintf("product-%d", i)), "test.product")
			rmq.Publish([]byte(fmt.Sprintf("category-%d", i)), "test.category")
			rmq.Publish([]byte(fmt.Sprintf("user-%d", i)), "test.user")
		}

		worker := queue.NewWorker(rmq.Connection(), queues, true, []string{}, "rabbitmq/consumer", []string{"./consumer", hostPort, queueChunkSize})
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

func setupRabbitMqService(ctx context.Context) (func(), string, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	deferFn := func() {
		dockerClient.Close()
	}

	_, err = dockerClient.ImagePull(ctx, rmqDockerImage, image.PullOptions{})
	if err != nil {
		return deferFn, "", err
	}

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:        rmqDockerImage,
		ExposedPorts: nat.PortSet{"5672": struct{}{}, "15672": struct{}{}},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port("5672"):  {{HostIP: "127.0.0.1", HostPort: "0"}},
			nat.Port("15672"): {{HostIP: "127.0.0.1", HostPort: "0"}},
		},
	}, nil, nil, "")

	if err != nil {
		return deferFn, "", err
	}

	deferFn = func() {
		dockerClient.Close()
		dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}

	if err = dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return deferFn, "", err
	}

	time.Sleep(3 * time.Second) //there is no healthcheck for rabbitmq

	info, err := dockerClient.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return deferFn, "", err
	}

	return deferFn, info.NetworkSettings.Ports["5672/tcp"][0].HostPort, nil
}
