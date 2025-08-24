package sradmin

import (
	"context"
	"github.com/charmbracelet/log"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"testing"
)

var ctx context.Context

var schemaRegistryPort nat.Port

func TestMain(m *testing.M) {
	ctx = context.Background()

	file, err := os.ReadFile("docker-compose.yml")
	if err != nil {
		log.Fatalf("Unable to read docker-compose.yml")
	}

	dComp, err := compose.NewDockerComposeWith(compose.WithStackReaders(strings.NewReader(string(file))))
	if err != nil {
		log.Fatalf("Failed to create stack: %v", err)
	}

	err = dComp.
		WaitForService("schema-registry", wait.NewHealthStrategy()).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Fatalf("Failed to start docker-compose: %v", err)
	}

	schemaRegistry, err := dComp.ServiceContainer(ctx, "schema-registry")
	if err != nil {
		log.Fatalf("Failed to get schema-registry container: %v", err)
	}
	schemaRegistryPort, err = schemaRegistry.MappedPort(ctx, "8081/tcp")

	exitCode := m.Run()

	if err := dComp.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal); err != nil {
		log.Fatalf("Failed to tear down docker-compose: %v", err)
	}

	os.Exit(exitCode)
}
