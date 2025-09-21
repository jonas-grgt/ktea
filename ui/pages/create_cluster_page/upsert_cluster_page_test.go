package create_cluster_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"ktea/ui/tabs"
	"testing"
)

var shortcuts []statusbar.Shortcut

var ktx = tests.NewKontext(
	tests.WithConfig(
		&config.Config{
			Clusters: []config.Cluster{},
		},
	),
)

func TestCreateInitialMessageWhenNoClusters(t *testing.T) {
	t.Run("Display info message when no clusters", func(t *testing.T) {
		// given
		createEnvPage := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)

		// then
		render := createEnvPage.View(ktx, tests.Renderer)
		assert.Contains(t, render, "No clusters configured. Please create your first cluster!")
	})

	t.Run("Do not display info message when there are clusters", func(t *testing.T) {
		createEnvPage := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)

		var ktx = tests.NewKontext(
			tests.WithConfig(
				&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
			),
		)
		render := createEnvPage.View(ktx, tests.Renderer)
		assert.NotContains(t, render, "No clusters configured. Please create your first cluster!")
	})
}

func TestTabs(t *testing.T) {
	t.Run("Switch to schema registry tab", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)
		// and: a cluster is registered
		cluster := config.Cluster{}
		page.clusterToEdit = &cluster

		// when
		page.Update(tests.Key(tea.KeyF5))

		// then: schema registry tab is visible
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "Schema Registry URL")
		assert.Contains(t, render, "Schema Registry Username")
		assert.Contains(t, render, "Schema Registry Password")
	})

	t.Run("Switch to kafka connect tab", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)
		// and: a cluster is registered
		cluster := config.Cluster{}
		page.clusterToEdit = &cluster

		// when
		page.Update(tests.Key(tea.KeyF6))

		// then: kafka connect tab is visible
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "Kafka Connect URL")
		assert.Contains(t, render, "Kafka Connect Username")
		assert.Contains(t, render, "Kafka Connect Password")
	})

	t.Run("switching back to clusters tab remembers previously entered state", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)
		// and: a cluster is registered
		cluster := config.Cluster{}
		page.clusterToEdit = &cluster
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// select Primary
		cmd = page.Update(tests.Key(tea.KeyUp))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())

		// when
		page.Update(tests.Key(tea.KeyF5))
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "Schema Registry URL")
		page.Update(tests.Key(tea.KeyF4))

		// then: previously entered details are visible
		render = page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "TST")
		assert.Contains(t, render, "localhost:9092")
	})

	t.Run("Cannot switch to schema registry tab when no cluster registered yet", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)

		// when
		page.Update(tests.Key(tea.KeyF5))

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "create a cluster before adding a schema registry")
	})
}

func TestValidation(t *testing.T) {
	t.Run("Name cannot be empty", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)

		// when
		page.Update(tests.Key(tea.KeyEnter))

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "name cannot be empty")
	})

	t.Run("Name must be unique", func(t *testing.T) {
		// given
		ktx := tests.NewKontext(
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "prd",
						Color:            "#808080",
						Active:           true,
						BootstrapServers: nil,
						SASLConfig:       nil,
					},
					{
						Name:             "tst",
						Color:            "#F0F0F0",
						Active:           false,
						BootstrapServers: nil,
						SASLConfig:       nil,
					},
				},
			}),
		)
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)

		// when
		tests.UpdateKeys(page, "prd")
		page.Update(tests.Key(tea.KeyEnter))

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "cluster prd already exists, name most be unique")
	})

	t.Run("Host cannot be empty", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())

		// when
		page.Update(tests.Key(tea.KeyEnter))

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "host cannot be empty")
	})
}

func TestCreateCluster(t *testing.T) {
	t.Run("Selecting none auth method creates cluster", func(t *testing.T) {
		// given
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			})),
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: next field
		page.Update(cmd())
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: next field
		page.Update(cmd())
		// and: select SSL disabled
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: next field
		page.Update(cmd())
		// and: select auth method none
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: submit
		msgs := tests.Submit(page)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "TST",
			Color:            styles.ColorGreen,
			Active:           false,
			BootstrapServers: []string{"localhost:9092"},
			SchemaRegistry:   nil,
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})
}

