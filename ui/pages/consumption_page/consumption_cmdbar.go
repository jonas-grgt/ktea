package consumption_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
)

type ConsumptionCmdBar struct {
	notifierWidget cmdbar.CmdBar
	active         cmdbar.CmdBar
}

func (c *ConsumptionCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if c.active != nil {
		return renderer.Render(c.active.View(ktx, renderer))
	}
	return renderer.Render("")
}

func (c *ConsumptionCmdBar) Update(msg tea.Msg) tea.Cmd {
	// when notifier is active it is receiving priority to handle messages
	// until a message comes in that deactivates the notifier
	if c.active == c.notifierWidget {
		c.active = c.notifierWidget
		active, _, cmd := c.active.Update(msg)
		if !active {
			c.active = nil
		}
		return cmd
	}

	switch msg := msg.(type) {
	case kadmin.ReadingStartedMsg:
		c.active = c.notifierWidget
		_, _, cmd := c.active.Update(msg)
		return cmd
	}

	return nil
}

func (c *ConsumptionCmdBar) Shortcuts() []statusbar.Shortcut {
	if c.active == nil {
		return nil
	}
	return c.active.Shortcuts()
}

func NewConsumptionCmdbar() *ConsumptionCmdBar {
	readingStartedNotifier := func(msg kadmin.ReadingStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Consuming")
	}
	c := func(msg nav.LoadCachedConsumptionPageMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Consuming")
	}
	consumptionEndedNotifier := func(msg ConsumptionEndedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	emptyTopicMsgHandler := func(_ EmptyTopicMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	notifierCmdBar := cmdbar.NewNotifierCmdBar("consumption-bar")
	cmdbar.WithMsgHandler(notifierCmdBar, readingStartedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, consumptionEndedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, emptyTopicMsgHandler)
	cmdbar.WithMsgHandler(notifierCmdBar, c)
	return &ConsumptionCmdBar{
		notifierWidget: notifierCmdBar,
	}
}
