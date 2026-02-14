package create_cluster_page

import (
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"ktea/ui/tabs"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

var shortcuts []statusbar.Shortcut

type option func(m *Model)

func createClusterPage(options ...option) (*Model, *kontext.ProgramKtx) {
	ktx := tests.NewKontext(
		tests.WithConfig(
			&config.Config{
				Clusters: []config.Cluster{},
			},
		),
	)
	model := NewCreateClusterPage(
		tabs.NewMockClustersTabNavigator(),
		kadmin.MockConnChecker,
		sradmin.MockConnChecker,
		config.MockClusterRegisterer{},
		ktx,
		shortcuts,
		func(certFile string) error {
			return nil
		},
	)
	for _, opt := range options {
		opt(model)
	}
	return model, model.ktx
}

func createEditClusterPage(options ...option) (*Model, *kontext.ProgramKtx) {

	model := &Model{}
	for _, opt := range options {
		opt(model)
	}
	var ktx *kontext.ProgramKtx
	if model.ktx == nil {
		ktx = tests.NewKontext(
			tests.WithConfig(
				&config.Config{
					Clusters: []config.Cluster{},
				},
			),
		)
	} else {
		ktx = model.ktx
	}

	var clusterToEdit config.Cluster
	if model.clusterToEdit != nil {
		clusterToEdit = *model.clusterToEdit
	} else {
		clusterToEdit = config.Cluster{
			Name:             "prd",
			Color:            styles.ColorGreen,
			Active:           false,
			BootstrapServers: []string{"localhost:9092"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
		}
	}

	model = NewEditClusterPage(
		tabs.NewMockClustersTabNavigator(),
		kadmin.MockConnChecker,
		sradmin.MockConnChecker,
		config.MockClusterRegisterer{},
		nil,
		ktx,
		clusterToEdit,
		func(certFile string) error {
			return nil
		},
		WithTitle("Edit Cluster"),
	)

	return model, model.ktx
}

func withContext(ktx *kontext.ProgramKtx) option {
	return func(m *Model) {
		m.ktx = ktx
	}
}

func withClusterToEdit(c *config.Cluster) option {
	return func(m *Model) {
		m.clusterToEdit = c
	}
}

func TestCreateInitialMessageWhenNoClusters(t *testing.T) {
	t.Run("Display info message when no clusters", func(t *testing.T) {
		// given
		page, ktx := createClusterPage()

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "No clusters configured. Please create your first cluster!")
	})

	t.Run("Do not display info message when there are clusters", func(t *testing.T) {
		page, ktx := createClusterPage(
			withContext(
				tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				}),
				),
			),
		)

		render := page.View(ktx, tests.Renderer)

		assert.NotContains(t, render, "No clusters configured. Please create your first cluster!")
	})
}

func TestTabs(t *testing.T) {
	t.Run("Switch to schema registry tab", func(t *testing.T) {
		// given
		page, ktx := createEditClusterPage()

		// when
		page.Update(tests.Key(tea.KeyF5))

		// then: schema-registry tab is visible
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "Schema Registry URL")
		assert.Contains(t, render, "Schema Registry Username")
		assert.Contains(t, render, "Schema Registry Password")
	})

	t.Run("Switch to kafka connect tab", func(t *testing.T) {
		// given
		page, ktx := createEditClusterPage()

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
		page, ktx := createEditClusterPage()
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
		page, ktx := createClusterPage()

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
		page, ktx := createClusterPage()

		// when
		page.Update(tests.Key(tea.KeyEnter))

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "name cannot be empty")
	})

	t.Run("Name must be unique", func(t *testing.T) {
		// given
		page, _ := createClusterPage(withContext(tests.NewKontext(
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "prd",
						Color:            "#808080",
						Active:           true,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
					{
						Name:             "tst",
						Color:            "#F0F0F0",
						Active:           false,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}),
		)))

		// when
		tests.UpdateKeys(page, "prd")
		page.Update(tests.Key(tea.KeyEnter))

		// then
		render := page.View(tests.NewKontext(
			tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "prd",
						Color:            "#808080",
						Active:           true,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
					{
						Name:             "tst",
						Color:            "#F0F0F0",
						Active:           false,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}),
		), tests.Renderer)
		assert.Contains(t, render, "cluster prd already exists, name most be unique")
	})

	t.Run("Host cannot be empty", func(t *testing.T) {
		// given
		page, ktx := createClusterPage()
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

	t.Run("Validate name is unique", func(t *testing.T) {
		// given
		clusterToEdit := config.Cluster{
			Name:             "prd",
			Color:            "#808080",
			Active:           true,
			BootstrapServers: []string{":19092"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			TLSConfig: config.TLSConfig{
				Enable: false,
			},
		}
		page, ktx := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								clusterToEdit,
								{
									Name:             "prod",
									Color:            "#808080",
									Active:           false,
									BootstrapServers: []string{":19092"},
									SASLConfig: config.SASLConfig{
										AuthMethod: config.AuthMethodNone,
									},
									TLSConfig: config.TLSConfig{
										Enable: false,
									},
								},
							},
						},
					),
				),
			),
			withClusterToEdit(&clusterToEdit),
		)

		kb := tests.NewKeyboard(page)
		// when: change name from prd to prod
		kb.Backspace().Backspace().Backspace().Type("prod").Enter()

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "prod already exists, name most be unique")
	})
}