func TestClusterForm(t *testing.T) {

	t.Run("Selecting SASL auth method displays username and password fields", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig:       nil,
				},
			},
		}))
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			programKtx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// select Primary
		cmd = page.Update(tests.Key(tea.KeyUp))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: SSL is disabled
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: auth method SASL is selected
		page.Update(tests.Key(tea.KeyDown))

		// then
		render := page.View(programKtx, tests.Renderer)
		assert.Contains(t, render, "SASL_PLAINTEXT")
		assert.Contains(t, render, "Username")
		assert.Contains(t, render, "Password")
	})

	t.Run("After selecting SASL_PLAINTEXT auth method and going back to select none username and password fields are not visible anymore", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig:       nil,
				},
			},
		}))
		createEnvPage := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			programKtx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(createEnvPage, "TST")
		cmd := createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// select Primary
		cmd = createEnvPage.Update(tests.Key(tea.KeyUp))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: select Color
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// and: Host is entered
		tests.UpdateKeys(createEnvPage, "localhost:9092")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: auth method none is selected
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// go back to auth method
		cmd = createEnvPage.Update(tests.Key(tea.KeyShiftTab))
		cmd = createEnvPage.Update(cmd())
		cmd = createEnvPage.Update(tests.Key(tea.KeyUp))

		// then
		render := createEnvPage.View(programKtx, tests.Renderer)
		assert.NotContains(t, render, "Username")
		assert.NotContains(t, render, "Password")
	})

	t.Run("Selecting SASL_PLAINTEXT auth method and filling all SASL fields creates cluster", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig:       nil,
				},
			},
		}))
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			programKtx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// select Primary
		cmd = page.Update(tests.Key(tea.KeyUp))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: SSL is enabled
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: auth method SASL is selected
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: security protocol SASL_PLAINTEXT
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(page, "username")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(page, "password")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// submit
		msgs := tests.Submit(page)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "TST",
			Color:            styles.ColorRed,
			Active:           false,
			BootstrapServers: []string{"localhost:9092"},
			SchemaRegistry:   nil,
			SSLEnabled:       true,
			SASLConfig: &config.SASLConfig{
				Username:         "username",
				Password:         "password",
				SecurityProtocol: config.SASLPlaintextSecurityProtocol,
			},
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("Enabling SSL", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig:       nil,
				},
			},
		}))
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			programKtx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// select Primary
		cmd = page.Update(tests.Key(tea.KeyUp))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: auth method none is selected
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: select SSL enabled
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: select SASL_SSL security protocol
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(page, "username")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(page, "password")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// submit
		msgs := tests.Submit(page)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "TST",
			Color:            styles.ColorRed,
			Active:           false,
			BootstrapServers: []string{"localhost:9092"},
			SchemaRegistry:   nil,
			SSLEnabled:       true,
			SASLConfig: &config.SASLConfig{
				Username:         "username",
				Password:         "password",
				SecurityProtocol: config.SASLPlaintextSecurityProtocol,
			},
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("C-r resets form", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig:       nil,
				},
			},
		}))
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			programKtx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select blue color
		page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: auth method none is selected
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: select SSL enabled
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and:SASL auth method
		page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(page, "username-john")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(page, "password")
		cmd = page.Update(tests.Key(tea.KeyEnter))

		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "TST")
		assert.Contains(t, render, ">  blue")
		assert.Contains(t, render, "localhost:9092")
		assert.Contains(t, render, "username-john")
		assert.Contains(t, render, "> Enable SSL")
		assert.Contains(t, render, "********")

		// when
		page.Update(tests.Key(tea.KeyCtrlR))

		// then
		render = page.View(ktx, tests.Renderer)
		assert.NotContains(t, render, "TST")
		assert.NotContains(t, render, ">  blue")
		assert.Contains(t, render, ">  green")
		assert.NotContains(t, render, "localhost:9092")
		assert.NotContains(t, render, "username-john")
		assert.NotContains(t, render, "> Enable SSL")
		assert.NotContains(t, render, "********")
	})

	t.Run("After creating a new cluster, display notification when checking connectivity", func(t *testing.T) {
		// given
		createEnvPage := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			ktx,
			shortcuts,
		)

		// when
		createEnvPage.Update(kadmin.ConnCheckStartedMsg{})

		// then
		render := createEnvPage.View(ktx, tests.Renderer)
		assert.Contains(t, render, "Testing cluster connectivity")
	})

}

