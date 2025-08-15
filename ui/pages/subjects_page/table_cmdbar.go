package subjects_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/statusbar"
)

type TableCmdsBar struct {
	mainCBar       *cmdbar.TableCmdsBar[sradmin.Subject]
	hardDeleteCBar *cmdbar.DeleteCmdBar[sradmin.Subject]
}

func (m *TableCmdsBar) Update(msg tea.Msg, selection *sradmin.Subject) (tea.Msg, tea.Cmd) {
	// If the hardDeleteCBar is already focused, pass all messages to it.
	if m.hardDeleteCBar.IsFocussed() {
		active, pmsg, cmd := m.hardDeleteCBar.Update(msg)
		if !active || pmsg != nil {
			// The hardDeleteCBar is no longer focused or the msg has not been handled by the hardDeleteCBar,
			// so reset its state and return to the main tcb.
			update, t := m.mainCBar.Update(pmsg, selection)
			if m.mainCBar.IsFocussed() {
				// tcb gained focus again we need to hide the hardDeleteCBar
				m.hardDeleteCBar.Hide()
			}
			return update, t
		}
		return pmsg, cmd
	}

	// Check for the F4 key press to activate the hardDeleteCBar.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "f4" && selection != nil {
			// Toggle the hardDeleteCBar.
			active, pmsg, cmd := m.hardDeleteCBar.Update(msg)
			if active {
				m.hardDeleteCBar.Delete(*selection)
				// hardDeleteCBar is active so we can deactivate tcb
				m.mainCBar.Hide()
			}
			return pmsg, cmd
		}
	}

	// Pass all other messages to the nested tcb.
	return m.mainCBar.Update(msg, selection)
}

func (m *TableCmdsBar) IsFocussed() bool {
	return m.mainCBar.IsFocussed() || m.hardDeleteCBar.IsFocussed()
}

func (m *TableCmdsBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.mainCBar.IsFocussed() {
		return m.mainCBar.View(ktx, renderer)
	}
	if m.hardDeleteCBar.IsFocussed() {
		return m.hardDeleteCBar.View(ktx, renderer)
	}
	return ""
}

func (m *TableCmdsBar) Shortcuts() []statusbar.Shortcut {
	shortcuts := m.hardDeleteCBar.Shortcuts()
	if shortcuts != nil {
		return shortcuts
	}
	return m.mainCBar.Shortcuts()
}

func (m *TableCmdsBar) HasSearchedAtLeastOneChar() bool {
	return m.mainCBar.HasSearchedAtLeastOneChar()
}

func (m *TableCmdsBar) GetSearchTerm() string {
	return m.mainCBar.GetSearchTerm()
}

func (m *TableCmdsBar) ResetSearch() {
	m.mainCBar.ResetSearch()
}

func NewTableCmdsBar(
	deleter sradmin.SubjectDeleter,
	deleteCmdbar *cmdbar.DeleteCmdBar[sradmin.Subject],
	searchCmdBar *cmdbar.SearchCmdBar,
	notifierCmdBar *cmdbar.NotifierCmdBar,
	sortByCmdBar *cmdbar.SortByCmdBar,
) *TableCmdsBar {
	deleteFn := func(subject sradmin.Subject) tea.Cmd {
		return func() tea.Msg {
			return deleter.HardDeleteSubject(subject.Name)
		}
	}
	deleteMsgFn := func(subject sradmin.Subject) string {
		message := subject.Name + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
			Bold(true).
			Render(" will be deleted permanently (hard)")
		return message
	}
	hardDeleteCBar := cmdbar.NewDeleteCmdBar[sradmin.Subject](
		deleteMsgFn,
		deleteFn,
		cmdbar.WithDeleteKey[sradmin.Subject]("f4"),
	)
	return &TableCmdsBar{
		cmdbar.NewTableCmdsBar[sradmin.Subject](
			deleteCmdbar,
			searchCmdBar,
			notifierCmdBar,
			sortByCmdBar,
		),
		hardDeleteCBar,
	}
}
