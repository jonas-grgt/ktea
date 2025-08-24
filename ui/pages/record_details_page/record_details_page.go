package record_details_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/clipper"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"sort"
	"strconv"
	"strings"
	"time"
)

type focus bool
type state bool

const (
	mainViewFocus    focus = true
	headersViewFocus focus = false
	recordView       state = true
	schemaView       state = false
)

type Model struct {
	notifierCmdbar *cmdbar.NotifierCmdBar
	record         *kadmin.ConsumerRecord
	recordVp       *viewport.Model
	headerValueVp  *viewport.Model
	topicName      string
	headerKeyTable *table.Model
	headerRows     []table.Row
	focus          focus
	state          state
	payload        string
	err            error
	metaInfo       string
	clipWriter     clipper.Writer
	config         *config.Config
	schemaVp       *viewport.Model
	border         *border.Model
}

type PayloadCopiedMsg struct {
}

type SchemaCopiedMsg struct {
}

type HeaderValueCopiedMsg struct {
}

type CopyErrorMsg struct {
	Err error
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {

	notifierCmdbarView := m.notifierCmdbar.View(ktx, renderer)

	width := int(float64(ktx.WindowWidth) * 0.70)
	height := ktx.AvailableHeight - 2

	mainView := m.mainView(width, height)
	sidebarView := m.sidebarView(ktx, width, height)

	return ui.JoinVertical(
		lipgloss.Top,
		notifierCmdbarView,
		lipgloss.NewStyle().Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.border.View(mainView),
			sidebarView)))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.recordVp == nil && m.err == nil {
		return nil
	}

	var cmds []tea.Cmd

	_, _, cmd := m.notifierCmdbar.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return ui.PublishMsg(nav.LoadCachedConsumptionPageMsg{})
		case "h", "left", "right":
			if len(m.record.Headers) >= 1 {
				m.focus = !m.focus
				m.border.Focused = m.focus == mainViewFocus
			}
		case "c":
			cmds = m.handleCopy(cmds)
		case "tab":
			if m.record.Payload.Schema != "" && m.focus == mainViewFocus {
				m.state = !m.state
				m.border.NextTab()
			}
		default:
			cmds = m.updatedFocussedArea(msg, cmds)
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) mainView(width int, height int) string {
	var mainView string
	if m.state == recordView {
		mainView = m.recordView(width, height)
	} else {
		mainView = m.schemaView(width, height)
	}

	return mainView
}

func (m *Model) schemaView(width int, height int) string {
	if m.schemaVp == nil {
		schemaVp := viewport.New(width, height)
		m.schemaVp = &schemaVp
		if m.err == nil {
			m.schemaVp.SetContent(lipgloss.NewStyle().
				Padding(0, 1).
				Render(ui.PrettyPrintJson(m.record.Payload.Schema)))
		}
	} else {
		m.schemaVp.Height = height
		m.schemaVp.Width = width
	}
	return m.schemaVp.View()
}

func (m *Model) sidebarView(ktx *kontext.ProgramKtx, payloadWidth int, height int) string {
	headersTableStyle := m.headerStyle()
	sideBarWidth := ktx.WindowWidth - (payloadWidth + 7)

	var headerSideBar string
	if len(m.record.Headers) == 0 {
		headerSideBar = ui.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().Padding(1).Render(m.metaInfo),
			lipgloss.JoinVertical(lipgloss.Center, lipgloss.NewStyle().Padding(1).Render("No headers present")),
		)
	} else {
		headerValueTableHeight := len(m.record.Headers) + 4

		headerValueVp := viewport.New(sideBarWidth, height-headerValueTableHeight-4)
		m.headerValueVp = &headerValueVp
		m.headerKeyTable.SetColumns([]table.Column{
			{"Header Key", sideBarWidth},
		})
		m.headerKeyTable.SetHeight(headerValueTableHeight)
		m.headerKeyTable.SetRows(m.headerRows)

		headerValueLine := strings.Builder{}
		for i := 0; i < sideBarWidth; i++ {
			headerValueLine.WriteString("─")
		}

		headerValue := m.selectedHeaderValue()
		m.headerValueVp.SetContent("Header Value\n" + headerValueLine.String() + "\n" + headerValue)

		headerSideBar = ui.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().Padding(1).Render(m.metaInfo),
			headersTableStyle.Render(lipgloss.JoinVertical(lipgloss.Top, m.headerKeyTable.View(), m.headerValueVp.View())),
		)
	}
	return headerSideBar
}

func (m *Model) selectedHeaderValue() string {
	selectedRow := m.headerKeyTable.SelectedRow()
	if selectedRow == nil {
		if len(m.record.Headers) > 0 {
			return m.record.Headers[0].Value.String()
		}
	} else {
		return m.record.Headers[m.headerKeyTable.Cursor()].Value.String()
	}
	return ""
}

func (m *Model) recordView(payloadWidth int, height int) string {
	if m.recordVp == nil {
		recordVp := viewport.New(payloadWidth, height)
		m.recordVp = &recordVp
		if m.err == nil {
			m.recordVp.SetContent(lipgloss.NewStyle().
				Padding(0, 1).
				Render(m.payload))
		} else {
			m.recordVp.SetContent(lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).
				Width(payloadWidth).
				Height(height).
				Render(lipgloss.NewStyle().
					Bold(true).
					Padding(1).
					Foreground(lipgloss.Color(styles.ColorGrey)).
					Render("Unable to render payload")))
		}
	} else {
		m.recordVp.Height = height
		m.recordVp.Width = payloadWidth
	}
	return m.recordVp.View()
}

