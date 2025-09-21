package kcon_clusters_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui/pages/kcon_page"
	"testing"
)

var (
	username = "john"
	pwd      = "doe"
)

func TestKConsPage(t *testing.T) {
	t.Run("List all available clusters", func(t *testing.T) {
		cluster := config.Cluster{
			Name:             "dev",
			Color:            styles.ColorRed,
			Active:           true,
			BootstrapServers: []string{"localhost:9092"},
			SASLConfig:       nil,
			SchemaRegistry:   nil,
			SSLEnabled:       false,
			KafkaConnectClusters: []config.KafkaConnectConfig{
				{
					Name:     "s3-sink",
					Url:      "http://localhost:8083",
					Username: &username,
					Password: &pwd,
				},
			},
		}
		page, _ := New(&cluster, kcon_page.LoadKConPageMock)

		ktx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				cluster,
			},
		}))

		render := page.View(ktx, tests.Renderer)

		assert.Contains(t, render, "s3-sink")
	})

	t.Run("Enter loads Kafka Connect Connectors Page", func(t *testing.T) {
		cluster := config.Cluster{
			Name:             "dev",
			Color:            styles.ColorRed,
			Active:           true,
			BootstrapServers: []string{"localhost:9092"},
			SASLConfig:       nil,
			SchemaRegistry:   nil,
			SSLEnabled:       false,
			KafkaConnectClusters: []config.KafkaConnectConfig{
				{
					Name:     "s3-sink",
					Url:      "http://localhost:8083",
					Username: &username,
					Password: &pwd,
				},
			},
		}
		page, _ := New(&cluster, kcon_page.LoadKConPageMock)

		ktx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				cluster,
			},
		}))

		page.View(ktx, tests.Renderer)
		cmd := page.Update(tests.Key(tea.KeyEnter))

		msg := cmd()
		assert.IsType(t, kcon_page.LoadKConPageMockCalled{}, msg)
		assert.Equal(t, config.KafkaConnectConfig{
			Name:     "s3-sink",
			Url:      "http://localhost:8083",
			Username: &username,
			Password: &pwd,
		}, msg.(kcon_page.LoadKConPageMockCalled).Config)
	})

}
