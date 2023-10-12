package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startWiremockContainer() testcontainers.Container {
	ctx := context.Background()

	// Start Wiremock container
	wiremockReq := testcontainers.ContainerRequest{
		Image:        "rodolpheche/wiremock:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp"),
	}
	wiremockContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: wiremockReq,
		Started:          true,
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
	stubsFile, err := os.Open("../stubs.json")
	if err != nil {
		log.Fatalf("Failed to open stubs file: %s", err)
	}
	defer stubsFile.Close()

	data, err := json.Marshal(stubsFile)
	if err != nil {
		log.Fatalf("Failed to marshal stubs file to json: %s", err)
	}

	log.Printf("Posting stubs to %s", wiremockURL)
	resp, err := http.Post("http://"+wiremockURL+"/__admin/mappings", "application/json", bytes.NewBuffer(data))

	if err != nil {
		log.Fatalf("Failed to post stub: %s", err)
	}
	defer resp.Body.Close()
}
