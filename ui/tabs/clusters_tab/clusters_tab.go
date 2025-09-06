package clusters_tab

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/pages/create_cluster_page"
)

type state int

type Model struct {
	state         state
	active        pages.Page
	createPage    pages.Page
	config        *config.Config
	statusbar     *statusbar.Model
	ktx           *kontext.ProgramKtx
	kConnChecker  kadmin.ConnChecker
	srConnChecker sradmin.ConnChecker
	escGoesBack   bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.statusbar != nil {
		views = append(views, m.statusbar.View(ktx, renderer))
	}

	views = append(views, m.active.View(ktx, renderer))

	return ui.JoinVertical(
		lipgloss.Top,
		views...,
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.active == nil {
		return nil
	}
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.escGoesBack {
				return m.active.Update(msg)
			}
			return nil
		case "ctrl+n":
			if _, ok := m.active.(*clusters_page.Model); ok {
				m.active = create_cluster_page.NewCreateClusterPage(
					m.GoBack,
					m.kConnChecker,
					m.srConnChecker,
					m.ktx.Config,
					m.ktx,
					[]statusbar.Shortcut{
						{"Confirm", "enter"},
						{"Next Field", "tab"},
						{"Prev. Field", "s-tab"},
						{"Reset Form", "C-r"},
						{"Go Back", "esc"},
					},
				)
			}
		case "ctrl+e":
			if clustersPage, ok := m.active.(*clusters_page.Model); ok {
				clusterName := clustersPage.SelectedCluster()
				selectedCluster := m.ktx.Config.FindClusterByName(*clusterName)
				m.active = create_cluster_page.NewEditClusterPage(
					m.GoBack,
					m.kConnChecker,
					m.srConnChecker,
					m.ktx.Config,
					m.ktx.Config,
					m.ktx,
					*selectedCluster,
					create_cluster_page.WithTitle(fmt.Sprintf("Clusters / %s / Edit", selectedCluster.Name)),
				)
			}
		}
	}

	// in case the active page might have changed, update the statusbar provider
	m.statusbar.SetProvider(m.active)

	return m.active.Update(msg)
}

func (m *Model) GoBack() tea.Cmd {
	var cmd tea.Cmd
	m.active, cmd = clusters_page.New(m.ktx, m.kConnChecker)
	m.statusbar.SetProvider(m.active)
	return cmd
}

func New(
	ktx *kontext.ProgramKtx,
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	stsBar *statusbar.Model,
) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	m := Model{}
	m.kConnChecker = kConnChecker
	m.srConnChecker = srConnChecker
	m.ktx = ktx
	m.config = ktx.Config
	m.statusbar = stsBar
	if m.config.HasClusters() {
		var listPage, c = clusters_page.New(ktx, m.kConnChecker)
		cmd = c
		m.escGoesBack = true
		m.active = listPage
	} else {
		m.active = create_cluster_page.NewCreateClusterPage(
			m.GoBack,
			m.kConnChecker,
			m.srConnChecker,
			m.ktx.Config,
			m.ktx,
			[]statusbar.Shortcut{
				{"Confirm", "enter"},
				{"Next Field", "tab"},
				{"Prev. Field", "s-tab"},
				{"Reset Form", "C-r"},
			},
			create_cluster_page.WithTitle("Clusters / Register"),
		)
		m.escGoesBack = false
	}
	m.statusbar.SetProvider(m.active)

	return &m, cmd
}
