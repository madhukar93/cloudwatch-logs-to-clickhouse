package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startWiremockContainer() testcontainers.Container {
	ctx := context.Background()

	// Start Wiremock container
	wiremockReq := testcontainers.ContainerRequest{
		Image:        "rodolpheche/wiremock:latest",
		Name:         "wiremock",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp"),
	}
	wiremockContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: wiremockReq,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		log.Fatalf("Failed to start Wiremock container: %s", err)
	}

	// Get the actual mapped port. This can be used to make HTTP requests to the mock server.
	mappedPort, err := wiremockContainer.MappedPort(ctx, "8080")
	if err != nil {
		log.Fatalf("Failed to get mapped port: %s", err)
	}
	log.Printf("Mock server is accessible at http://localhost:%s", mappedPort.Port())

	return wiremockContainer
}

func setupStubs(wiremockURL string) {
	// Successful scenario
	var err error
	stubs, err := ioutil.ReadFile("../stubs.json")
	if err != nil {
		log.Fatalf("Failed to open stubs file: %s", err)
	}

	log.Printf("Posting stubs to %s", wiremockURL)
	resp, err := http.Post("http://"+wiremockURL+"/__admin/mappings/import", "application/json", bytes.NewBuffer(stubs))

	if err != nil {
		log.Fatalf("Failed to post stub: %s", err)
	} else {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %s", err)
		}
		log.Printf("Successfully posted stub: %s", string(bytes))
	}
	defer resp.Body.Close()
}