func TestCreateCluster(t *testing.T) {
	t.Run("Transport Plaintext and Auth Method None", func(t *testing.T) {
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
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			})),
			shortcuts,
			func(certFile string) error {
				return nil
			},
		)
		kb := tests.NewKeyboard(page)
		// and: name is entered
		kb.Type("TST").Enter()
		// and: select Color
		kb.Enter()
		// and: host is entered
		kb.Type("localhost:9092").Enter()
		// and: transport Plaintext
		kb.Enter()
		// and
		msgs := kb.Submit()

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:                 "TST",
			Color:                styles.ColorGreen,
			Active:               false,
			BootstrapServers:     []string{"localhost:9092"},
			TLSConfig:            config.TLSConfig{Enable: false},
			SASLConfig:           config.SASLConfig{AuthMethod: config.AuthMethodNone},
			KafkaConnectClusters: nil,
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("Transport SSL and Auth Method SASL_PLAINTEXT", func(t *testing.T) {
		// given
		page, _ := createClusterPage(
			withContext(tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))))
		kb := tests.NewKeyboard(page)
		// and: name is entered
		kb.Type("TST").Enter()
		// and: select Color
		kb.Right().Enter()
		// and: host is entered
		kb.Type("localhost:9092").Enter()
		// and: transport SSL
		kb.Down().Enter()
		// and: verification broker cert
		kb.Enter()
		// and: CA Cert path is entered
		kb.Type("path/to/cert.crt").Enter()
		// and: Auth method SASL_PLAINTEXT
		kb.Down().Enter()
		// and: Auth method username
		kb.Type("john").Enter()
		// and: Auth method pwd
		kb.Type("secret")

		msgs := kb.Submit()

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "TST",
			Color:            styles.ColorBlue,
			Active:           false,
			BootstrapServers: []string{"localhost:9092"},
			SchemaRegistry:   nil,
			TLSConfig: config.TLSConfig{
				Enable:     true,
				SkipVerify: false,
				CACertPath: "path/to/cert.crt",
			},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodSASLPlaintext,
				Username:   "john",
				Password:   "secret",
			},
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("Display notification when cluster has been created", func(t *testing.T) {
		// given
		page, _ := createClusterPage()

		// when
		cluster := config.Cluster{
			Name:             "production",
			Color:            styles.ColorGreen,
			Active:           false,
			BootstrapServers: []string{"localhost:9093"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			SchemaRegistry: nil,
			TLSConfig:      config.TLSConfig{Enable: false},
		}
		page.Update(
			config.ClusterRegisteredMsg{
				Cluster: &cluster,
			},
		)

		// then
		render := page.View(
			tests.NewKontext(
				tests.WithConfig(
					&config.Config{
						Clusters: []config.Cluster{cluster},
					},
				),
			),
			tests.Renderer)

		assert.Contains(t, render, "Cluster registered! <ESC> to go back or <F5> to add a schema registry.")
	})
}

