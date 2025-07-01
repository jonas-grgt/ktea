package cluster_config_page

import (
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"

	"github.com/charmbracelet/bubbles/spinner"
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

type Model struct {
	state    state
	spinner  spinner.Model
	viewport *viewport.Model
	cluster  *config.Cluster
	configs  kadmin.ClusterConfig
	err      error
}

func New(cluster *config.Cluster, ka kadmin.Kadmin) (*Model, tea.Cmd) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &Model{
			state:   loadingState,
			spinner: s,
			cluster: cluster,
		}, func() tea.Msg {
			log.Info("Fetching cluster configuration", "clusterName", cluster.Name)
			cfg, err := ka.GetClusterConfig()
			if err != nil {
				return kadmin.ClusterConfigListingErrorMsg{Err: err}
			}
			return kadmin.ClusterConfigListedMsg{Config: cfg}
		}
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	switch m.state {
	case loadingState:
		return fmt.Sprintf("%s Loading cluster configuration...", m.spinner.View())
	case errorState:
		return fmt.Sprintf("Error: %s", m.err.Error())
	default:
		if m.viewport == nil {
			vp := viewport.New(ktx.WindowWidth-2, ktx.AvailableHeight-1)
			m.viewport = &vp
			content := ""
			for _, broker := range m.configs.Brokers {
				content += fmt.Sprintf("\nBroker ID: %d\nAddress: %s\n", broker.ID, broker.Addr)

				// Marshal the broker configs to YAML for display
				configBytes, err := yaml.Marshal(broker.Configs)
				if err != nil {
					content += fmt.Sprintf("  Error marshalling configs: %v\n", err)
				} else {
					content += fmt.Sprintf("Configs:\n%s\n", string(configBytes))
				}
			}
			m.viewport.SetContent(lipgloss.NewStyle().Padding(1).Render(content))
		} else {
			m.viewport.Width = ktx.WindowWidth - 2
			m.viewport.Height = ktx.AvailableHeight - 1
		}
		return m.viewport.View()
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case kadmin.ClusterConfigListedMsg:
		m.state = loadedState
		m.configs = msg.Config

	case kadmin.ClusterConfigListingErrorMsg:
		m.state = errorState
		m.err = msg.Err

	default:
		if m.viewport != nil {
			vp, cmd := m.viewport.Update(msg)
			m.viewport = &vp
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) Title() string {
	return "Cluster Configuration"
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Go Back", "esc"},
	}
}
