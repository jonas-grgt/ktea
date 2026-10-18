package sradmin

import (
	"charm.land/log/v2"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"sync"
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

	// The compose module only registers with Ryuk (testcontainers' reaper) after Up
	// returns. Packages tested in parallel share one Ryuk session, so if another
	// package (e.g. kadmin) exits while this stack is still starting, Ryuk reaps it.
	// Hold our own connection for the whole run to prevent that.
	releaseReaper := holdReaperConnection()
	defer releaseReaper()

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

	releaseReaper()
	os.Exit(exitCode)
}

func holdReaperConnection() func() {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		log.Fatalf("Failed to create docker provider: %v", err)
	}
	if provider.Config().Config.RyukDisabled {
		return func() {}
	}
	//nolint:staticcheck // same call the compose module uses; there is no public replacement.
	reaper, err := testcontainers.NewReaper(ctx, testcontainers.SessionID(), provider, "")
	if err != nil {
		log.Fatalf("Failed to create reaper: %v", err)
	}
	termSignal, err := reaper.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to reaper: %v", err)
	}
	var once sync.Once
	return func() { once.Do(func() { termSignal <- true }) }
}
