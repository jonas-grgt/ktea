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

	t.Run("Render title func", func(t *testing.T) {
		title := "My Title"
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTitleFunc(func() string {
				return title
			}),
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
			WithTabs(
				Tab{Title: "tab1", Label: "tab1"},
				Tab{Title: "tab2", Label: "tab2"},
				Tab{Title: "tab3", Label: "tab3"},
				Tab{Title: "tab4", Label: "tab4"},
				Tab{Title: "tab5", Label: "tab5"},
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
				Tab{Title: "tab1", Label: "tab1"},
				Tab{Title: "tab2", Label: "tab2"},
			),
		)

		render := bt.View(content())

		tests.TrimAndEqual(t, render, `
╭ | tab1  tab2 | ──────────── My Title ────────────╮
│content                                           │
╰──────────────────────────────────────────────────╯`)
	})

	t.Run("Next tab", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs(
				Tab{Title: "tab1", Label: "tab1"},
				Tab{Title: "tab2", Label: "tab2"},
			),
		)

		assert.Equal(t, 0, bt.activeTabIdx)

		bt.NextTab()
		assert.Equal(t, 1, bt.activeTabIdx)

		bt.NextTab()
		assert.Equal(t, 0, bt.activeTabIdx)
	})

	t.Run("Go to tab", func(t *testing.T) {
		bt := New(
			WithOnTabChanged(func(t string, m *Model) {}),
			WithTabs(
				Tab{Title: "tab1", Label: "tab1"},
				Tab{Title: "tab2", Label: "tab2"},
				Tab{Title: "tabX", Label: "tabX"},
				Tab{Title: "tab4", Label: "tab4"},
			),
		)

		bt.GoTo("tabX")

		assert.Equal(t, 2, bt.activeTabIdx)

		// non existent tab
		bt.GoTo("tabY")

		assert.Equal(t, 2, bt.activeTabIdx)
	})
}

func content() string {
	return lipgloss.NewStyle().Width(50).Render("content")
}
