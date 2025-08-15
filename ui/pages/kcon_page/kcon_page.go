package kcon_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	border     *border.Model
	cmdBar     *cmdbar.TableCmdsBar[string]
	table      *table.Model
	stsSpinner *spinner.Model
	sort       cmdbar.SortLabel

	connectors *kcadmin.Connectors
	rows       []table.Row

	sortByCmdBar *cmdbar.SortByCmdBar

	kca     kcadmin.Admin
	navBack ui.NavBack

	connectorName string
	state
	stateChangingConnectorName string
	resumeDeadline             *time.Time
	connectorChangeState       string
}

type state int

type LoadPage func(c config.KafkaConnectConfig) tea.Cmd

const (
	loading state = iota
	loaded
	errored
)

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.rows = m.createRows()

	available := ktx.WindowWidth - 8
	nameCol := int(float64(available) * 0.8)
	tasksCol := int(float64(available) * 0.08)
	compCol := available - nameCol - tasksCol
	m.table.SetColumns([]table.Column{
		{m.sortByCmdBar.PrefixSortIcon("Name"), nameCol},
		{m.sortByCmdBar.PrefixSortIcon("Tasks"), tasksCol},
		{m.sortByCmdBar.PrefixSortIcon("Status"), compCol},
	})
	m.table.SetRows(m.rows)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetHeight(ktx.AvailableHeight - 2)

	tableView := m.border.View(m.table.View())
	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

type ConnectorStateAlreadyChanging struct {
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if !m.cmdBar.IsFocussed() {
				return m.navBack()
			}
		case "P":
			if m.stateChangingConnectorName == "" {
				var name = m.selectedConnector()
				log.Debug("Pausing", "connector name", name)
				m.connectorChangeState = "PAUSING"
				return func() tea.Msg {
					return m.kca.Pause(name)
				}
			} else {
				return func() tea.Msg {
					return ConnectorStateAlreadyChanging{}
				}
			}
		case "R":
			if m.stateChangingConnectorName == "" {
				var name = m.selectedConnector()
				m.connectorChangeState = "RESUMING"
				log.Debug("Resuming", "connector name", name)
				return func() tea.Msg {
					return m.kca.Resume(name)
				}
			} else {
				return func() tea.Msg {
					return ConnectorStateAlreadyChanging{}
				}
			}
		case "f5":
			if m.state == loading {
				log.Debug("Not refreshing connectors due to loading state")
				return nil
			}
			m.connectors = nil
			m.rows = m.createRows()
			m.state = loading
			log.Debug("Refreshing connectors")
			return m.kca.ListActiveConnectors
		}

	case spinner.TickMsg:
		if m.stsSpinner != nil {
			sm, cmd := m.stsSpinner.Update(msg)
			m.stsSpinner = &sm
			cmds = append(cmds, cmd)
		}

	case kcadmin.ResumingStartedMsg:
		var name = m.selectedConnector()
		cmd := m.newSpinner(name)
		return tea.Batch(cmd, msg.AwaitCompletion)
	case kcadmin.ResumeRequestedMsg:
		return func() tea.Msg {
			return m.waitForConnectorState("RUNNING")
		}

	case kcadmin.PausingStartedMsg:
		var name = m.selectedConnector()
		cmd := m.newSpinner(name)
		return tea.Batch(cmd, msg.AwaitCompletion)
	case kcadmin.PauseRequestedMsg:
		return func() tea.Msg {
			return m.waitForConnectorState("PAUSED")
		}

	case ConnectorStateChanged:
		m.stsSpinner = nil
		m.connectors = msg.Connectors
		m.stateChangingConnectorName = ""

	case kcadmin.ConnectorListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kcadmin.ConnectorsListedMsg:
		m.connectors = &msg.Connectors
		m.state = loaded
	case kcadmin.ConnectorDeletionStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case kcadmin.ConnectorDeletedMsg:
		m.state = loaded
		delete(*m.connectors, msg.Name)
	}

	var (
		selection = m.selectedConnector()
		cmd       tea.Cmd
	)
	_, cmd = m.cmdBar.Update(msg, &selection)
	m.border.Focused = !m.cmdBar.IsFocussed()
	if cmd != nil {
		cmds = append(cmds, cmd)
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

func (m *Model) newSpinner(name string) tea.Cmd {
	model := spinner.New()
	m.stsSpinner = &model
	m.stsSpinner.Spinner = spinner.Dot
	m.stateChangingConnectorName = name
	return m.stsSpinner.Tick
}

type ConnectorStateChanged struct {
	Connectors *kcadmin.Connectors
}

func (m *Model) waitForConnectorState(state string) tea.Msg {
	if m.resumeDeadline == nil {
		resumeDeadline := time.Now().Add(30 * time.Second)
		m.resumeDeadline = &resumeDeadline
	}
	if time.Now().After(*m.resumeDeadline) {
		log.Warn("Timeout waiting for connector to resume", "connector", m.stateChangingConnectorName)
		m.resumeDeadline = nil
		return kcadmin.ConnectorListingErrMsg{
			Err: fmt.Errorf("timeout waiting for connector %s to resume", m.stateChangingConnectorName),
		}
	}

	msg := m.kca.ListActiveConnectors()

	startedMsg, ok := msg.(kcadmin.ConnectorListingStartedMsg)
	if !ok {
		log.Error("Expected ConnectorListingStartedMsg but got something else", "msg", msg)
		return m.waitForConnectorState
	}

	completedMsg := startedMsg.AwaitCompletion()

	listedMsg, ok := completedMsg.(kcadmin.ConnectorsListedMsg)
	if !ok {
		log.Error("Expected ConnectorsListedMsg but got something else", "msg", completedMsg)
		return m.waitForConnectorState
	}

	res := listedMsg
	log.Error(res.Connectors)
	if res.Connectors[m.stateChangingConnectorName].Status.Connector.State != state {
		return m.waitForConnectorState(state)
	}

	return ConnectorStateChanged{
		Connectors: &res.Connectors,
	}
}

func (m *Model) selectedConnector() string {
	selectedRow := m.table.SelectedRow()
	var selectedConnector string
	if selectedRow != nil {
		selectedConnector = selectedRow[0]
	}
	return selectedConnector
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
		{"Delete", "F2"},
		{"Sort", "F3"},
		{"Refresh", "F5"},
		{"Pause", "S-p"},
		{"Resume", "S-r"},
	}
}

