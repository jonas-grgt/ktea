package kadmin

import (
	"context"
	"ktea/config"
	"ktea/styles"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func TestSaramaKadminWithSCRAM(t *testing.T) {
	ctx := context.Background()

	redpandaContainer, err := redpanda.Run(
		ctx,
		"docker.redpanda.com/redpandadata/redpanda:v23.3.10",
		redpanda.WithEnableSASL(),
		redpanda.WithNewServiceAccount("admin", "adminadmin"),
		redpanda.WithSuperusers("admin"),
	)
	if err != nil {
		t.Fatalf("Failed to start Redpanda container: %v", err)
	}

	t.Cleanup(func() {
		if err := redpandaContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Redpanda container: %v", err)
		}
	})

	brokerAddress, err := redpandaContainer.KafkaSeedBroker(ctx)
	if err != nil {
		t.Fatalf("Failed to get Kafka seed broker: %v", err)
	}

	t.Logf("Broker address: %s", brokerAddress)

	t.Run("SCRAMSHA256", func(t *testing.T) {
		cluster := &config.Cluster{
			Name:             "test-scram",
			Color:            styles.ColorGreen,
			Active:           true,
			BootstrapServers: []string{brokerAddress},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodSASLSCRAMSHA256,
				Username:   "admin",
				Password:   "adminadmin",
			},
			TLSConfig: config.TLSConfig{
				Enable: false,
			},
		}

		_, err := NewSaramaKadmin(cluster)
		if err != nil {
			t.Fatalf("Failed to connect with SCRAM-SHA256: %v", err)
		}
	})
}
