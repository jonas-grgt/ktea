package tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

func TestTabs(t *testing.T) {

	t.Run("Movements", func(t *testing.T) {

		t.Run("Next (C-→) tab when available", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			// when
			tabs.Update(tests.Key(tea.KeyCtrlRight))
			// then
			assert.Equal(t, Label("tab2"), tabs.ActiveTab().Label)
		})

		t.Run("Next (C-l) tab when available", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			// when
			tabs.Update(tests.Key(tea.KeyCtrlRight))
			// then
			assert.Equal(t, Label("tab2"), tabs.ActiveTab().Label)
		})

		t.Run("Next tab when last", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			tabs.Update(tests.Key(tea.KeyCtrlRight))
			// when
			tabs.Update(tests.Key(tea.KeyCtrlRight))
			// then
			assert.Equal(t, Label("tab2"), tabs.ActiveTab().Label)
		})

		t.Run("Previous (C-h) tab when available", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			tabs.Update(tests.Key(tea.KeyCtrlRight))
			// when
			tabs.Update(tests.Key(tea.KeyCtrlH))
			// then
			assert.Equal(t, Label("tab1"), tabs.ActiveTab().Label)
		})

		t.Run("Previous (C-←) tab when available", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			tabs.Update(tests.Key(tea.KeyCtrlRight))
			// when
			tabs.Update(tests.Key(tea.KeyCtrlLeft))
			// then
			assert.Equal(t, Label("tab1"), tabs.ActiveTab().Label)
		})

		t.Run("Previous tab when first", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			// when
			tabs.Update(tests.Key(tea.KeyCtrlLeft))
			// then
			assert.Equal(t, Label("tab1"), tabs.ActiveTab().Label)
		})
	})

	t.Run("Go To Tab", func(t *testing.T) {

		t.Run("Based on string", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			// when
			tabs.GoToTab("tab2")
			// then
			assert.Equal(t, Label("tab2"), tabs.ActiveTab().Label)
		})

		t.Run("Based on Label", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			// when
			tabs.GoToTab(Label("tab2"))
			// then
			assert.Equal(t, Label("tab2"), tabs.ActiveTab().Label)
		})

	})

	t.Run("Rendering", func(t *testing.T) {
		t.Run("Multiple tabs", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"}, Tab{Title: "tab2", Label: "tab2"})
			// when
			actual := tabs.View(tests.Kontext, tests.Renderer)
			// then
			assert.Equal(t, "╭──────╮╭──────╮               \n│ tab1 ││ tab2 │  ≪ F1 » help  \n┘      └┴──────┴────────────────────────────────────────────────────────────────────────────────────", actual)
		})

		t.Run("Single tab no shortcut", func(t *testing.T) {
			// given
			tabs := New(Tab{Title: "tab1", Label: "tab1"})
			// when
			actual := tabs.View(tests.Kontext, tests.Renderer)
			// then
			assert.NotContains(t, actual, "Meta")
		})

	})
}
