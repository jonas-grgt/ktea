package sradmin

import (
	"context"
	"github.com/charmbracelet/log"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"time"
)

var ctx context.Context
var kc *kafka.KafkaContainer
var brokers []string

var sra SrAdmin

func init() {
	ctx = context.Background()

	file, err := os.ReadFile("docker-compose.yml")
	if err != nil {
		log.Fatal("Unable to read docker-compose.yml")
	}

	dComp, err := compose.NewDockerComposeWith(compose.WithStackReaders(strings.NewReader(string(file))))
	if err != nil {
		log.Printf("Failed to create stack: %v", err)
		return
	}

	err = dComp.
		WaitForService("schema-registry", wait.NewHTTPStrategy("/").WithPort("8081/tcp").WithStartupTimeout(10*time.Second)).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Printf("Failed to start docker-compose: %v", err)
		return
	}
}
