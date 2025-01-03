package cgroups_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/cgroups_page"
	"ktea/ui/pages/cgroups_topics_page"
	"ktea/ui/pages/navigation"
)

type Model struct {
	active              navigation.Page
	list                *cgroups_page.Model
	statusbar           *statusbar.Model
	offsetLister        kadmin.OffsetLister
	consumerGroupLister kadmin.CGroupLister
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.statusbar.View(ktx, renderer),
		m.active.View(ktx, renderer),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case navigation.LoadCGroupTopicsPageMsg:
		cgroupsTopicsPage, cmd := cgroups_topics_page.New(m.offsetLister, msg.GroupName)
		cmds = append(cmds, cmd)
		m.active = cgroupsTopicsPage
		return tea.Batch(cmds...)
	case navigation.LoadCGroupsPageMsg:
		cgroupsPage, cmd := cgroups_page.New(m.consumerGroupLister)
		m.active = cgroupsPage
		return cmd
	case kadmin.ConsumerGroupListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	}

	cmd := m.active.Update(msg)

	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.active)

	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func New(
	consumerGroupLister kadmin.CGroupLister,
	consumerGroupOffsetLister kadmin.OffsetLister,
) (*Model, tea.Cmd) {
	cgroupsPage, cmd := cgroups_page.New(consumerGroupLister)

	m := &Model{}
	m.offsetLister = consumerGroupOffsetLister
	m.consumerGroupLister = consumerGroupLister
	m.list = cgroupsPage
	m.active = cgroupsPage
	m.statusbar = statusbar.New(m.active)

	return m, cmd
}
