package border

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

func TestBorderTab(t *testing.T) {
	t.Run("Render only border", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭──────────────────────────────────────────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("Render title", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTitle("My Title"),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭──────────────────── My Title ────────────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("Render tabs", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs("tab1", "tab2", "tab3", "tab4", "tab5"),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭ [ tab1 tab2 tab3 tab4 tab5 ] ────────────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("When no tabs do not render tabs section", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs([]string{}...),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭──────────────────────────────────────────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("Render title and tabs", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTitle("My Title"),
			WithTabs("tab1", "tab2"),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭ [ tab1 tab2 ] ──────────── My Title ─────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("Next tab", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs("tab1", "tab2"),
		)

		assert.Equal(t, 0, bt.activeTabIdx)

		bt.NextTab()
		assert.Equal(t, 1, bt.activeTabIdx)

		bt.NextTab()
		assert.Equal(t, 0, bt.activeTabIdx)
	})
}

func content() string {
	return lipgloss.NewStyle().Width(50).Render("content")
}
