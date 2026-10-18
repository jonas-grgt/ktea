package consume_page

import (
	"context"
	"fmt"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/tabs"
	"sort"
	"strconv"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	table              *table.Model
	border             *border.Model
	cmdBar             *ConsumptionCmdBar
	cancelConsumption  context.CancelFunc
	reader             kadmin.RecordReader
	rows               []table.Row
	records            []kadmin.ConsumerRecord
	readDetails        kadmin.ReadDetails
	consuming          bool
	noRecordsAvailable bool
	noRecordsFound     bool
	topic              *kadmin.ListedTopic
	navigator          tabs.TopicsTabNavigator
	search             *searchModal
	// readStarted is false while the page waits for a search to be submitted
	readStarted bool
	// source identifies the current consumption, see isStale
	source  chan kadmin.ConsumerRecord
	goToTop bool
}

// isStale reports whether msg belongs to a consumption that has been replaced
// by a newer search.
func (m *Model) isStale(msg tea.Msg) bool {
	var source chan kadmin.ConsumerRecord
	switch msg := msg.(type) {
	case kadmin.ConsumerRecordReceived:
		source = msg.Source()
	case kadmin.ConsumptionEndedMsg:
		source = msg.Source()
	}
	return source != nil && source != m.source
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	views = append(views, m.cmdBar.View(ktx, renderer))

	if m.noRecordsAvailable {
		views = append(views, styles.CenterText(ktx.WindowWidth, ktx.AvailableHeight).
			Render("👀 Empty topic"))
	} else if m.noRecordsFound {
		views = append(views, styles.CenterText(ktx.WindowWidth, ktx.AvailableHeight).
			Render("👀 No records found for the given criteria"))
	} else {
		keyCol := int(float64(ktx.WindowWidth) * 0.4)
		tsCol := int(float64(ktx.WindowWidth) * 0.2)
		PCol := int(float64(ktx.WindowWidth) * 0.2)
		oCol := ktx.WindowWidth - keyCol - tsCol - PCol - 10
		m.table.SetColumns([]table.Column{
			{Title: m.cmdBar.sortByCBar.PrefixSortIcon("Key"), Width: keyCol},
			{Title: m.cmdBar.sortByCBar.PrefixSortIcon("Timestamp"), Width: tsCol},
			{Title: m.cmdBar.sortByCBar.PrefixSortIcon("Partition"), Width: PCol},
			{Title: m.cmdBar.sortByCBar.PrefixSortIcon("Offset"), Width: oCol},
		})
		m.table.SetRows(m.rows)
		// a new search starts at the top; the cursor needs rows to move
		if m.goToTop && len(m.rows) > 0 {
			m.table.GotoTop()
			m.goToTop = false
		}
		m.table.SetWidth(ktx.WindowWidth - 2)
		m.table.SetHeight(ktx.AvailableTableHeight())

		views = append(views, m.border.View(m.table.View()))
	}

	content := ui.JoinVertical(lipgloss.Top, views...)
	if m.search != nil {
		content = ui.Overlay(content, m.search.View(ktx.WindowWidth), ktx.WindowWidth, ktx.AvailableHeight)
	}
	return content
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case searchClosedMsg:
		m.search = nil
		if !m.readStarted {
			return m.navigator.ToTopicsPage()
		}
		return nil
	case searchSubmittedMsg:
		m.search = nil
		return m.startReading(msg.details)
	}

	if m.isStale(msg) {
		return nil
	}

	// the search modal captures all key presses while open, other messages
	// still reach the page so a running consumption keeps going
	if m.search != nil {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			return m.search.Update(msg)
		}
		cmds = append(cmds, m.search.Update(msg))
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			m.cancelConsumption()
			return m.navigator.ToTopicsPage()
		} else if (msg.String() == "/" || msg.String() == "ctrl+f") && !m.cmdBar.IsFocussed() {
			m.search = newSearchModal(m.topic, m.readDetails, msg.String() == "ctrl+f")
			return nil
		} else if msg.String() == "f2" {
			m.cancelConsumption()
			m.consuming = false
			cmds = append(cmds, ui.PublishMsg(kadmin.ConsumptionEndedMsg{}))
		} else if msg.String() == "enter" {
			if !m.cmdBar.IsFocussed() {
				if len(m.records) > 0 {
					selectedRow := m.rows[m.table.Cursor()]
					selectedRecord := m.recordForRow(selectedRow)
					m.consuming = false
					recordIndex := m.recordIndexForRow(selectedRow)
					return m.navigator.ToRecordDetailsPage(
						tabs.LoadRecordDetailPageMsg{
							Record:    &selectedRecord,
							TopicName: m.readDetails.TopicName,
							Records:   m.records,
							Index:     recordIndex,
						})
				}
			}
		}
	case kadmin.EmptyTopicMsg:
		m.noRecordsAvailable = true
		m.consuming = false
	case kadmin.NoRecordsFound:
		m.noRecordsFound = true
		m.consuming = false
	case *kadmin.ReadingStartedMsg:
		m.source = msg.Source()
		m.consuming = true
		cmds = append(cmds, msg.AwaitRecord)
	case kadmin.ConsumptionEndedMsg:
		m.consuming = false
	case kadmin.ConsumerRecordReceived:
		m.records = append(m.records, msg.Records...)
		cmds = append(cmds, msg.AwaitNextRecord)
	}

	cmd := m.cmdBar.Update(msg)
	cmds = append(cmds, cmd)

	m.rows = m.createRows()

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = &t
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) recordForRow(row table.Row) kadmin.ConsumerRecord {
	offset, _ := strconv.ParseInt(row[3], 10, 64)
	partition, _ := strconv.ParseInt(row[2], 10, 32)
	for _, rec := range m.records {
		if rec.Partition == partition &&
			rec.Offset == offset {
			return rec
		}
	}
	panic(fmt.Sprintf("Record not found for row: %v", row))
}

