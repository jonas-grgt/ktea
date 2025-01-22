package notifier

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"strings"
	"time"
)

type state int

const (
	idle     state = 0
	err      state = 1
	success  state = 2
	spinning state = 3
)

type Model struct {
	spinner    spinner.Model
	successMsg string
	msg        string
	state      state
}

type HideNotificationMsg struct{}

type NotificationHiddenMsg struct{}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.state == spinning {
		return renderer.RenderWithStyle(
			lg.JoinHorizontal(lg.Top, m.spinner.View(), m.msg),
			styles.Notifier.Spinner,
		)
	} else if m.state == success {
		return renderer.RenderWithStyle(
			wordwrap.String(m.msg, ktx.WindowWidth),
			styles.Notifier.Success,
		)
	} else if m.state == err {
		return renderer.RenderWithStyle(
			wordwrap.String(m.msg, ktx.WindowWidth),
			styles.Notifier.Error,
		)
	}
	return renderer.Render("")
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.state != spinning {
			return nil
		}
		s, cmd := m.spinner.Update(msg)
		m.spinner = s
		return cmd
	case HideNotificationMsg:
		return func() tea.Msg {
			m.Idle()
			return NotificationHiddenMsg{}
		}
	}
	return nil
}

func (m *Model) SpinWithLoadingMsg(msg string) tea.Cmd {
	m.state = spinning
	m.msg = "⏳ " + msg
	return m.spinner.Tick
}

func (m *Model) SpinWithRocketMsg(msg string) tea.Cmd {
	m.state = spinning
	m.msg = "🚀 " + msg
	return m.spinner.Tick
}

func (m *Model) ShowErrorMsg(msg string, error error) tea.Cmd {
	m.state = err
	m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(msg+": ") +
		styles.FG(styles.ColorWhite).Render(strings.TrimSuffix(error.Error(), "\n"))
	return nil
}

func (m *Model) ShowSuccessMsg(msg string) tea.Cmd {
	m.state = success
	m.msg = "🎉 " + msg
	return nil
}

func (m *Model) Idle() {
	m.state = idle
	m.msg = ""
}

func (m *Model) AutoHideCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(5 * time.Second)
		return HideNotificationMsg{}
	}
}

func (m *Model) HasPriority() bool {
	return m.state == spinning
}

func New() *Model {
	l := Model{}
	l.state = idle
	l.spinner = spinner.New()
	l.spinner.Spinner = spinner.Dot
	return &l
}
