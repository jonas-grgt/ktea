package kcon_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/kcon_clusters_page"
	"ktea/ui/pages/kcon_page"
	"net/http"
)

type Model struct {
	active    pages.Page
	statusbar *statusbar.Model
	kconsPage *kcon_clusters_page.Model
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.statusbar != nil {
		views = append(views, m.statusbar.View(ktx, renderer))
	}

	views = append(views, m.active.View(ktx, renderer))

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	// in case the active page might have changed, update the statusbar provider
	m.statusbar.SetProvider(m.active)

	return m.active.Update(msg)
}

func (m *Model) navBack() tea.Cmd {
	m.active = m.kconsPage
	m.statusbar.SetProvider(m.active)
	return nil
}

func (m *Model) loadKConPage(c config.KafkaConnectConfig) tea.Cmd {
	kca := kcadmin.New(http.DefaultClient, &c)
	var cmd tea.Cmd
	m.active, cmd = kcon_page.New(m.navBack, kca, c.Name)
	return cmd
}

func New(cluster *config.Cluster, stsBar *statusbar.Model) (*Model, tea.Cmd) {
	m := Model{}
	kconsPage, cmd := kcon_clusters_page.New(cluster, m.loadKConPage)
	m.kconsPage = kconsPage
	m.active = kconsPage
	m.statusbar = stsBar
	return &m, cmd
}
