package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MockTopicsTabNavigator struct {
}

type ToConsumeFormPageCalledMsg struct {
	Details ConsumeFormPageDetails
}

type ToConsumePageCalledMsg struct {
	Details ConsumePageDetails
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

type ToRecordDetailsPageCalledMsg struct {
	Msg LoadRecordDetailPageMsg
}

func (m *MockTopicsTabNavigator) ToRecordDetailsPage(msg LoadRecordDetailPageMsg) tea.Cmd {
	return func() tea.Msg {
		return ToRecordDetailsPageCalledMsg{msg}
	}
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

func (m *MockTopicsTabNavigator) ToConsumeFormPage(d ConsumeFormPageDetails) tea.Cmd {
	return func() tea.Msg {
		return ToConsumeFormPageCalledMsg{d}
	}
}

func (m *MockTopicsTabNavigator) ToConsumePage(d ConsumePageDetails) tea.Cmd {
	return func() tea.Msg {
		return ToConsumePageCalledMsg{d}
	}
}

func (m *MockTopicsTabNavigator) ToTopicsPage() tea.Cmd {
	return func() tea.Msg {
		return ToTopicsPageCalledMsg{}
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
