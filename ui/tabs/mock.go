package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/ui/pages/nav"
)

type MockTopicsTabNavigator struct {
}

type ToConsumeFormPageCalledMsg struct {
	Details nav.ConsumeFormPageDetails
}

type ToConsumePageCalledMsg struct {
	Details nav.ConsumePageDetails
}

type ToTopicsPageCalledMsg struct {
}

type MockClustersTabNavigator struct {
}

type ToClustersPageCalledMsg struct {
}

type MockKConTabNavigator struct {
}

type ToKConsPageCalledMsg struct {
}

func (m *MockKConTabNavigator) ToKConsPage() tea.Cmd {
	return func() tea.Msg {
		return ToKConsPageCalledMsg{}
	}
}

func (m *MockClustersTabNavigator) ToClustersPage() tea.Cmd {
	return func() tea.Msg {
		return ToClustersPageCalledMsg{}
	}
}

func (m *MockTopicsTabNavigator) ToConsumeFormPage(d nav.ConsumeFormPageDetails) tea.Cmd {
	return func() tea.Msg {
		return ToConsumeFormPageCalledMsg{d}
	}
}

func (m *MockTopicsTabNavigator) ToConsumePage(d nav.ConsumePageDetails) tea.Cmd {
	return func() tea.Msg {
		return ToConsumePageCalledMsg{d}
	}
}

func (m *MockTopicsTabNavigator) ToTopicsPage() tea.Cmd {
	return func() tea.Msg {
		return ToConsumePageCalledMsg{}
	}
}

func NewMockTopicsTabNavigator() TopicsTabNavigator {
	return &MockTopicsTabNavigator{}
}

func NewMockClustersTabNavigator() ClustersTabNavigator {
	return &MockClustersTabNavigator{}
}

func NewMockKConTabNavigator() KConTabNavigator {
	return &MockKConTabNavigator{}
}
