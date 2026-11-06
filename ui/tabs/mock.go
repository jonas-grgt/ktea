package tabs

import (
	tea "charm.land/bubbletea/v2"
	"ktea/kadmin"
	"ktea/sradmin"
)

// Each mock navigator method returns a cmd producing a *CalledMsg,
// so tests can assert which navigation was requested.

type MockTopicsTabNavigator struct {
}

type ToTopicsPageCalledMsg struct {
	Refresh bool
}

type ToCreateTopicPageCalledMsg struct {
}

type ToTopicConfigPageCalledMsg struct {
	Topic string
}

type ToPublishPageCalledMsg struct {
	Topic *kadmin.ListedTopic
}

type ToConsumePageCalledMsg struct {
	Details ConsumePageDetails
}

type ToCachedConsumptionPageCalledMsg struct {
}

type ToRecordDetailsPageCalledMsg struct {
	Details RecordDetailsPageDetails
}

func (m *MockTopicsTabNavigator) ToTopicsPage(refresh bool) tea.Cmd {
	return func() tea.Msg {
		return ToTopicsPageCalledMsg{Refresh: refresh}
	}
}

func (m *MockTopicsTabNavigator) ToCreateTopicPage() tea.Cmd {
	return func() tea.Msg {
		return ToCreateTopicPageCalledMsg{}
	}
}

func (m *MockTopicsTabNavigator) ToTopicConfigPage(topic string) tea.Cmd {
	return func() tea.Msg {
		return ToTopicConfigPageCalledMsg{Topic: topic}
	}
}

func (m *MockTopicsTabNavigator) ToPublishPage(topic *kadmin.ListedTopic) tea.Cmd {
	return func() tea.Msg {
		return ToPublishPageCalledMsg{Topic: topic}
	}
}

func (m *MockTopicsTabNavigator) ToConsumePage(d ConsumePageDetails) tea.Cmd {
	return func() tea.Msg {
		return ToConsumePageCalledMsg{d}
	}
}

func (m *MockTopicsTabNavigator) ToCachedConsumptionPage() tea.Cmd {
	return func() tea.Msg {
		return ToCachedConsumptionPageCalledMsg{}
	}
}

func (m *MockTopicsTabNavigator) ToRecordDetailsPage(d RecordDetailsPageDetails) tea.Cmd {
	return func() tea.Msg {
		return ToRecordDetailsPageCalledMsg{d}
	}
}

type MockCGroupsTabNavigator struct {
}

type ToCGroupsPageCalledMsg struct {
}

type ToCGroupTopicsPageCalledMsg struct {
	Group string
}

func (m *MockCGroupsTabNavigator) ToCGroupsPage() tea.Cmd {
	return func() tea.Msg {
		return ToCGroupsPageCalledMsg{}
	}
}

func (m *MockCGroupsTabNavigator) ToCGroupTopicsPage(group string) tea.Cmd {
	return func() tea.Msg {
		return ToCGroupTopicsPageCalledMsg{Group: group}
	}
}

type MockSrTabNavigator struct {
}

type ToSubjectsPageCalledMsg struct {
	Refresh bool
}

type ToCreateSubjectPageCalledMsg struct {
}

type ToSchemaDetailsPageCalledMsg struct {
	Subject sradmin.Subject
}

func (m *MockSrTabNavigator) ToSubjectsPage(refresh bool) tea.Cmd {
	return func() tea.Msg {
		return ToSubjectsPageCalledMsg{Refresh: refresh}
	}
}

func (m *MockSrTabNavigator) ToCreateSubjectPage() tea.Cmd {
	return func() tea.Msg {
		return ToCreateSubjectPageCalledMsg{}
	}
}

func (m *MockSrTabNavigator) ToSchemaDetailsPage(subject sradmin.Subject) tea.Cmd {
	return func() tea.Msg {
		return ToSchemaDetailsPageCalledMsg{Subject: subject}
	}
}

type MockClustersTabNavigator struct {
}

type ToClustersPageCalledMsg struct {
}

func (m *MockClustersTabNavigator) ToClustersPage() tea.Cmd {
	return func() tea.Msg {
		return ToClustersPageCalledMsg{}
	}
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

func NewMockTopicsTabNavigator() TopicsTabNavigator {
	return &MockTopicsTabNavigator{}
}

func NewMockCGroupsTabNavigator() CGroupsTabNavigator {
	return &MockCGroupsTabNavigator{}
}

func NewMockSrTabNavigator() SrTabNavigator {
	return &MockSrTabNavigator{}
}

func NewMockClustersTabNavigator() ClustersTabNavigator {
	return &MockClustersTabNavigator{}
}

func NewMockKConTabNavigator() KConTabNavigator {
	return &MockKConTabNavigator{}
}
