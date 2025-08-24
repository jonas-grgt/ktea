package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

const BorderedPadding = 2

type CmdBar interface {
	View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string

	Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd)

	// Shortcuts returns a slice of shortcuts
	// or nil if there are none, i.e. due to CmdBar not being active
	Shortcuts() []statusbar.Shortcut

	// IsFocussed returns true if the CmdBar is currently focussed
	// as in accepting input and can be interacted with
	IsFocussed() bool
}
