package clusters_tab

import (
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/cluster_config_page"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/pages/create_cluster_page"
	"ktea/ui/pages/nav"

	"github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listState state = iota
	viewState
)

type state int

type Model struct {
	state       state
	active      nav.Page
	createPage  nav.Page
	config      *config.Config
	statusbar   *statusbar.Model
	ktx         *kontext.ProgramKtx
	connChecker kadmin.ConnChecker
	ka          kadmin.Kadmin
	escGoesBack bool
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
	case config.ClusterRegisteredMsg:
		listPage, _ := clusters_page.New(m.ktx, m.connChecker)
		m.active = listPage
		m.statusbar = statusbar.New(m.active)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.state == viewState {
				listPage, _ := clusters_page.New(m.ktx, m.connChecker)
				m.active = listPage
				m.state = listState
			} else if m.escGoesBack {
				m.active, _ = clusters_page.New(m.ktx, m.connChecker)
			}
		case "ctrl+n":
			if _, ok := m.active.(*clusters_page.Model); ok {
				m.active = create_cluster_page.NewForm(
					m.connChecker,
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
				formValues := &create_cluster_page.FormValues{
					Name:  selectedCluster.Name,
					Color: selectedCluster.Color,
					Host:  selectedCluster.BootstrapServers[0],
				}
				if selectedCluster.SASLConfig != nil {
					formValues.SecurityProtocol = selectedCluster.SASLConfig.SecurityProtocol
					formValues.Username = selectedCluster.SASLConfig.Username
					formValues.Password = selectedCluster.SASLConfig.Password
					formValues.AuthMethod = config.SASLAuthMethod
					formValues.SSLEnabled = selectedCluster.SSLEnabled
				}
				if selectedCluster.SchemaRegistry != nil {
					formValues.SrEnabled = true
					formValues.SrUrl = selectedCluster.SchemaRegistry.Url
					formValues.SrUsername = selectedCluster.SchemaRegistry.Username
					formValues.SrPassword = selectedCluster.SchemaRegistry.Password
				}
				m.active = create_cluster_page.NewEditForm(
					m.connChecker,
					m.ktx.Config,
					m.ktx,
					formValues,
				)
			}
		case "ctrl+v":
			log.Info("Ctrl+V pressed, switching to view state")
			clustersPage, ok := m.active.(*clusters_page.Model)
			if ok {
				clusterName := clustersPage.SelectedCluster()
				cluster := m.ktx.Config.FindClusterByName(*clusterName)
				var cmd tea.Cmd
				m.active, cmd = cluster_config_page.New(cluster, m.ka)
				m.state = viewState
				return cmd
			}
		}
	}

	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.active)

	return m.active.Update(msg)
}

func New(
	ktx *kontext.ProgramKtx,
	connChecker kadmin.ConnChecker,
	ka kadmin.Kadmin,
) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	m := Model{}
	m.connChecker = connChecker
	m.ka = ka
	m.ktx = ktx
	m.config = ktx.Config
	if m.config.HasClusters() {
		var listPage, c = clusters_page.New(ktx, m.connChecker)
		cmd = c
		m.escGoesBack = true
		m.active = listPage
		m.statusbar = statusbar.New(m.active)
	} else {
		m.active = create_cluster_page.NewForm(
			m.connChecker,
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
