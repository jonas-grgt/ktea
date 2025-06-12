package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

func newTestBar(callback SortSelectedCallback) *SortByCmdBar {
	return NewSortByCmdBar(
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
		WithSortSelectedCallback(callback),
	)
}

func TestSortByCmdBar(t *testing.T) {

	t.Run("Renders all options", func(t *testing.T) {
		bar := newTestBar(func(label SortLabel) {})

		render := bar.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃   Name ▼     Size ▼     Date ▼     Type ▼                                                        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`, render)
	})

	t.Run("Sorted by", func(t *testing.T) {
		bar := newTestBar(func(label SortLabel) {
		})

		bar.Update(tests.Key(tea.KeyRight))
		bar.Update(tests.Key(tea.KeyRight))
		bar.Update(tests.Key(tea.KeyLeft))
		bar.Update(tests.Key('h'))
		bar.Update(tests.Key('l'))

		assert.Equal(t, SortLabel{
			Label:     "Size",
			Direction: Desc,
		}, bar.SortedBy())
	})

	t.Run("Enter selects new sorting", func(t *testing.T) {
		var selectedLabel SortLabel
		bar := newTestBar(func(label SortLabel) {
			selectedLabel = label
		})
		bar.Update(tests.Key(tea.KeyRight))

		bar.Update(tests.Key(tea.KeyEnter))

		assert.Equal(t, SortLabel{
			Label:     "Size",
			Direction: Desc,
		}, selectedLabel)
	})

	t.Run("Enter selected toggles sorting directions", func(t *testing.T) {
		var selectedLabel SortLabel
		bar := newTestBar(func(label SortLabel) {
			selectedLabel = label
		})

		bar.Update(tests.Key(tea.KeyRight))

		bar.Update(tests.Key(tea.KeyEnter))
		bar.Update(tests.Key(tea.KeyEnter))

		assert.Equal(t, SortLabel{
			Label:     "Size",
			Direction: Asc,
		}, selectedLabel)
	})
}
