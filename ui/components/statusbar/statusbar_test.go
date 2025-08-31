package statusbar_test

import (
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"testing"
)

type TestProvider struct {
}

func (t TestProvider) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{
			Name:       "Open",
			Keybinding: "C-o",
		},
	}
}

func (t TestProvider) Title() string {
	return "test provider"
}

func TestStatusbar(t *testing.T) {
	t.Run("do not show shortcuts by default", func(t *testing.T) {
		sb := statusbar.New()
		sb.SetProvider(TestProvider{})

		render := sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{},
				ConfigIO: nil,
			})), tests.Renderer)

		assert.Equal(t, "                              \n  test provider               ", render)
	})

	t.Run("toggle shortcuts", func(t *testing.T) {
		sb := statusbar.New()
		sb.SetProvider(TestProvider{})

		render := sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{},
				ConfigIO: nil,
			})), tests.Renderer)

		assert.Equal(t, "                              \n  test provider               ", render)

		sb.ToggleShortcuts()

		render = sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{},
				ConfigIO: nil,
			})), tests.Renderer)

		assert.Equal(t, "                                                    \n  Switch Tabs: ≪ C-←/→/h/l »   Open: ≪ C-o »        \n                                                    \n  test provider                                     ", render)
	})

	t.Run("with active cluster", func(t *testing.T) {
		sb := statusbar.New()
		sb.SetProvider(TestProvider{})

		render := sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:                 "prd",
						Color:                styles.ColorRed,
						Active:               false,
						BootstrapServers:     nil,
						SASLConfig:           nil,
						SchemaRegistry:       nil,
						SSLEnabled:           false,
						KafkaConnectClusters: nil,
					},
					{
						Name:                 "dev",
						Color:                styles.ColorGreen,
						Active:               true,
						BootstrapServers:     nil,
						SASLConfig:           nil,
						SchemaRegistry:       nil,
						SSLEnabled:           false,
						KafkaConnectClusters: nil,
					},
				},
				ConfigIO: nil,
			})), tests.Renderer)

		assert.Contains(t, render, "dev")
		assert.NotContains(t, render, "prd")
	})
}
