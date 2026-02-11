//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/hgajjar/toolbox/integration/rabbitmq"
	"github.com/pkg/errors"
)

const (
	rmqDockerImage      = "docker.io/library/rabbitmq:3.9-management-alpine"
	postgresDockerImage = "postgres:14.8"
)

func setupRabbitMqService(ctx context.Context) (func(), string, error) {
	return setupAndStartContainer(ctx, rmqDockerImage, []string{"5672", "15672"}, nil, &container.HealthConfig{
		Test:     []string{"CMD", "rabbitmq-diagnostics", "-q", "check_running"},
		Interval: 5 * time.Second,
		Timeout:  2 * time.Second,
		Retries:  5,
	})
}

func setupRabbitMqConnection(hostPort string) (*rabbitmq.Rabbitmq, error) {
	for {
		select {
		case <-time.After(5 * time.Second):
			return nil, errors.New("timed-out while waiting for rabbitmq service to start")
		default:
			rmq, err := rabbitmq.New(rabbitmq.Config{
				URL: fmt.Sprintf("amqp://guest:guest@127.0.0.1:%s/", hostPort),
			})
			if err == nil {
				return rmq, nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func setupPostgresService(ctx context.Context) (func(), string, error) {
	return setupAndStartContainer(ctx, postgresDockerImage, []string{"5432"}, []string{"POSTGRES_PASSWORD=postgres"}, &container.HealthConfig{
		Test:     []string{"CMD", "pg_isready", "-U", "postgres"},
		Interval: 5 * time.Second,
		Timeout:  2 * time.Second,
		Retries:  5,
	})
}

func buildPostgresConnStr(hostPort string) string {
	return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%s/postgres?sslmode=disable", hostPort)
}

func setupAndStartContainer(ctx context.Context, imageName string, ports []string, env []string, healthcheck *container.HealthConfig) (func(), string, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	deferFn := func() {
		dockerClient.Close()
	}

	imagePullResp, err := dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return deferFn, "", errors.Wrap(err, "failed to pull rabbitmq image")
	}
	defer imagePullResp.Close()
	io.ReadAll(imagePullResp)

	exposedPorts := nat.PortSet{}
	portBindings := map[nat.Port][]nat.PortBinding{}

	for _, port := range ports {
		natPort := nat.Port(port)
		exposedPorts[natPort] = struct{}{}
		portBindings[natPort] = []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "0"}}
	}

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: exposedPorts,
		Env:          env,
		Healthcheck:  healthcheck,
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, "")

	if err != nil {
		return deferFn, "", errors.Wrap(err, "failed to create container")
	}

	deferFn = func() {
		dockerClient.Close()
		dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}

	if err = dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return deferFn, "", errors.Wrap(err, "failed to start container")
	}

	info, err := dockerClient.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return deferFn, "", errors.Wrap(err, "failed to inspect container")
	}
	if info.State == nil || !info.State.Running {
		logs, err := dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			return deferFn, "", errors.Wrap(err, "container is not running and failed to get logs")
		}
		defer logs.Close()
		logData, _ := io.ReadAll(logs)
		return deferFn, "", errors.Errorf("container is not running. Logs: %s", string(logData))
	}

	bindingPort := nat.Port(ports[0] + "/tcp")

	err = waitUntilContainerHealthy(ctx, dockerClient, resp.ID)
	if err != nil {
		return deferFn, "", errors.Wrap(err, "container did not become healthy in time")
	}

	return deferFn, info.NetworkSettings.Ports[bindingPort][0].HostPort, nil
}

func waitUntilContainerHealthy(ctx context.Context, dockerClient *client.Client, containerID string) error {
	for {
		select {
		case <-time.After(30 * time.Second):
			return errors.New("timed-out while waiting for container to become healthy")
		default:
			info, err := dockerClient.ContainerInspect(ctx, containerID)
			if err != nil {
				return errors.Wrap(err, "failed to inspect container")
			}
			if info.State != nil && info.State.Health != nil && info.State.Health.Status == "healthy" {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func defineQueues(queueNames []string, rmq *rabbitmq.Rabbitmq) error {
	for _, queue := range queueNames {
		if err := rmq.Queue(queue); err != nil {
			return err
		}
	}

	return nil
}
