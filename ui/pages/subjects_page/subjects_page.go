package subjects_page

import (
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	initialized     state = 0
	subjectsLoaded  state = 1
	loading         state = 2
	noSubjectsFound state = 3
	deleting        state = 4
)

type Model struct {
	table            table.Model
	rows             []table.Row
	cmdBar           *cmdbar.TableCmdsBar[sradmin.Subject]
	subjects         []sradmin.Subject
	renderedSubjects []sradmin.Subject
	tableFocussed    bool
	lister           sradmin.SubjectLister
	state            state
	// when last subject in table is deleted no subject is focussed anymore
	deletedLast bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.state == noSubjectsFound {
		return lipgloss.NewStyle().
			Width(ktx.WindowWidth - 2).
			Height(ktx.AvailableHeight - 2).
			AlignVertical(lipgloss.Center).
			AlignHorizontal(lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			Render("No Subjects Found")
	}

	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.table.SetColumns([]table.Column{
		{"Subject Name", int(float64(ktx.WindowWidth-5) * 0.9)},
		{"Version Count", int(float64(ktx.WindowWidth-5) * 0.1)},
	})
	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetRows(m.rows)

	if m.deletedLast && m.table.SelectedRow() == nil {
		m.table.GotoBottom()
		m.deletedLast = false
	}
	if m.table.SelectedRow() == nil {
		m.table.GotoTop()
	}

	tableView := styles.Borderize(m.table.View(), m.tableFocussed, nil)
	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f5":
			m.state = loading
			m.subjects = nil
			return m.lister.ListSubjects
		case "ctrl+n":
			if m.state != loading && m.state != deleting {
				return ui.PublishMsg(nav.LoadCreateSubjectPageMsg{})
			}
		case "enter":
			// only accept enter when the table is focussed
			if !m.cmdBar.IsFocussed() {
				// ignore enter when there are no schemas loaded
				if m.state == subjectsLoaded && len(m.subjects) > 0 {
					return ui.PublishMsg(
						nav.LoadSchemaDetailsPageMsg{
							Subject: *m.SelectedSubject(),
						},
					)
				}
			}
		}
	case sradmin.SubjectListingStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.SubjectsListedMsg:
		if len(msg.Subjects) > 0 {
			m.state = subjectsLoaded
			m.subjects = msg.Subjects
		} else {
			m.state = noSubjectsFound
		}
	case sradmin.SubjectDeletionStartedMsg:
		m.state = deleting
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.SubjectDeletedMsg:
		// set state back to loaded after removing the deleted subject
		m.state = subjectsLoaded
		m.removeDeletedSubjectFromModel(msg.SubjectName)
		if len(m.subjects) == 0 {
			m.state = noSubjectsFound
		}
	}

	msg, cmd := m.cmdBar.Update(msg, m.SelectedSubject())
	m.tableFocussed = !m.cmdBar.IsFocussed()
	cmds = append(cmds, cmd)

	subjects := m.filterSubjectsBySearchTerm()
	subjects = m.sortSubjects(subjects)
	m.renderedSubjects = subjects
	m.rows = m.createRows(subjects)

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = t
		cmds = append(cmds, cmd)
	}

	if m.cmdBar.HasSearchedAtLeastOneChar() {
		m.table.GotoTop()
	}

	return tea.Batch(cmds...)
}

func (m *Model) removeDeletedSubjectFromModel(subjectName string) {
	for i, subject := range m.subjects {
		if subject.Name == subjectName {
			if i == len(m.subjects)-1 {
				m.deletedLast = true
			}
			m.subjects = append(m.subjects[:i], m.subjects[i+1:]...)
		}
	}
}

func (m *Model) sortSubjects(subjects []sradmin.Subject) []sradmin.Subject {
	sort.Slice(subjects, func(i int, y int) bool {
		return subjects[i].Name < subjects[y].Name
	})
	return subjects
}

func (m *Model) createRows(subjects []sradmin.Subject) []table.Row {
	var rows []table.Row
	for _, subject := range subjects {
		rows = append(rows, table.Row{
			subject.Name,
			strconv.Itoa(len(subject.Versions)),
		})
	}
	return rows
}

func (m *Model) filterSubjectsBySearchTerm() []sradmin.Subject {
	var subjects []sradmin.Subject
	searchTerm := m.cmdBar.GetSearchTerm()
	for _, subject := range m.subjects {
		if searchTerm != "" {
			if strings.Contains(strings.ToUpper(subject.Name), strings.ToUpper(searchTerm)) {
				subjects = append(subjects, subject)
			}
		} else {
			subjects = append(subjects, subject)
		}
	}
	return subjects
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	shortcuts := m.cmdBar.Shortcuts()
	if shortcuts == nil {
		return []statusbar.Shortcut{
			{
				Name:       "Search",
				Keybinding: "/",
			},
			{
				Name:       "Delete",
				Keybinding: "F2",
			},
			{
				Name:       "Register New Schema",
				Keybinding: "C-n",
			},
			//{
			//	Name:       "Evolve Selected Schema",
			//	Keybinding: "C-e",
			//},
			{
				Name:       "Refresh",
				Keybinding: "F5",
			},
		}
	} else {
		return shortcuts
	}
}

func (m *Model) SelectedSubject() *sradmin.Subject {
	if len(m.renderedSubjects) > 0 {
		selectedRow := m.table.SelectedRow()
		if selectedRow != nil {
			return &m.renderedSubjects[m.table.Cursor()]
		}
		return nil
	}
	return nil
}

func (m *Model) Title() string {
	return "Subjects"
}

func New(lister sradmin.SubjectLister, deleter sradmin.SubjectDeleter) (*Model, tea.Cmd) {
	deleteMsgFunc := func(subject sradmin.Subject) string {
		message := subject.Name + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7571F9")).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(subject sradmin.Subject) tea.Cmd {
		return func() tea.Msg {
			return deleter.DeleteSubject(subject.Name)
		}
	}

	notifierCmdBar := cmdbar.NewNotifierCmdBar("subjects-page")

	subjectListingStartedNotifier := func(msg sradmin.SubjectListingStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Loading subjects")
		return true, cmd
	}
	subjectsListedNotifier := func(msg sradmin.SubjectsListedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	subjectDeletionStartedNotifier := func(msg sradmin.SubjectDeletionStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Deleting Subject " + msg.Subject)
		return true, cmd
	}
	subjectListingErrorMsg := func(msg sradmin.SubjectListingErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Error listing subjects", msg.Err)
		return true, nil
	}
	subjectDeletedNotifier := func(msg sradmin.SubjectDeletedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Subject deleted")
		return true, m.AutoHideCmd("subjects-page")
	}
	subjectDeletionErrorNotifier := func(msg sradmin.SubjectDeletionErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Failed to delete subject", msg.Err)
		return true, m.AutoHideCmd("subjects-page")
	}
	cmdbar.WithMsgHandler(notifierCmdBar, subjectListingStartedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, subjectsListedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, subjectDeletionStartedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, subjectListingErrorMsg)
	cmdbar.WithMsgHandler(notifierCmdBar, subjectDeletedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, subjectDeletionErrorNotifier)

	return &Model{
		cmdBar: cmdbar.NewTableCmdsBar[sradmin.Subject](
			cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc, nil),
			cmdbar.NewSearchCmdBar("Search subjects by name"),
			notifierCmdBar,
		),
		table:         ktable.NewDefaultTable(),
		tableFocussed: true,
		lister:        lister,
		state:         initialized,
	}, lister.ListSubjects
}
