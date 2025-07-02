package cluster_config_page

import (
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

type state int

const (
	loadingState state = iota
	loadedState
	errorState
)

type viewMode int

const (
	brokerListMode viewMode = iota
	brokerDetailMode
)

type Model struct {
	state          state
	spinner        spinner.Model
	viewport       *viewport.Model
	cluster        *config.Cluster
	clusterConfigs kadmin.ClusterConfig
	brokerConfig   kadmin.BrokerConfig
	err            error

	mode        viewMode
	brokerTable table.Model
	ka          kadmin.Kadmin
}

func New(cluster *config.Cluster, ka kadmin.Kadmin) (*Model, tea.Cmd) {
	s := spinner.New()
	s.Spinner = spinner.Dot

	t := ktable.NewDefaultTable()
	t.SetColumns([]table.Column{
		{Title: "ID", Width: 5},
		{Title: "Address", Width: 30},
	})

	return &Model{
		state:       loadingState,
		spinner:     s,
		cluster:     cluster,
		mode:        brokerListMode,
		brokerTable: t,
		ka:          ka,
	}, ka.GetClusterConfig
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	switch m.state {
	case loadingState:
		return fmt.Sprintf("%s Loading cluster configuration...", m.spinner.View())
	case errorState:
		return fmt.Sprintf("Error: %s", m.err.Error())
	case loadedState:
		switch m.mode {
		case brokerListMode:
			m.brokerTable.SetHeight(ktx.AvailableHeight - 1)
			m.brokerTable.SetWidth(ktx.WindowWidth - 2)
			return m.brokerTable.View()
		case brokerDetailMode:
			if m.viewport == nil {
				vp := viewport.New(ktx.WindowWidth-2, ktx.AvailableHeight-1)
				m.viewport = &vp

				content := fmt.Sprintf("Broker ID: %d\n\n", m.brokerConfig.ID)
				configBytes, err := yaml.Marshal(m.brokerConfig.Configs)
				if err != nil {
					content += fmt.Sprintf("  Error marshalling configs: %v\n", err)
				} else {
					content += fmt.Sprintf("Configs:\n%s\n", string(configBytes))
				}
				log.Debug("Broker Configs", "configs", m.brokerConfig.Configs)
				m.viewport.SetContent(lipgloss.NewStyle().Padding(1).Render(content))
			} else {
				m.viewport.Width = ktx.WindowWidth - 2
				m.viewport.Height = ktx.AvailableHeight - 1
			}
			return m.viewport.View()
		}
	}
	return ""
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case kadmin.ClusterConfigStartedMsg:
		m.state = loadingState
		return msg.AwaitCompletion

	case kadmin.ClusterConfigMsg:
		m.state = loadedState
		m.clusterConfigs = msg.Config
		rows := make([]table.Row, len(m.clusterConfigs.Brokers))
		for i, broker := range m.clusterConfigs.Brokers {
			rows[i] = table.Row{fmt.Sprintf("%d", broker.ID), broker.Address}
		}
		m.brokerTable.SetRows(rows)

	case kadmin.ClusterConfigErrorMsg:
		m.state = errorState
		m.err = msg.Err

	case kadmin.BrokerConfigListingStartedMsg:
		m.state = loadingState
		return msg.AwaitCompletion

	case kadmin.BrokerConfigListedMsg:
		m.state = loadedState
		m.brokerConfig = msg.Config
		m.mode = brokerDetailMode
		m.viewport = nil

	case kadmin.BrokerConfigErrorMsg:
		m.state = errorState
		m.err = msg.Err

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.mode == brokerListMode {
				selectedRow := m.brokerTable.SelectedRow()
				if selectedRow != nil {
					brokerIDStr := selectedRow[0]
					brokerID := 0
					fmt.Sscanf(brokerIDStr, "%d", &brokerID)

					// Trigger command to fetch single broker config
					cmds = append(cmds, func() tea.Msg { return m.ka.GetBrokerConfig(int32(brokerID)) })
				}
			}
		case "esc":
			if m.mode == brokerDetailMode {
				m.mode = brokerListMode
				m.viewport = nil
			} else if m.mode == brokerListMode {
				return func() tea.Msg { return GoBackMsg{} }
			}
		}
	}

	if m.mode == brokerListMode {
		t, cmd := m.brokerTable.Update(msg)
		m.brokerTable = t
		cmds = append(cmds, cmd)
	} else if m.viewport != nil {
		vp, cmd := m.viewport.Update(msg)
		m.viewport = &vp
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

type GoBackMsg struct{}

func (m *Model) Title() string {
	if m.mode == brokerDetailMode && m.brokerConfig.ID != 0 {
		return fmt.Sprintf("Broker %d Configuration", m.brokerConfig.ID)
	}
	return "Cluster Configuration"
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.mode == brokerDetailMode {
		return []statusbar.Shortcut{
			{"Go Back", "esc"},
		}
	}
	return []statusbar.Shortcut{
		{"View Broker Config", "enter"},
		{"Go Back", "esc"},
	}
}
