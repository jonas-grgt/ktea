package sr_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/clipper"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/create_schema_page"
	"ktea/ui/pages/nav"
	"ktea/ui/pages/schema_details_page"
	"ktea/ui/pages/subjects_page"
)

type Model struct {
	active            nav.Page
	statusbar         *statusbar.Model
	ktx               *kontext.ProgramKtx
	compLister        sradmin.GlobalCompatibilityLister
	subjectsPage      *subjects_page.Model
	schemaDetailsPage *schema_details_page.Model
	srClient          sradmin.Client
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	statusBarView := m.statusbar.View(ktx, renderer)
	return ui.JoinVertical(
		lipgloss.Top,
		statusBarView,
		m.active.View(ktx, renderer),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case sradmin.SubjectsListedMsg:
		return m.subjectsPage.Update(msg)
	case nav.LoadCreateSubjectPageMsg:
		createPage, cmd := create_schema_page.New(m.srClient, m.ktx)
		cmds = append(cmds, cmd)
		m.active = createPage
	case nav.LoadSubjectsPageMsg:
		if m.subjectsPage == nil || msg.Refresh && m.active != m.subjectsPage {
			var cmd tea.Cmd
			m.subjectsPage, cmd = subjects_page.New(m.srClient)
			cmds = append(cmds, cmd)
		}
		m.active = m.subjectsPage
	case nav.LoadSchemaDetailsPageMsg:
		var cmd tea.Cmd
		m.schemaDetailsPage, cmd = schema_details_page.New(m.srClient, m.srClient, msg.Subject, clipper.New())
		m.active = m.schemaDetailsPage
		cmds = append(cmds, cmd)
	}

	m.statusbar = statusbar.New(m.active)

	cmds = append(cmds, m.active.Update(msg))

	return tea.Batch(cmds...)
}

func New(
	srClient sradmin.Client,
	ktx *kontext.ProgramKtx,
) (*Model, tea.Cmd) {
	subjectsPage, cmd := subjects_page.New(srClient)
	model := Model{active: subjectsPage}
	model.subjectsPage = subjectsPage
	model.statusbar = statusbar.New(subjectsPage)
	model.srClient = srClient
	model.ktx = ktx
	return &model, cmd
}
