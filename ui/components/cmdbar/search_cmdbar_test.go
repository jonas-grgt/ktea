package cmdbar

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

func TestSearchCmdBar(t *testing.T) {
	t.Run("Is hidden by default", func(t *testing.T) {
		model := NewSearchCmdBar(">")

		render := ansi.Strip(model.View(tests.NewKontext(), tests.Renderer))

		assert.Empty(t, render)

		t.Run("/ activates search", func(t *testing.T) {
			model.Update(tests.Key('/'))

			render = ansi.Strip(model.View(tests.NewKontext(), tests.Renderer))

			assert.Contains(t, render, ">")
		})
	})

	t.Run("Resets search value upon toggle", func(t *testing.T) {
		model := NewSearchCmdBar(">")

		model.Update(tests.Key('/'))

		model.Update(tests.Key('s'))
		model.Update(tests.Key('e'))
		model.Update(tests.Key('a'))
		model.Update(tests.Key('r'))
		model.Update(tests.Key('c'))
		model.Update(tests.Key('h'))

		render := ansi.Strip(model.View(tests.NewKontext(), tests.Renderer))

		assert.Contains(t, render, "> search")

		model.Update(tests.Key('/'))

		render = ansi.Strip(model.View(tests.NewKontext(), tests.Renderer))

		assert.NotContains(t, render, "> search")
	})
}
