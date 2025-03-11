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

type HideNotificationMsg struct {
	Tag string
}

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
		m.Idle()
		return func() tea.Msg {
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
	s := ": "
	if msg == "" {
		s = ""
	}
	m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(msg+s) +
		styles.FG(styles.ColorWhite).Render(strings.TrimSuffix(error.Error(), "\n"))
	return nil
}

func (m *Model) ShowError(error error) tea.Cmd {
	m.state = err
	msg := error.Error()
	split := strings.Split(msg, ":")
	if len(split) > 1 {
		m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(split[0]) + ": " +
			styles.FG(styles.ColorWhite).Render(strings.TrimSuffix(split[1], "\n"))
	} else {
		m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(msg) +
			styles.FG(styles.ColorWhite).Render(strings.TrimSuffix(error.Error(), "\n"))
	}
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

func (m *Model) AutoHideCmd(tag string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(5 * time.Second)
		return HideNotificationMsg{Tag: tag}
	}
}

func (m *Model) HasPriority() bool {
	return m.state == spinning
}

func (m *Model) IsIdle() bool {
	return m.state == idle
}

func New() *Model {
	l := Model{}
	l.state = idle
	l.spinner = spinner.New()
	l.spinner.Spinner = spinner.Dot
	return &l
}
