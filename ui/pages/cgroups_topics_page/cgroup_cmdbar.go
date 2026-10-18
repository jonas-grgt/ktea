package cgroups_topics_page

import (
	tea "charm.land/bubbletea/v2"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/statusbar"
)

type CGroupCmdbar[T any] struct {
	searchWidget     cmdbar.CmdBar
	notifierWidget   cmdbar.CmdBar
	active           cmdbar.CmdBar
	searchPrevActive bool
}

type NotifierConfigurerFunc func(notifier *cmdbar.NotifierCmdBar)

func (m *CGroupCmdbar[T]) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.active != nil {
		return m.active.View(ktx, renderer)
	}
	return ""
}

func (m *CGroupCmdbar[T]) Update(msg tea.Msg) (tea.Msg, tea.Cmd) {
	// when the notifier is active and has priority (because of a loading spinner) it should handle all msgs
	if m.active == m.notifierWidget {
		if m.notifierWidget.(*cmdbar.NotifierCmdBar).Notifier.HasPriority() {
			active, pmsg, cmd := m.active.Update(msg)
			if !active {
				m.active = nil
			}
			return pmsg, cmd
		}
	}

	// notifier was not actively spinning
	// if it is able to handle the msg it will return nil and the processing can stop
	active, pmsg, cmd := m.notifierWidget.Update(msg)
	if active && pmsg == nil {
		m.active = m.notifierWidget
		return msg, cmd
	}

	if _, ok := m.active.(*cmdbar.SearchCmdBar); ok {
		m.searchPrevActive = true
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "/":
			active, pmsg, cmd := m.searchWidget.Update(msg)
			if active {
				m.active = m.searchWidget
			} else {
				m.active = nil
			}
			return pmsg, cmd
		}
	}

	if m.active != nil {
		active, pmsg, cmd := m.active.Update(msg)
		if !active {
			if m.searchPrevActive {
				m.searchPrevActive = false
				m.active = m.searchWidget
			} else {
				m.active = nil
			}
		}
		return pmsg, cmd
	}

	return msg, nil
}

func (m *CGroupCmdbar[T]) HasSearchedAtLeastOneChar() bool {
	return m.searchWidget.(*cmdbar.SearchCmdBar).IsSearching() && len(m.GetSearchTerm()) > 0
}

func (m *CGroupCmdbar[T]) IsFocussed() bool {
	return m.active != nil && m.active.IsFocussed()
}

func (m *CGroupCmdbar[T]) GetSearchTerm() string {
	if searchBar, ok := m.searchWidget.(*cmdbar.SearchCmdBar); ok {
		return searchBar.GetSearchTerm()
	}
	return ""
}

func (m *CGroupCmdbar[T]) Shortcuts() []statusbar.Shortcut {
	if m.active == nil {
		return nil
	}
	return m.active.Shortcuts()
}

func NewCGroupCmdbar[T any](
	searchCmdBar *cmdbar.SearchCmdBar,
	notifierCmdBar *cmdbar.NotifierCmdBar,
) *CGroupCmdbar[T] {
	return &CGroupCmdbar[T]{
		searchCmdBar,
		notifierCmdBar,
		notifierCmdBar,
		false,
	}
}
