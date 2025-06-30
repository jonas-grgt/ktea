package tab

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"regexp"
	"strconv"
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

func (m *Model) View(ctx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if len(m.tabs) == 0 {
		return ""
	}
	tabsToRender := make([]string, len(m.tabs))
	if len(m.tabs) == 1 {
		tabsToRender = append(tabsToRender, styles.Tab.ActiveTab.Render(m.tabs[0].Title))
	} else {
		for i, t := range m.tabs {
			tabName := fmt.Sprintf("%s (Meta-%d)", t.Title, i+1)
			if i == m.activeTab {
				tabsToRender = append(tabsToRender, styles.Tab.ActiveTab.Render(tabName))
			} else {
				tabsToRender = append(tabsToRender, styles.Tab.Tab.Render(tabName))
			}
		}
	}
	renderedTabs := lipgloss.JoinHorizontal(lipgloss.Top, tabsToRender...)
	tabLine := strings.Builder{}
	leftOverSpace := ctx.WindowWidth - lipgloss.Width(renderedTabs)
	for i := 0; i < leftOverSpace; i++ {
		tabLine.WriteString("─")
	}
	s := renderedTabs + tabLine.String()
	return renderer.Render(s)
}

func (m *Model) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyPattern, _ := regexp.Compile("alt\\+(\\d)")
		if keyPattern.MatchString(msg.String()) {
			subMatch := keyPattern.FindAllStringSubmatch(msg.String(), -1)[0]
			tabNr, _ := strconv.Atoi(subMatch[1])
			if tabNr <= len(m.tabs) {
				m.GoToTab(m.tabs[tabNr-1].Label)
			}
		}
	}
}

func (m *Model) Next() {
	if m.activeTab < m.numberOfTabs()-1 {
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

func (m *Model) GoToTab(tab interface{}) {
	switch tab := tab.(type) {
	case Label:
		for i, t := range m.tabs {
			if t.Label == tab {
				m.activeTab = i
			}
		}
	case int:
		if tab <= m.numberOfTabs()-1 {
			m.activeTab = tab
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
		tabs: tabs,
	}
}
