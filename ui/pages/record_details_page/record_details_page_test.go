package record_details_page

import (
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/serdes"
	"ktea/tests"
	"ktea/ui/clipper"
	"ktea/ui/components/statusbar"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRecordDetailsPage(t *testing.T) {
	t.Run("c-h or arrows toggles focus between content and headers", func(t *testing.T) {
		record := &kadmin.ConsumerRecord{
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
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)
		// init ui
		m.View(tests.Kontext, tests.Renderer)

		assert.Equal(t, mainViewFocus, m.focus)

		m.Update(tests.Key('h'))

		assert.Equal(t, headersViewFocus, m.focus)

		m.Update(tests.Key(tea.KeyRight))

		assert.Equal(t, mainViewFocus, m.focus)

		m.Update(tests.Key(tea.KeyRight))

		assert.Equal(t, headersViewFocus, m.focus)

		m.Update(tests.Key(tea.KeyLeft))

		assert.Equal(t, mainViewFocus, m.focus)
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
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
						SchemaRegistry: nil,
						TLSConfig:      config.TLSConfig{Enable: false},
					},
				},
				ConfigIO: nil,
			}))
			record := &kadmin.ConsumerRecord{
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
			}
			m := New(record,
				"",
				[]kadmin.ConsumerRecord{*record},
				0,
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
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
						SchemaRegistry: &config.SchemaRegistryConfig{
							Url:      "http://localhost:8080",
							Username: "john",
							Password: "doe",
						},
						TLSConfig: config.TLSConfig{Enable: false},
					},
				},
				ConfigIO: nil,
			}))
			record := &kadmin.ConsumerRecord{
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
			}
			m := New(record,
				"",
				[]kadmin.ConsumerRecord{*record},
				0,
				clipper.NewMock(),
				&ktx,
			)

			shortcuts := m.Shortcuts()

			assert.Contains(t, shortcuts, statusbar.Shortcut{
				Name:       "Toggle Record/Schema",
				Keybinding: "<tab>",
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
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
						SchemaRegistry: &config.SchemaRegistryConfig{
							Url:      "http://localhost:8080",
							Username: "john",
							Password: "doe",
						},
						TLSConfig: config.TLSConfig{Enable: false},
					},
				},
				ConfigIO: nil,
			}))
			record := &kadmin.ConsumerRecord{
				Key: "",
				Payload: serdes.DesData{Value: `{"Name": "john", "Age": 12"}`, Schema: `
{
   "type" : "record",
   "namespace" : "ktea.test",
   "name" : "Person",
   "fields" : [
      { "name" : "Name" , "type" : "string" },
      { "name" : "Age" , "type" : "int" }
   ]
}
`},
				Partition: 0,
				Offset:    0,
				Headers: []kadmin.Header{
					{
						Key:   "h1",
						Value: kadmin.NewHeaderValue("v2"),
					},
				},
			}
			m := New(record,
				"",
				[]kadmin.ConsumerRecord{*record},
				0,
				clipper.NewMock(),
				&ktx,
			)

			m.View(tests.NewKontext(), tests.Renderer)
			m.Update(tests.Key(tea.KeyTab))

			render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))

			assert.Contains(t, render, `"namespace": "ktea.test"`)
		})
	})

	t.Run("Display record without headers", func(t *testing.T) {
		record := &kadmin.ConsumerRecord{
			Key:       "",
			Payload:   serdes.DesData{Value: ""},
			Partition: 0,
			Offset:    0,
			Headers:   nil,
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "No headers present")
	})

	t.Run("Title contains topic name, partition and offset", func(t *testing.T) {
		record := &kadmin.ConsumerRecord{
			Key:       "ABC",
			Payload:   serdes.DesData{Value: ""},
			Partition: 88,
			Offset:    123,
			Headers:   nil,
		}
		m := New(record,
			"dev.title.test",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)

		title := m.Title()

		assert.Equal(t, title, "Topics / dev.title.test / Partition / 88 / Offset / 123")
	})

	t.Run("Copy payload", func(t *testing.T) {
		var clippedText string
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			clippedText = text
			return nil
		}
		record := &kadmin.ConsumerRecord{
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
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))

		assert.Equal(t, "{\n\t\"name\": \"John\"\n}", clippedText)
		assert.Contains(t, render, "Payload copied")
	})

	t.Run("Copy schema", func(t *testing.T) {
		var clippedText string
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			clippedText = text
			return nil
		}
		record := &kadmin.ConsumerRecord{
			Key: "740ed9fd-195f-427e-8e0d-adb63d9c16ed",
			Payload: serdes.DesData{Value: `{"name":"John"}`, Schema: `
{
  "type"": "record",
  "name": "Person",
  "namespace": "io.jonasg.ktea",
  "fields": [ {"name": "name", "type": "string"} ]
}`},
			Partition: 0,
			Offset:    123,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: kadmin.NewHeaderValue("v1"),
				},
			},
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(tests.Key(tea.KeyTab))
		cmds = m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))

		tests.TrimAndEqual(t, clippedText, `
{
  "type"": "record",
  "name": "Person",
  "namespace": "io.jonasg.ktea",
  "fields": [ {"name": "name", "type": "string"} ]
}`)
		assert.Contains(t, render, "Schema copied")
	})

	t.Run("Copy header value", func(t *testing.T) {
		var clippedText string
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			clippedText = text
			return nil
		}
		record := &kadmin.ConsumerRecord{
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
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.Renderer)

		m.Update(tests.Key('h'))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))

		assert.Equal(t, "v3\nv3", clippedText)
		assert.Contains(t, render, "Header Value copied")
	})

	t.Run("Copy header value failed", func(t *testing.T) {
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			return fmt.Errorf("unable to access clipboard")
		}
		record := &kadmin.ConsumerRecord{
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
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.Renderer)

		m.Update(tests.Key('h'))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))

		assert.Contains(t, render, "Copy failed: unable to access clipboard")
	})

	t.Run("Copy payload failed", func(t *testing.T) {
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			return fmt.Errorf("unable to access clipboard")
		}
		record := &kadmin.ConsumerRecord{
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
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipMock,
			tests.NewKontext(),
		)

		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(tests.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))

		assert.Contains(t, render, "Copy failed: unable to access clipboard")
	})

	t.Run("on deserialization error", func(t *testing.T) {
		record := &kadmin.ConsumerRecord{
			Key:       "",
			Payload:   serdes.DesData{Value: ""},
			Err:       fmt.Errorf("deserialization error"),
			Partition: 0,
			Offset:    0,
			Headers:   []kadmin.Header{},
		}
		m := New(record,
			"",
			[]kadmin.ConsumerRecord{*record},
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)

		// init ui
		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "deserialization error")
		assert.Contains(t, render, "Unable to render payload")

		t.Run("do not update viewport", func(t *testing.T) {
			// do not crash but ignore the update
			m.Update(tests.Key(tea.KeyF2))
		})
	})
}

