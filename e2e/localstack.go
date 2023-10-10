package main

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startlocalStackContainer() testcontainers.Container {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		// Name:         "localstack",
		Image:        "localstack/localstack",
		ExposedPorts: []string{"4566/tcp"}, // LocalStack now uses a single edge port
		WaitingFor:   wait.ForLog("Ready."),
		// mount /var/run/docker.sock:/var/run/docker.sock
		Mounts: testcontainers.Mounts(
			// Mount the docker socket so that we can run docker commands from within the container
			testcontainers.ContainerMount{
				Source: testcontainers.DockerBindMountSource{HostPath: "/var/run/docker.sock"},
				Target: testcontainers.ContainerMountTarget("/var/run/docker.sock"),
			},
		),
		Env: map[string]string{
			"SERVICES":      "apigateway,lambda,cloudwatch",
			"IAM_SOFT_MODE": "1",
		},
	}

	localstackContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		// Reuse:            true,
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start container: %s", err)
	}

	return localstackContainer
}