func TestEditClusterForm(t *testing.T) {

	t.Run("Sets title", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig:       nil,
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig:       nil,
				},
			},
		}))
		page := NewEditClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			nil,
			programKtx,
			config.Cluster{
				Name:             "prd",
				Color:            styles.ColorGreen,
				Active:           false,
				BootstrapServers: []string{"localhost:9092"},
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
			WithTitle("Edit Cluster"),
		)

		// when
		title := page.Title()

		// then
		assert.Equal(t, "Edit Cluster", title)
	})

	t.Run("Checks connection upon updating", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig:       nil,
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig:       nil,
				},
			},
		}))
		page := NewEditClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			nil,
			programKtx,
			config.Cluster{
				Name:             "prd",
				Color:            "#808080",
				BootstrapServers: []string{"localhost:9092"},
			},
		)

		// when
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: Host is entered
		for i := 0; i < len("localhost:9092"); i++ {
			page.Update(tests.Key(tea.KeyBackspace))
		}
		tests.UpdateKeys(page, "localhost:9091")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: auth method none is selected
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: select SSL disabled
		cmd = page.Update(tests.Key(tea.KeyEnter))
		msgs := tests.Submit(page)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "prd",
			Color:            styles.ColorGreen,
			Active:           false,
			BootstrapServers: []string{"localhost:9091"},
			SchemaRegistry:   nil,
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("After editing, display notification when checking connectivity", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig:       nil,
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig:       nil,
				},
			},
		}))
		page := NewEditClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			nil,
			programKtx,
			config.Cluster{
				Name:             "prd",
				Color:            "#808080",
				Active:           false,
				BootstrapServers: []string{":9092"},
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
		)

		// when
		page.Update(kadmin.ConnCheckStartedMsg{})

		// then
		render := page.View(programKtx, tests.Renderer)
		assert.Contains(t, render, "Testing cluster connectivity")
	})

	t.Run("Edit when there was no initial schema registry created", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig:       nil,
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig:       nil,
				},
			},
		}))
		page := NewEditClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			nil,
			programKtx,
			config.Cluster{
				Name:             "prd",
				Color:            "#808080",
				Active:           false,
				BootstrapServers: []string{":9092"},
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
		)
		page.Update(tests.Key(tea.KeyF5))

		// when
		tests.UpdateKeys(page, "https://localhost:8081")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and
		tests.UpdateKeys(page, "sr-username")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and
		tests.UpdateKeys(page, "sr-password")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and
		msgs := tests.Submit(page)

		// then
		assert.IsType(t, sradmin.MockConnectionCheckedMsg{}, msgs[0])
		assert.Equal(t, &config.SchemaRegistryConfig{
			Url:      "https://localhost:8081",
			Username: "sr-username",
			Password: "sr-password",
		}, msgs[0].(sradmin.MockConnectionCheckedMsg).Config)
	})

	t.Run("Display notification when connection has failed", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig:       nil,
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig:       nil,
				},
			},
		}))
		page := NewEditClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			nil,
			programKtx,
			config.Cluster{
				Name:             "prd",
				Color:            "#808080",
				Active:           false,
				BootstrapServers: []string{":9092"},
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
		)

		// when
		page.Update(kadmin.ConnCheckErrMsg{Err: fmt.Errorf("kafka: client has run out of available brokers to talk to")})

		// then
		render := page.View(programKtx, tests.Renderer)
		assert.Contains(t, render, "Cluster not updated: kafka: client has run out of available brokers to talk to")
	})

	t.Run("Display notification when cluster has been updated", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig:       nil,
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig:       nil,
				},
			},
		}))
		page := NewEditClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			nil,
			programKtx,
			config.Cluster{
				Name:             "prd",
				Color:            "#808080",
				Active:           false,
				BootstrapServers: []string{":9092"},
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
		)

		// when
		page.Update(config.ClusterRegisteredMsg{
			Cluster: &config.Cluster{
				Name:             "production",
				Color:            styles.ColorGreen,
				Active:           false,
				BootstrapServers: []string{"localhost:9093"},
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
		})

		// then
		render := page.View(programKtx, tests.Renderer)
		assert.Contains(t, render, "Cluster updated!")
	})
}