func (m *Model) Title() string {
	return "Kafka Connect Clusters / " + m.connectorName
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	if m.connectors != nil {
		for k, c := range *m.connectors {
			if m.cmdBar.GetSearchTerm() != "" {
				if strings.Contains(strings.ToLower(k), strings.ToLower(m.cmdBar.GetSearchTerm())) {
					rows = append(rows, table.Row{k, strconv.Itoa(len(c.Status.Tasks)), c.Status.Connector.State})
				}
			} else {
				status := c.Status.Connector.State
				if m.stateChangingConnectorName == k {
					status = m.stsSpinner.View() + fmt.Sprintf(" %s ", m.connectorChangeState) + m.stsSpinner.View()
				}
				rows = append(rows, table.Row{k, strconv.Itoa(len(c.Status.Tasks)), status})
			}
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		switch m.sortByCmdBar.SortedBy().Label {
		case "Name":
			if m.sortByCmdBar.SortedBy().Direction == cmdbar.Asc {
				return rows[i][0] < rows[j][0]
			}
			return rows[i][0] > rows[j][0]
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.sortByCmdBar.SortedBy().Label))
		}
	})

	return rows
}

func New(
	navBack ui.NavBack,
	kca kcadmin.Admin,
	connectorName string,
) (*Model, tea.Cmd) {
	m := Model{}
	m.connectorName = connectorName
	m.navBack = navBack
	m.kca = kca
	m.state = loading

	m.border = border.New(
		border.WithTitleFunc(func() string {
			if m.connectors == nil {
				return styles.BorderTopTitle("Loading Connectors", "", false)
			}
			return styles.BorderTopTitle("Total Connectors", fmt.Sprintf("%d/%d", len(m.rows), len(*m.connectors)), true)
		}))
	m.border.Focused = false

	t := ktable.NewDefaultTable()
	m.table = &t

	sortByCmdBar := cmdbar.NewSortByCmdBar(
		[]cmdbar.SortLabel{
			{
				Label:     "Name",
				Direction: cmdbar.Asc,
			},
		},
		cmdbar.WithSortSelectedCallback(func(label cmdbar.SortLabel) {
			m.sort = label
		}),
	)
	m.sortByCmdBar = sortByCmdBar

	notifierCmdBar := cmdbar.NewNotifierCmdBar("kcons-page")
	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorListingStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Loading connectors")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorsListedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return true, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorListingErrMsg,
			model *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := model.ShowError(fmt.Errorf("unable to reach connect cluster: %v", msg.Err))
			m.state = errored
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorDeletionStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Deleting Connector")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorDeletedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.ShowSuccessMsg("Connector deleted")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorDeletionErrMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.ShowError(msg.Err)
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg ui.RegainedFocusMsg,
			model *notifier.Model,
		) (bool, tea.Cmd) {
			if model.State == notifier.Spinning {
				cmd := model.SpinWithLoadingMsg("Loading connectors")
				return true, cmd
			}
			if model.State == notifier.Err {
				return true, nil
			}
			return false, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.PauseRequestedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return true, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg ConnectorStateAlreadyChanging,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.ShowError(fmt.Errorf("Other connector already in progress of changing state, wait until it completes."))
			return true, m.AutoHideCmd("kcons-page")
		},
	)

	deleteMsgFunc := func(connector string) string {
		message := connector + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(name string) tea.Cmd {
		return func() tea.Msg {
			return kca.DeleteConnector(name)
		}
	}

	m.cmdBar = cmdbar.NewTableCmdsBar(
		cmdbar.NewDeleteCmdBar[string](deleteMsgFunc, deleteFunc),
		cmdbar.NewSearchCmdBar("Search connectors by name"),
		notifierCmdBar,
		sortByCmdBar)

	m.sort = sortByCmdBar.SortedBy()

	return &m, kca.ListActiveConnectors
}
