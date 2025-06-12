package create_cluster_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/tests"
	"testing"
)

type mockClusterRegisterer struct {
}

type capturedRegistrationDetails struct {
	config.RegistrationDetails
}

func (m mockClusterRegisterer) RegisterCluster(d config.RegistrationDetails) tea.Msg {
	return capturedRegistrationDetails{d}
}

func mockConnChecker(cluster *config.Cluster) tea.Msg {
	return cluster
}

func TestCreateClusterPage(t *testing.T) {
	ktx := kontext.ProgramKtx{
		WindowWidth:  100,
		WindowHeight: 100,
		Config: &config.Config{
			Clusters: []config.Cluster{},
		},
	}

	t.Run("Display info message when no clusters", func(t *testing.T) {
		// given
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &ktx)

		// then
		render := createEnvPage.View(&ktx, tests.TestRenderer)
		assert.Contains(t, render, "No clusters configured. Please create your first cluster!")
	})

	t.Run("Do not display info message when no clusters", func(t *testing.T) {
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &ktx)

		render := createEnvPage.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}, tests.TestRenderer)
		assert.NotContains(t, render, "No clusters configured. Please create your first cluster!")
	})

	t.Run("Name cannot be empty", func(t *testing.T) {
		// given
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &ktx)

		// when
		createEnvPage.Update(tests.Key(tea.KeyEnter))

		// then
		render := createEnvPage.View(&ktx, tests.TestRenderer)
		assert.Contains(t, render, "name cannot be empty")
	})

	t.Run("Name must be unique", func(t *testing.T) {
		// given
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
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
			},
		})

		// when
		tests.UpdateKeys(createEnvPage, "prd")
		createEnvPage.Update(tests.Key(tea.KeyEnter))

		// then
		render := createEnvPage.View(&ktx, tests.TestRenderer)
		assert.Contains(t, render, "cluster prd already exists, name most be unique")
	})

	t.Run("When updating", func(t *testing.T) {
		t.Run("updates existing cluster fields", func(t *testing.T) {
			// given
			createEnvPage := NewEditForm(mockConnChecker, mockClusterRegisterer{}, &kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
				Config: &config.Config{
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
				},
			}, &FormValues{
				Name:  "prd",
				Color: "#808080",
				Host:  ":9092",
			})

			// when
			cmd := createEnvPage.Update(tests.Key(tea.KeyEnter))
			createEnvPage.Update(cmd())
			// and: select Color
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			createEnvPage.Update(cmd())
			// and: Host is entered
			for i := 0; i < len("localhost:9092"); i++ {
				createEnvPage.Update(tests.Key(tea.KeyBackspace))
			}
			tests.UpdateKeys(createEnvPage, "localhost:9091")
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			// next field
			createEnvPage.Update(cmd())
			// and: auth method none is selected
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			// next field
			cmd = createEnvPage.Update(cmd())
			// and: select SSL disabled
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			// next field
			cmd = createEnvPage.Update(cmd())
			// next group
			createEnvPage.Update(cmd())
			// and: select disabled schema registry and in doing so submitting the form
			msgs := tests.Submit(createEnvPage)

			// then
			assert.Len(t, msgs, 1)
			assert.IsType(t, &config.Cluster{}, msgs[0])
			// and
			assert.Equal(t, &config.Cluster{
				Name:             "prd",
				Color:            styles.ColorGreen,
				Active:           false,
				BootstrapServers: []string{"localhost:9091"},
				SchemaRegistry:   nil,
			}, msgs[0])
		})

		t.Run("name still has to be unique", func(t *testing.T) {
			// given
			createEnvPage := NewEditForm(mockConnChecker, mockClusterRegisterer{}, &kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
				Config: &config.Config{
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
				},
			}, &FormValues{
				Name:  "prd",
				Color: "#808080",
				Host:  ":9092",
			})

			// when
			// delete existing prd name
			createEnvPage.Update(tests.Key(tea.KeyBackspace))
			createEnvPage.Update(tests.Key(tea.KeyBackspace))
			createEnvPage.Update(tests.Key(tea.KeyBackspace))
			// enter already existing tst name
			tests.UpdateKeys(createEnvPage, "tst")
			createEnvPage.Update(tests.Key(tea.KeyEnter))

			// then
			render := createEnvPage.View(&ktx, tests.TestRenderer)
			assert.Contains(t, render, "cluster tst already exists, name most be unique")
		})
	})

	t.Run("Host cannot be empty", func(t *testing.T) {
		// given
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		})
		// and: enter name
		tests.UpdateKeys(createEnvPage, "TST")
		cmd := createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: select Color
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())

		// when
		createEnvPage.Update(tests.Key(tea.KeyEnter))

		// then
		render := createEnvPage.View(&ktx, tests.TestRenderer)
		assert.Contains(t, render, "Host cannot be empty")
	})

	t.Run("Selecting none auth method creates cluster", func(t *testing.T) {
		// given
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		})
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
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// and: select SSL disabled
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// next group
		createEnvPage.Update(cmd())
		// and: select disabled schema registry and in doing so submitting the form
		msgs := tests.Submit(createEnvPage)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, &config.Cluster{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "TST",
			Color:            styles.ColorRed,
			Active:           false,
			BootstrapServers: []string{"localhost:9092"},
			SchemaRegistry:   nil,
		}, msgs[0])
	})

	t.Run("Selecting SASL auth method displays username and password fields", func(t *testing.T) {
		// given
		programKtx := kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &programKtx)
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
		// next field
		createEnvPage.Update(cmd())
		// and: SSL is disabled
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// and: auth method SASL is selected
		createEnvPage.Update(tests.Key(tea.KeyDown))

		// then
		render := createEnvPage.View(&programKtx, tests.TestRenderer)
		assert.Contains(t, render, "SASL_PLAINTEXT")
		assert.Contains(t, render, "Username")
		assert.Contains(t, render, "Password")
	})

	t.Run("After selecting SASL_PLAINTEXT auth method and going back to select none username and password fields are not visible anymore", func(t *testing.T) {
		// given
		programKtx := kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &programKtx)
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
		render := createEnvPage.View(&programKtx, tests.TestRenderer)
		assert.NotContains(t, render, "Username")
		assert.NotContains(t, render, "Password")
	})

	t.Run("Selecting SASL_PLAINTEXT auth method and filling all SASL fields creates cluster", func(t *testing.T) {
		// given
		programKtx := kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &programKtx)
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
		// and: SSL is enabled
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// and: auth method SASL is selected
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		createEnvPage.Update(cmd())
		// and: security protocol SASL_PLAINTEXT
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		createEnvPage.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(createEnvPage, "username")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(createEnvPage, "password")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// next group
		createEnvPage.Update(cmd())
		// submit
		msgs := tests.Submit(createEnvPage)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, &config.Cluster{}, msgs[0])
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
		}, msgs[0])
	})

	t.Run("Enabling SSL", func(t *testing.T) {
		// given
		programKtx := kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &programKtx)
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
		// and: select SSL enabled
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		createEnvPage.Update(cmd())
		// and: select SASL_SSL security protocol
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(createEnvPage, "username")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(createEnvPage, "password")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// next group
		createEnvPage.Update(cmd())
		// submit
		msgs := tests.Submit(createEnvPage)

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, &config.Cluster{}, msgs[0])
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
		}, msgs[0])
	})

	t.Run("Selecting Schema Registry Enabled opens up schema registry credentials fields", func(t *testing.T) {
		// given
		programKtx := kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &programKtx)
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
		// and: SSL is enabled
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// and: auth method SASL is selected
		cmd = createEnvPage.Update(tests.Key(tea.KeyDown))
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// and: SASL_PLAINTEXT protocol is selected
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// and: enter SASL username
		tests.UpdateKeys(createEnvPage, "username")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		createEnvPage.Update(cmd())
		// and: enter SASL password
		tests.UpdateKeys(createEnvPage, "password")
		cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = createEnvPage.Update(cmd())
		// next group
		createEnvPage.Update(cmd())
		// select schema-registry enabled
		createEnvPage.Update(tests.Key(tea.KeyDown))

		render := createEnvPage.View(&ktx, tests.TestRenderer)
		assert.Contains(t, render, "Schema Registry URL")
		assert.Contains(t, render, "Schema Registry Username")
		assert.Contains(t, render, "Schema Registry Password")

		t.Run("Filling in the credentials and submitting creates the cluster", func(t *testing.T) {
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			// next field
			createEnvPage.Update(cmd())

			// url
			tests.UpdateKeys(createEnvPage, "sr-url")
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			// next field
			createEnvPage.Update(cmd())

			// username
			tests.UpdateKeys(createEnvPage, "sr-user")
			cmd = createEnvPage.Update(tests.Key(tea.KeyEnter))
			// next field
			createEnvPage.Update(cmd())

			// pwd
			tests.UpdateKeys(createEnvPage, "sr-pwd")
			msgs := tests.Submit(createEnvPage)

			// then
			assert.Len(t, msgs, 1)
			assert.IsType(t, &config.Cluster{}, msgs[0])
			// and
			assert.Equal(t, &config.Cluster{
				Name:             "TST",
				Color:            styles.ColorRed,
				Active:           false,
				BootstrapServers: []string{"localhost:9092"},
				SSLEnabled:       true,
				SASLConfig: &config.SASLConfig{
					Username:         "username",
					Password:         "password",
					SecurityProtocol: config.SASLPlaintextSecurityProtocol,
				},
				SchemaRegistry: &config.SchemaRegistryConfig{
					Url:      "sr-url",
					Username: "sr-user",
					Password: "sr-pwd",
				},
			}, msgs[0])
		})
	})

	t.Run("After editing, display notification when checking connectivity", func(t *testing.T) {
		// given
		programKtx := &kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
			Config: &config.Config{
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
			},
		}
		createEnvPage := NewEditForm(mockConnChecker, mockClusterRegisterer{}, programKtx, &FormValues{
			Name:  "prd",
			Color: "#808080",
			Host:  ":9092",
		})

		// when
		createEnvPage.Update(kadmin.ConnCheckStartedMsg{})

		// then
		render := createEnvPage.View(programKtx, tests.TestRenderer)
		assert.Contains(t, render, "Testing cluster connectivity")
	})

	t.Run("After creating a new cluster, display notification when checking connectivity", func(t *testing.T) {
		// given
		createEnvPage := NewForm(mockConnChecker, mockClusterRegisterer{}, &ktx)

		// when
		createEnvPage.Update(kadmin.ConnCheckStartedMsg{})

		// then
		render := createEnvPage.View(&ktx, tests.TestRenderer)
		assert.Contains(t, render, "Testing cluster connectivity")
	})

}
