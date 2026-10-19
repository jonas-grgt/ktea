package settings_tab

import (
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/settings_page"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	page      *settings_page.Model
	statusbar *statusbar.Model
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return ui.JoinVertical(lipgloss.Top,
		m.statusbar.View(ktx, renderer),
		m.page.View(ktx, renderer),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.statusbar.SetProvider(m.page)
	return m.page.Update(msg)
}

func New(ktx *kontext.ProgramKtx, themesDir string, isDark bool, stsBar *statusbar.Model) *Model {
	return &Model{
		page:      settings_page.New(ktx, themesDir, isDark),
		statusbar: stsBar,
	}
}
