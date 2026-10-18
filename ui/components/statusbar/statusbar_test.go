package statusbar_test

import (
	"github.com/charmbracelet/x/ansi"
	"ktea/config"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"testing"

	"github.com/stretchr/testify/assert"
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

		render := ansi.Strip(sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{},
				ConfigIO: nil,
			})), tests.Renderer))

		assert.Contains(t, render, "\n  test provider  \ue0b4           \ue0b4\n")
	})

	t.Run("exposes global and provider shortcuts", func(t *testing.T) {
		sb := statusbar.New()
		sb.SetProvider(TestProvider{})

		assert.Equal(t, []statusbar.Shortcut{
			{Name: "Switch Tabs", Keybinding: "C-←/→/h/l"},
			{Name: "Open", Keybinding: "C-o"},
		}, sb.Shortcuts())
		assert.Equal(t, "test provider", sb.Title())

		render := ansi.Strip(sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{})), tests.Renderer))
		assert.NotContains(t, render, "Open")
	})

	t.Run("with active cluster", func(t *testing.T) {
		sb := statusbar.New()
		sb.SetProvider(TestProvider{})

		render := ansi.Strip(sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "prd",
						Color:            styles.ColorRed,
						Active:           false,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
						SchemaRegistry:       nil,
						TLSConfig:            config.TLSConfig{Enable: false},
						KafkaConnectClusters: nil,
					},
					{
						Name:             "dev",
						Color:            styles.ColorGreen,
						Active:           true,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
						SchemaRegistry:       nil,
						TLSConfig:            config.TLSConfig{Enable: false},
						KafkaConnectClusters: nil,
					},
				},
				ConfigIO: nil,
			})), tests.Renderer))

		assert.Contains(t, render, "dev")
		assert.NotContains(t, render, "prd")
	})

	t.Run("plain fonts", func(t *testing.T) {
		sb := statusbar.New()
		sb.SetProvider(TestProvider{})

		render := ansi.Strip(sb.View(tests.NewKontext(
			tests.WithWindowWidth(30),
			tests.WithConfig(&config.Config{
				PlainFonts: true,
				Clusters:   []config.Cluster{},
				ConfigIO:   nil,
			})), tests.Renderer))

		assert.NotContains(t, render, "\uE0B4")
		assert.NotContains(t, render, "\uE0B0")
		assert.NotContains(t, render, "\uE0B6")
	})
}
