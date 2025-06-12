package record_details_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/serdes"
	"ktea/tests"
	"ktea/ui/clipper"
	"ktea/ui/components/statusbar"
	"testing"
)

func TestRecordDetailsPage(t *testing.T) {
	t.Run("c-h or arrows toggles focus between content and headers", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Payload:   serdes.DesData{Value: ""},
			Partition: 0,
			Offset:    0,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: kadmin.NewHeaderValue("v2"),
				},
			},
		},
			"",
			clipper.NewMock(),
			tests.NewKontext(),
		)
		// init ui
		m.View(tests.TestKontext, tests.TestRenderer)

		assert.Equal(t, payloadFocus, m.focus)

		m.Update(tests.Key(tea.KeyCtrlH))

		assert.Equal(t, headersFocus, m.focus)

		m.Update(tests.Key(tea.KeyCtrlH))

		assert.Equal(t, payloadFocus, m.focus)

		m.Update(tests.Key(tea.KeyRight))

		assert.Equal(t, headersFocus, m.focus)

		m.Update(tests.Key(tea.KeyLeft))

		assert.Equal(t, payloadFocus, m.focus)
	})

	t.Run("view schema", func(t *testing.T) {
		t.Run("Shortcut not visible when cluster has no SchemaRegistry", func(t *testing.T) {
			ktx := *tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "",
						Color:            "",
						Active:           true,
						BootstrapServers: nil,
						SASLConfig:       nil,
						SchemaRegistry:   nil,
						SSLEnabled:       false,
					},
				},
				ConfigIO: nil,
			}))
			m := New(&kadmin.ConsumerRecord{
				Key:       "",
				Payload:   serdes.DesData{Value: ""},
				Partition: 0,
				Offset:    0,
				Headers: []kadmin.Header{
					{
						Key:   "h1",
						Value: kadmin.NewHeaderValue("v2"),
					},
				},
			},
				"",
				clipper.NewMock(),
				&ktx,
			)

			shortcuts := m.Shortcuts()

			assert.NotContains(t, shortcuts, statusbar.Shortcut{
				Name:       "View Schema",
				Keybinding: "C-s",
			})
		})

		t.Run("Shortcut visible when cluster has SchemaRegistry", func(t *testing.T) {
			ktx := *tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "",
						Color:            "",
						Active:           true,
						BootstrapServers: nil,
						SASLConfig:       nil,
						SchemaRegistry: &config.SchemaRegistryConfig{
							Url:      "http://localhost:8080",
							Username: "john",
							Password: "doe",
						},
						SSLEnabled: false,
					},
				},
				ConfigIO: nil,
			}))
			m := New(&kadmin.ConsumerRecord{
				Key:       "",
				Payload:   serdes.DesData{Value: ""},
				Partition: 0,
				Offset:    0,
				Headers: []kadmin.Header{
					{
						Key:   "h1",
						Value: kadmin.NewHeaderValue("v2"),
					},
				},
			},
				"",
				clipper.NewMock(),
				&ktx,
			)

			shortcuts := m.Shortcuts()

			assert.Contains(t, shortcuts, statusbar.Shortcut{
				Name:       "View Schema",
				Keybinding: "C-s",
			})
		})

		t.Run("Shortcut leads to Schema", func(t *testing.T) {
			ktx := *tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "",
						Color:            "",
						Active:           true,
						BootstrapServers: nil,
						SASLConfig:       nil,
						SchemaRegistry: &config.SchemaRegistryConfig{
							Url:      "http://localhost:8080",
							Username: "john",
							Password: "doe",
						},
						SSLEnabled: false,
					},
				},
				ConfigIO: nil,
			}))
			m := New(&kadmin.ConsumerRecord{
				Key:       "",
				Payload:   serdes.DesData{Value: ""},
				Partition: 0,
				Offset:    0,
				Headers: []kadmin.Header{
					{
						Key:   "h1",
						Value: kadmin.NewHeaderValue("v2"),
					},
				},
			},
				"",
				clipper.NewMock(),
				&ktx,
			)

			m.Update(tests.Key(tea.KeyCtrlS))

			//assert.IsType(msg)
		})
	})

	t.Run("Display record without headers", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Payload:   serdes.DesData{Value: ""},
			Partition: 0,
			Offset:    0,
			Headers:   nil,
		},
			"",
			clipper.NewMock(),
			tests.NewKontext(),
		)

		render := m.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "No headers present")
	})

	t.Run("Copy payload", func(t *testing.T) {
		var clippedText string
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			clippedText = text
			return nil
		}
		m := New(&kadmin.ConsumerRecord{
			Key:       "740ed9fd-195f-427e-8e0d-adb63d9c16ed",
			Payload:   serdes.DesData{Value: `{"name":"John"}`},
			Partition: 0,
			Offset:    123,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: kadmin.NewHeaderValue("v1"),
				},
			},
		},
			"",
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.TestRenderer)

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.TestRenderer))

		assert.Equal(t, "{\n\t\"name\": \"John\"\n}", clippedText)
		assert.Contains(t, render, "Payload copied")
	})

	t.Run("Copy header value", func(t *testing.T) {
		var clippedText string
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			clippedText = text
			return nil
		}
		m := New(&kadmin.ConsumerRecord{
			Key:       "740ed9fd-195f-427e-8e0d-adb63d9c16ed",
			Payload:   serdes.DesData{Value: `{"name":"John"}`},
			Partition: 0,
			Offset:    123,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: kadmin.NewHeaderValue("v1"),
				},
				{
					Key:   "h2",
					Value: kadmin.NewHeaderValue("v2"),
				},
				{
					Key:   "h3",
					Value: kadmin.NewHeaderValue("v3\nv3"),
				},
			},
		},
			"",
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.TestRenderer)

		m.Update(tests.Key(tea.KeyCtrlH))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.TestRenderer))

		assert.Equal(t, "v3\nv3", clippedText)
		assert.Contains(t, render, "Header Value copied")
	})

	t.Run("Copy header value failed", func(t *testing.T) {
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			return fmt.Errorf("unable to access clipboard")
		}
		m := New(&kadmin.ConsumerRecord{
			Key:       "740ed9fd-195f-427e-8e0d-adb63d9c16ed",
			Payload:   serdes.DesData{Value: `{"name":"John"}`},
			Partition: 0,
			Offset:    123,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: kadmin.NewHeaderValue("v1"),
				},
				{
					Key:   "h2",
					Value: kadmin.NewHeaderValue("v2"),
				},
				{
					Key:   "h3",
					Value: kadmin.NewHeaderValue("v3\nv3"),
				},
			},
		},
			"",
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.TestRenderer)

		m.Update(tests.Key(tea.KeyCtrlH))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.TestRenderer))

		assert.Contains(t, render, "Copy failed: unable to access clipboard")
	})

	t.Run("Copy payload failed", func(t *testing.T) {
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			return fmt.Errorf("unable to access clipboard")
		}
		m := New(&kadmin.ConsumerRecord{
			Key:       "740ed9fd-195f-427e-8e0d-adb63d9c16ed",
			Payload:   serdes.DesData{Value: `{"name":"John"}`},
			Partition: 0,
			Offset:    123,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: kadmin.NewHeaderValue("v1"),
				},
			},
		},
			"",
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.TestRenderer)

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.TestRenderer))

		assert.Contains(t, render, "Copy failed: unable to access clipboard")
	})

	t.Run("on deserialization error", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Payload:   serdes.DesData{Value: ""},
			Err:       fmt.Errorf("deserialization error"),
			Partition: 0,
			Offset:    0,
			Headers:   []kadmin.Header{},
		},
			"",
			clipper.NewMock(),
			tests.NewKontext(),
		)

		// init ui
		render := m.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "deserialization error")
		assert.Contains(t, render, "Unable to render payload")

		t.Run("do not update viewport", func(t *testing.T) {
			// do not crash but ignore the update
			m.Update(tests.Key(tea.KeyF2))
		})
	})
}
