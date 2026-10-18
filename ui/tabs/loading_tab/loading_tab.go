package loading_tab

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/notifier"
)

type Model struct {
	notifier *notifier.Model
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return renderer.Render(lipgloss.NewStyle().
		Width(ktx.WindowWidth).
		Height(ktx.AvailableHeight).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(m.notifier.View(ktx, renderer)))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	return m.notifier.Update(msg)
}

func New() (*Model, tea.Cmd) {
	n := notifier.New()
	cmd := n.SpinWithRocketMsg("loading")
	return &Model{n}, cmd
}
