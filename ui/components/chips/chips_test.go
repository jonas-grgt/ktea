package chips

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

func TestChips(t *testing.T) {
	t.Run("Render all chips", func(t *testing.T) {
		chips := New("label", "a1", "b2", "c3", "d4", "e5")

		render := ansi.Strip(chips.View(tests.NewKontext(), tests.TestRenderer))

		assert.Equal(t, "label:  «a1»    b2     c3     d4     e5  ", render)
	})

	t.Run("Activate element", func(t *testing.T) {
		chips := New("label", "a1", "b2", "c3", "d4", "e5")

		chips.ActivateByLabel("d4")

		assert.Equal(t, 3, chips.selectedIdx)

		t.Run("previous", func(t *testing.T) {
			chips.Update(tests.Key('h'))
			chips.Update(tests.Key(tea.KeyLeft))

			assert.Equal(t, 1, chips.selectedIdx)
		})

		t.Run("previous when first", func(t *testing.T) {
			chips.Update(tests.Key('h'))

			assert.Equal(t, 0, chips.selectedIdx)

			chips.Update(tests.Key('h'))
			chips.Update(tests.Key(tea.KeyLeft))

			assert.Equal(t, 0, chips.selectedIdx)
		})

		t.Run("next", func(t *testing.T) {
			chips.Update(tests.Key('l'))
			chips.Update(tests.Key(tea.KeyRight))

			assert.Equal(t, 2, chips.selectedIdx)
		})

		t.Run("next when last", func(t *testing.T) {
			chips.Update(tests.Key('l'))
			chips.Update(tests.Key(tea.KeyRight))

			assert.Equal(t, 4, chips.selectedIdx)

			chips.Update(tests.Key('l'))
			chips.Update(tests.Key(tea.KeyRight))

			assert.Equal(t, 4, chips.selectedIdx)
		})

	})
}
