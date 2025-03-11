package configs_page

import (
	"fmt"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	rows    []table.Row
	table   *table.Model
	cmdBar  *CmdBarModel
	configs map[string]string
	topic   string
	err     error
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := m.cmdBar.View(ktx, renderer)
	views = append(views, cmdBarView)

	// TODO errors should not be checked here
	//if m.err != nil {
	//	builder.WriteString(m.err.Error())
	//}
	m.table.SetColumns([]table.Column{
		{Title: "Config", Width: int(float64(ktx.WindowWidth-5) * 0.5)},
		{Title: "Value", Width: int(float64(ktx.WindowWidth-5) * 0.5)},
	})
	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetRows(m.rows)
	m.table.Focus()

	borderedView := styles.Borderize(m.table.View(), m.cmdBar.IsFocused(), nil)
	views = append(views, borderedView)

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.cmdBar.IsLoading() {
			return nil
		}
		selectedRow := m.table.SelectedRow()
		var config SelectedTopicConfig
		if selectedRow != nil {
			config = SelectedTopicConfig{
				Topic:       m.topic,
				ConfigKey:   selectedRow[0],
				ConfigValue: selectedRow[1],
			}
		}
		um, c := m.cmdBar.Update(msg, config)
		if c != nil {
			cmds = append(cmds, c)
		}
		if um != nil {
			t, c := m.table.Update(um)
			if c != nil {
				cmds = append(cmds, c)
			}
			m.table = &t
		}
	case kadmin.TopicConfigListingStartedMsg:
		_, cmd := m.cmdBar.Update(msg, SelectedTopicConfig{})
		return tea.Batch(cmd, msg.AwaitCompletion)
	case kadmin.TopicConfigsListedMsg:
		m.cmdBar.Update(msg, SelectedTopicConfig{})
		m.configs = msg.Configs
	default:
		selectedRow := m.table.SelectedRow()
		var config SelectedTopicConfig
		if selectedRow != nil {
			config = SelectedTopicConfig{
				Topic:       m.topic,
				ConfigKey:   selectedRow[0],
				ConfigValue: selectedRow[1],
			}
		}
		_, c := m.cmdBar.Update(msg, config)
		return c
	}

	keys := make([]string, 0, len(m.configs))
	for k := range m.configs {
		if m.cmdBar.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(k), strings.ToLower(m.cmdBar.GetSearchTerm())) {
				keys = append(keys, k)
			}
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var rows []table.Row
	for _, k := range keys {
		rows = append(rows, table.Row{k, m.configs[k]})
	}
	m.rows = rows
	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Search", "/"},
		{"Edit", "e"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return fmt.Sprintf("Topics / %s / Configuration", m.topic)
}

func New(configUpdater kadmin.ConfigUpdater, topicConfigLister kadmin.TopicConfigLister, topic string) (*Model, tea.Cmd) {
	m := &Model{}
	m.cmdBar = NewCmdBar(configUpdater, topicConfigLister, topic)
	t := table.New(
		table.WithStyles(styles.Table.Styles),
	)
	m.table = &t
	m.topic = topic
	return m, func() tea.Msg { return topicConfigLister.ListConfigs(topic) }
}
