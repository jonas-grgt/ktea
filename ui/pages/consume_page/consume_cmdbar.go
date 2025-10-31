package consume_page

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
	sortByCBar     *cmdbar.SortByCmdBar
	searchCBar     *cmdbar.SearchCmdBar
}

func (c *ConsumptionCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if c.active != nil {
		return c.active.View(ktx, renderer)
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
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			active, _, cmd := c.searchCBar.Update(msg)
			if !active {
				c.active = nil
			} else {
				c.active = c.searchCBar
				c.sortByCBar.Active = false
			}
			return cmd
		case "f3":
			active, _, cmd := c.sortByCBar.Update(msg)
			if !active {
				c.active = nil
			} else {
				c.active = c.sortByCBar
				c.searchCBar.Hide()
			}
			return cmd
		default:
			if c.active != nil {
				active, _, cmd := c.active.Update(msg)
				if !active {
					c.active = nil
				}
				return cmd
			}
		}
	}

	switch msg := msg.(type) {
	case *kadmin.ReadingStartedMsg:
		c.active = c.notifierWidget
		_, _, cmd := c.active.Update(msg)
		return cmd
	}

	return nil
}

func (c *ConsumptionCmdBar) IsFocussed() bool {
	return c.active != nil && c.active.IsFocussed()
}

func (c *ConsumptionCmdBar) Shortcuts() []statusbar.Shortcut {
	if c.active == nil {
		return nil
	}
	return c.active.Shortcuts()
}

func (c *ConsumptionCmdBar) GetSearchTerm() string {
	return c.searchCBar.GetSearchTerm()
}

func (c *ConsumptionCmdBar) IsSorting() bool {
	return c.active == c.sortByCBar
}

func NewConsumptionCmdbar() *ConsumptionCmdBar {
	readingStartedNotifier := func(msg *kadmin.ReadingStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Consuming")
	}
	c := func(msg nav.LoadCachedConsumptionPageMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Consuming")
	}
	consumptionEndedNotifier := func(msg kadmin.ConsumptionEndedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	emptyTopicMsgHandler := func(_ kadmin.EmptyTopicMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	noRecordFoundMsgHandler := func(_ kadmin.NoRecordsFound, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	notifierCmdBar := cmdbar.NewNotifierCmdBar("consumption-bar")
	cmdbar.BindNotificationHandler(notifierCmdBar, readingStartedNotifier)
	cmdbar.BindNotificationHandler(notifierCmdBar, consumptionEndedNotifier)
	cmdbar.BindNotificationHandler(notifierCmdBar, emptyTopicMsgHandler)
	cmdbar.BindNotificationHandler(notifierCmdBar, noRecordFoundMsgHandler)
	cmdbar.BindNotificationHandler(notifierCmdBar, c)

	sortByCmdBar := cmdbar.NewSortByCmdBar(
		[]cmdbar.SortLabel{
			{
				Label:     "Key",
				Direction: cmdbar.Asc,
			},
			{
				Label:     "Timestamp",
				Direction: cmdbar.Desc,
			},
			{
				Label:     "Partition",
				Direction: cmdbar.Desc,
			},
			{
				Label:     "Offset",
				Direction: cmdbar.Desc,
			},
		},
		cmdbar.WithInitialSortColumn("Timestamp", cmdbar.Desc),
	)

	return &ConsumptionCmdBar{
		notifierWidget: notifierCmdBar,
		sortByCBar:     sortByCmdBar,
		searchCBar:     cmdbar.NewSearchCmdBar("Search by key or record value"),
	}
}
