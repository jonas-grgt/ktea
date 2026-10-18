package schema_details_page

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/clipper"
	"ktea/ui/components/chips"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"slices"
	"sort"
	"strconv"
	"time"
)

type Model struct {
	cmdbar                  *CmdBar
	schemas                 []sradmin.Schema
	vp                      *viewport.Model
	subject                 sradmin.Subject
	versionChips            *chips.Model
	schemaLister            sradmin.VersionLister
	activeSchema            *sradmin.Schema
	atLeastOneSchemaDeleted bool
	updatedSchemas          []sradmin.Schema
	clipWriter              clipper.Writer
}

type SchemaCopiedMsg struct {
}

type CopyErrorMsg struct {
	Err error
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string

	views = append(views, m.cmdbar.View(ktx, renderer))

	if m.schemas != nil {
		if m.vp == nil {
			vp := viewport.New(viewport.WithWidth(ktx.WindowWidth-3), viewport.WithHeight(ktx.AvailableHeight-4))
			m.vp = &vp
			m.createVersionChipsView()
		}
		if m.vp != nil {
			if len(m.updatedSchemas) != 0 && len(m.schemas) != len(m.updatedSchemas) {
				m.schemas = m.updatedSchemas
				m.createVersionChipsView()
			}
			m.vp.SetHeight(ktx.AvailableHeight - 5)
			m.vp.SetWidth(ktx.WindowWidth - 3)
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
	} else {
		m.versionChips = chips.New("Versions")
	}

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) createVersionChipsView() {
	var versions []string
	for _, schema := range m.schemas {
		versions = append(versions, strconv.Itoa(schema.Version))
	}
	m.versionChips = chips.New("Versions", versions...)
	m.versionChips.ActivateByLabel(strconv.Itoa(m.activeSchema.Version))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if m.cmdbar.active == nil {
		cmd := m.versionChips.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.cmdbar.active != nil {
				pmsg, _ := m.cmdbar.Update(msg, m.versionChips.SelectedLabel())
				if pmsg == nil {
					return nil
				}
			}
			return ui.PublishMsg(nav.LoadSubjectsPageMsg{Refresh: m.atLeastOneSchemaDeleted})
		case "enter":
			if m.cmdbar.active == nil {
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
		case "c":
			err := m.clipWriter.Write(m.activeSchema.Value)
			if err != nil {
				cmds = append(cmds, ui.PublishMsg(CopyErrorMsg{Err: err}))
			} else {
				cmds = append(cmds, ui.PublishMsg(SchemaCopiedMsg{}))
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
	case sradmin.SchemaDeletionStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.SchemaDeletedMsg:
		m.atLeastOneSchemaDeleted = true
		for i, schema := range m.schemas {
			if schema.Version == msg.Version {
				m.updatedSchemas = slices.Delete(m.schemas, i, i+1)
				break
			}
		}
		if len(m.updatedSchemas) == 0 {
			cmds = append(cmds, func() tea.Msg {
				time.Sleep(5 * time.Second)
				return nav.LoadSubjectsPageMsg{Refresh: true}
			})
		}
	}

	msg, cmd := m.cmdbar.Update(msg, m.versionChips.SelectedLabel())
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
			Name:       "Copy Schema",
			Keybinding: "c",
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
	schemaDeleter sradmin.SchemaDeleter,
	subject sradmin.Subject,
	clipWriter clipper.Writer,
) (*Model, tea.Cmd) {

	model := &Model{
		subject:      subject,
		schemaLister: schemaLister,
		clipWriter:   clipWriter,
	}

	deleteFunc := func(version int) tea.Cmd {
		return func() tea.Msg {
			return schemaDeleter.DeleteSchema(subject.Name, version)
		}
	}

	schemaDeletedMsg := func(msg sradmin.SchemaDeletedMsg, m *notifier.Model) (bool, tea.Cmd) {
		if len(model.updatedSchemas) == 0 {
			m.ShowSuccessMsg(fmt.Sprintf("Deleted last schema with version %d of subject, returning to subjects list.", msg.Version))
		} else {
			m.ShowSuccessMsg(fmt.Sprintf("Schema with version %d has been deleted", msg.Version))
		}
		return true, m.AutoHideCmd("schema-details-cmd-bar")
	}

	model.cmdbar = NewCmdBar(deleteFunc, schemaDeletedMsg)

	return model, func() tea.Msg {
		return schemaLister.ListVersions(subject.Name, subject.Versions)
	}
}
