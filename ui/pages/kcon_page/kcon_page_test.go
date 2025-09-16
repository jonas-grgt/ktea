package kcon_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kcadmin"
	"ktea/tests"
	"ktea/ui/tabs"
	"strings"
	"testing"
	"time"
)

func TestKconsPage(t *testing.T) {

	t.Run("esc goes back to Kafka Connect Clusters page", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		cmd := page.Update(tests.Key(tea.KeyEsc))

		assert.IsType(t, tabs.ToKConsPageCalledMsg{}, cmd())
	})

	t.Run("F5 refreshes connectors list", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		cmd := page.Update(tests.Key(tea.KeyF5))

		assert.Contains(t, tests.ExecuteBatchCmd(cmd), kcadmin.ListActiveConnectorsCalledMsg{})

		t.Run("F5 blocks until connectors are loaded", func(t *testing.T) {
			cmd := page.Update(tests.Key(tea.KeyF5))

			assert.Nil(t, cmd)
		})

	})

	t.Run("Default sort by Name Asc", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		render := page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Name")

		t1Idx := strings.Index(render, "a-my-connector")
		t2Idx := strings.Index(render, "b-my-connector")
		t3Idx := strings.Index(render, "z-my-connector")

		assert.Less(t, t1Idx, t3Idx)
		assert.Less(t, t1Idx, t3Idx)
		assert.Less(t, t2Idx, t3Idx)
	})

	t.Run("Toggle sort by Name", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.Renderer)

		render = page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Name")

		t1Idx := strings.Index(render, "c-my-connector")
		t2Idx := strings.Index(render, "b-my-connector")
		t3Idx := strings.Index(render, "a-my-connector")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t1Idx, t3Idx)
		assert.Less(t, t2Idx, t3Idx)
	})

	t.Run("Esc hides toggled sort by", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))

		render := page.View(tests.NewKontext(), tests.Renderer)
		assert.Contains(t, render, "Name ▲")

		cmd := page.Update(tests.Key(tea.KeyEsc))

		assert.Nil(t, cmd)
		render = page.View(tests.NewKontext(), tests.Renderer)
		assert.NotContains(t, render, "Name ▲")
	})

	t.Run("Search for connector", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		page.Update(tests.Key('/'))
		page.Update(tests.Key('a'))

		render := page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "1/3")
		assert.Greater(t, strings.Index(render, "a-my-connector"), -1)
		assert.Equal(t, strings.Index(render, "b-my-connector"), -1)
		assert.Equal(t, strings.Index(render, "z-my-connector"), -1)
	})

	t.Run("Manage connector", func(t *testing.T) {
		page, _ := New(
			tabs.NewMockKConTabNavigator(),
			kcadmin.NewMock(),
			"connector-name",
		)

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "PAUSED",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		t.Run("cannot pause an already paused connector", func(t *testing.T) {
			page.Update(tests.Key('P'))

		})

		t.Run("pause running connector", func(t *testing.T) {
			page.View(tests.Kontext, tests.Renderer)
			cmds := page.Update(tests.Key('P'))

			msgs := tests.ExecuteBatchCmd(cmds)
			assert.Len(t, msgs, 1)
			assert.IsType(t, kcadmin.ConnectorPauseCalledMsg{}, msgs[0])
			assert.Equal(t, "a-my-connector", msgs[0].(kcadmin.ConnectorPauseCalledMsg).Name)
		})

		t.Run("resume running connector", func(t *testing.T) {
			page.View(tests.Kontext, tests.Renderer)
			cmds := page.Update(tests.Key('R'))

			msgs := tests.ExecuteBatchCmd(cmds)
			assert.Len(t, msgs, 1)
			assert.IsType(t, kcadmin.ConnectorResumeCalledMsg{}, msgs[0])
			assert.Equal(t, "a-my-connector", msgs[0].(kcadmin.ConnectorResumeCalledMsg).Name)

			page.Update(kcadmin.ResumingStartedMsg{})

			render := page.View(tests.Kontext, tests.Renderer)
			assert.Contains(t, render, "RESUMING")

			page.Update(ConnectorStateChanged{
				&kcadmin.Connectors{
					"z-my-connector": {
						Status: kcadmin.ConnectorStatus{
							Name: "z-my-connector",
							Connector: kcadmin.ConnectorState{
								State:    "RUNNING",
								WorkerID: "10.0.0.6:8083",
							},
							Tasks: nil,
							Type:  "sink",
						},
					},
					"b-my-connector": {
						Status: kcadmin.ConnectorStatus{
							Name: "b-my-connector",
							Connector: kcadmin.ConnectorState{
								State:    "RUNNING",
								WorkerID: "10.0.0.6:8083",
							},
							Tasks: nil,
							Type:  "sink",
						},
					},
					"a-my-connector": {
						Status: kcadmin.ConnectorStatus{
							Name: "a-my-connector",
							Connector: kcadmin.ConnectorState{
								State:    "RUNNING",
								WorkerID: "10.0.0.6:8083",
							},
							Tasks: nil,
							Type:  "sink",
						},
					},
				},
			})

			render = page.View(tests.Kontext, tests.Renderer)
			assert.NotContains(t, render, "RESUMING")
		})
	})
}

