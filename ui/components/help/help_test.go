package help_test

import (
	"ktea/ui/components/help"
	"ktea/ui/components/statusbar"
	"strings"
	"testing"

	lg "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

var shortcuts = []statusbar.Shortcut{
	{Name: "Search", Keybinding: "/"},
	{Name: "Delete", Keybinding: "F2"},
	{Name: "Refresh", Keybinding: "C-r"},
}

func TestHelp(t *testing.T) {
	t.Run("renders title and shortcuts", func(t *testing.T) {
		render := ansi.Strip(help.View("Topics", shortcuts, 100))

		assert.Contains(t, render, "Keybindings · Topics")
		assert.Contains(t, render, "  /  Search")
		assert.Contains(t, render, " F2  Delete")
		assert.Contains(t, render, "C-r  Refresh")
		assert.Contains(t, render, "esc / F1 close")
	})

	t.Run("splits into columns to fit the height", func(t *testing.T) {
		render := ansi.Strip(help.View("Topics", shortcuts, 10))

		assert.Equal(t, 10, lg.Height(render))
		var row string
		for _, line := range strings.Split(render, "\n") {
			if strings.Contains(line, "Search") {
				row = line
			}
		}
		assert.Contains(t, row, "Refresh")
	})
}
