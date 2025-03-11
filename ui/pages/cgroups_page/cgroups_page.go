package cgroups_page

import (
	"fmt"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	lister        kadmin.CGroupLister
	table         table.Model
	cmdBar        *cmdbar.TableCmdsBar[string]
	groups        []*kadmin.ConsumerGroup
	rows          []table.Row
	tableFocussed bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := m.cmdBar.View(ktx, renderer)
	views = append(views, cmdBarView)

	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 3)
	m.table.SetColumns([]table.Column{
		{"Consumer Group", int(float64(ktx.WindowWidth-6) * 0.7)},
		{"Members", int(float64(ktx.WindowWidth-6) * 0.3)},
	})
	m.table.SetRows(m.rows)

	var tableView string
	if m.tableFocussed {
		// Apply focus style
		styledTable := renderer.RenderWithStyle(m.table.View(), styles.Table.Focus)

		embeddedText := map[styles.BorderPosition]string{
			styles.TopMiddleBorder: lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorPink)).
				Bold(true).
				Render(fmt.Sprintf("Total Consumer Groups: %d", len(m.rows))),
		}
		tableView = styles.Borderize(styledTable, true, embeddedText)
	} else {
		// Regular border approach for unfocused state
		styledTable := renderer.RenderWithStyle(m.table.View(), styles.Table.Blur)

		embeddedText := map[styles.BorderPosition]string{
			styles.TopMiddleBorder: lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorPink)).
				Bold(true).
				Render(fmt.Sprintf("Total Consumer Groups: %d", len(m.rows))),
		}
		tableView = styles.Borderize(styledTable, false, embeddedText)
	}

	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// only accept enter when the table is focussed
			if !m.cmdBar.IsFocussed() {
				// TODO ignore enter when there are no groups loaded
				return ui.PublishMsg(nav.LoadCGroupTopicsPageMsg{GroupName: *m.SelectedCGroup()})
			}
		case "f5":
			return m.lister.ListCGroups
		}
	case kadmin.ConsumerGroupsListedMsg:
		m.groups = msg.ConsumerGroups
	case kadmin.CGroupDeletionStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	}

	var cmd tea.Cmd

	msg, cmd = m.cmdBar.Update(msg, m.SelectedCGroup())
	m.tableFocussed = !m.cmdBar.IsFocussed()
	cmds = append(cmds, cmd)

	m.rows = m.filterCGroupsBySearchTerm()

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = t
		cmds = append(cmds, cmd)
	}

	if m.cmdBar.HasSearchedAtLeastOneChar() {
		m.table.GotoTop()
	}

	return tea.Batch(cmds...)
}

func (m *Model) filterCGroupsBySearchTerm() []table.Row {
	var rows []table.Row
	for _, group := range m.groups {
		if m.cmdBar.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(group.Name), strings.ToLower(m.cmdBar.GetSearchTerm())) {
				rows = m.apppendGroupToRows(rows, group)
			}
		} else {
			rows = m.apppendGroupToRows(rows, group)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	return rows
}

func (m *Model) apppendGroupToRows(rows []table.Row, group *kadmin.ConsumerGroup) []table.Row {
	rows = append(
		rows,
		table.Row{
			group.Name,
			strconv.Itoa(len(group.Members)),
		},
	)
	return rows
}

func (m *Model) SelectedCGroup() *string {
	selectedRow := m.table.SelectedRow()
	var selectedTopic string
	if selectedRow != nil {
		selectedTopic = selectedRow[0]
	}
	return &selectedTopic
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Search", "/"},
		{"View", "enter"},
		{"Refresh", "F5"},
	}
}

func (m *Model) Title() string {
	return "Consumer Groups"
}

func New(lister kadmin.CGroupLister, deleter kadmin.CGroupDeleter) (*Model, tea.Cmd) {
	m := &Model{}
	m.lister = lister

	// Use ktable.NewDefaultTable() instead of direct initialization
	t := ktable.NewDefaultTable()
	m.table = t

	deleteMsgFunc := func(topic string) string {
		message := topic + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7571F9")).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(group string) tea.Cmd {
		return func() tea.Msg {
			return deleter.DeleteCGroup(group)
		}
	}
	notifierCmdBar := cmdbar.NewNotifierCmdBar("cgroups-page")
	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.ConsumerGroupListingStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Loading Consumer Groups")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.ConsumerGroupsListedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return true, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.CGroupDeletionStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Deleting Consumer Group")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.CGroupDeletedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return false, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.CGroupDeletionErrMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.ShowErrorMsg("Failed to delete group", msg.Err)
			return true, cmd
		},
	)

	m.cmdBar = cmdbar.NewTableCmdsBar[string](
		cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc, nil),
		cmdbar.NewSearchCmdBar("Search Consumer Group"),
		notifierCmdBar,
	)
	return m, m.lister.ListCGroups
}
