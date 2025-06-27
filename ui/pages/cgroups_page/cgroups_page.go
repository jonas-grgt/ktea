package cgroups_page

import (
	"fmt"
	"github.com/charmbracelet/log"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateRefreshing state = iota
	stateLoading
	stateLoaded
)

type Model struct {
	lister        kadmin.CGroupLister
	table         table.Model
	cmdBar        *cmdbar.TableCmdsBar[string]
	groups        []*kadmin.ConsumerGroup
	rows          []table.Row
	tableFocussed bool
	sort          cmdbar.SortLabel
	state         state
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := m.cmdBar.View(ktx, renderer)
	views = append(views, cmdBarView)

	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{m.columnTitle("Consumer Group"), int(float64(ktx.WindowWidth-5) * 0.7)},
		{m.columnTitle("Members"), int(float64(ktx.WindowWidth-5) * 0.3)},
	})
	m.table.SetRows(m.rows)
	m.table.SetHeight(ktx.AvailableHeight - 2)

	var tableView string

	styledTable := renderer.RenderWithStyle(m.table.View(), styles.Table.Blur)

	embeddedText := map[styles.BorderPosition]styles.EmbeddedTextFunc{
		styles.TopMiddleBorder:    styles.BorderKeyValueTitle("Total Consumer Groups", fmt.Sprintf(" %d/%d", len(m.rows), len(m.groups))),
		styles.BottomMiddleBorder: styles.BorderKeyValueTitle("Total Consumer Groups", fmt.Sprintf(" %d/%d", len(m.rows), len(m.groups))),
	}
	tableView = styles.Borderize(styledTable, m.tableFocussed, embeddedText)

	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

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
			m.groups = nil
			m.state = stateRefreshing
			return m.lister.ListCGroups
		}
	case kadmin.ConsumerGroupsListedMsg:
		m.state = stateLoaded
		m.groups = msg.ConsumerGroups
	case kadmin.CGroupDeletionStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.CGroupDeletedMsg:
		for i, group := range m.groups {
			if group.Name == msg.GroupName {
				m.groups = slices.Delete(m.groups, i, i+1)
				break
			}
		}
	}

	var cmd tea.Cmd

	msg, cmd = m.cmdBar.Update(msg, m.SelectedCGroup())
	m.tableFocussed = !m.cmdBar.IsFocussed()
	cmds = append(cmds, cmd)

	m.rows = m.createRows()

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

func (m *Model) columnTitle(title string) string {
	if m.sort.Label == title {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorPink)).
			Bold(true).
			Render(m.sort.Direction.String()) + " " + title
	}
	return title
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	for _, group := range m.groups {
		if m.cmdBar.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(group.Name), strings.ToLower(m.cmdBar.GetSearchTerm())) {
				rows = m.appendGroupToRows(rows, group)
			}
		} else {
			rows = m.appendGroupToRows(rows, group)
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		switch m.sort.Label {
		case "Consumer Group":
			if m.sort.Direction == cmdbar.Asc {
				return rows[i][0] < rows[j][0]
			}
			return rows[i][0] > rows[j][0]
		case "Members":
			partitionI, _ := strconv.Atoi(rows[i][1])
			partitionJ, _ := strconv.Atoi(rows[j][1])
			if m.sort.Direction == cmdbar.Asc {
				return partitionI < partitionJ
			}
			return partitionI > partitionJ
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.sort.Label))
		}
	})
	return rows
}

func (m *Model) appendGroupToRows(rows []table.Row, group *kadmin.ConsumerGroup) []table.Row {
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
	if m.cmdBar.IsFocussed() {
		shortCuts := m.cmdBar.Shortcuts()
		if shortCuts != nil {
			return shortCuts
		}
	}
	return []statusbar.Shortcut{
		{"Search", "/"},
		{"View", "enter"},
		{"Delete", "F2"},
		{"Sort", "F3"},
		{"Refresh", "F5"},
	}
}

func (m *Model) Title() string {
	return "Consumer Groups"
}

func New(
	lister kadmin.CGroupLister,
	deleter kadmin.CGroupDeleter,
) (*Model, tea.Cmd) {
	m := &Model{}
	m.lister = lister

	// Use ktable.NewDefaultTable() instead of direct initialization
	t := ktable.NewDefaultTable()
	m.table = t

	deleteMsgFunc := func(topic string) string {
		message := topic + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
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
			msg ui.RegainedFocusMsg,
			model *notifier.Model,
		) (bool, tea.Cmd) {
			if m.state == stateRefreshing || m.state == stateLoading {
				cmd := model.SpinWithLoadingMsg("Loading Consumer Groups")
				return true, cmd
			}
			return false, nil
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

	sortByBar := cmdbar.NewSortByCmdBar(
		[]cmdbar.SortLabel{
			{
				Label:     "Consumer Group",
				Direction: cmdbar.Asc,
			},
			{
				Label:     "Members",
				Direction: cmdbar.Desc,
			},
		},
		cmdbar.WithSortSelectedCallback(func(label cmdbar.SortLabel) {
			m.sort = label
		}),
	)

	m.cmdBar = cmdbar.NewTableCmdsBar[string](
		cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc, nil),
		cmdbar.NewSearchCmdBar("Search Consumer Group"),
		notifierCmdBar,
		sortByBar,
	)
	m.sort = sortByBar.SortedBy()
	m.state = stateLoading
	return m, m.lister.ListCGroups
}
