package record_details_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests"
	"ktea/tests/keys"
	"ktea/ui"
	"ktea/ui/clipper"
	"testing"
)

func TestRecordDetailsPage(t *testing.T) {
	t.Run("c-h or arrows toggles focus between content and headers", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Value:     "",
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
		)
		// init ui
		m.View(ui.TestKontext, ui.TestRenderer)

		assert.Equal(t, payloadFocus, m.focus)

		m.Update(keys.Key(tea.KeyCtrlH))

		assert.Equal(t, headersFocus, m.focus)

		m.Update(keys.Key(tea.KeyCtrlH))

		assert.Equal(t, payloadFocus, m.focus)

		m.Update(keys.Key(tea.KeyRight))

		assert.Equal(t, headersFocus, m.focus)

		m.Update(keys.Key(tea.KeyLeft))

		assert.Equal(t, payloadFocus, m.focus)
	})

	t.Run("Display record without headers", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Value:     "",
			Partition: 0,
			Offset:    0,
			Headers:   nil,
		},
			"",
			clipper.NewMock(),
		)

		render := m.View(ui.NewTestKontext(), ui.TestRenderer)

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
			Value:     `{"name":"John"}`,
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
		)

		m.View(ui.NewTestKontext(), ui.TestRenderer)

		cmds := m.Update(keys.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(ui.NewTestKontext(), ui.TestRenderer))

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
			Value:     `{"name":"John"}`,
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
		)

		m.View(ui.NewTestKontext(), ui.TestRenderer)

		m.Update(keys.Key(tea.KeyCtrlH))
		m.Update(keys.Key(tea.KeyDown))
		m.Update(keys.Key(tea.KeyDown))

		cmds := m.Update(keys.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(ui.NewTestKontext(), ui.TestRenderer))

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
			Value:     `{"name":"John"}`,
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
		)

		m.View(ui.NewTestKontext(), ui.TestRenderer)

		m.Update(keys.Key(tea.KeyCtrlH))
		m.Update(keys.Key(tea.KeyDown))
		m.Update(keys.Key(tea.KeyDown))

		cmds := m.Update(keys.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(ui.NewTestKontext(), ui.TestRenderer))

		assert.Contains(t, render, "Copy failed: unable to access clipboard")
	})

	t.Run("Copy payload failed", func(t *testing.T) {
		clipMock := clipper.NewMock()
		clipMock.WriteFunc = func(text string) error {
			return fmt.Errorf("unable to access clipboard")
		}
		m := New(&kadmin.ConsumerRecord{
			Key:       "740ed9fd-195f-427e-8e0d-adb63d9c16ed",
			Value:     `{"name":"John"}`,
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
		)

		m.View(ui.NewTestKontext(), ui.TestRenderer)

		cmds := m.Update(keys.Key('c'))
		for _, msg := range tests.ExecuteBatchCmd(cmds) {
			m.Update(msg)
		}

		render := ansi.Strip(m.View(ui.NewTestKontext(), ui.TestRenderer))

		assert.Contains(t, render, "Copy failed: unable to access clipboard")
	})

	t.Run("on deserialization error", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Value:     "",
			Err:       fmt.Errorf("deserialization error"),
			Partition: 0,
			Offset:    0,
			Headers:   []kadmin.Header{},
		},
			"",
			clipper.NewMock(),
		)

		// init ui
		render := m.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, render, "deserialization error")
		assert.Contains(t, render, "Unable to render payload")

		t.Run("do not update viewport", func(t *testing.T) {
			// do not crash but ignore the update
			m.Update(keys.Key(tea.KeyF2))
		})
	})
}
