package sr_tab

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/clipper"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/create_schema_page"
	"ktea/ui/pages/schema_details_page"
	"ktea/ui/pages/subjects_page"
)

type Model struct {
	active            pages.Page
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
	if msg, ok := msg.(sradmin.SubjectsListedMsg); ok {
		return m.subjectsPage.Update(msg)
	}

	cmd := m.active.Update(msg)

	// in case the active page might have changed, update the statusbar provider
	m.statusbar.SetProvider(m.active)

	return cmd
}

func (m *Model) ToSubjectsPage(refresh bool) tea.Cmd {
	var cmd tea.Cmd
	if refresh && m.active != m.subjectsPage {
		m.subjectsPage, cmd = subjects_page.New(m.srClient, m)
	}
	m.active = m.subjectsPage
	return cmd
}

func (m *Model) ToCreateSubjectPage() tea.Cmd {
	page, cmd := create_schema_page.New(m.srClient, m.ktx, m)
	m.active = page
	return cmd
}

func (m *Model) ToSchemaDetailsPage(subject sradmin.Subject) tea.Cmd {
	var cmd tea.Cmd
	m.schemaDetailsPage, cmd = schema_details_page.New(m.srClient, m.srClient, subject, clipper.New(), m)
	m.active = m.schemaDetailsPage
	return cmd
}

func New(srClient sradmin.Client, ktx *kontext.ProgramKtx, stsBar *statusbar.Model) (*Model, tea.Cmd) {
	model := Model{}
	subjectsPage, cmd := subjects_page.New(srClient, &model)
	model.active = subjectsPage
	model.subjectsPage = subjectsPage
	model.statusbar = stsBar
	model.srClient = srClient
	model.ktx = ktx
	return &model, cmd
}
