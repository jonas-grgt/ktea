package consumption_page

import (
	"context"
	"fmt"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	table              *table.Model
	cmdBar             *ConsumptionCmdBar
	consumerRecordChan chan kadmin.ConsumerRecord
	cancelConsumption  context.CancelFunc
	errChan            chan error
	reader             kadmin.RecordReader
	rows               []table.Row
	records            []kadmin.ConsumerRecord
	readDetails        kadmin.ReadDetails
	consuming          bool
	noRecordsAvailable bool
	topic              *kadmin.ListedTopic
}

type ConsumerRecordReceived struct {
	Record kadmin.ConsumerRecord
}

type ConsumptionEndedMsg struct{}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	views = append(views, m.cmdBar.View(ktx, renderer))

	if m.noRecordsAvailable {
		views = append(views, styles.CenterText(ktx.WindowWidth, ktx.AvailableHeight).
			Render("👀 Empty topic"))
	} else if len(m.rows) > 0 {
		m.table.SetColumns([]table.Column{
			{Title: "Key", Width: int(float64(ktx.WindowWidth-9) * 0.5)},
			{Title: "Timestamp", Width: int(float64(ktx.WindowWidth-9) * 0.30)},
			{Title: "Partition", Width: int(float64(ktx.WindowWidth-9) * 0.10)},
			{Title: "Offset", Width: int(float64(ktx.WindowWidth-9) * 0.10)},
		})
		m.table.SetHeight(ktx.AvailableHeight - 2)
		m.table.SetRows(m.rows)

		embeddedText := map[styles.BorderPosition]styles.EmbeddedTextFunc{
			styles.TopMiddleBorder: func(active bool) string {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color(styles.ColorPink)).
					Bold(true).
					Render(fmt.Sprintf("Records: %d", len(m.rows)))
			},
		}
		borderedView := styles.Borderize(m.table.View(), true, embeddedText)
		views = append(views, borderedView)
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
			return ui.PublishMsg(nav.LoadConsumptionFormPageMsg{ReadDetails: &m.readDetails, Topic: m.topic})
		} else if msg.String() == "f2" {
			m.cancelConsumption()
			m.consuming = false
			cmds = append(cmds, ui.PublishMsg(ConsumptionEndedMsg{}))
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
	case kadmin.ReadingStartedMsg:
		m.consuming = true
		m.consumerRecordChan = msg.ConsumerRecord
		m.errChan = msg.Err
		cmds = append(cmds, m.waitForActivity())
	case ConsumptionEndedMsg:
		m.consuming = false
		return nil
	case ConsumerRecordReceived:
		var key string
		if msg.Record.Key == "" {
			key = "<null>"
		} else {
			key = msg.Record.Key
		}
		m.records = append(m.records, msg.Record)
		m.rows = append(
			[]table.Row{
				{
					key,
					msg.Record.Timestamp.Format("2006-01-02 15:04:05"),
					strconv.FormatInt(msg.Record.Partition, 10),
					strconv.FormatInt(msg.Record.Offset, 10),
				},
			},
			m.rows...,
		)
		return m.waitForActivity()
	}

	return tea.Batch(cmds...)
}

func (m *Model) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		for {
			select {
			case record, ok := <-m.consumerRecordChan:
				if !ok {
					return ConsumptionEndedMsg{}
				}
				return ConsumerRecordReceived{Record: record}
			case err := <-m.errChan:
				return err
			}
		}
	}
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
) (nav.Page, tea.Cmd) {
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

	ctx, cancelFunc := context.WithCancel(context.Background())
	m.cancelConsumption = cancelFunc

	return m, func() tea.Msg {
		return m.reader.ReadRecords(ctx, readDetails)
	}
}