func TestTickMsgHandled(t *testing.T) {
	initialSpinner := spinner.New(spinner.WithSpinner(spinner.Dot))
	model, _ := New(
		tabs.NewMockKConTabNavigator(),
		kcadmin.NewMock(),
		"connector-name",
	)
	model.stsSpinner = &initialSpinner

	cmd := model.Update(spinner.TickMsg{
		ID:   initialSpinner.ID(),
		Time: time.Now(),
	})

	assert.NotNil(t, model.stsSpinner, "expected stsSpinner to still be set after update")
	assert.Equal(t, model.stsSpinner.ID(), initialSpinner.ID(),
		fmt.Sprintf("spinner ID should not change; got %d want %d", model.stsSpinner.ID(), initialSpinner.ID()))
	assert.NotNil(t, cmd, "expected a tea.Cmd to be returned after TickMsg")
	assert.IsType(t, spinner.TickMsg{}, cmd())
}

func TestTickMsgIgnored(t *testing.T) {
	initialSpinner := spinner.New(spinner.WithSpinner(spinner.Dot))
	model, _ := New(
		tabs.NewMockKConTabNavigator(),
		kcadmin.NewMock(),
		"connector-name",
	)
	model.stsSpinner = nil

	cmd := model.Update(spinner.TickMsg{
		ID:   initialSpinner.ID(),
		Time: time.Now(),
	})

	assert.Nil(t, model.stsSpinner, "expected stsSpinner to still be nil")
	assert.Nil(t, cmd, "expected no tea.Cmd to be returned after TickMsg")
}

func TestWaitForConnectorResumed_Success(t *testing.T) {
	connectorsChan := make(chan kcadmin.Connectors, 1)
	errChan := make(chan error, 1)
	connectorsChan <- kcadmin.Connectors{
		"my-connector": {
			Status: kcadmin.ConnectorStatus{
				Name: "my-connector",
				Connector: kcadmin.ConnectorState{
					State:    "RUNNING",
					WorkerID: "10.0.0.6:8083",
				},
				Tasks: nil,
				Type:  "sink",
			},
		},
	}

	mockConnectorListingStartedMsg := kcadmin.ConnectorListingStartedMsg{
		Connectors: connectorsChan,
		Err:        errChan,
	}

	var listActiveConnectorsResponse = kcadmin.WithListActiveConnectorsResponse(
		mockConnectorListingStartedMsg,
	)
	resumeDeadline := time.Now().Add(30 * time.Second)

	model := &Model{
		kca:                        kcadmin.NewMock(listActiveConnectorsResponse),
		stateChangingConnectorName: "my-connector",
		resumeDeadline:             &resumeDeadline,
	}

	msg := model.waitForConnectorState("RUNNING")

	resumedMsg, ok := msg.(ConnectorStateChanged)
	if !ok {
		t.Fatalf("Expected ResumedMsg, got %T", msg)
	}
	if resumedMsg.Connectors == nil || (*resumedMsg.Connectors)["my-connector"].Status.Connector.State != "RUNNING" {
		t.Errorf("Connector state not RUNNING")
	}
}

