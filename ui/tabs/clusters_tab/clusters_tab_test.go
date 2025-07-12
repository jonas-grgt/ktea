package clusters_tab

import (
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui/pages/clusters_page"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestClustersTab(t *testing.T) {
	var ktx = kontext.ProgramKtx{
		Config: &config.Config{
			Clusters: []config.Cluster{},
		},
		WindowWidth:  100,
		WindowHeight: 100,
	}
	t.Run("When no cluster exists, open create new form", func(t *testing.T) {
		// given
		clustersTab, _ := New(&kontext.ProgramKtx{
			Config:       &config.Config{},
			WindowWidth:  0,
			WindowHeight: 0,
		}, kadmin.MockConnChecker, sradmin.MockConnChecker)

		// when
		render := clustersTab.View(&ktx, tests.TestRenderer)

		// then
		assert.Contains(t, render, "┃ Name")
	})

	t.Run("List clusters page", func(t *testing.T) {
		t.Run("opens when at least one env exists", func(t *testing.T) {
			// given
			programKtx := &kontext.ProgramKtx{
				Config: &config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
				WindowWidth:  100,
				WindowHeight: 100,
			}
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)

			// when
			render := clustersTab.View(programKtx, tests.TestRenderer)

			// then
			assert.Contains(t, render, "prd")
		})

		t.Run("indicates active cluster in list", func(t *testing.T) {
			// given
			programKtx := &kontext.ProgramKtx{
				Config: &config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            styles.ColorRed,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "tst",
							Color:            styles.ColorGreen,
							Active:           true,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "uat",
							Color:            styles.ColorBlue,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
				WindowWidth:     100,
				WindowHeight:    100,
				AvailableHeight: 100,
			}
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)

			// when
			render := clustersTab.View(programKtx, tests.TestRenderer)

			// then
			assert.Regexp(t, "X\\W+tst", render)
			assert.Regexp(t, "│\\W+prd", render)
			assert.Regexp(t, "│\\W+uat", render)
			assert.NotRegexp(t, "X\\W+prd", render)
			assert.NotRegexp(t, "X\\W+uat", render)
		})

		t.Run("enter starts a connectivity check for the selected cluster", func(t *testing.T) {
			// given
			programKtx := &kontext.ProgramKtx{
				Config: &config.Config{
					ConfigIO: &config.InMemoryConfigIO{},
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            styles.ColorRed,
							Active:           true,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "tst",
							Color:            styles.ColorGreen,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "uat",
							Color:            styles.ColorBlue,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
				WindowWidth:     100,
				WindowHeight:    100,
				AvailableHeight: 100,
			}
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			clustersTab.View(programKtx, tests.TestRenderer)

			// when
			clustersTab.Update(tests.Key(tea.KeyDown))
			cmds := clustersTab.Update(tests.Key(tea.KeyEnter))
			msgs := executeBatchCmd(cmds)
			assert.NotNil(t, msgs)

			// then the connectivity check has been started
			assert.IsType(t, kadmin.MockConnectionCheckedMsg{}, msgs[0])

			t.Run("and shows a spinner", func(t *testing.T) {
				clustersTab.Update(kadmin.ConnCheckStartedMsg{
					Cluster: &config.Cluster{
						Name: "tst",
					},
				})

				rendered := clustersTab.View(programKtx, tests.TestRenderer)

				assert.Contains(t, rendered, "Checking connectivity to tst")
			})

			t.Run("and shows failure msg upon connectivity error", func(t *testing.T) {
				clustersTab.Update(kadmin.ConnCheckErrMsg{
					Err: fmt.Errorf("kafka: client has run out of available brokers to talk to: EOF"),
				})

				rendered := clustersTab.View(programKtx, tests.TestRenderer)

				assert.Contains(t, rendered, "Connection check failed: kafka: client has run out of available brokers to talk to: EOF")
			})

			t.Run("and shows success msg upon connection ", func(t *testing.T) {
				clustersTab.Update(kadmin.ConnCheckSucceededMsg{})

				rendered := clustersTab.View(programKtx, tests.TestRenderer)

				assert.Contains(t, rendered, "Connection check succeeded, switching cluster")
			})

			t.Run("and activated is indicated", func(t *testing.T) {
				programKtx.Config.SwitchCluster("tst")
				clustersTab.Update(clusters_page.ClusterSwitchedMsg{
					Cluster: &config.Cluster{
						Name: "tst",
					},
				})

				rendered := clustersTab.View(programKtx, tests.TestRenderer)

				assert.Regexp(t, "X\\W+tst", rendered)
				assert.Regexp(t, "│\\W+prd", rendered)
				assert.Regexp(t, "│\\W+uat", rendered)
				assert.NotRegexp(t, "X\\W+prd", rendered)
			})

			t.Run("Activated cluster is selected", func(t *testing.T) {
				assert.Equal(t, "tst",
					*(clustersTab.active.(*clusters_page.Model)).SelectedCluster())
			})
		})
	})

	t.Run("Delete cluster", func(t *testing.T) {
		cfg := config.New(&config.InMemoryConfigIO{})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "prd",
			Color:      styles.ColorRed,
			Host:       "localhost:9092",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "tst",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "uat",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		programKtx := &kontext.ProgramKtx{
			Config:          cfg,
			WindowWidth:     100,
			WindowHeight:    100,
			AvailableHeight: 100,
		}

		t.Run("/ raises search prompt", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, tests.TestRenderer)

			// when
			clustersTab.Update(tests.Key('/'))

			// then
			render = clustersTab.View(programKtx, tests.TestRenderer)
			assert.Contains(t, render, "┃ >")
		})

		t.Run("F2 raises delete confirmation", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, tests.TestRenderer)

			// when
			clustersTab.Update(tests.Key(tea.KeyF2))

			// then
			render = clustersTab.View(programKtx, tests.TestRenderer)
			assert.Contains(t, render, "🗑️  prd will be deleted permanently")
		})

		t.Run("esc cancels raised delete confirmation", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, tests.TestRenderer)
			// and delete confirmation has been raised
			clustersTab.Update(tests.Key(tea.KeyF2))
			render = clustersTab.View(programKtx, tests.TestRenderer)
			assert.Contains(t, render, "🗑️  prd will be deleted permanently")

			// when
			clustersTab.Update(tests.Key(tea.KeyEsc))

			// then
			render = clustersTab.View(programKtx, tests.TestRenderer)
			assert.NotContains(t, render, "🗑️  prd will be deleted permanently")
		})

		t.Run("enter deletes cluster", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, tests.TestRenderer)

			// when
			clustersTab.Update(tests.Key(tea.KeyDown))
			clustersTab.Update(tests.Key(tea.KeyF2))
			clustersTab.Update(tests.Key('d'))
			cmds := clustersTab.Update(tests.Key(tea.KeyEnter))
			msgs := executeBatchCmd(cmds)
			for _, msg := range msgs {
				clustersTab.Update(msg)
			}

			// then
			render = clustersTab.View(programKtx, tests.TestRenderer)
			assert.NotContains(t, render, "tst")
		})

		t.Run("unable to delete active cluster", func(t *testing.T) {
			// given
			cfg := config.New(&config.InMemoryConfigIO{})
			cfg.RegisterCluster(config.RegistrationDetails{
				Name:       "prd",
				Color:      styles.ColorRed,
				Host:       "localhost:9092",
				AuthMethod: config.NoneAuthMethod,
			})
			cfg.RegisterCluster(config.RegistrationDetails{
				Name:       "tst",
				Color:      styles.ColorRed,
				Host:       "localhost:9093",
				AuthMethod: config.NoneAuthMethod,
			})
			cfg.RegisterCluster(config.RegistrationDetails{
				Name:       "uat",
				Color:      styles.ColorRed,
				Host:       "localhost:9093",
				AuthMethod: config.NoneAuthMethod,
			})
			programKtx := &kontext.ProgramKtx{
				Config:          cfg,
				WindowWidth:     100,
				WindowHeight:    100,
				AvailableHeight: 100,
			}
			// and
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, tests.TestRenderer)

			// when
			clustersTab.Update(tests.Key(tea.KeyF2))
			clustersTab.Update(tests.Key('d'))
			cmds := clustersTab.Update(tests.Key(tea.KeyEnter))
			msgs := executeBatchCmd(cmds)
			for _, msg := range msgs {
				clustersTab.Update(msg)
			}

			// then
			render = clustersTab.View(programKtx, tests.TestRenderer)
			assert.Contains(t, render, "Unable to delete: active cluster")
		})
	})

	t.Run("Edit cluster", func(t *testing.T) {
		cfg := config.New(&config.InMemoryConfigIO{})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:             "prd",
			Color:            styles.ColorRed,
			Host:             "localhost:9092",
			AuthMethod:       config.SASLAuthMethod,
			SecurityProtocol: config.SASLPlaintextSecurityProtocol,
			SSLEnabled:       true,
			Username:         "John",
			Password:         "Doe",
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "tst",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "uat",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		programKtx := &kontext.ProgramKtx{
			Config:       cfg,
			WindowWidth:  100,
			WindowHeight: 100,
		}

		t.Run("c-e opens edit page", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
			// and table has been initialized
			clustersTab.View(programKtx, tests.TestRenderer)

			// when
			clustersTab.Update(tests.Key(tea.KeyCtrlE))

			// then
			render := clustersTab.View(programKtx, tests.TestRenderer)
			assert.Contains(t, render, "> prd")
			assert.Contains(t, render, "> localhost:9092")
			assert.Contains(t, render, "> Enable SSL")

			t.Run("updates shortcuts", func(t *testing.T) {
				assert.Contains(t, render, "Prev. Field:")
			})

			t.Run("c-e on edit page does nothing", func(t *testing.T) {
				cmd := clustersTab.Update(tests.Key(tea.KeyCtrlE))

				tests.ExecuteBatchCmd(cmd)

				render := clustersTab.View(programKtx, tests.TestRenderer)
				assert.Contains(t, render, "> prd")
			})

			t.Run("c-n on edit page does nothing", func(t *testing.T) {
				cmd := clustersTab.Update(tests.Key(tea.KeyCtrlN))

				tests.ExecuteBatchCmd(cmd)

				render := clustersTab.View(programKtx, tests.TestRenderer)
				assert.Contains(t, render, "> prd")
			})
		})
	})

	t.Run("esc does not go back when there are no clusters", func(t *testing.T) {
		// given
		programKtx := &kontext.ProgramKtx{
			Config: &config.Config{
				Clusters: []config.Cluster{},
			},
			WindowWidth:  100,
			WindowHeight: 100,
		}
		var clustersTab, _ = New(programKtx, kadmin.MockConnChecker, sradmin.MockConnChecker)
		clustersTab.View(programKtx, tests.TestRenderer)

		// when
		clustersTab.Update(tests.Key(tea.KeyEsc))

		// then
		render := clustersTab.View(programKtx, tests.TestRenderer)
		assert.Contains(t, render, "┃ Name")
	})
}

func executeBatchCmd(cmd tea.Cmd) []tea.Msg {
	var msgs []tea.Msg
	if cmd == nil {
		return msgs
	}

	msg := cmd()
	if msg == nil {
		return msgs
	}

	// If the message is a batch, process its commands
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, subCmd := range batch {
			if subCmd != nil {
				msgs = append(msgs, executeBatchCmd(subCmd)...)
			}
		}
		return msgs
	}

	// Otherwise, it's a normal message
	msgs = append(msgs, msg)
	return msgs
}