func TestClusterForm(t *testing.T) {

	t.Run("Transport", func(t *testing.T) {

		t.Run("selecting TLS displays TLS fields", func(t *testing.T) {
			// given
			ktx := tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))
			page, _ := createClusterPage(withContext(ktx))
			kb := tests.NewKeyboard(page)
			// and: name is entered
			kb.Type("TST").Enter()
			// and: select Color
			kb.Enter()
			// and: Host is entered
			kb.Type("localhost:9092").Enter()
			// verify
			render := page.View(ktx, tests.Renderer)
			assert.NotContains(t, render, "> TLS")
			assert.NotContains(t, render, "Verification")
			assert.NotContains(t, render, "> Verify Broker Certificate")
			assert.NotContains(t, render, "Skip verification (INSECURE)")

			// when: selecting TLS
			kb.Down()

			// then
			render = page.View(ktx, tests.Renderer)
			assert.Contains(t, render, "> TLS")
			assert.Contains(t, render, "Verification")
			assert.Contains(t, render, "> Verify Broker Certificate")
			assert.Contains(t, render, "Skip verification (INSECURE)")
			assert.Contains(t, render, "Path to Broker CA Certificate")

			t.Run("deselecting TLS hides TLS fields", func(t *testing.T) {
				// and: deselecting TLS (selecting Plaintext)
				kb.Up()

				// then
				render = page.View(ktx, tests.Renderer)
				assert.Contains(t, render, "> Plaintext")
				assert.NotContains(t, render, "> TLS")
				assert.NotContains(t, render, "Verification")
				assert.NotContains(t, render, "> Verify Broker Certificate")
				assert.NotContains(t, render, "Skip verification (INSECURE)")
				assert.NotContains(t, render, "Path to Broker CA Certificate")
			})
		})
	})

	t.Run("Authentication Method", func(t *testing.T) {

		t.Run("selecting SASL_PLAINTEXT displays username and password fields", func(t *testing.T) {
			// given
			ktx := tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))
			page, _ := createClusterPage(withContext(ktx))
			kb := tests.NewKeyboard(page)
			// and: name is entered
			kb.Type("TST").Enter()
			// and: select Color
			kb.Enter()
			// and: Host is entered
			kb.Type("localhost:9092").Enter()
			// and: Transport Plaintext
			kb.Enter()
			// and: Auth Method SASL_PLAINTEXT
			kb.Down()

			// then
			render := page.View(ktx, tests.Renderer)
			assert.Contains(t, render, "SASL_PLAINTEXT")
			assert.Contains(t, render, "SASL username")
			assert.Contains(t, render, "SASL password")

			t.Run("deselecting SASL_PLAINTEXT to select NONE hides username and password fields", func(t *testing.T) {
				// given
				page, _ := createClusterPage(withContext(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				}))))
				kb := tests.NewKeyboard(page)
				// and: name is entered
				kb.Type("TST").Enter()
				// and: select Color
				kb.Enter()
				// and: Host is entered
				kb.Type("localhost:9092").Enter()
				// and: Transport TLS
				kb.Down().Enter()
				// and: Verify Broker
				kb.Enter()
				// and: CA cert
				kb.Type("path/to/cert.crt").Enter()
				// and: Auth Method SASL_PLAINTEXT
				kb.Down()

				assert.Contains(t, render, "SASL username")
				assert.Contains(t, render, "SASL password")

				// and: Auth Method NONE
				kb.Up()

				// then
				render := page.View(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				})), tests.Renderer)
				// then: Auth Method NONE is still selected
				assert.Contains(t, render, "> NONE")
				assert.NotContains(t, render, "SASL username")
				assert.NotContains(t, render, "SASL password")
				assert.Contains(t, render, "> Verify Broker Certificate")
				assert.Contains(t, render, "> path/to/cert.crt")
			})

			t.Run("deselecting SASL_PLAINTEXT to select NONE keeps previous selected fields state", func(t *testing.T) {
				// given
				page, _ := createClusterPage(withContext(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				}))))
				kb := tests.NewKeyboard(page)
				// and: name is entered
				kb.Type("TST").Enter()
				// and: select Color
				kb.Enter()
				// and: Host is entered
				kb.Type("localhost:9092").Enter()
				// and: Transport TLS
				kb.Down().Enter()
				// and: Skip Verify
				kb.Down().Enter()
				// and: Auth Method SASL_PLAINTEXT
				kb.Down()

				render := page.View(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				})), tests.Renderer)

				assert.Contains(t, render, "SASL username")
				assert.Contains(t, render, "SASL password")

				kb.Up()

				// then
				render = page.View(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				})), tests.Renderer)
				assert.Contains(t, render, "┃ > NONE")
				assert.Contains(t, render, "> Skip verification (INSECURE)")

				page, _ = createClusterPage(withContext(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				}))))
				kb = tests.NewKeyboard(page)
				// and: name is entered
				kb.Type("TST").Enter()
				// and: select Color
				kb.Enter()
				// and: Host is entered
				kb.Type("localhost:9092").Enter()
				// and: Transport Plaintext
				kb.Enter()
				// and: Auth Method SASL_PLAINTEXT
				kb.Down()

				render = page.View(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				})), tests.Renderer)

				assert.Contains(t, render, "SASL username")
				assert.Contains(t, render, "SASL password")

				kb.Up()

				render = page.View(tests.NewKontext(tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "PRD",
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				})), tests.Renderer)
				assert.Contains(t, render, "> Plaintext")
				assert.Contains(t, render, "┃ > NONE")
			})
		})

		t.Run("filling all SASL fields creates cluster", func(t *testing.T) {
			// given
			page, _ := createClusterPage(
				withContext(
					tests.NewKontext(tests.WithConfig(&config.Config{
						Clusters: []config.Cluster{
							{
								Name:             "PRD",
								BootstrapServers: []string{"localhost:9092"},
								SASLConfig: config.SASLConfig{
									AuthMethod: config.AuthMethodNone,
								},
							},
						},
					}))))
			kb := tests.NewKeyboard(page)
			// and: name is entered
			kb.Type("TST").Enter()
			// and: select Color
			kb.Enter()
			// and: Host is entered
			kb.Type("localhost:9092").Enter()
			// and: Transport Plaintext
			kb.Enter()
			// and: Auth Method SASL_PLAINTEXT
			kb.Down().Enter()
			// and: username is entered
			kb.Type("SASL username").Enter()
			// and: username is entered
			kb.Type("SASL password").Enter()
			// and
			msgs := kb.Submit()

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
				TLSConfig: config.TLSConfig{
					Enable:     false,
					SkipVerify: false,
					CACertPath: "",
				},
				SASLConfig: config.SASLConfig{
					Username:   "SASL username",
					Password:   "SASL password",
					AuthMethod: config.AuthMethodSASLPlaintext,
				},
			}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
		})
	})

	t.Run("C-r resets form", func(t *testing.T) {
		// given
		page, ktx := createClusterPage(
			withContext(tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))))
		kb := tests.NewKeyboard(page)
		// and: name is entered
		kb.Type("TST").Enter()
		// and: select Color
		kb.Right().Enter()
		// and: host is entered
		kb.Type("localhost:9092").Enter()
		// and: transport SSL
		kb.Down().Enter()
		// and: verification broker cert
		kb.Enter()
		// and: CA Cert path is entered
		kb.Type("/").Enter()
		// and: Auth Method SASL_PLAINTEXT
		kb.Down().Enter()
		// and: username is entered
		kb.Type("username-john").Enter()
		// and: username is entered
		kb.Type("SASL password").Enter()

		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "TST")
		assert.Contains(t, render, "blue")
		assert.Contains(t, render, "localhost:9092")
		assert.Contains(t, render, "> TLS")
		assert.Contains(t, render, "> SASL_PLAINTEXT")
		assert.Contains(t, render, "username-john")
		assert.Contains(t, render, "********")

		// when
		page.Update(tests.Key(tea.KeyCtrlR))

		// then
		render = page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "┃ Name")
		assert.NotContains(t, render, "TST")
		assert.NotContains(t, render, "blue")
		assert.Contains(t, render, "green")
		assert.NotContains(t, render, "localhost:9092")
		assert.NotContains(t, render, "> TLS")
		assert.NotContains(t, render, "> SASL_PLAINTEXT")
		assert.NotContains(t, render, "username-john")
		assert.NotContains(t, render, "********")
	})

	t.Run("Check connectivity after submitting cluster", func(t *testing.T) {
		// given
		page, ktx := createClusterPage()

		// when
		page.Update(kadmin.ConnCheckStartedMsg{})

		// then
		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "Testing cluster connectivity")
	})
}