func TestWaitForConnectorResumed_RetriesUntilRunning(t *testing.T) {
	// Step 1: simulate first call returns connector in PAUSED state
	pausedConnectors := make(chan kcadmin.Connectors, 1)
	pausedConnectors <- kcadmin.Connectors{
		"my-connector": {
			Status: kcadmin.ConnectorStatus{
				Name: "my-connector",
				Connector: kcadmin.ConnectorState{
					State: "PAUSED",
				},
			},
		},
	}

	// Step 2: simulate second call returns connector in RUNNING state
	runningConnectors := make(chan kcadmin.Connectors, 1)
	runningConnectors <- kcadmin.Connectors{
		"my-connector": {
			Status: kcadmin.ConnectorStatus{
				Name: "my-connector",
				Connector: kcadmin.ConnectorState{
					State: "RUNNING",
				},
			},
		},
	}

	// Step 3: inject logic that returns first paused, then running
	callCount := 0
	mock := kcadmin.NewMock(
		kcadmin.WithListActiveConnectorsFunc(func() tea.Msg {
			callCount++
			switch callCount {
			case 1:
				return kcadmin.ConnectorListingStartedMsg{
					Connectors: pausedConnectors,
					Err:        make(chan error, 1),
				}
			case 2:
				return kcadmin.ConnectorListingStartedMsg{
					Connectors: runningConnectors,
					Err:        make(chan error, 1),
				}
			default:
				t.Fatalf("ListActiveConnectors called more than expected")
				return nil
			}
		}),
	)

	resumeDeadline := time.Now().Add(30 * time.Second)
	model := &Model{
		kca:                        mock,
		stateChangingConnectorName: "my-connector",
		resumeDeadline:             &resumeDeadline,
	}

	start := time.Now()
	timeout := 5 * time.Second

	var msg tea.Msg
	for {
		msg = model.waitForConnectorState("RUNNING")
		if _, ok := msg.(ConnectorStateChanged); ok {
			break
		}
		if time.Since(start) > timeout {
			t.Fatal("Timed out waiting for connector to resume")
		}
	}

	if resumedMsg, ok := msg.(ConnectorStateChanged); !ok {
		t.Fatalf("Expected ResumedMsg, got %T", msg)
	} else {
		state := (*resumedMsg.Connectors)["my-connector"].Status.Connector.State
		if state != "RUNNING" {
			t.Errorf("Expected RUNNING, got %s", state)
		}
	}

	if callCount != 2 {
		t.Errorf("Expected ListActiveConnectors to be called twice, got %d", callCount)
	}
}

func TestWaitForConnectorResumed_SetsResumeDeadline(t *testing.T) {
	connectorsChan := make(chan kcadmin.Connectors, 1)
	errChan := make(chan error, 1)
	connectorsChan <- kcadmin.Connectors{
		"my-connector": {
			Status: kcadmin.ConnectorStatus{
				Name: "my-connector",
				Connector: kcadmin.ConnectorState{
					State:    "RUNNING",
					WorkerID: "10.0.0.6:8083",
				},
				Tasks: nil,
				Type:  "sink",
			},
		},
	}

	mockConnectorListingStartedMsg := kcadmin.ConnectorListingStartedMsg{
		Connectors: connectorsChan,
		Err:        errChan,
	}

	mock := kcadmin.NewMock(
		kcadmin.WithListActiveConnectorsResponse(mockConnectorListingStartedMsg),
	)

	model := &Model{
		kca:                        mock,
		stateChangingConnectorName: "my-connector",
		resumeDeadline:             nil,
	}

	_ = model.waitForConnectorState("RUNNING")

	if model.resumeDeadline == nil {
		t.Fatal("Expected resumeDeadline to be set, but it is still nil")
	}

	now := time.Now()
	if model.resumeDeadline.Before(now) || model.resumeDeadline.After(now.Add(31*time.Second)) {
		t.Errorf("resumeDeadline was set to an unexpected value: %v", *model.resumeDeadline)
	}
}

func TestWaitForConnectorResumed_TimesOut(t *testing.T) {
	mock := kcadmin.NewMock(
		kcadmin.WithListActiveConnectorsFunc(func() tea.Msg {
			t.Fatal("ListActiveConnectors should not be called after timeout")
			return nil
		}),
	)

	// Set the deadline to a time in the past
	past := time.Now().Add(-1 * time.Second)

	model := &Model{
		kca:                        mock,
		stateChangingConnectorName: "my-connector",
		resumeDeadline:             &past,
	}

	msg := model.waitForConnectorState("RUNNING")

	errMsg, ok := msg.(kcadmin.ConnectorListingErrMsg)
	if !ok {
		t.Fatalf("Expected ConnectorListingErrMsg, got %T", msg)
	}

	if errMsg.Err == nil || !strings.Contains(errMsg.Err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", errMsg.Err)
	}

	if model.resumeDeadline != nil {
		t.Errorf("Expected resumeDeadline to be cleared (nil), but got: %v", model.resumeDeadline)
	}
}
