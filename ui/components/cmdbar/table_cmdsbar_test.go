package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

func TestTableCmdsBar(t *testing.T) {
	t.Run("Toggle between widgets resets previous widget state", func(t *testing.T) {
		sortByCmdBar := NewSortByCmdBar(
			[]SortLabel{
				{
					Label:     "Name",
					Direction: Desc,
				},
				{
					Label:     "Size",
					Direction: Desc,
				},
				{
					Label:     "Date",
					Direction: Desc,
				},
				{
					Label:     "Type",
					Direction: Desc,
				},
			},
			WithSortSelectedCallback(func(label SortLabel) {}),
		)
		cmdBar := NewTableCmdsBar[string](
			NewDeleteCmdBar[string](
				func(s string) string { return "The rabbit will be deleted" },
				func(s string) tea.Cmd { return nil },
				func(string) (bool, tea.Cmd) {
					return false, func() tea.Msg {
						return TestMsg{}
					}
				}),
			NewSearchCmdBar("Search Consumer Group"),
			NewNotifierCmdBar("notifier"),
			sortByCmdBar,
		)

		selection := "SelectedTopic"

		cmdBar.Update(tests.Key(tea.KeyF2), &selection)
		render := cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "The rabbit will be deleted")

		cmdBar.Update(tests.Key(tea.KeyF3), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "Name ▼")

		cmdBar.Update(tests.Key(tea.KeyF2), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "The rabbit will be deleted")

		cmdBar.Update(tests.Key('/'), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "┃ >")

		cmdBar.Update(tests.Key(tea.KeyF3), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "Name ▼")

		cmdBar.Update(tests.Key('/'), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "┃ >")

		cmdBar.Update(tests.Key('/'), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.NotContains(t, render, "┃ >")

		cmdBar.Update(tests.Key(tea.KeyF2), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "The rabbit will be deleted")

		cmdBar.Update(tests.Key('/'), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "┃ >")

		cmdBar.Update(tests.Key(tea.KeyF3), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "Name ▼")

		cmdBar.Update(tests.Key(tea.KeyF2), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.Contains(t, render, "The rabbit will be deleted")

		cmdBar.Update(tests.Key(tea.KeyF2), &selection)
		render = cmdBar.View(tests.TestKontext, tests.TestRenderer)
		assert.NotContains(t, render, "The rabbit will be deleted")
		assert.NotContains(t, render, "┃ >")
		assert.NotContains(t, render, "Name ▼")
	})
}
