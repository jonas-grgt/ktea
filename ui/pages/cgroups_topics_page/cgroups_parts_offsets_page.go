package cgroups_topics_page

import (
	"fmt"
	"github.com/charmbracelet/log"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

type tableFocus int

type state int

const (
	topicFocus  tableFocus = 0
	offsetFocus tableFocus = 1

	stateNoOffsets      state = 0
	stateOffsetsLoading state = 1
	stateOffsetsLoaded  state = 2
)

type Model struct {
	tableFocus        tableFocus
	topicsTable       table.Model
	offsetsTable      table.Model
	offsetsBorder     *border.Model
	topicsBorder      *border.Model
	topicsRows        []table.Row
	offsetRows        []table.Row
	groupName         string
	topicByPartOffset map[string][]partOffset
	cmdBar            *CGroupCmdbar[string]
	offsets           []kadmin.TopicPartitionOffset
	state             state
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {

	if m.state == stateNoOffsets {
		return styles.
			CenterText(ktx.WindowWidth, ktx.AvailableHeight).
			Render("👀 No Committed Offsets Found")
	}

	cmdBarView := m.cmdBar.View(ktx, renderer)

	halfWidth := int(float64(ktx.WindowWidth / 2))
	m.topicsTable.SetHeight(ktx.AvailableHeight - 4)
	m.topicsTable.SetWidth(halfWidth - 2)
	m.topicsTable.SetColumns([]table.Column{
		{"Topic Name", int(float64(halfWidth - 4))},
	})
	m.topicsTable.SetRows(m.topicsRows)

	m.offsetsTable.SetHeight(ktx.AvailableHeight - 4)
	m.offsetsTable.SetColumns([]table.Column{
		{"Partition", int(float64(halfWidth-6) * 0.5)},
		{"Offset", int(float64(halfWidth-5) * 0.5)},
	})
	m.offsetsTable.SetRows(m.offsetRows)

	topicTableStyle := styles.Table.Blur
	offsetTableStyle := styles.Table.Blur
	if m.tableFocus == topicFocus {
		topicTableStyle = styles.Table.Focus
		offsetTableStyle = styles.Table.Blur
	}

	topicsView := m.topicsBorder.View(
		renderer.RenderWithStyle(m.topicsTable.View(), topicTableStyle),
	)
	offsetsView := m.offsetsBorder.View(
		renderer.RenderWithStyle(m.offsetsTable.View(), offsetTableStyle),
	)

	return ui.JoinVertical(lg.Left,
		cmdBarView,
		lg.JoinHorizontal(
			lg.Top,
			[]string{
				topicsView,
				offsetsView,
			}...,
		),
	)
}

type partOffset struct {
	partition string
	offset    int64
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// only accept when the table is focussed
			if !m.cmdBar.IsFocussed() {
				return ui.PublishMsg(nav.LoadCGroupsPageMsg{})
			}
		}
	case kadmin.OffsetListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.OffsetListedMsg:
		if msg.Offsets == nil {
			m.state = stateNoOffsets
		} else {
			m.state = stateOffsetsLoaded
			m.offsets = msg.Offsets
		}
	}

	var cmd tea.Cmd
	msg, cmd = m.cmdBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		if m.tableFocus == topicFocus {
			m.topicsTable, cmd = m.topicsTable.Update(msg)
		} else {
			m.offsetsTable, cmd = m.offsetsTable.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// recreate offset rows after topic table has been updated
	m.recreateTopicRows()
	m.recreateOffsetRows()

	return tea.Batch(cmds...)
}

func (m *Model) recreateOffsetRows() {
	// if topics aren't listed yet
	if m.topicsRows == nil {
		return
	}

	selectedTopic := m.selectedRow()
	if selectedTopic != "" {
		m.offsetRows = []table.Row{}
		for _, partOffset := range m.topicByPartOffset[selectedTopic] {
			m.offsetRows = append(m.offsetRows, table.Row{
				partOffset.partition,
				humanize.Comma(partOffset.offset),
			})
		}
		sort.SliceStable(m.offsetRows, func(i, j int) bool {
			a, _ := strconv.Atoi(m.offsetRows[i][0])
			b, _ := strconv.Atoi(m.offsetRows[j][0])
			return a < b
		})
	}
}

func (m *Model) recreateTopicRows() {
	if m.offsets == nil || len(m.offsets) == 0 {
		return
	}

	var topics []string
	m.topicByPartOffset = make(map[string][]partOffset)
	for _, offset := range m.offsets {
		if m.cmdBar.GetSearchTerm() != "" {
			if !strings.Contains(offset.Topic, m.cmdBar.GetSearchTerm()) {
				continue
			}
		}
		if !slices.Contains(topics, offset.Topic) {
			topics = append(topics, offset.Topic)
		}
		partOffset := partOffset{
			partition: strconv.FormatInt(int64(offset.Partition), 10),
			offset:    offset.Offset,
		}
		m.topicByPartOffset[offset.Topic] = append(m.topicByPartOffset[offset.Topic], partOffset)
	}
	m.topicsRows = []table.Row{}
	for _, topic := range topics {
		m.topicsRows = append(m.topicsRows, table.Row{topic})
	}
	sort.SliceStable(m.topicsRows, func(i, j int) bool {
		return m.topicsRows[i][0] < m.topicsRows[j][0]
	})
}

func (m *Model) selectedRow() string {
	row := m.topicsTable.SelectedRow()
	if row == nil {
		return m.topicsRows[0][0]
	}
	return row[0]
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Go Back", "esc"},
		{"Search", "/"},
		{"Refresh", "F5"},
	}
}

func (m *Model) Title() string {
	return "Consumer Groups / " + m.groupName
}

func New(lister kadmin.OffsetLister, group string) (*Model, tea.Cmd) {
	tt := table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	ot := table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)

	notifierCmdBar := cmdbar.NewNotifierCmdBar("cgroup")

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.OffsetListingStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Loading Offsets")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.OffsetListedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return false, m.AutoHideCmd("cgroup")
		},
	)

	model := Model{
		cmdBar: NewCGroupCmdbar[string](
			cmdbar.NewSearchCmdBar("Search groups by name"),
			notifierCmdBar,
		),
		tableFocus:   topicFocus,
		groupName:    group,
		topicsTable:  tt,
		offsetsTable: ot,
		state:        stateOffsetsLoading,
	}
	model.topicsBorder = border.New(
		border.WithInnerPaddingTop(),
		border.WithTitleFn(func() string {
			return border.KeyValueTitle("Total Topics", fmt.Sprintf(" %d", len(model.topicsRows)), true)
		}))
	model.offsetsBorder = border.New(border.WithInnerPaddingTop())
	return &model, func() tea.Msg {
		return lister.ListOffsets(group)
	}
}