func (m *Model) recordIndexForRow(row table.Row) int {
	offset, _ := strconv.ParseInt(row[3], 10, 64)
	partition, _ := strconv.ParseInt(row[2], 10, 32)
	for i, rec := range m.records {
		if rec.Partition == partition &&
			rec.Offset == offset {
			return i
		}
	}
	panic(fmt.Sprintf("Record not found for row: %v", row))
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	for _, rec := range m.records {

		var key string
		if rec.Key == "" {
			key = "<null>"
		} else {
			key = rec.Key
		}

		rows = append(
			rows,
			table.Row{
				key,
				rec.Timestamp.Format("2006-01-02 15:04:05"),
				strconv.FormatInt(rec.Partition, 10),
				strconv.FormatInt(rec.Offset, 10),
			},
		)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		var col int
		switch m.cmdBar.sortByCBar.SortedBy().Label {
		case "Key":
			col = 0
		case "Timestamp":
			col = 1
		case "Partition":
			col = 2
		case "Offset":
			col = 3
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.cmdBar.sortByCBar.SortedBy().Label))
		}

		if m.cmdBar.sortByCBar.SortedBy().Direction == cmdbar.Asc {
			return rows[i][col] < rows[j][col]
		}
		return rows[i][col] > rows[j][col]
	})

	return rows
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.search != nil {
		return m.search.Shortcuts()
	}
	if m.consuming {
		return []statusbar.Shortcut{
			{"View Record", "enter"},
			{"Stop consuming", "F2"},
			{"Go Back", "esc"},
		}
	} else if m.noRecordsAvailable || m.noRecordsFound {
		return []statusbar.Shortcut{
			{"Search", "/"},
			{"Advanced Search", "C-f"},
			{"Go Back", "esc"},
		}
	} else if m.cmdBar.IsSorting() {
		return []statusbar.Shortcut{
			{"Cancel Sorting", "F3"},
			{"Select Sorting Column", "←/→/h/l"},
			{"Apply Sorting Column", "enter"},
		}
	}

	return []statusbar.Shortcut{
		{"View Record", "enter"},
		{"Search", "/"},
		{"Advanced Search", "C-f"},
		{"Sort", "F3"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Topics / " + m.readDetails.TopicName + " / Records"
}

// New creates the records page. With openSearch it opens the expanded search
// modal and waits for it to be submitted instead of reading right away.
func New(
	reader kadmin.RecordReader,
	readDetails kadmin.ReadDetails,
	topic *kadmin.ListedTopic,
	openSearch bool,
	navigator tabs.TopicsTabNavigator,
) (pages.Page, tea.Cmd) {
	m := &Model{}

	t := table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	m.table = &t
	m.reader = reader
	m.cmdBar = NewConsumptionCmdbar()
	m.readDetails = readDetails
	m.topic = topic
	m.navigator = navigator
	m.cancelConsumption = func() {}

	m.border = border.New(
		border.WithInnerPaddingTop(),
		border.WithTitleFn(func() string {
			return border.KeyValueTitle("Records", fmt.Sprintf(" %d", len(m.rows)), true)
		}))

	if openSearch {
		m.search = newSearchModal(topic, readDetails, true)
		return m, nil
	}
	return m, m.startReading(readDetails)
}

// startReading cancels any running consumption and reads the topic anew.
func (m *Model) startReading(details kadmin.ReadDetails) tea.Cmd {
	m.cancelConsumption()
	ctx, cancelFn := context.WithCancel(context.Background())
	m.cancelConsumption = cancelFn

	m.readDetails = details
	m.readStarted = true
	m.source = nil
	m.records = nil
	m.rows = nil
	m.noRecordsAvailable = false
	m.noRecordsFound = false
	m.goToTop = true

	return func() tea.Msg {
		msg := m.reader.ReadRecords(ctx, details)
		if ctx.Err() != nil {
			// replaced by a newer search while starting
			return nil
		}
		return msg
	}
}
