package cgroups_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/cgroups_page"
	"ktea/ui/pages/cgroups_topics_page"
	"ktea/ui/pages/nav"
)

type Model struct {
	active        pages.Page
	statusbar     *statusbar.Model
	offsetLister  kadmin.OffsetLister
	cgroupLister  kadmin.CGroupLister
	cgroupDeleter kadmin.CGroupDeleter
	cgroupsPage   *cgroups_page.Model
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
	switch msg := msg.(type) {
	case nav.LoadCGroupTopicsPageMsg:
		cgroupsTopicsPage, cmd := cgroups_topics_page.New(m.offsetLister, msg.GroupName)
		cmds = append(cmds, cmd)
		m.active = cgroupsTopicsPage
		return tea.Batch(cmds...)
	case nav.LoadCGroupsPageMsg:
		var cmd tea.Cmd
		if m.cgroupsPage == nil {
			m.cgroupsPage, cmd = cgroups_page.New(m.cgroupLister, m.cgroupDeleter)
		}
		m.active = m.cgroupsPage
		cmds = append(cmds, cmd)
	case kadmin.ConsumerGroupListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	}

	cmd := m.active.Update(msg)

	// in case the active page might have changed, update the statusbar provider
	m.statusbar.SetProvider(m.active)

	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func New(
	cgroupLister kadmin.CGroupLister,
	cgroupDeleter kadmin.CGroupDeleter,
	consumerGroupOffsetLister kadmin.OffsetLister,
	statusbar *statusbar.Model,
) (*Model, tea.Cmd) {
	cgroupsPage, cmd := cgroups_page.New(cgroupLister, cgroupDeleter)

	m := &Model{}
	m.offsetLister = consumerGroupOffsetLister
	m.cgroupLister = cgroupLister
	m.cgroupDeleter = cgroupDeleter
	m.cgroupsPage = cgroupsPage
	m.active = cgroupsPage
	m.statusbar = statusbar

	return m, cmd
}
