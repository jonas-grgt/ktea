package ui

import (
	tea "charm.land/bubbletea/v2"
	"ktea/kontext"
)

type View interface {
	View(ktx *kontext.ProgramKtx, renderer *Renderer) string

	Update(msg tea.Msg) tea.Cmd
}