func (m *Model) headerStyle() lipgloss.Style {
	var headersTableStyle lipgloss.Style
	if m.focus == mainViewFocus {
		headersTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color(styles.ColorBlurBorder))
	} else {
		headersTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color(styles.ColorFocusBorder))
	}
	return headersTableStyle
}

func (m *Model) handleCopy(cmds []tea.Cmd) []tea.Cmd {
	if m.focus == mainViewFocus {
		var copiedValue string
		if m.state == schemaView {
			copiedValue = ansi.Strip(m.record.Payload.Schema)
		} else {
			copiedValue = ansi.Strip(m.payload)
		}

		err := m.clipWriter.Write(copiedValue)

		if err != nil {
			cmds = append(cmds, ui.PublishMsg(CopyErrorMsg{Err: err}))
		} else if m.state == recordView {
			cmds = append(cmds, ui.PublishMsg(PayloadCopiedMsg{}))
		} else {
			cmds = append(cmds, ui.PublishMsg(SchemaCopiedMsg{}))
		}
	} else {
		err := m.clipWriter.Write(m.selectedHeaderValue())
		if err != nil {
			cmds = append(cmds, ui.PublishMsg(CopyErrorMsg{Err: err}))
		} else {
			cmds = append(cmds, ui.PublishMsg(HeaderValueCopiedMsg{}))
		}
	}
	return cmds
}

func (m *Model) updatedFocussedArea(msg tea.Msg, cmds []tea.Cmd) []tea.Cmd {
	// only update component if no error is present
	if m.err != nil {
		return cmds
	}

	if m.focus == mainViewFocus {
		if m.state == recordView {
			vp, cmd := m.recordVp.Update(msg)
			cmds = append(cmds, cmd)
			m.recordVp = &vp
		} else {
			vp, cmd := m.schemaVp.Update(msg)
			cmds = append(cmds, cmd)
			m.schemaVp = &vp
		}
	} else {
		t, cmd := m.headerKeyTable.Update(msg)
		cmds = append(cmds, cmd)
		m.headerKeyTable = &t
	}
	return cmds
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	whatToCopy := "Header Value"
	if m.focus == mainViewFocus {
		if m.state == schemaView {
			whatToCopy = "Schema"
		} else {
			whatToCopy = "Record"
		}
	}
	if m.err == nil {
		shortcuts := []statusbar.Shortcut{
			{"Toggle Headers/Content", "h/left/right"},
			{"Go Back", "esc"},
			{"Copy " + whatToCopy, "c"},
		}

		if m.config.ActiveCluster().HasSchemaRegistry() && m.focus == mainViewFocus {
			shortcuts = append(shortcuts, statusbar.Shortcut{
				Name:       "Toggle Record/Schema",
				Keybinding: "<tab>",
			})
		}

		return shortcuts
	} else {
		return []statusbar.Shortcut{
			{"Go Back", "esc"},
		}
	}
}

func (m *Model) Title() string {
	return "Topics / " + m.topicName + " / Records / " + strconv.FormatInt(m.record.Offset, 10)
}

func New(
	record *kadmin.ConsumerRecord,
	topicName string,
	clipWriter clipper.Writer,
	ktx *kontext.ProgramKtx,
) *Model {
	headersTable := ktable.NewDefaultTable()

	var headerRows []table.Row
	sort.SliceStable(record.Headers, func(i, j int) bool {
		return record.Headers[i].Key < record.Headers[j].Key
	})
	for _, header := range record.Headers {
		headerRows = append(headerRows, table.Row{header.Key})
	}

	notifierCmdBar := cmdbar.NewNotifierCmdBar("record-details-page")

	var (
		payload string
		err     error
	)
	if record.Err == nil {
		payload = ui.PrettyPrintJson(record.Payload.Value)
	} else {
		err = record.Err
		notifierCmdBar.Notifier.ShowError(record.Err)
	}

	key := record.Key
	if key == "" {
		key = "<null>"
	}

	metaInfo := fmt.Sprintf("key: %s\ntimestamp: %s", key, record.Timestamp.Format(time.UnixDate))

	cmdbar.WithMsgHandler(notifierCmdBar, func(msg PayloadCopiedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Payload copied")
		return true, m.AutoHideCmd("record-details-page")
	})
	cmdbar.WithMsgHandler(notifierCmdBar, func(msg SchemaCopiedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Schema copied")
		return true, m.AutoHideCmd("record-details-page")
	})
	cmdbar.WithMsgHandler(notifierCmdBar, func(msg HeaderValueCopiedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Header Value copied")
		return true, m.AutoHideCmd("record-details-page")
	})
	cmdbar.WithMsgHandler(notifierCmdBar, func(msg CopyErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Copy failed", msg.Err)
		return true, m.AutoHideCmd("record-details-page")
	})

	var tabs []border.Tab
	if record.Payload.Schema != "" {
		tabs = []border.Tab{
			{Title: "Record", TabLabel: "record"},
			{Title: "Schema", TabLabel: "record"},
		}
	}
	b := border.New(
		border.WithTabs(tabs...),
		border.WithTitle("AVRO Record"))

	return &Model{
		record:         record,
		topicName:      topicName,
		headerKeyTable: &headersTable,
		focus:          mainViewFocus,
		headerRows:     headerRows,
		payload:        payload,
		err:            err,
		metaInfo:       metaInfo,
		clipWriter:     clipWriter,
		notifierCmdbar: notifierCmdBar,
		config:         ktx.Config,
		state:          recordView,
		border:         b,
	}
}
