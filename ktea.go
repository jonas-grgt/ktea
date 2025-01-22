package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/tab"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/tabs"
	"ktea/ui/tabs/cgroups_tab"
	"ktea/ui/tabs/clusters_tab"
	"ktea/ui/tabs/con_err_tab"
	"ktea/ui/tabs/loading_tab"
	"ktea/ui/tabs/sr_tab"
	"ktea/ui/tabs/topics_tab"
	"os"
	"reflect"
	"time"
)

var version string

type Model struct {
	tabs                  tab.Model
	tabCtrl               tabs.TabController
	ktx                   *kontext.ProgramKtx
	activeTab             int
	topicsTabCtrl         *topics_tab.Model
	cgroupsTabCtrl        *cgroups_tab.Model
	kaInstantiator        kadmin.Instantiator
	ka                    kadmin.Kadmin
	sra                   sradmin.SrAdmin
	renderer              *ui.Renderer
	schemaRegistryTabCtrl *sr_tab.Model
	clustersTabCtrl       *clusters_tab.Model
	configIO              config.IO
}

// RetryClusterConnectionMsg is an internal Msg
// to actually retry the cluster connection
type RetryClusterConnectionMsg struct {
	Cluster *config.Cluster
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		return config.LoadedMsg{Config: config.New(m.configIO)}
	}, tea.WindowSize())
}

func (m *Model) View() string {
	m.ktx = kontext.WithNewAvailableDimensions(m.ktx)
	if m.renderer == nil {
		m.renderer = ui.NewRenderer(m.ktx)
	}

	var views []string
	logoView := m.renderer.Render("   ___        \n |/ |  _   _.\n |\\ | (/_ (_|  " + version)
	views = append(views, logoView)

	tabsView := m.tabs.View(m.ktx, m.renderer)
	views = append(views, tabsView)

	if m.tabCtrl != nil {
		view := m.tabCtrl.View(m.ktx, m.renderer)
		views = append(views, view)
	}

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug("Update ktea", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

		// Make sure the events, because of their async nature,
		// are explicitly captured and properly propagated
		// in the case when the tabCtrl hence the page isn't focussed anymore
	case kadmin.TopicListedMsg:
		return m, m.topicsTabCtrl.Update(msg)
	case kadmin.ConsumerGroupsListedMsg:
		return m, m.cgroupsTabCtrl.Update(msg)
	case sradmin.SubjectDeletedMsg:
		return m, m.schemaRegistryTabCtrl.Update(msg)
	case sradmin.SubjectsListedMsg:
		if m.schemaRegistryTabCtrl != nil {
			return m, m.schemaRegistryTabCtrl.Update(msg)
		}
	case sradmin.SubjectListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)

	case config.ClusterRegisteredMsg:
		// if the active cluster has been updated it needs to be reloaded
		if msg.Cluster.Active {
			// TODO check err
			m.activateCluster(msg.Cluster)
			// keep clusters tab focussed after recreating tabs
			if msg.Cluster.HasSchemaRegistry() {
				m.tabs.GoToTab(tabs.ClustersTab)
			} else {
				m.tabs.GoToTab(tabs.ClustersTab)
			}

		}
	case con_err_tab.RetryClusterConnectionMsg:
		var cmd tea.Cmd
		m.tabCtrl, cmd = loading_tab.New()
		return m, tea.Batch(cmd, func() tea.Msg {
			return RetryClusterConnectionMsg{msg.Cluster}
		})

	case RetryClusterConnectionMsg:
		c, _ := m.initTopicsTabOrError(msg.Cluster)
		return m, c

	case config.LoadedMsg:
		m.ktx.Config = msg.Config
		if m.ktx.Config.HasClusters() {
			m.tabs.GoToTab(tabs.TopicsTab)
			cmds := []tea.Cmd{}
			cmd, err := m.initTopicsTabOrError(msg.Config.ActiveCluster())
			if err == nil {
				// cluster has been activated and sradmin has been loaded only if a
				// schema registry has been configured
				if m.ktx.Config.ActiveCluster().HasSchemaRegistry() {
					cmds = append(cmds, m.sra.ListSubjects)
				}
			}
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		} else {
			t, c := clusters_tab.New(m.ktx)
			m.tabCtrl = t
			m.tabs.GoToTab(tabs.ClustersTab)
			return m, c
		}

	case clusters_page.ClusterSwitchedMsg:
		// TODO check err
		m.activateCluster(msg.Cluster)
		// tabs were recreated due to cluster switch,
		// make sure we stay on the clusters tab because,
		// which might have introduced or removed the schema-registry tab
		if msg.Cluster.HasSchemaRegistry() {
			m.tabs.GoToTab(3)
		} else {
			m.tabs.GoToTab(2)
		}
		// reset all cached tabs, so they are loaded again for the new cluster
		m.topicsTabCtrl = nil
		m.cgroupsTabCtrl = nil
		m.schemaRegistryTabCtrl = nil

	case tea.WindowSizeMsg:
		m.onWindowSizeUpdated(msg)
	}

	// if no clusters configured,
	// do not allow to move away from create cluster form
	if m.ktx.Config != nil && m.ktx.Config.HasClusters() {
		m.tabs.Update(msg)
	}
	if m.tabs.ActiveTab() != m.activeTab {
		m.activeTab = m.tabs.ActiveTab()
		switch m.activeTab {
		case 0:
			if m.topicsTabCtrl == nil {
				var cmd tea.Cmd
				m.topicsTabCtrl, cmd = topics_tab.New(m.ktx, m.ka)
				cmds = append(cmds, cmd)
			}
			m.tabCtrl = m.topicsTabCtrl
		case 1:
			if m.cgroupsTabCtrl == nil {
				var cmd tea.Cmd
				m.cgroupsTabCtrl, cmd = cgroups_tab.New(m.ka, m.ka)
				cmds = append(cmds, cmd)
			}
			m.tabCtrl = m.cgroupsTabCtrl
		case 2:
			if m.ktx.Config.ActiveCluster().HasSchemaRegistry() {
				if m.schemaRegistryTabCtrl == nil {
					var cmd tea.Cmd
					m.schemaRegistryTabCtrl, cmd = sr_tab.New(m.sra, m.sra, m.sra, m.sra, m.ktx)
					cmds = append(cmds, cmd)
				}
				m.tabCtrl = m.schemaRegistryTabCtrl
				break
			}
			fallthrough
		case 3:
			if m.clustersTabCtrl == nil {
				var cmd tea.Cmd
				m.clustersTabCtrl, cmd = clusters_tab.New(m.ktx)
				cmds = append(cmds, cmd)
			}
			m.tabCtrl = m.clustersTabCtrl
		}
	}

	if m.tabCtrl == nil {
		var cmd tea.Cmd
		m.tabCtrl, cmd = loading_tab.New()
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	cmd = m.tabCtrl.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) createTabs(cluster *config.Cluster) {
	if cluster.HasSchemaRegistry() {
		m.tabs = tab.New("Topics", "Consumer Groups", "Schema Registry", "Clusters")
		tabs.ClustersTab = 3
	} else {
		m.tabs = tab.New("Topics", "Consumer Groups", "Clusters")
		tabs.ClustersTab = 2
	}
}

