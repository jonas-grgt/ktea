package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/ui/pages/nav"
)

type TopicsTabNavigator interface {
	ToConsumeFormPage(d nav.ConsumeFormPageDetails) tea.Cmd

	ToConsumePage(msg nav.ConsumePageDetails) tea.Cmd

	ToTopicsPage() tea.Cmd
}

type ClustersTabNavigator interface {
	ToClustersPage() tea.Cmd
}

type KConTabNavigator interface {
	ToKConsPage() tea.Cmd
}