func TestCreateSchemaRegistry(t *testing.T) {
	// given

	programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
		Clusters: []config.Cluster{
			{
				Name:             "PRD",
				BootstrapServers: []string{"localhost:9092"},
				SASLConfig:       nil,
			},
		},
	}))
	page := NewCreateClusterPage(
		tabs.NewMockClustersTabNavigator(),
		kadmin.MockConnChecker,
		sradmin.MockConnChecker,
		config.MockClusterRegisterer{},
		programKtx,
		shortcuts,
	)

	t.Run("Check connectivity before registering the schema registry", func(t *testing.T) {
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// select Primary
		cmd = page.Update(tests.Key(tea.KeyUp))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select Color
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: auth method none is selected
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: select SSL enabled
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: select SASL_SSL security protocol
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(page, "username")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(page, "password")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// submit
		tests.Submit(page)
		page.Update(config.ClusterRegisteredMsg{
			Cluster: &config.Cluster{
				Name:             "cluster-name",
				Color:            styles.ColorGreen,
				Active:           false,
				BootstrapServers: nil,
				SASLConfig:       nil,
				SchemaRegistry:   nil,
				SSLEnabled:       false,
			},
		})

		// and: switch to schema registry tab
		page.Update(tests.Key(tea.KeyF5))

		// and: schema registry url
		tests.UpdateKeys(page, "sr-url")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: schema registry username
		tests.UpdateKeys(page, "sr-username")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: schema registry pwd
		tests.UpdateKeys(page, "sr-password")
		cmd = page.Update(tests.Key(tea.KeyEnter))

		// when: submit
		msgs := tests.Submit(page)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, sradmin.MockConnectionCheckedMsg{}, msgs[0])
		assert.EqualValues(t, &config.SchemaRegistryConfig{
			Url:      "sr-url",
			Username: "sr-username",
			Password: "sr-password",
		}, msgs[0].(sradmin.MockConnectionCheckedMsg).Config)
	})

	t.Run("Display error notification when connection cannot be made", func(t *testing.T) {
		page.Update(sradmin.ConnCheckErrMsg{Err: fmt.Errorf("cannot connect")})

		render := page.View(ktx, tests.Renderer)

		assert.Contains(t, render, "unable to reach the schema registry")
	})
}

func TestSchemaRegistryForm(t *testing.T) {
	t.Run("C-r resets form", func(t *testing.T) {
		// given
		programKtx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig:       nil,
				},
			},
		}))
		page := NewCreateClusterPage(
			tabs.NewMockClustersTabNavigator(),
			kadmin.MockConnChecker,
			sradmin.MockConnChecker,
			config.MockClusterRegisterer{},
			programKtx,
			shortcuts,
		)
		// and: enter name
		tests.UpdateKeys(page, "TST")
		cmd := page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: select blue color
		page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and: Host is entered
		tests.UpdateKeys(page, "localhost:9092")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: auth method none is selected
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = page.Update(cmd())
		// and: select SSL enabled
		cmd = page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		// next field
		page.Update(cmd())
		// and:SASL auth method
		page.Update(tests.Key(tea.KeyDown))
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(page, "username-john")
		cmd = page.Update(tests.Key(tea.KeyEnter))
		page.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(page, "password")
		cmd = page.Update(tests.Key(tea.KeyEnter))

		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "TST")
		assert.Contains(t, render, ">  blue")
		assert.Contains(t, render, "localhost:9092")
		assert.Contains(t, render, "username-john")
		assert.Contains(t, render, "> Enable SSL")
		assert.Contains(t, render, "********")

		// when
		page.Update(tests.Key(tea.KeyCtrlR))

		// then
		render = page.View(ktx, tests.Renderer)
		assert.NotContains(t, render, "TST")
		assert.NotContains(t, render, ">  blue")
		assert.Contains(t, render, ">  green")
		assert.NotContains(t, render, "localhost:9092")
		assert.NotContains(t, render, "username-john")
		assert.NotContains(t, render, "> Enable SSL")
		assert.NotContains(t, render, "********")
	})
}
