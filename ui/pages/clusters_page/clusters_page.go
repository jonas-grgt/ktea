package clusters_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"reflect"
	"sort"
	"strings"
)

type Model struct {
	table         *table.Model
	rows          []table.Row
	ktx           *kontext.ProgramKtx
	cmdBar        *cmdbar.TableCmdsBar[string]
	tableFocussed bool
}

func (m *Model) Title() string {
	return "Clusters"
}

type ClusterSwitchedMsg struct {
	Cluster *config.Cluster
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.table.SetColumns([]table.Column{
		{"Active", int(float64(ktx.WindowWidth-5) * 0.05)},
		{"Name", int(float64(ktx.WindowWidth-5) * 0.95)},
	})
	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetRows(m.rows)

	var ts lipgloss.Style
	if m.tableFocussed {
		ts = styles.Table.Focus
	} else {
		ts = styles.Table.Blur
	}
	tableView := renderer.RenderWithStyle(m.table.View(), ts)

	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Switch Cluster", "enter"},
		{"Edit", "C-e"},
		{"Delete", "F2"},
		{"Create", "C-n"},
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	log.Debug(reflect.TypeOf(msg))

	var cmds []tea.Cmd

	msg, cmd := m.cmdBar.Update(msg, m.SelectedCluster())
	m.tableFocussed = !m.cmdBar.IsFocussed()
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case ClusterSwitchedMsg:
		// immediately recreate the rows updating the active cluster
		m.rows = m.createRows()
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			cmds = append(cmds, func() tea.Msg {
				activeCluster := m.ktx.Config.SwitchCluster(*m.SelectedCluster())
				m.rows = m.createRows()
				return ClusterSwitchedMsg{activeCluster}
			})
		}
	}

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = &t
		cmds = append(cmds, cmd)
	}

	m.rows = m.createRows()

	return tea.Batch(cmds...)
}

func (m *Model) SelectedCluster() *string {
	row := m.table.SelectedRow()
	if row == nil {
		return nil
	}
	return &row[1]
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	for _, c := range m.ktx.Config.Clusters {
		if m.cmdBar.GetSearchTerm() != "" {
			if !strings.Contains(strings.ToUpper(c.Name), strings.ToUpper(m.cmdBar.GetSearchTerm())) {
				continue
			}
		}
		var activeCell string
		if c.Active {
			activeCell = "X"
		} else {
			activeCell = ""
		}
		rows = append(rows, table.Row{activeCell, c.Name})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][1] < rows[j][1]
	})
	return rows
}

type ActiveClusterDeleteErrMsg struct {
}

func New(ktx *kontext.ProgramKtx) (nav.Page, tea.Cmd) {

	model := Model{}

	deleteFunc := func(subject string) tea.Cmd {
		return func() tea.Msg {
			selectedCluster := *model.SelectedCluster()
			model.ktx.Config.DeleteCluster(selectedCluster)
			return config.ClusterDeletedMsg{Name: selectedCluster}
		}
	}
	deleteMsgFunc := func(subject string) string {
		message := subject + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7571F9")).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	validateFunc := func(subject string) (bool, tea.Cmd) {
		if ktx.Config.ActiveCluster().Name == subject {
			return false, func() tea.Msg {
				return ActiveClusterDeleteErrMsg{}
			}
		}
		return true, nil
	}

	searchCmdBar := cmdbar.NewSearchCmdBar("Search clusters by name")
	deleteCmdBar := cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc, validateFunc)
	notifierCmdBar := cmdbar.NewNotifierCmdBar()

	clusterDeletedHandler := func(msg config.ClusterDeletedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Cluster has been deleted")
		return true, m.AutoHideCmd()
	}
	activeClusterDeleteErrMsgHandler := func(msg ActiveClusterDeleteErrMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Unable to delete", fmt.Errorf("active cluster"))
		return true, m.AutoHideCmd()
	}
	cmdbar.WithMsgHandler(notifierCmdBar, clusterDeletedHandler)
	cmdbar.WithMsgHandler(notifierCmdBar, activeClusterDeleteErrMsgHandler)

	model.ktx = ktx
	t := ktable.NewDefaultTable()
	model.table = &t
	model.cmdBar = cmdbar.NewTableCmdsBar(
		deleteCmdBar,
		searchCmdBar,
		notifierCmdBar,
	)
	model.rows = model.createRows()
	return &model, nil
}