func TestEditClusterForm(t *testing.T) {

	t.Run("Sets title", func(t *testing.T) {
		// given
		page, _ := createEditClusterPage(
			withContext(tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "prd",
						Color:            "#808080",
						Active:           true,
						BootstrapServers: []string{":19092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
					{
						Name:             "tst",
						Color:            "#F0F0F0",
						Active:           false,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))))

		// when
		title := page.Title()

		// then
		assert.Equal(t, "Edit Cluster", title)
	})

	t.Run("Initial focus on first (name) field", func(t *testing.T) {
		// given
		cluster := config.Cluster{
			Name:             "prd",
			Color:            "#808080",
			Active:           true,
			BootstrapServers: []string{":19092"},
			TLSConfig: config.TLSConfig{
				Enable:     true,
				SkipVerify: false,
				CACertPath: "path/to/ca.cert",
			},
			SASLConfig: config.SASLConfig{
				Username:   "john",
				Password:   "doe",
				AuthMethod: config.AuthMethodSASLPlaintext,
			},
		}
		page, ktx := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								cluster,
							},
						},
					),
				),
			),
			withClusterToEdit(&cluster),
		)

		// when
		page.Update(tests.Key(tea.KeyCtrlE))
		render := page.View(ktx, tests.Renderer)

		// then
		assert.Contains(t, render, "┃ Name")
	})

	t.Run("Update field values", func(t *testing.T) {
		// given
		cluster := config.Cluster{
			Name:             "prd",
			Color:            "#808080",
			Active:           true,
			BootstrapServers: []string{":19092"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			TLSConfig: config.TLSConfig{
				Enable: false,
			},
		}
		page, _ := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								cluster,
							},
						},
					),
				),
			),
			withClusterToEdit(&cluster),
		)

		kb := tests.NewKeyboard(page)
		// when: change name from prd to prod
		kb.Backspace().Backspace().Backspace().Type("prod").Enter()
		// and: change color to orange
		kb.Right().Right().Enter()
		// and: update host
		for i := 0; i < len("localhost:9092"); i++ {
			kb.Backspace()
		}
		kb.Type("localhost:9091").Enter()
		// and: change transport to TLS
		kb.Down().Enter()
		// and: verify broker
		kb.Enter()
		// and: and ca file path

		kb.Type("path/to/ca.cert").Enter()
		// and: auth method SASL_PLAINTEXT
		kb.Down().Enter()
		// and: username john
		kb.Type("john").Enter()
		// and: pwd secret
		kb.Type("secret")

		msgs := kb.Submit()

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "prd",
			Color:            styles.ColorOrange,
			Active:           false,
			BootstrapServers: []string{"localhost:9091"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodSASLPlaintext,
				Username:   "john",
				Password:   "secret",
			},
			TLSConfig: config.TLSConfig{
				Enable:     true,
				SkipVerify: false,
				CACertPath: "path/to/ca.cert",
			},
			SchemaRegistry: nil,
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("Pre-fill fields for Auth Method None", func(t *testing.T) {
		// given
		cluster := config.Cluster{
			Name:             "prd",
			Color:            "#808080",
			Active:           true,
			BootstrapServers: []string{":19092"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			TLSConfig: config.TLSConfig{
				Enable: false,
			},
		}
		page, ktx := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								cluster,
							},
						},
					),
				),
			),
			withClusterToEdit(&cluster),
		)

		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "> NONE")
	})

	t.Run("Pre-fill fields for TLSConfig", func(t *testing.T) {
		// given
		cluster := config.Cluster{
			Name:             "prd",
			Color:            "#808080",
			Active:           true,
			BootstrapServers: []string{":19092"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			TLSConfig: config.TLSConfig{
				Enable:     true,
				SkipVerify: false,
				CACertPath: "path/to/ca.cert",
			},
		}
		page, ktx := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								cluster,
							},
						},
					),
				),
			),
			withClusterToEdit(&cluster),
		)

		render := page.View(ktx, tests.Renderer)
		assert.Contains(t, render, "> TLS")
		assert.Contains(t, render, "> Verify Broker Certificate")
		assert.Contains(t, render, "> path/to/ca.cert")
	})

	t.Run("Check connectivity upon updating", func(t *testing.T) {
		// given
		clusterToEdit := config.Cluster{
			Name:             "prd",
			Color:            "#808080",
			Active:           true,
			BootstrapServers: []string{"localhost:9092"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			TLSConfig: config.TLSConfig{
				Enable:     false,
				SkipVerify: false,
				CACertPath: "",
			},
		}
		page, _ := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								clusterToEdit,
								{
									Name:             "tst",
									Color:            "#F0F0F0",
									Active:           false,
									BootstrapServers: nil,
									SASLConfig: config.SASLConfig{
										AuthMethod: config.AuthMethodNone,
									},
								},
							},
						},
					),
				),
			),
			withClusterToEdit(&clusterToEdit),
		)

		kb := tests.NewKeyboard(page)
		// when: leave name as is
		kb.Enter()
		// and: leave color as is
		kb.Enter()
		// and: update host
		for i := 0; i < len("localhost:9092"); i++ {
			kb.Backspace()
		}
		kb.Type("localhost:9091").Enter()
		// and: leave transport as is
		kb.Enter()
		// and: leave auth method as is
		kb.Enter()
		// and
		msgs := kb.Submit()

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])
		// and
		assert.Equal(t, &config.Cluster{
			Name:             "prd",
			Color:            styles.ColorGreen,
			Active:           false,
			BootstrapServers: []string{"localhost:9091"},
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			TLSConfig: config.TLSConfig{
				Enable: false,
			},
			SchemaRegistry: nil,
		}, msgs[0].(kadmin.MockConnectionCheckedMsg).Cluster)
	})

	t.Run("After editing, display notification when checking connectivity", func(t *testing.T) {
		// given
		page, _ := createEditClusterPage(
			withContext(tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "prd",
						Color:            "#808080",
						Active:           true,
						BootstrapServers: []string{":19092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
					{
						Name:             "tst",
						Color:            "#F0F0F0",
						Active:           false,
						BootstrapServers: nil,
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))))

		// when
		page.Update(kadmin.ConnCheckStartedMsg{})

		// then
		render := page.View(
			tests.NewKontext(
				tests.WithConfig(&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{":19092"},
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
						{
							Name:             "tst",
							Color:            "#F0F0F0",
							Active:           false,
							BootstrapServers: nil,
							SASLConfig: config.SASLConfig{
								AuthMethod: config.AuthMethodNone,
							},
						},
					},
				})),
			tests.Renderer,
		)
		assert.Contains(t, render, "Testing cluster connectivity")
	})

	t.Run("Edit when there was no initial schema registry created", func(t *testing.T) {
		// given
		page, _ := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								{
									Name:             "prd",
									Color:            "#808080",
									Active:           true,
									BootstrapServers: []string{":19092"},
									SASLConfig: config.SASLConfig{
										AuthMethod: config.AuthMethodNone,
									},
								},
								{
									Name:             "tst",
									Color:            "#F0F0F0",
									Active:           false,
									BootstrapServers: nil,
									SASLConfig: config.SASLConfig{
										AuthMethod: config.AuthMethodNone,
									},
								},
							},
						},
					),
				),
			),
		)

		kb := tests.NewKeyboard(page)
		kb.F5()

		// when
		kb.Type("https://localhost:8081").Enter()
		// and
		kb.Type("sr-username").Enter()
		// and
		kb.Type("sr-password").Enter()
		// and
		msgs := kb.Submit()

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
		clusters := []config.Cluster{
			{
				Name:             "prd",
				Color:            "#808080",
				Active:           true,
				BootstrapServers: []string{":19092"},
				SASLConfig: config.SASLConfig{
					AuthMethod: config.AuthMethodNone,
				},
			},
			{
				Name:             "tst",
				Color:            "#F0F0F0",
				Active:           false,
				BootstrapServers: nil,
				SASLConfig: config.SASLConfig{
					AuthMethod: config.AuthMethodNone,
				},
			},
		}
		page, _ := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{Clusters: clusters},
					),
				),
			),
		)

		// when
		page.Update(kadmin.ConnCheckErrMsg{Err: fmt.Errorf("kafka: client has run out of available brokers to talk to")})

		// then
		render := page.View(
			tests.NewKontext(
				tests.WithConfig(
					&config.Config{Clusters: clusters},
				),
			),
			tests.Renderer,
		)
		assert.Contains(t, render, "Failed to Update Cluster: kafka: client has run out of available brokers to talk to")
	})

	t.Run("Display notification when cluster has been updated", func(t *testing.T) {
		// given
		page, _ := createEditClusterPage(
			withContext(
				tests.NewKontext(
					tests.WithConfig(
						&config.Config{
							Clusters: []config.Cluster{
								{
									Name:             "prd",
									Color:            "#808080",
									Active:           true,
									BootstrapServers: []string{":19092"},
									SASLConfig: config.SASLConfig{
										AuthMethod: config.AuthMethodNone,
									},
								},
								{
									Name:             "tst",
									Color:            "#F0F0F0",
									Active:           false,
									BootstrapServers: nil,
									SASLConfig: config.SASLConfig{
										AuthMethod: config.AuthMethodNone,
									},
								},
							},
						},
					),
				),
			),
		)

		// when
		page.Update(config.ClusterRegisteredMsg{
			Cluster: &config.Cluster{
				Name:             "production",
				Color:            styles.ColorGreen,
				Active:           false,
				BootstrapServers: []string{"localhost:9093"},
				SASLConfig: config.SASLConfig{
					AuthMethod: config.AuthMethodNone,
				},
				SchemaRegistry: nil,
				TLSConfig:      config.TLSConfig{Enable: false},
			},
		})

		// then
		render := page.View(tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "prd",
					Color:            "#808080",
					Active:           true,
					BootstrapServers: []string{":19092"},
					SASLConfig: config.SASLConfig{
						AuthMethod: config.AuthMethodNone,
					},
				},
				{
					Name:             "tst",
					Color:            "#F0F0F0",
					Active:           false,
					BootstrapServers: nil,
					SASLConfig: config.SASLConfig{
						AuthMethod: config.AuthMethodNone,
					},
				},
			},
		})), tests.Renderer)
		assert.Contains(t, render, "Cluster updated!")
	})
}

