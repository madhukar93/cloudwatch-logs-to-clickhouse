package main

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startDozzleContainer() (testcontainers.Container, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "amir20/dozzle:latest",
		ExposedPorts: []string{"8080/tcp"},
		Env:          map[string]string{},
		Mounts: testcontainers.Mounts(
			// Mount the docker socket so that we can run docker commands from within the container
			testcontainers.ContainerMount{
				Source: testcontainers.DockerBindMountSource{HostPath: "/var/run/docker.sock"},
				Target: testcontainers.ContainerMountTarget("/var/run/docker.sock"),
			},
		),
		WaitingFor: wait.ForHTTP("/").WithPort("8080/tcp"),
		Name:       "dozzle",
	}

	dozzleContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, err
	}

	return dozzleContainer, nil
}
