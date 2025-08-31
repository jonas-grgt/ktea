package tab

import (
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"strings"
)

type Label string

type Tab struct {
	Title string
	Label
}

type Model struct {
	tabs []Tab
	// zero indexed
	activeTab int
}

var helpTab = Tab{Title: lg.NewStyle().Foreground(lg.Color(styles.ColorYellow)).Render("≪ F1 »") + " help", Label: "help"}

func (m *Model) View(ctx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if len(m.tabs) == 0 {
		return ""
	}
	tabsToRender := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		var tab string
		if i == m.activeTab {
			tab = styles.Tab.Active.Render(t.Title)
		} else if m.lastTab(i) {
			tab = styles.Tab.Help.Render(t.Title)
		} else {
			tab = styles.Tab.Tab.Render(t.Title)
		}
		tabsToRender = append(tabsToRender, tab)
	}
	renderedTabs := lg.JoinHorizontal(lg.Top, tabsToRender...)
	tabLine := strings.Builder{}
	leftOverSpace := ctx.WindowWidth - lg.Width(renderedTabs)
	for i := 0; i < leftOverSpace; i++ {
		tabLine.WriteString("─")
	}
	s := renderedTabs + tabLine.String()
	return renderer.Render(s)
}

func (m *Model) lastTab(i int) bool {
	return i == len(m.tabs)-1
}

func (m *Model) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlLeft, tea.KeyCtrlH:
			m.Prev()
		case tea.KeyCtrlRight, tea.KeyCtrlL:
			m.Next()
		}
	}
}

func (m *Model) Next() {
	if m.activeTab < m.numberOfTabs()-2 {
		m.activeTab++
	}
}

func (m *Model) Prev() {
	if m.activeTab > 0 {
		m.activeTab--
	}
}

func (m *Model) numberOfTabs() int {
	return len(m.tabs)
}

func (m *Model) GoToTab(label Label) {
	for i, t := range m.tabs {
		if t.Label == label {
			m.activeTab = i
		}
	}
}

func (m *Model) ActiveTab() Tab {
	if m.tabs == nil {
		return Tab{}
	}
	return m.tabs[m.activeTab]
}

func New(tabs ...Tab) Model {
	return Model{
		tabs: append(tabs, helpTab),
	}
}
