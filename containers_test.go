package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	// defaultProtocol = "udp"
	// defaultPort     = 51820
	defaultProtocol = "tcp"
	defaultPort     = 80
)

func createUbuntuContainer() (func() error, error) {
	ctx := context.Background()
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %w", err)
	}
	sharedDir := path.Join(pwd, "shared")

	portProtocol := fmt.Sprintf("%d/%s", defaultPort, defaultProtocol)
	req := testcontainers.ContainerRequest{
		Image:        "ubuntu:latest",
		ExposedPorts: []string{portProtocol},
		WaitingFor:   wait.ForListeningPort(nat.Port(portProtocol)),
		Cmd: []string{
			"bash", "-c",
			"bash /shared/init.sh",
		},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(sharedDir, "/shared"),
		),
	}

	ubuntuContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	// stream logs
	logs, err := ubuntuContainer.Logs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get container logs: %w", err)
	}
	go func() {
		_, err := io.Copy(os.Stdout, logs)
		if err != nil {
			slog.Error("streaming container logs", "error", err)
		}
	}()

	// closeFunc to pass to caller
	closeFunc := func() error {
		if err := ubuntuContainer.Terminate(ctx); err != nil {
			return fmt.Errorf("terminate container: %w", err)
		}
		return nil
	}
	ip, err := ubuntuContainer.Host(ctx)
	if err != nil {
		errClose := closeFunc()
		if errClose != nil {
			return nil, fmt.Errorf("close failure after other failure: %w: %w", errClose, err)
		}
		return nil, fmt.Errorf("get host: %w", err)
	}

	port, err := ubuntuContainer.MappedPort(ctx, nat.Port(portProtocol))
	if err != nil {
		return nil, fmt.Errorf("get mapped port: %w", err)
	}
	fmt.Printf("container started at %s:%s\n", ip, port.Port())
	slog.Info("container started at %s:%s\n", ip, port.Port())

	return closeFunc, nil

}

func TestNew(t *testing.T) {
	closeFunc, err := createUbuntuContainer()
	t.Log("created container")
	require.NoError(t, err)
	t.Log("no error")
	require.NoError(t, closeFunc())
	t.Log("closed container")
}