func TestCreateSchemaRegistry(t *testing.T) {
	// given
	page, ktx := createClusterPage()

	kb := tests.NewKeyboard(page)

	t.Run("Check connectivity before registering the schema registry", func(t *testing.T) {
		// and: enter name
		kb.Type("TST").Enter()
		// and: select Color
		kb.Enter()
		// and: Host is entered
		kb.Type("localhost:9092").Enter()
		// and: transport Plaintext
		kb.Enter()
		// and: auth method SASL_PLAINTEXT
		kb.Down().Enter()
		// and: enter SASL username
		kb.Type("SASL username").Enter()
		// and: enter SASL password
		kb.Type("SASL password").Enter()
		// submit
		kb.Submit()
		cluster := config.Cluster{
			Name:             "cluster-name",
			Color:            styles.ColorGreen,
			Active:           false,
			BootstrapServers: nil,
			SASLConfig: config.SASLConfig{
				AuthMethod: config.AuthMethodNone,
			},
			SchemaRegistry: nil,
			TLSConfig:      config.TLSConfig{Enable: false},
		}
		page.Update(config.ClusterRegisteredMsg{
			Cluster: &cluster,
		})

		// and: switch to the schema-registry tab
		kb.F5()

		// and: schema registry url
		kb.Type("sr-url").Enter()
		// and: schema registry username
		kb.Type("sr-username").Enter()
		// and: schema registry pwd
		kb.Type("sr-password")
		// and
		msgs := kb.Submit()

		// then
		assert.Len(t, msgs, 1)
		assert.IsType(t, sradmin.MockConnectionCheckedMsg{}, msgs[0])
		assert.EqualValues(t, &config.SchemaRegistryConfig{
			Url:      "sr-url",
			Username: "sr-username",
			Password: "sr-password",
		}, msgs[0].(sradmin.MockConnectionCheckedMsg).Config)

		t.Run("Display error notification when connection cannot be made", func(t *testing.T) {
			page.Update(sradmin.ConnCheckErrMsg{Err: fmt.Errorf("cannot connect")})

			render := page.View(ktx, tests.Renderer)

			assert.Contains(t, render, "Failed to register schema registry")
		})

		t.Run("Display success notification upon registration", func(t *testing.T) {
			page.Update(config.ClusterRegisteredMsg{
				Cluster: &cluster,
			})

			render := page.View(ktx, tests.Renderer)

			assert.Contains(t, render, "Schema registry registered! <ESC> to go back.")
		})
	})
}
