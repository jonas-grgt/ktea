package create_schema_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
)

type state int

const (
	entering state = 0
	creating state = 1
)

type values struct {
	subject string
	schema  string
}

type Model struct {
	values
	form                    *huh.Form
	schemaCreator           sradmin.SchemaCreator
	cmdBar                  *cmdbar.NotifierCmdBar
	state                   state
	ktx                     *kontext.ProgramKtx
	schemaInput             *huh.Text
	createdAtLeastOneSchema bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdbarView := m.cmdBar.View(ktx, renderer)

	if m.form == nil {
		m.form = newForm(m)
	}

	return ui.JoinVertical(
		lipgloss.Top,
		cmdbarView,
		renderer.RenderWithStyle(m.form.View(), styles.Form),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if m.form != nil {
		form, cmd := m.form.Update(msg)
		cmds = append(cmds, cmd)

		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		if m.form.State == huh.StateCompleted && m.state == entering {
			m.state = creating
			return func() tea.Msg {
				return m.schemaCreator.CreateSchema(sradmin.SubjectCreationDetails{
					Subject: m.subject,
					Schema:  m.schema,
				})
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		{
			switch msg.String() {
			case "esc":
				if m.state == creating {
					return nil
				}
				cmds = append(cmds, ui.PublishMsg(nav.LoadSubjectsPageMsg{
					Refresh: m.createdAtLeastOneSchema,
				}))
			default:
				m.cmdBar.Notifier.Idle()
			}
		}
	case sradmin.SchemaCreatedMsg:
		m.createdAtLeastOneSchema = true
		m.state = entering
		m.form = nil
	case sradmin.SchemaCreationErrMsg:
		m.state = entering
		m.form = nil
	case sradmin.SchemaCreationStartedMsg:
		m.state = creating
		cmds = append(cmds, msg.AwaitCompletion)
	}

	_, _, cmd := m.cmdBar.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Subjects / Register New"
}

func newForm(model *Model) *huh.Form {
	model.subject = ""
	model.schema = ""
	schemaInput := huh.NewText().
		Value(&model.values.schema).
		Title("Schema").
		Validate(func(v string) error {
			if v == "" {
				return fmt.Errorf("schema cannot be empty")
			}
			return nil
		}).
		WithHeight(model.ktx.AvailableHeight - 9).(*huh.Text)
	model.schemaInput = schemaInput
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Value(&model.values.subject).
			Title("Subject").
			Validate(func(v string) error {
				if v == "" {
					return fmt.Errorf("subject cannot be empty")
				}
				return nil
			}),
		schemaInput,
	))
	form.Init()
	form.QuitAfterSubmit = false
	return form
}

func New(schemaCreator sradmin.SchemaCreator, ktx *kontext.ProgramKtx) (*Model, tea.Cmd) {
	model := &Model{}
	model.ktx = ktx
	model.schemaCreator = schemaCreator
	model.state = entering
	notifierCmdBar := cmdbar.NewNotifierCmdBar("create-schema-page")
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg sradmin.SchemaCreationStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Creating Schema")
		return true, cmd
	})
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg sradmin.SchemaCreatedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Schema created")
		return true, nil
	})
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg notifier.HideNotificationMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return true, nil
	})
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg sradmin.SchemaCreationErrMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Schema creation failed", msg.Err)
		return true, nil
	})
	model.cmdBar = notifierCmdBar
	return model, nil
}