func TestRecordNavigation(t *testing.T) {
	ctrlN := tea.KeyMsg{Type: tea.KeyCtrlN, Runes: []rune{}, Alt: false}
	ctrlP := tea.KeyMsg{Type: tea.KeyCtrlP, Runes: []rune{}, Alt: false}

	t.Run("Ctrl+n navigates to next record", func(t *testing.T) {
		record1 := &kadmin.ConsumerRecord{
			Key:       "key-0",
			Payload:   serdes.DesData{Value: `{"value":"first"}`},
			Partition: 0,
			Offset:    0,
		}
		record2 := &kadmin.ConsumerRecord{
			Key:       "key-1",
			Payload:   serdes.DesData{Value: `{"value":"second"}`},
			Partition: 0,
			Offset:    1,
		}
		records := []kadmin.ConsumerRecord{*record1, *record2}

		m := New(record1,
			"topic1",
			records,
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)
		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(ctrlN)
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		assert.Equal(t, 1, m.recordIndex)
		assert.Equal(t, "key-1", m.record.Key)
	})

	t.Run("Ctrl+n shows error when no more records", func(t *testing.T) {
		record := &kadmin.ConsumerRecord{
			Key:       "key-0",
			Payload:   serdes.DesData{Value: `{"value":"first"}`},
			Partition: 0,
			Offset:    0,
		}
		records := []kadmin.ConsumerRecord{*record}

		m := New(record,
			"topic1",
			records,
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)
		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(ctrlN)

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))
		assert.Contains(t, render, "no more records")
		assert.Empty(t, cmds)
	})

	t.Run("Ctrl+p navigates to previous record", func(t *testing.T) {
		record1 := &kadmin.ConsumerRecord{
			Key:       "key-0",
			Payload:   serdes.DesData{Value: `{"value":"first"}`},
			Partition: 0,
			Offset:    0,
		}
		record2 := &kadmin.ConsumerRecord{
			Key:       "key-1",
			Payload:   serdes.DesData{Value: `{"value":"second"}`},
			Partition: 0,
			Offset:    1,
		}
		records := []kadmin.ConsumerRecord{*record1, *record2}

		m := New(record2,
			"topic1",
			records,
			1,
			clipper.NewMock(),
			tests.NewKontext(),
		)
		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(ctrlP)
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		assert.Equal(t, 0, m.recordIndex)
		assert.Equal(t, "key-0", m.record.Key)
	})

	t.Run("Ctrl+p shows error when no previous records", func(t *testing.T) {
		record := &kadmin.ConsumerRecord{
			Key:       "key-0",
			Payload:   serdes.DesData{Value: `{"value":"first"}`},
			Partition: 0,
			Offset:    0,
		}
		records := []kadmin.ConsumerRecord{*record}

		m := New(record,
			"topic1",
			records,
			0,
			clipper.NewMock(),
			tests.NewKontext(),
		)
		m.View(tests.NewKontext(), tests.Renderer)

		cmds := m.Update(ctrlP)

		render := ansi.Strip(m.View(tests.NewKontext(), tests.Renderer))
		assert.Contains(t, render, "no previous records")
		assert.Empty(t, cmds)
	})

	t.Run("Shortcuts show next/prev when multiple records", func(t *testing.T) {
		ktx := *tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "cluster1",
					Color:            "",
					Active:           true,
					BootstrapServers: nil,
					SASLConfig:       config.SASLConfig{AuthMethod: config.AuthMethodNone},
					SchemaRegistry:   nil,
					TLSConfig:        config.TLSConfig{Enable: false},
				},
			},
			ConfigIO: nil,
		}))
		record1 := &kadmin.ConsumerRecord{
			Key:       "key-0",
			Payload:   serdes.DesData{Value: `{"value":"first"}`},
			Partition: 0,
			Offset:    0,
		}
		record2 := &kadmin.ConsumerRecord{
			Key:       "key-1",
			Payload:   serdes.DesData{Value: `{"value":"second"}`},
			Partition: 0,
			Offset:    1,
		}
		records := []kadmin.ConsumerRecord{*record1, *record2}

		m := New(record1,
			"topic1",
			records,
			0,
			clipper.NewMock(),
			&ktx,
		)

		shortcuts := m.Shortcuts()

		foundNext := false
		foundPrev := false
		for _, sc := range shortcuts {
			if sc.Name == "Next Record" {
				foundNext = true
			}
			if sc.Name == "Prev Record" {
				foundPrev = true
			}
		}
		assert.True(t, foundNext, "Next Record shortcut should be present")
		assert.True(t, foundPrev, "Prev Record shortcut should be present")
	})

	t.Run("Shortcuts hidden when single record", func(t *testing.T) {
		ktx := *tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "cluster1",
					Color:            "",
					Active:           true,
					BootstrapServers: nil,
					SASLConfig:       config.SASLConfig{AuthMethod: config.AuthMethodNone},
					SchemaRegistry:   nil,
					TLSConfig:        config.TLSConfig{Enable: false},
				},
			},
			ConfigIO: nil,
		}))
		record := &kadmin.ConsumerRecord{
			Key:       "key-0",
			Payload:   serdes.DesData{Value: `{"value":"first"}`},
			Partition: 0,
			Offset:    0,
		}
		records := []kadmin.ConsumerRecord{*record}

		m := New(record,
			"topic1",
			records,
			0,
			clipper.NewMock(),
			&ktx,
		)

		shortcuts := m.Shortcuts()

		for _, sc := range shortcuts {
			assert.NotEqual(t, "Next Record", sc.Name)
			assert.NotEqual(t, "Prev Record", sc.Name)
		}
	})
}
