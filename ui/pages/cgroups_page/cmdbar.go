package cgroups_page

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/notifier"
)

type CmdBar struct {
	notifier *notifier.Model
}

func (c *CmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return c.notifier.View(ktx, renderer)
}

func (c *CmdBar) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.(type) {
	case spinner.TickMsg:
		cmd = c.notifier.Update(msg)
	case kadmin.ConsumerGroupListingStartedMsg:
		cmd = c.notifier.SpinWithLoadingMsg("Loading Consumer Groups")
	case kadmin.ConsumerGroupsListedMsg:
		c.notifier.Idle()
	}

	return cmd
}

func NewCmdBar() *CmdBar {
	return &CmdBar{
		notifier: notifier.New(),
	}
}
