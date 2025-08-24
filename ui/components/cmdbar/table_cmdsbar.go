package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type TableCmdsBar[T any] struct {
	notifierCBar     CmdBar
	deleteCBar       *DeleteCmdBar[T]
	searchCBar       *SearchCmdBar
	sortByCBar       *SortByCmdBar
	activeCBar       CmdBar
	searchPrevActive bool
}

type NotifierConfigurerFunc func(notifier *NotifierCmdBar)

func (m *TableCmdsBar[T]) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.activeCBar != nil {
		return m.activeCBar.View(ktx, renderer)
	}
	return ""
}

func (m *TableCmdsBar[T]) Update(msg tea.Msg, selection *T) (tea.Msg, tea.Cmd) {
	// when the notifier is active
	if m.activeCBar == m.notifierCBar {
		// and has priority (because of a loading spinner) it should handle all msgs
		if m.notifierCBar.(*NotifierCmdBar).Notifier.HasPriority() {
			active, pmsg, cmd := m.activeCBar.Update(msg)
			if !active {
				m.activeCBar = nil
			}
			return pmsg, cmd
		}
	}

	// notifier was not actively spinning
	// if it is able to handle the msg it will return nil and the processing can stop
	active, pmsg, cmd := m.notifierCBar.Update(msg)
	if active && pmsg == nil {
		m.activeCBar = m.notifierCBar
		return msg, cmd
	}

	if _, ok := m.activeCBar.(*SearchCmdBar); ok {
		m.searchPrevActive = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			return m.handleSlash(msg)
		case "f2":
			if selection != nil {
				return m.handleF2(selection, msg)
			}
			return nil, nil
		case "f3":
			if selection != nil && m.sortByCBar != nil {
				return m.handleF3(msg, pmsg, cmd)
			}
			return pmsg, cmd
		}
	}

	if m.activeCBar != nil {
		active, pmsg, cmd := m.activeCBar.Update(msg)
		if !active {
			if m.searchPrevActive {
				m.searchPrevActive = false
				m.activeCBar = m.searchCBar
			} else {
				m.activeCBar = nil
			}
		}
		return pmsg, cmd
	}

	return msg, nil
}

func (m *TableCmdsBar[T]) handleSlash(msg tea.Msg) (tea.Msg, tea.Cmd) {
	active, pmsg, cmd := m.searchCBar.Update(msg)
	if active {
		m.activeCBar = m.searchCBar
		m.deleteCBar.active = false
		if m.sortByCBar != nil {
			m.sortByCBar.active = false
		}
	} else {
		m.activeCBar = nil
	}
	return pmsg, cmd
}

func (m *TableCmdsBar[T]) handleF3(msg tea.Msg, pmsg tea.Msg, cmd tea.Cmd) (tea.Msg, tea.Cmd) {
	active, pmsg, cmd := m.sortByCBar.Update(msg)
	if !active {
		m.activeCBar = nil
	} else {
		m.activeCBar = m.sortByCBar
		m.searchCBar.state = hidden
		m.deleteCBar.active = false
	}
	return pmsg, cmd
}

func (m *TableCmdsBar[T]) handleF2(selection *T, msg tea.Msg) (tea.Msg, tea.Cmd) {
	active, pmsg, cmd := m.deleteCBar.Update(msg)
	if active {
		m.activeCBar = m.deleteCBar
		m.deleteCBar.Delete(*selection)
		m.searchCBar.state = hidden
		if m.sortByCBar != nil {
			m.sortByCBar.active = false
		}
	} else {
		m.activeCBar = nil
	}
	return pmsg, cmd
}

func (m *TableCmdsBar[T]) HasSearchedAtLeastOneChar() bool {
	return m.searchCBar.IsSearching() && len(m.GetSearchTerm()) > 0
}

func (m *TableCmdsBar[T]) IsFocussed() bool {
	return m.activeCBar != nil && m.activeCBar.IsFocussed()
}

func (m *TableCmdsBar[T]) GetSearchTerm() string {
	return m.searchCBar.GetSearchTerm()
}

func (m *TableCmdsBar[T]) Shortcuts() []statusbar.Shortcut {
	if m.activeCBar == nil {
		return nil
	}
	return m.activeCBar.Shortcuts()
}

func (m *TableCmdsBar[T]) ResetSearch() {
	m.searchCBar.Reset()
}

func (m *TableCmdsBar[T]) Hide() {
	m.searchCBar.state = hidden
	m.deleteCBar.Hide()
	m.sortByCBar.active = false
	m.activeCBar = nil
}

func NewTableCmdsBar[T any](
	deleteCmdBar *DeleteCmdBar[T],
	searchCmdBar *SearchCmdBar,
	notifierCmdBar *NotifierCmdBar,
	sortByCmdBar *SortByCmdBar,
) *TableCmdsBar[T] {
	return &TableCmdsBar[T]{
		notifierCmdBar,
		deleteCmdBar,
		searchCmdBar,
		sortByCmdBar,
		notifierCmdBar,
		false,
	}
}