// activateCluster creates the kadmin.Model and kadmin.SrAdmin
// based on the given cluster
func (m *Model) activateCluster(cluster *config.Cluster) error {
	var saslConfig *kadmin.SASLConfig
	if cluster.SASLConfig != nil {
		saslConfig = &kadmin.SASLConfig{
			Username: cluster.SASLConfig.Username,
			Password: cluster.SASLConfig.Password,
			Protocol: kadmin.SSL,
		}
	}

	connDetails := kadmin.ConnectionDetails{
		BootstrapServers: cluster.BootstrapServers,
		SASLConfig:       saslConfig,
	}
	if ka, err := m.kaInstantiator(connDetails); err != nil {
		return err
	} else {
		m.ka = ka
	}

	if cluster.HasSchemaRegistry() {
		m.sra = sradmin.New(m.ktx)
		m.ka.SetSra(m.sra)
	}

	m.createTabs(cluster)

	return nil
}

func (m *Model) onWindowSizeUpdated(msg tea.WindowSizeMsg) {
	m.ktx.WindowWidth = msg.Width
	m.ktx.WindowHeight = msg.Height
	m.ktx.AvailableHeight = msg.Height
}

func (m *Model) initTopicsTabOrError(cluster *config.Cluster) (tea.Cmd, error) {
	var cmd tea.Cmd
	if err := m.activateCluster(cluster); err != nil {
		m.tabCtrl, cmd = con_err_tab.New(err, cluster)
		return cmd, err
	} else {
		m.topicsTabCtrl, cmd = topics_tab.New(m.ktx, m.ka)
		m.tabCtrl = m.topicsTabCtrl
		return cmd, nil
	}
}

func NewModel(kai kadmin.Instantiator, configIO config.IO) *Model {
	return &Model{
		kaInstantiator: kai,
		ktx:            kontext.New(),
		configIO:       configIO,
	}
}

func main() {
	p := tea.NewProgram(
		NewModel(
			kadmin.SaramaInstantiator(),
			config.NewDefaultIO(),
		),
		tea.WithAltScreen(),
	)
	var fileErr error
	newConfigFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if fileErr == nil {
		log.SetOutput(newConfigFile)
		log.SetTimeFormat(time.Kitchen)
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		log.Debug("Logging to debug.log")
		log.Info("started")
		if _, err := p.Run(); err != nil {
			log.Fatal("Failed starting the TUI", err)
		}
	}
}
