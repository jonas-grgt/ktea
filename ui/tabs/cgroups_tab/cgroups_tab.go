package cgroups_tab

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/cgroups_page"
	"ktea/ui/pages/cgroups_topics_page"
)

type Model struct {
	active       pages.Page
	statusbar    *statusbar.Model
	offsetLister kadmin.OffsetLister
	cgroupsPage  *cgroups_page.Model
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return ui.JoinVertical(
		lipgloss.Top,
		m.statusbar.View(ktx, renderer),
		m.active.View(ktx, renderer),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	if msg, ok := msg.(kadmin.ConsumerGroupListingStartedMsg); ok {
		cmds = append(cmds, msg.AwaitCompletion)
	}

	cmd := m.active.Update(msg)

	// in case the active page might have changed, update the statusbar provider
	m.statusbar.SetProvider(m.active)

	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (m *Model) ToCGroupsPage() tea.Cmd {
	m.active = m.cgroupsPage
	return nil
}

func (m *Model) ToCGroupTopicsPage(group string) tea.Cmd {
	page, cmd := cgroups_topics_page.New(m.offsetLister, group, m)
	m.active = page
	return cmd
}

func New(
	cgroupLister kadmin.CGroupLister,
	cgroupDeleter kadmin.CGroupDeleter,
	consumerGroupOffsetLister kadmin.OffsetLister,
	statusbar *statusbar.Model,
) (*Model, tea.Cmd) {
	m := &Model{}
	cgroupsPage, cmd := cgroups_page.New(cgroupLister, cgroupDeleter, m)

	m.offsetLister = consumerGroupOffsetLister
	m.cgroupsPage = cgroupsPage
	m.active = cgroupsPage
	m.statusbar = statusbar

	return m, cmd
}
