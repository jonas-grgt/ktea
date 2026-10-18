package main

import (
	"github.com/charmbracelet/x/ansi"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestKtea(t *testing.T) {

	disableNerdFonts := true

	t.Run("No clusters configured", func(t *testing.T) {
		t.Run("Shows create cluster page", func(t *testing.T) {
			model := NewModel(&disableNerdFonts, kadmin.NewMockKadminInstantiator(), config.NewInMemoryConfigIO(&config.Config{}))
			model.ktx = &kontext.ProgramKtx{
				WindowWidth:     200,
				WindowHeight:    200,
				AvailableHeight: 200,
			}
			model.Update(config.LoadedMsg{
				Config: &config.Config{},
			})
			view := ansi.Strip(model.View().Content)

			assert.Contains(t, view, "┃ Name")
			// do not show cluster upsert tabs
			assert.NotContains(t, view, "f1")
			assert.NotContains(t, view, "f2")
			assert.NotContains(t, view, "f3")
		})
	})

	t.Run("Tabs", func(t *testing.T) {
		t.Run("Cycle through tabs", func(t *testing.T) {
			io := config.NewInMemoryConfigIO(
				&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{":19092"},
							SchemaRegistry: &config.SchemaRegistryConfig{
								Url:      "",
								Username: "",
								Password: "",
							},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
						{
							Name:             "tst",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{":19092"},
							SchemaRegistry:   nil,
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				},
			)
			model := NewModel(&disableNerdFonts, kadmin.NewMockKadminInstantiator(), io)
			model.Update(config.LoadedMsg{Config: config.New(io)})

			model.Update(tea.WindowSizeMsg{
				Width:  100,
				Height: 100,
			})

			model.Update(tea.KeyPressMsg{Code: '2', Mod: tea.ModAlt})

			view := ansi.Strip(model.View().Content)

			var expectedLayout = "" +
				"╭────────╮╭─────────────────╮╭─────────────────╮╭──────────╮                                        \n" +
				"│ Topics ││ Consumer Groups ││ Schema Registry ││ Clusters │  ≪ F1 » help                           \n" +
				"┘        └┴─────────────────┴┴─────────────────┴┴──────────┴────────────────────────────────────────"
			assert.Contains(t, view, expectedLayout)

			model.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModCtrl})

			model.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl})

			view = ansi.Strip(model.View().Content)

			expectedLayout = "" +
				"╭────────╮╭─────────────────╮╭─────────────────╮╭──────────╮                                        \n" +
				"│ Topics ││ Consumer Groups ││ Schema Registry ││ Clusters │  ≪ F1 » help                           \n" +
				"┴────────┴┴─────────────────┴┘                 └┴──────────┴────────────────────────────────────────\n"
			assert.Contains(t, view, expectedLayout)

			model.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl})

			view = ansi.Strip(model.View().Content)

			expectedLayout = "" +
				"╭────────╮╭─────────────────╮╭─────────────────╮╭──────────╮                                        \n" +
				"│ Topics ││ Consumer Groups ││ Schema Registry ││ Clusters │  ≪ F1 » help                           \n" +
				"┴────────┴┴─────────────────┴┴─────────────────┴┘          └────────────────────────────────────────\n"
			assert.Contains(t, view, expectedLayout)

			model.Update(tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModCtrl})

			model.Update(tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModCtrl})

			model.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl})

			view = ansi.Strip(model.View().Content)

			expectedLayout = "" +
				"╭────────╮╭─────────────────╮╭─────────────────╮╭──────────╮                                        \n" +
				"│ Topics ││ Consumer Groups ││ Schema Registry ││ Clusters │  ≪ F1 » help                           \n" +
				"┘        └┴─────────────────┴┴─────────────────┴┴──────────┴────────────────────────────────────────\n"

			assert.Contains(t, view, expectedLayout)
		})
	})
}
