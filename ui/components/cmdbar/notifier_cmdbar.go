package cmdbar

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"reflect"
)

// NotificationHandler handles a specific notification and
// returns if the Cmdbar should be considered active or not and an optional tea.Cmd to be executed
type NotificationHandler[T any] func(T, *notifier.Model) (bool, tea.Cmd)

type NotifierCmdBar struct {
	active            bool
	Notifier          *notifier.Model
	msgByNotification map[reflect.Type]NotificationHandler[any]
	tag               string
}

func (n *NotifierCmdBar) IsFocussed() bool {
	return n.active && n.Notifier.HasPriority()
}

func (n *NotifierCmdBar) Shortcuts() []statusbar.Shortcut {
	return nil
}

func (n *NotifierCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	view := n.Notifier.View(ktx, renderer)
	// when empty no border style should be applied
	if view == "" {
		return view
	}
	// subtract padding (because of the rounded border of the cmdbar) and actual height
	ktx.AvailableHeight -= BorderedPadding + lipgloss.Height(view)
	return styles.CmdBarWithWidth(ktx.WindowWidth - BorderedPadding).Render(view)
}

func (n *NotifierCmdBar) Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		cmd := n.Notifier.Update(msg)
		return n.active, nil, cmd
	case notifier.HideNotificationMsg:
		if n.tag == msg.Tag {
			cmd := n.Notifier.Update(msg)
			return false, nil, cmd
		}
		return n.active, msg, nil
	}

	msgType := reflect.TypeOf(msg)
	if notification, ok := n.msgByNotification[msgType]; ok {
		active, cmd := notification(msg, n.Notifier)
		n.active = active
		return n.active, nil, cmd
	}
	return n.active, msg, nil
}

// TODO rename
func WithMsgHandler[T any](bar *NotifierCmdBar, notification NotificationHandler[T]) *NotifierCmdBar {
	msgType := reflect.TypeOf((*T)(nil)).Elem()
	bar.msgByNotification[msgType] = WrapNotification(notification)
	return bar
}

func WrapNotification[T any](n NotificationHandler[T]) NotificationHandler[any] {
	return func(msg any, m *notifier.Model) (bool, tea.Cmd) {
		typedMsg, ok := msg.(T)
		if !ok {
			return false, nil
		}
		return n(typedMsg, m)
	}
}

func NewNotifierCmdBar(tag string) *NotifierCmdBar {
	return &NotifierCmdBar{
		tag:               tag,
		msgByNotification: make(map[reflect.Type]NotificationHandler[any]),
		Notifier:          notifier.New(),
	}
}
