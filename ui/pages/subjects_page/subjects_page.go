package subjects_page

import (
	"fmt"
	"github.com/charmbracelet/log"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
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
	initialized state = iota
	subjectsLoaded
	loading
	noSubjectsFound
	deleting
	activeTabLbl  = border.TabLabel("active")
	deletedTabLbl = border.TabLabel("deleted")
)

type Model struct {
	table           table.Model
	rows            []table.Row
	tcb             *TableCmdsBar
	border          *border.Model
	subjects        []sradmin.Subject
	visibleSubjects []sradmin.Subject
	tableFocussed   bool
	lister          sradmin.SubjectLister
	gCompLister     sradmin.GlobalCompatibilityLister
	state           state
	// when last subject in table is deleted no subject is focussed anymore
	deletedLast     bool
	sort            cmdbar.SortLabel
	globalCompLevel string
	goToTop         bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {

	if m.state == noSubjectsFound {
		return m.border.View(lipgloss.NewStyle().
			Width(ktx.WindowWidth - 2).
			Height(ktx.AvailableHeight - 3).
			AlignVertical(lipgloss.Center).
			AlignHorizontal(lipgloss.Center).
			Render("No Subjects Found"))
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
	m.table.SetHeight(ktx.AvailableHeight - 3)
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

	return ui.JoinVertical(lipgloss.Top, cmdBarView, m.border.View(m.table.View()))
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
		case "tab":
			m.border.NextTab()
			m.table.GotoTop()
			m.tcb.Reset()
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

	var visSubjects []sradmin.Subject
	for _, subject := range m.subjects {
		if m.border.ActiveTab() == deletedTabLbl && subject.Deleted {
			visSubjects = append(visSubjects, subject)
		} else if m.border.ActiveTab() == activeTabLbl && !subject.Deleted {
			visSubjects = append(visSubjects, subject)
		}
	}
	visSubjects = m.filterSubjectsBySearchTerm(visSubjects)
	m.visibleSubjects = visSubjects
	m.rows = m.createRows(visSubjects)

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

func (m *Model) createRows(subjects []sradmin.Subject) []table.Row {
	var rows []table.Row
	for _, subject := range subjects {
		rows = append(rows, table.Row{
			subject.Name,
			strconv.Itoa(len(subject.Versions)),
			subject.Compatibility,
		})
		rows = rows
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
			return rows[i][0] < rows[j][0]
		}
	})

	return rows
}
func (m *Model) filterSubjectsBySearchTerm(subjects []sradmin.Subject) []sradmin.Subject {
	var resSubjects []sradmin.Subject
	searchTerm := m.tcb.GetSearchTerm()
	for _, subject := range subjects {
		if searchTerm != "" {
			if strings.Contains(strings.ToUpper(subject.Name), strings.ToUpper(searchTerm)) {
				resSubjects = append(resSubjects, subject)
			}
		} else {
			resSubjects = append(resSubjects, subject)
		}
	}
	return resSubjects
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	shortcuts := m.tcb.Shortcuts()
	if shortcuts == nil {
		var extraShortcuts []statusbar.Shortcut
		if len(m.visibleSubjects) > 0 {
			var deleteType string
			var tabShortcut statusbar.Shortcut
			if m.border.ActiveTab() == activeTabLbl {
				deleteType = "soft"
				tabShortcut = statusbar.Shortcut{
					Name:       "Deleted Subjects",
					Keybinding: "tab",
				}
			} else {
				deleteType = "hard"
				tabShortcut = statusbar.Shortcut{
					Name:       "Active Subjects",
					Keybinding: "tab",
				}
			}
			extraShortcuts = []statusbar.Shortcut{
				tabShortcut,
				{
					Name:       "Search",
					Keybinding: "/",
				},
				{
					Name:       fmt.Sprintf("Delete (%s)", deleteType),
					Keybinding: "F2",
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
	if len(m.visibleSubjects) > 0 {
		selectedRow := m.table.SelectedRow()
		if selectedRow != nil {
			for _, subject := range m.visibleSubjects {
				if subject.Name == selectedRow[0] {
					return &subject
				}
			}
		}
		return nil
	}
	return nil
}

func (m *Model) Title() string {
	return "Subjects"
}

func New(srClient sradmin.Client) (*Model, tea.Cmd) {
	model := Model{
		table:         ktable.NewDefaultTable(),
		tableFocussed: true,
		lister:        srClient,
		state:         initialized,
	}

	deleteMsgFn := func(subject sradmin.Subject) string {
		var deleteType string
		if model.border.ActiveTab() == activeTabLbl {
			deleteType = "soft"
		} else {
			deleteType = "hard"
		}
		message := subject.Name + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
			Bold(true).
			Render(fmt.Sprintf(" will be deleted (%s)", deleteType))
		return message
	}

	deleteFn := func(subject sradmin.Subject) tea.Cmd {
		return func() tea.Msg {
			if model.border.ActiveTab() == activeTabLbl {
				return srClient.SoftDeleteSubject(subject.Name)
			} else {
				return srClient.HardDeleteSubject(subject.Name)
			}
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

	model.border = border.New(
		border.WithInnerPaddingTop(),
		border.WithTabs(
			border.Tab{Title: "Active Subjects", TabLabel: activeTabLbl},
			border.Tab{Title: "Deleted Subjects (soft)", TabLabel: deletedTabLbl},
		),
		border.WithTitleFn(func() string {
			var compLevel string
			if model.globalCompLevel == "" {
				compLevel = ""
			} else {
				compLevel = border.KeyValueTitle("Global Compatibility", model.globalCompLevel, model.tableFocussed)
			}
			return border.KeyValueTitle("Total Subjects", fmt.Sprintf(" %d/%d", len(model.rows), len(model.subjects)), model.tableFocussed) + compLevel
		}))

	model.tcb = NewTableCmdsBar(
		srClient,
		cmdbar.NewDeleteCmdBar(deleteMsgFn, deleteFn),
		cmdbar.NewSearchCmdBar("Search subjects by name"),
		notifierCmdBar,
		sortByBar,
	)
	return &model, tea.Batch(
		srClient.ListSubjects,
		srClient.ListGlobalCompatibility,
	)
}
