package main

import (
	"flag"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/components/tab"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/tabs"
	"ktea/ui/tabs/cgroups_tab"
	"ktea/ui/tabs/clusters_tab"
	"ktea/ui/tabs/kcon_tab"
	"ktea/ui/tabs/loading_tab"
	"ktea/ui/tabs/sr_tab"
	"ktea/ui/tabs/topics_tab"
	"os"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var version string

const (
	topicsTabLbl    tab.Label = "topics"
	cgroupsTabLbl             = "cgroups"
	schemaRegTabLbl           = "schemaReg"
	clustersTabLbl            = "clusters"
	kconnectTabLbl            = "kconnect"
)

var topicsTab = tab.Tab{Title: "Topics", Label: topicsTabLbl}
var cgroupsTab = tab.Tab{Title: "Consumer Groups", Label: cgroupsTabLbl}
var schemaRegTab = tab.Tab{Title: "Schema Registry", Label: schemaRegTabLbl}
var kconnectTab = tab.Tab{Title: "Kafka Connect", Label: kconnectTabLbl}
var clustersTab = tab.Tab{Title: "Clusters", Label: clustersTabLbl}

type Model struct {
	tabs                  tab.Model
	activeTab             tab.Tab
	tabCtrl               tabs.TabController
	ktx                   *kontext.ProgramKtx
	topicsTabCtrl         *topics_tab.Model
	cgroupsTabCtrl        *cgroups_tab.Model
	kaInstantiator        kadmin.Instantiator
	ka                    kadmin.Kadmin
	sra                   sradmin.Client
	renderer              *ui.Renderer
	schemaRegistryTabCtrl *sr_tab.Model
	clustersTabCtrl       *clusters_tab.Model
	kconTabCtrl           *kcon_tab.Model
	configIO              config.IO
	switchingCluster      bool
	startupConnErr        bool
	statusbar             *statusbar.Model
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyF1:
			m.statusbar.ToggleShortcuts()
			return m, nil
		}

		// Make sure the events, because of their async nature,
		// are explicitly captured and properly propagated
		// in the case when the tabCtrl hence the page isn't focussed anymore
	case kadmin.TopicsListedMsg,
		kadmin.TopicListingStartedMsg:
		if m.topicsTabCtrl != nil {
			return m, m.topicsTabCtrl.Update(msg)
		}
	case kadmin.ConsumerGroupsListedMsg,
		kadmin.ConsumerGroupListingStartedMsg:
		return m, m.cgroupsTabCtrl.Update(msg)
	case sradmin.SubjectsListedMsg,
		sradmin.GlobalCompatibilityListingStartedMsg,
		sradmin.GlobalCompatibilityListedMsg,
		sradmin.SubjectDeletedMsg:
		if m.schemaRegistryTabCtrl != nil {
			return m, m.schemaRegistryTabCtrl.Update(msg)
		}
	case sradmin.SubjectListingStartedMsg:
		if m.schemaRegistryTabCtrl != nil {
			cmd := m.schemaRegistryTabCtrl.Update(msg)
			cmds = append(cmds, cmd)
		}
	case kcadmin.ConnectorListingStartedMsg,
		kcadmin.ConnectorsListedMsg,
		kcadmin.ConnectorListingErrMsg:
		return m, m.kconTabCtrl.Update(msg)

	case kadmin.ConnCheckStartedMsg:
		m.switchingCluster = true
	case kadmin.ConnCheckErrMsg, kadmin.ConnCheckSucceededMsg:
		m.switchingCluster = false

	case config.ClusterRegisteredMsg:
		// if the active cluster has been updated it needs to be reloaded
		if msg.Cluster.Active {
			cmd := m.boostrapUI(msg.Cluster)
			cmds = append(cmds, cmd)

			// keep clusters tab focussed after recreating tabs
			m.tabs.GoToTab(clustersTabLbl)
			m.tabCtrl = m.clustersTabCtrl
			// TODO navigate to active cluster form
		}

	case RetryClusterConnectionMsg:
		c := m.boostrapUI(msg.Cluster)
		return m, c

	case kadmin.ClusterConfigMsg, kadmin.ClusterConfigStartedMsg:
		if m.clustersTabCtrl != nil {
			return m, m.clustersTabCtrl.Update(msg)
		}
	case config.LoadedMsg:
		m.ktx.RegisterConfig(msg.Config)
		if m.ktx.Config().HasClusters() {
			cmd := m.boostrapUI(msg.Config.ActiveCluster())
			cmds = append(cmds, cmd)

			m.tabs.GoToTab(topicsTabLbl)

			return m, tea.Batch(cmds...)
		} else {
			m.statusbar = statusbar.New()
			tCtrl, cmd := clusters_tab.New(m.ktx, kadmin.CheckKafkaConnectivity, sradmin.CheckSchemaRegistryConn, m.statusbar)
			m.tabCtrl = tCtrl
			m.tabs = tab.New(clustersTab)
			return m, cmd
		}

	case clusters_page.ClusterSwitchedMsg:
		cmd := m.boostrapUI(msg.Cluster)
		cmds = append(cmds, cmd)

		if m.startupConnErr {
			m.startupConnErr = false
			m.tabs.GoToTab(topicsTabLbl)
			m.tabCtrl = m.topicsTabCtrl
		} else {
			// tabs were recreated due to cluster switch,
			// make sure we stay on the clusters tab because,
			// which might have introduced or removed the schema-registry tab
			m.tabs.GoToTab(clustersTabLbl)
			m.tabCtrl = m.clustersTabCtrl
		}

	case tea.WindowSizeMsg:
		m.onWindowSizeUpdated(msg)
	}

	if !m.switchingCluster {
		m.tabs.Update(msg)

		if m.tabs.ActiveTab() != m.activeTab {
			m.activeTab = m.tabs.ActiveTab()
			switch m.activeTab.Label {
			case topicsTabLbl:
				m.tabCtrl = m.topicsTabCtrl
			case cgroupsTabLbl:
				m.tabCtrl = m.cgroupsTabCtrl
			case schemaRegTabLbl:
				if m.ktx.Config().ActiveCluster().HasSchemaRegistry() {
					m.tabCtrl = m.schemaRegistryTabCtrl
					break
				}
				fallthrough
			case clustersTabLbl:
				if m.clustersTabCtrl == nil {
					var cmd tea.Cmd
					m.clustersTabCtrl, cmd = clusters_tab.New(m.ktx, kadmin.CheckKafkaConnectivity, sradmin.CheckSchemaRegistryConn, m.statusbar)
					cmds = append(cmds, cmd)
				}
				m.tabCtrl = m.clustersTabCtrl
			case kconnectTabLbl:
				m.tabCtrl = m.kconTabCtrl
			}
			// can only be nil when ktea has not been fully loaded yet (config.LoadedMsg not been processed)
			if m.tabCtrl != nil {
				cmds = append(cmds, m.tabCtrl.Update(ui.RegainedFocusMsg{}))
			}
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

func (m *Model) recreateTabs(cluster *config.Cluster) {
	titles := []tab.Tab{topicsTab, cgroupsTab, clustersTab}

	if cluster.HasSchemaRegistry() {
		titles = slices.Insert(titles, 2, schemaRegTab)
	}

	if cluster.HasKafkaConnect() {
		titles = slices.Insert(titles, len(titles)-1, kconnectTab)
	}

	m.tabs = tab.New(titles...)
}

// recreateAdminClients (re)creates the kadmin.Model and kadmin.SrAdmin
// based on the given cluster
func (m *Model) recreateAdminClients(cluster *config.Cluster) error {
	connDetails := kadmin.ToConnectionDetails(cluster)
	if ka, err := m.kaInstantiator(connDetails); err != nil {
		return err
	} else {
		m.ka = ka
	}

	if cluster.HasSchemaRegistry() {
		m.sra = sradmin.New(m.ktx.Config().ActiveCluster().SchemaRegistry)
		m.ka.SetSra(m.sra)
	}

	return nil
}

func (m *Model) onWindowSizeUpdated(msg tea.WindowSizeMsg) {
	m.ktx.WindowWidth = msg.Width
	m.ktx.WindowHeight = msg.Height
	m.ktx.AvailableHeight = msg.Height
}

func (m *Model) boostrapUI(cluster *config.Cluster) tea.Cmd {
	var cmd tea.Cmd
	m.statusbar = statusbar.New()
	if err := m.recreateAdminClients(cluster); err != nil {
		m.tabs = tab.New(clustersTab)
		m.clustersTabCtrl, cmd = clusters_tab.New(m.ktx, kadmin.CheckKafkaConnectivity, sradmin.CheckSchemaRegistryConn, m.statusbar)
		m.startupConnErr = true
		m.tabCtrl = m.clustersTabCtrl
		return tea.Batch(cmd, func() tea.Msg {
			return kadmin.ConnErrMsg{
				Err: err,
			}
		})
	} else {
		var cmds []tea.Cmd
		m.recreateTabs(cluster)
		m.topicsTabCtrl, cmd = topics_tab.New(m.ktx, m.ka, m.statusbar)
		cmds = append(cmds, cmd)
		m.cgroupsTabCtrl, cmd = cgroups_tab.New(m.ka, m.ka, m.ka, m.statusbar)
		cmds = append(cmds, cmd)
		m.clustersTabCtrl, cmd = clusters_tab.New(m.ktx, kadmin.CheckKafkaConnectivity, sradmin.CheckSchemaRegistryConn, m.statusbar)
		cmds = append(cmds, cmd)
		m.kconTabCtrl, cmd = kcon_tab.New(m.ktx.Config().ActiveCluster(), m.statusbar)
		cmds = append(cmds, cmd)

		if m.ktx.Config().ActiveCluster().HasSchemaRegistry() {
			m.schemaRegistryTabCtrl, cmd = sr_tab.New(m.sra, m.ktx, m.statusbar)
			cmds = append(cmds, cmd)
		}

		m.tabCtrl = m.topicsTabCtrl

		return tea.Batch(cmds...)
	}
}

func NewModel(
	disableNerdFonts *bool,
	kai kadmin.Instantiator,
	configIO config.IO,
) *Model {
	return &Model{
		kaInstantiator: kai,
		ktx:            kontext.New(disableNerdFonts),
		configIO:       configIO,
	}
}

func main() {
	var (
		debugParam      bool
		plainFontsParam bool
	)
	flag.BoolVar(&debugParam, "debug", false, "enable debug")
	flag.BoolVar(&plainFontsParam, "plain-fonts", false, "disable NerdFonts (if you see weird icons)")
	flag.Parse()

	plainFontParamSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "plain-fonts" {
			plainFontParamSet = true
		}
	})

	var disableNerdFonts *bool
	if plainFontParamSet {
		disableNerdFonts = &plainFontsParam
	}

	p := tea.NewProgram(
		NewModel(
			disableNerdFonts,
			kadmin.SaramaInstantiator(),
			config.NewDefaultIO(),
		),
		tea.WithAltScreen(),
	)

	if debugParam {
		var fileErr error
		debugFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if fileErr == nil {
			log.SetOutput(debugFile)
			log.SetTimeFormat(time.Kitchen)
			log.SetReportCaller(true)
			log.SetLevel(log.DebugLevel)
			log.Debug("Logging to debug.log")
			log.Info("started")
		}
	} else {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.FatalLevel)
	}

	if _, err := p.Run(); err != nil {
		log.Fatal("Failed starting the TUI", err)
	}
}
