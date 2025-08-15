package subjects_page

import (
	"fmt"
	"github.com/charmbracelet/log"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"reflect"
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
	tcb              *TableCmdsBar
	subjects         []sradmin.Subject
	renderedSubjects []sradmin.Subject
	tableFocussed    bool
	lister           sradmin.SubjectLister
	gCompLister      sradmin.GlobalCompatibilityLister
	state            state
	// when last subject in table is deleted no subject is focussed anymore
	deletedLast     bool
	sort            cmdbar.SortLabel
	globalCompLevel string
	goToTop         bool
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

	cmdBarView := m.tcb.View(ktx, renderer)

	available := ktx.WindowWidth - 8
	subjCol := int(float64(available) * 0.8)
	versionCol := int(float64(available) * 0.08)
	compCol := available - subjCol - versionCol
	m.table.SetColumns([]table.Column{
		{m.columnTitle("Subject Name"), subjCol},
		{m.columnTitle("Versions"), versionCol},
		{m.columnTitle("Compatibility"), compCol},
	})
	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetRows(m.rows)

	if m.deletedLast && m.table.SelectedRow() == nil {
		m.table.GotoBottom()
		m.deletedLast = false
	}
	if m.table.SelectedRow() == nil && len(m.table.Rows()) > 0 {
		m.table.GotoTop()
	}
	if m.goToTop {
		m.table.GotoTop()
		m.goToTop = false
	}

	embeddedText := map[styles.BorderPosition]styles.EmbeddedTextFunc{
		styles.TopMiddleBorder: func(active bool) string {
			var compLevel string
			if m.globalCompLevel == "" {
				compLevel = ""
			} else {
				compLevel = " ── " +
					styles.EmbeddedBorderText("Global Compatibility", m.globalCompLevel)(active)
			}
			return styles.EmbeddedBorderText("Total Subjects", fmt.Sprintf(" %d/%d", len(m.rows), len(m.subjects)))(active) +
				compLevel
		},
		styles.BottomMiddleBorder: styles.EmbeddedBorderText("Total Subjects", fmt.Sprintf(" %d/%d", len(m.rows), len(m.subjects))),
	}

	tableView := styles.Borderize(m.table.View(), m.tableFocussed, embeddedText)
	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) columnTitle(title string) string {
	if m.sort.Label == title {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorPink)).
			Bold(true).
			Render(m.sort.Direction.String()) + " " + title
	}
	return title
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

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
			if !m.tcb.IsFocussed() {
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
			m.goToTop = true
			m.tcb.ResetSearch()
		} else {
			m.state = noSubjectsFound
		}
	case sradmin.GlobalCompatibilityListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.GlobalCompatibilityListedMsg:
		m.globalCompLevel = msg.Compatibility
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

	_, cmd := m.tcb.Update(msg, m.SelectedSubject())
	m.tableFocussed = !m.tcb.IsFocussed()
	cmds = append(cmds, cmd)

	subjects := m.filterSubjectsBySearchTerm()
	subjects = m.sortSubjects(subjects)
	m.renderedSubjects = subjects
	m.rows = m.createRows(subjects)

	// make sure table navigation is off when the cmdbar is focussed
	if !m.tcb.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = t
		cmds = append(cmds, cmd)
	}

	if m.tcb.HasSearchedAtLeastOneChar() {
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
			subject.Compatibility,
		})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		switch m.sort.Label {
		case "Subject Name":
			if m.sort.Direction == cmdbar.Asc {
				return rows[i][0] < rows[j][0]
			}
			return rows[i][0] > rows[j][0]
		case "Versions":
			countI, _ := strconv.Atoi(rows[i][1])
			countJ, _ := strconv.Atoi(rows[j][1])
			if m.sort.Direction == cmdbar.Asc {
				return countI < countJ
			}
			return countI > countJ
		case "Compatibility":
			if m.sort.Direction == cmdbar.Asc {
				return rows[i][2] < rows[j][2]
			}
			return rows[i][2] > rows[j][2]
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.sort.Label))
		}
	})

	return rows
}

func (m *Model) filterSubjectsBySearchTerm() []sradmin.Subject {
	var subjects []sradmin.Subject
	searchTerm := m.tcb.GetSearchTerm()
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
	shortcuts := m.tcb.Shortcuts()
	if shortcuts == nil {
		var extraShortcuts []statusbar.Shortcut
		if len(m.subjects) > 0 {
			extraShortcuts = []statusbar.Shortcut{
				{
					Name:       "Search",
					Keybinding: "/",
				},
				{
					Name:       "Delete (soft)",
					Keybinding: "F2",
				},
				{
					Name:       "Delete (hard)",
					Keybinding: "F4",
				},
			}
		}

		return append([]statusbar.Shortcut{
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
		}, extraShortcuts...)
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

func New(
	lister sradmin.SubjectLister,
	compLister sradmin.GlobalCompatibilityLister,
	deleter sradmin.SubjectDeleter,
) (*Model, tea.Cmd) {
	model := Model{
		table:         ktable.NewDefaultTable(),
		tableFocussed: true,
		lister:        lister,
		state:         initialized,
	}

	deleteMsgFn := func(subject sradmin.Subject) string {
		message := subject.Name + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
			Bold(true).
			Render(" will be deleted (soft)")
		return message
	}

	deleteFn := func(subject sradmin.Subject) tea.Cmd {
		return func() tea.Msg {
			return deleter.SoftDeleteSubject(subject.Name)
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
	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg ui.RegainedFocusMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			if model.state == loading {
				cmd := m.SpinWithLoadingMsg("Loading subjects")
				return true, cmd
			}
			if model.state == deleting {
				cmd := m.SpinWithLoadingMsg("Deleting Subject")
				return true, cmd
			}
			return false, nil
		},
	)

	sortByBar := cmdbar.NewSortByCmdBar(
		[]cmdbar.SortLabel{
			{
				Label:     "Subject Name",
				Direction: cmdbar.Asc,
			},
			{
				Label:     "Versions",
				Direction: cmdbar.Desc,
			},
			{
				Label:     "Compatibility",
				Direction: cmdbar.Asc,
			},
		},
		cmdbar.WithSortSelectedCallback(func(label cmdbar.SortLabel) {
			model.sort = label
		}),
	)

	model.sort = sortByBar.SortedBy()

	model.tcb = NewTableCmdsBar(
		deleter,
		cmdbar.NewDeleteCmdBar(deleteMsgFn, deleteFn),
		cmdbar.NewSearchCmdBar("Search subjects by name"),
		notifierCmdBar,
		sortByBar,
	)
	return &model, tea.Batch(lister.ListSubjects, compLister.ListGlobalCompatibility)
}
