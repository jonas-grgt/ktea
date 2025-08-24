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
╭──────────────────────────────────────── My Title ╮
│content                                           │
╰──────────────────────────────────────── My Title ╯`)
	})

	t.Run("Render title func", func(t *testing.T) {
		title := "My Title"
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTitleFn(func() string {
				return title
			}),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭──────────────────────────────────────── My Title ╮
│content                                           │
╰──────────────────────────────────────── My Title ╯`)
	})

	t.Run("Render tabs", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs(
				Tab{Title: "tab1", TabLabel: "tab1"},
				Tab{Title: "tab2", TabLabel: "tab2"},
				Tab{Title: "tab3", TabLabel: "tab3"},
				Tab{Title: "tab4", TabLabel: "tab4"},
				Tab{Title: "tab5", TabLabel: "tab5"},
			),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭ | tab1  tab2  tab3  tab4  tab5 | ────────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("When no tabs do not render tabs section", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs([]Tab{}...),
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
			WithTabs(
				Tab{Title: "tab1", TabLabel: "tab1"},
				Tab{Title: "tab2", TabLabel: "tab2"},
			),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭ | tab1  tab2 | ──────────────────────── My Title ╮
│content                                           │
╰──────────────────────────────────────── My Title ╯`)
	})

	t.Run("Next tab", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs(
				Tab{Title: "tab1", TabLabel: "tab1"},
				Tab{Title: "tab2", TabLabel: "tab2"},
			),
		)

		assert.Equal(t, TabLabel("tab1"), bt.ActiveTab())

		bt.NextTab()
		assert.Equal(t, TabLabel("tab2"), bt.ActiveTab())

		bt.NextTab()
		assert.Equal(t, TabLabel("tab1"), bt.ActiveTab())
	})

	t.Run("Go to tab", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs(
				Tab{Title: "tab1", TabLabel: "tab1"},
				Tab{Title: "tab2", TabLabel: "tab2"},
				Tab{Title: "tabX", TabLabel: "tabX"},
				Tab{Title: "tab4", TabLabel: "tab4"},
			),
		)

		bt.GoTo("tabX")

		assert.Equal(t, 2, bt.activeTabIdx)

		// non existent tab
		bt.GoTo("tabY")

		assert.Equal(t, 2, bt.activeTabIdx)
	})

	t.Run("WithInnerPaddingTop", func(t *testing.T) {
		bt := New(
			WithInnerPaddingTop(),
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs(
				Tab{Title: "tab1", TabLabel: "tab1"},
				Tab{Title: "tab2", TabLabel: "tab2"},
				Tab{Title: "tab3", TabLabel: "tab3"},
				Tab{Title: "tab4", TabLabel: "tab4"},
				Tab{Title: "tab5", TabLabel: "tab5"},
			),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭ | tab1  tab2  tab3  tab4  tab5 | ────────────────╮
│                                                  │
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})
}

func content() string {
	return lipgloss.NewStyle().Width(50).Render("content")
}
