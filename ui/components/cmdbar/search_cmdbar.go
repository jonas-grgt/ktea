package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type state int

const (
	hidden state = iota
	searching
	searched
)

type SearchCmdBar struct {
	searchInput *huh.Input
	state       state
	placeholder string
}

func (s *SearchCmdBar) IsFocussed() bool {
	return s.state == searching
}

func (s *SearchCmdBar) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{Name: "Confirm", Keybinding: "enter"},
		{Name: "Cancel", Keybinding: "esc"},
		{Name: "Cancel", Keybinding: "/"},
		{Name: "Clear", Keybinding: "C-u"},
	}
}

func (s *SearchCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if s.state != hidden {
		style := styles.CmdBarWithWidth(ktx.WindowWidth - BorderedPadding)
		if s.state == searching {
			style = style.BorderForeground(lipgloss.Color(styles.ColorFocusBorder))
		} else {
			style = style.BorderForeground(lipgloss.Color(styles.ColorBlurBorder))
		}
		return renderer.RenderWithStyle(s.searchInput.View(), style)
	}
	return ""
}

func (s *SearchCmdBar) Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return s.handleKeyMsg(msg)
	}

	return s.isActive(), msg, nil
}

func (s *SearchCmdBar) handleKeyMsg(msg tea.KeyMsg) (bool, tea.Msg, tea.Cmd) {
	switch msg.String() {
	case "/":
		s.toggleSearch()
	case "enter":
		return s.confirmSearch(msg)
	case "esc":
		s.cancelSearch()
	default:
		if s.state == searching {
			s.updateSearchInput(msg)
		}
	}

	return s.isActive(), msg, nil
}

func (s *SearchCmdBar) toggleSearch() {
	if s.state == searching {
		log.Debug("Toggling search off")
		s.state = hidden
		emptyString := ""
		s.searchInput.Value(&emptyString)
		s.searchInput.Blur()
	} else {
		log.Debug("Toggling search on")
		s.state = searching
		s.searchInput.Focus()
	}
}

func (s *SearchCmdBar) confirmSearch(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	if s.state == searched {
		return true, msg, nil
	}

	s.searchInput.Blur()
	if s.GetSearchTerm() == "" {
		s.state = hidden
	} else {
		s.state = searched
	}

	return s.isActive(), nil, nil
}

func (s *SearchCmdBar) cancelSearch() {
	if s.state == searching {
		s.searchInput.Blur()
		s.state = hidden
		s.resetSearchInput()
	} else if s.state == searched {
	} else {
		s.state = hidden
	}
}

func (s *SearchCmdBar) updateSearchInput(msg tea.Msg) {
	input, _ := s.searchInput.Update(msg)
	if i, ok := input.(*huh.Input); ok {
		s.searchInput = i
	}
}

func (s *SearchCmdBar) resetSearchInput() {
	s.searchInput = newSearchInput(s.placeholder)
}

func (s *SearchCmdBar) GetSearchTerm() string {
	return s.searchInput.GetValue().(string)
}

func (s *SearchCmdBar) IsSearching() bool {
	return s.state == searching
}

func (s *SearchCmdBar) isActive() bool {
	return s.state == searching || s.state == searched
}

func (s *SearchCmdBar) Reset() {
	s.state = hidden
	s.searchInput = newSearchInput(s.placeholder)
}

func (s *SearchCmdBar) Hide() {
	s.state = hidden
}

func newSearchInput(placeholder string) *huh.Input {
	searchInput := huh.NewInput()
	searchInput.Init()
	return searchInput
}

func NewSearchCmdBar(placeholder string) *SearchCmdBar {
	return &SearchCmdBar{
		state:       hidden,
		searchInput: newSearchInput(placeholder),
		placeholder: placeholder,
	}
}
