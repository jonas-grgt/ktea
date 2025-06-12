package schema_details_page

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/chips"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"slices"
	"sort"
	"strconv"
)

type Model struct {
	cmdbar       *CmdBar
	schemas      []sradmin.Schema
	vp           *viewport.Model
	subject      sradmin.Subject
	versionChips *chips.Model
	schemaLister sradmin.VersionLister
	activeSchema *sradmin.Schema
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string

	views = append(views, m.cmdbar.View(ktx, renderer))

	if m.schemas != nil {
		if m.vp == nil {
			var versions []string
			for _, schema := range m.schemas {
				versions = append(versions, strconv.Itoa(schema.Version))
			}
			m.versionChips = chips.New("Versions", versions...)
			m.versionChips.ActivateByLabel(strconv.Itoa(m.activeSchema.Version))
			vp := viewport.New(ktx.WindowWidth-3, ktx.AvailableHeight-4)
			m.vp = &vp
		}
		if m.vp != nil {
			m.vp.Height = ktx.AvailableHeight - 5
			m.vp.Width = ktx.WindowWidth - 3
			views = append(views, lipgloss.NewStyle().
				PaddingTop(1).
				PaddingLeft(1).
				Render(m.versionChips.View(ktx, renderer)))
			views = append(views, lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().
					PaddingTop(0).
					PaddingLeft(1).
					Render("ID      : "),
				lipgloss.NewStyle().
					Bold(true).
					Render(m.activeSchema.Id)))

			m.vp.SetContent(ui.PrettyPrintJson(m.activeSchema.Value))
			views = append(views, renderer.RenderWithStyle(m.vp.View(), styles.TextViewPort))
		}
	}

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if m.versionChips != nil {
		cmd := m.versionChips.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return ui.PublishMsg(nav.LoadSubjectsPageMsg{})
		case "enter":
			version, _ := strconv.Atoi(m.versionChips.SelectedLabel())
			m.activeSchema = nil
			for _, schema := range m.schemas {
				if schema.Version == version {
					m.activeSchema = &schema
				}
			}
			if m.activeSchema == nil {
				panic("No schema found that matches " + m.versionChips.SelectedLabel())
			}
		}
	case sradmin.SchemasListed:
		m.schemas = msg.Schemas
		sort.Slice(m.schemas, func(i int, j int) bool {
			return m.schemas[i].Version < m.schemas[j].Version
		})
		m.activeSchema = m.latestSchema()
	case sradmin.SchemaListingStarted:
		cmds = append(cmds, msg.AwaitCompletion)
	}

	msg, cmd := m.cmdbar.Update(msg)
	cmds = append(cmds, cmd)

	if m.vp != nil {
		vp, cmd := m.vp.Update(msg)
		m.vp = &vp
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{
			Name:       "Prev Version",
			Keybinding: "h/←",
		},
		{
			Name:       "Next Version",
			Keybinding: "l/→",
		},
		{
			Name:       "Select Version",
			Keybinding: "enter",
		},
		{
			Name:       "Delete Version",
			Keybinding: "F2",
		},
		{
			Name:       "Prev Version",
			Keybinding: "h/←",
		},
		{
			Name:       "Go Back",
			Keybinding: "esc",
		},
	}
}

func (m *Model) Title() string {
	schema := m.latestSchema()
	if schema != nil {
		// wait until schemas have been loaded
		return "Subjects / " + m.subject.Name + " / Versions / " + strconv.Itoa(schema.Version)
	}
	return ""
}

func (m *Model) latestSchema() *sradmin.Schema {
	if m.schemas == nil {
		return nil
	}
	var latest = slices.MaxFunc(m.schemas, func(a sradmin.Schema, b sradmin.Schema) int {
		if a.Version >= b.Version {
			return a.Version
		}
		return b.Version
	})
	return &latest
}

func New(
	schemaLister sradmin.VersionLister,
	subject sradmin.Subject,
) (*Model, tea.Cmd) {
	model := &Model{
		cmdbar:       NewCmdBar(),
		subject:      subject,
		schemaLister: schemaLister,
	}
	return model, func() tea.Msg {
		return schemaLister.ListVersions(subject.Name, subject.Versions)
	}
}
