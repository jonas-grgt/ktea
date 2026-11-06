package ui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/styles"
	"ktea/tests"
	"ktea/ui"
)

func submittedForm(t *testing.T) *huh.Form {
	value := "my-topic"
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Topic name").Value(&value),
	)).WithTheme(styles.Huh)
	form.Init()

	_, cmd := form.Update(tests.Key(tea.KeyEnter))
	// drive the form through next field / next group until it submits
	for i := 0; i < 5 && form.State != huh.StateCompleted && cmd != nil; i++ {
		for _, msg := range tests.ExecuteBatchCmd(cmd) {
			_, cmd = form.Update(msg)
		}
	}
	assert.Equal(t, huh.StateCompleted, form.State)
	return form
}

func TestRenderForm(t *testing.T) {
	t.Run("renders a submitted form that huh itself blanks", func(t *testing.T) {
		form := submittedForm(t)

		assert.Empty(t, form.View(), "huh no longer blanks submitted forms, RenderForm may be obsolete")

		render := ansi.Strip(ui.RenderForm(form, huh.LayoutDefault))
		assert.Contains(t, render, "Topic name")
		assert.Contains(t, render, "my-topic")
	})

	t.Run("submitted form ignores input", func(t *testing.T) {
		form := submittedForm(t)

		form.Update(tests.Key('x'))

		assert.NotContains(t, ansi.Strip(ui.RenderForm(form, huh.LayoutDefault)), "my-topicx")
	})

	t.Run("renders an active form like huh does", func(t *testing.T) {
		value := ""
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Topic name").Value(&value),
		)).WithTheme(styles.Huh)
		form.Init()

		assert.Equal(t, form.View(), ui.RenderForm(form, huh.LayoutDefault))
	})
}
