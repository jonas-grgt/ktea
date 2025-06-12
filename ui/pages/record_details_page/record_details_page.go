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

const (
	payloadFocus focus = true
	headersFocus focus = false
)

type Model struct {
	notifierCmdbar *cmdbar.NotifierCmdBar
	record         *kadmin.ConsumerRecord
	payloadVp      *viewport.Model
	headerValueVp  *viewport.Model
	topicName      string
	headerKeyTable *table.Model
	headerRows     []table.Row
	focus          focus
	payload        string
	err            error
	metaInfo       string
	clipWriter     clipper.Writer
	config         *config.Config
}

type PayloadCopiedMsg struct {
}

type HeaderValueCopiedMsg struct {
}

type CopyErrorMsg struct {
	Err error
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	contentStyle, headersTableStyle := m.determineStyles()

	notifierCmdbarView := m.notifierCmdbar.View(ktx, renderer)

	payloadWidth := int(float64(ktx.WindowWidth) * 0.70)
	height := ktx.AvailableHeight - 2

	m.createPayloadViewPort(payloadWidth, height)
	contentView := renderer.RenderWithStyle(m.payloadVp.View(), contentStyle)

	headerSideBar := m.createSidebar(ktx, payloadWidth, height, headersTableStyle)

	return ui.JoinVertical(
		lipgloss.Top,
		notifierCmdbarView,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			contentView,
			headerSideBar,
		))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.payloadVp == nil && m.err == nil {
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
		case "ctrl+h", "left", "right":
			m.focus = !m.focus
		case "c":
			cmds = m.handleCopy(cmds)
		case "ctrl+s":
			m.focus = !m.focus
		default:
			cmds = m.updatedFocussedArea(msg, cmds)
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) createSidebar(ktx *kontext.ProgramKtx, payloadWidth int, height int, headersTableStyle lipgloss.Style) string {
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

func (m *Model) createPayloadViewPort(payloadWidth int, height int) {
	if m.payloadVp == nil {
		payloadVp := viewport.New(payloadWidth, height)
		m.payloadVp = &payloadVp
		if m.err == nil {
			m.payloadVp.SetContent(m.payload)
		} else {
			m.payloadVp.SetContent(lipgloss.NewStyle().
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
		m.payloadVp.Height = height
		m.payloadVp.Width = payloadWidth
	}
}

func (m *Model) determineStyles() (lipgloss.Style, lipgloss.Style) {
	var contentStyle lipgloss.Style
	var headersTableStyle lipgloss.Style
	if m.focus == payloadFocus {
		contentStyle = lipgloss.NewStyle().
			Inherit(styles.TextViewPort).
			BorderForeground(lipgloss.Color(styles.ColorFocusBorder))
		headersTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color(styles.ColorBlurBorder))
	} else {
		contentStyle = lipgloss.NewStyle().
			Inherit(styles.TextViewPort).
			BorderForeground(lipgloss.Color(styles.ColorBlurBorder))
		headersTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color(styles.ColorFocusBorder))
	}
	return contentStyle, headersTableStyle
}

func (m *Model) handleCopy(cmds []tea.Cmd) []tea.Cmd {
	if m.focus == payloadFocus {
		err := m.clipWriter.Write(ansi.Strip(m.payload))
		if err != nil {
			cmds = append(cmds, ui.PublishMsg(CopyErrorMsg{Err: err}))
		} else {
			cmds = append(cmds, ui.PublishMsg(PayloadCopiedMsg{}))
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

	if m.focus == payloadFocus {
		vp, cmd := m.payloadVp.Update(msg)
		cmds = append(cmds, cmd)
		m.payloadVp = &vp
	} else {
		t, cmd := m.headerKeyTable.Update(msg)
		cmds = append(cmds, cmd)
		m.headerKeyTable = &t
	}
	return cmds
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	whatToCopy := "Header Value"
	if m.focus == payloadFocus {
		whatToCopy = "Content"
	}
	if m.err == nil {
		shortcuts := []statusbar.Shortcut{
			{"Toggle Headers/Content", "C-h/Arrows"},
			{"Go Back", "esc"},
			{"Copy " + whatToCopy, "c"},
		}

		if m.config.ActiveCluster().HasSchemaRegistry() {
			shortcuts = append(shortcuts, statusbar.Shortcut{
				Name:       "View Schema",
				Keybinding: "C-s",
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
	cmdbar.WithMsgHandler(notifierCmdBar, func(msg HeaderValueCopiedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Header Value copied")
		return true, m.AutoHideCmd("record-details-page")
	})
	cmdbar.WithMsgHandler(notifierCmdBar, func(msg CopyErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Copy failed", msg.Err)
		return true, m.AutoHideCmd("record-details-page")
	})

	return &Model{
		record:         record,
		topicName:      topicName,
		headerKeyTable: &headersTable,
		focus:          payloadFocus,
		headerRows:     headerRows,
		payload:        payload,
		err:            err,
		metaInfo:       metaInfo,
		clipWriter:     clipWriter,
		notifierCmdbar: notifierCmdBar,
		config:         ktx.Config,
	}
}
