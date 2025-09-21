package consume_page

import (
	"context"
	"fmt"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/nav"
	"ktea/ui/tabs"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	origin             nav.Origin
	navigator          tabs.TopicsTabNavigator
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
	} else if len(m.rows) > 0 {
		m.table.SetColumns([]table.Column{
			{Title: "Key", Width: int(float64(ktx.WindowWidth-9) * 0.5)},
			{Title: "Timestamp", Width: int(float64(ktx.WindowWidth-9) * 0.30)},
			{Title: "Partition", Width: int(float64(ktx.WindowWidth-9) * 0.10)},
			{Title: "Offset", Width: int(float64(ktx.WindowWidth-9) * 0.10)},
		})
		m.table.SetRows(m.rows)
		m.table.SetWidth(ktx.WindowWidth - 2)
		m.table.SetHeight(ktx.AvailableTableHeight())

		views = append(views, m.border.View(m.table.View()))
	}

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	cmd := m.cmdBar.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.cancelConsumption()
			if m.readDetails.StartPoint == kadmin.Live {
				return ui.PublishMsg(nav.LoadTopicsPageMsg{})
			}
			if m.origin == nav.OriginTopicsPage {
				return m.navigator.ToTopicsPage()
			} else {
				return m.navigator.ToConsumeFormPage(
					nav.ConsumeFormPageDetails{
						ReadDetails: &m.readDetails,
						Topic:       m.topic,
					},
				)
			}
		} else if msg.String() == "f2" {
			m.cancelConsumption()
			m.consuming = false
			//cmds = append(cmds, ui.PublishMsg(ConsumptionEndedMsg{}))
		} else if msg.String() == "enter" {
			if len(m.records) > 0 {
				selectedRow := m.records[len(m.records)-m.table.Cursor()-1]
				m.consuming = false
				return ui.PublishMsg(nav.LoadRecordDetailPageMsg{
					Record:    &selectedRow,
					TopicName: m.readDetails.TopicName,
				})
			}
		} else {
			t, cmd := m.table.Update(msg)
			m.table = &t
			cmds = append(cmds, cmd)
		}
	case kadmin.EmptyTopicMsg:
		m.noRecordsAvailable = true
		m.consuming = false
	case kadmin.NoRecordsFound:
		m.noRecordsFound = true
		m.consuming = false
	case *kadmin.ReadingStartedMsg:
		m.consuming = true
		cmds = append(cmds, msg.AwaitRecord)
	case kadmin.ConsumptionEndedMsg:
		m.consuming = false
		return nil
	case kadmin.ConsumerRecordReceived:
		var key string
		for _, rec := range msg.Record {
			if rec.Key == "" {
				key = "<null>"
			} else {
				key = rec.Key
			}
			m.records = append(m.records, rec)
			m.rows = append(
				[]table.Row{
					{
						key,
						rec.Timestamp.Format("2006-01-02 15:04:05"),
						strconv.FormatInt(rec.Partition, 10),
						strconv.FormatInt(rec.Offset, 10),
					},
				},
				m.rows...,
			)
		}
		return msg.AwaitNextRecord
	}

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.consuming {
		return []statusbar.Shortcut{
			{"View Record", "enter"},
			{"Stop consuming", "F2"},
			{"Go Back", "esc"},
		}
	} else if m.noRecordsAvailable {
		return []statusbar.Shortcut{
			{"Go Back", "esc"},
		}
	} else {
		return []statusbar.Shortcut{
			{"View Record", "enter"},
			{"Go Back", "esc"},
		}
	}
}

func (m *Model) Title() string {
	return "Topics / " + m.readDetails.TopicName + " / Records"
}

func New(
	reader kadmin.RecordReader,
	readDetails kadmin.ReadDetails,
	topic *kadmin.ListedTopic,
	origin nav.Origin,
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
	m.origin = origin

	ctx, cancelFunc := context.WithCancel(context.Background())
	m.cancelConsumption = cancelFunc

	m.border = border.New(
		border.WithInnerPaddingTop(),
		border.WithTitleFn(func() string {
			return border.KeyValueTitle("Records", fmt.Sprintf(" %d", len(m.rows)), true)
		}))
	return m, func() tea.Msg {
		return m.reader.ReadRecords(ctx, readDetails)
	}
}
