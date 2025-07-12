package clusters_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/pages/create_cluster_page"
	"ktea/ui/pages/nav"
)

type state int

type Model struct {
	state         state
	active        nav.Page
	createPage    nav.Page
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
				m.active, _ = clusters_page.New(m.ktx, m.kConnChecker)
			}
		case "ctrl+n":
			if _, ok := m.active.(*clusters_page.Model); ok {
				m.active = create_cluster_page.NewCreateClusterPage(
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
					m.kConnChecker,
					m.ktx.Config,
					m.ktx,
					selectedCluster,
					create_cluster_page.WithTitle("Edit cluster"),
				)
			}
		}
	}

	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.active)

	return m.active.Update(msg)
}

func New(
	ktx *kontext.ProgramKtx,
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	m := Model{}
	m.kConnChecker = kConnChecker
	m.srConnChecker = srConnChecker
	m.ktx = ktx
	m.config = ktx.Config
	if m.config.HasClusters() {
		var listPage, c = clusters_page.New(ktx, m.kConnChecker)
		cmd = c
		m.escGoesBack = true
		m.active = listPage
		m.statusbar = statusbar.New(m.active)
	} else {
		m.active = create_cluster_page.NewCreateClusterPage(
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
			create_cluster_page.WithTitle("Register your first Cluster"),
		)
		m.escGoesBack = false
	}

	return &m, cmd
}
