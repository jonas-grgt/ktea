package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"ktea/config"
)

type MockKcAdmin struct {
	listActiveConnectorsFunc func() tea.Msg
}

type ListActiveConnectorsCalledMsg struct{}

type ConnectorPauseCalledMsg struct {
	Name string
}

type ConnectorResumeCalledMsg struct {
	Name string
}

type Option func(*MockKcAdmin)

type MockConnectorListingStartedMsg struct {
	Connectors Connectors
}

func (c *MockConnectorListingStartedMsg) AwaitCompletion() tea.Msg {
	return ConnectorsListedMsg{c.Connectors}
}

func (m *MockKcAdmin) ListActiveConnectors() tea.Msg {
	if m.listActiveConnectorsFunc != nil {
		return m.listActiveConnectorsFunc()
	}
	return ListActiveConnectorsCalledMsg{}
}

func (m *MockKcAdmin) DeleteConnector(name string) tea.Msg {
	return nil
}

func (m *MockKcAdmin) ListVersion() tea.Msg {
	return nil
}

func (m *MockKcAdmin) Pause(name string) tea.Msg {
	return ConnectorPauseCalledMsg{name}
}

func (m *MockKcAdmin) Resume(name string) tea.Msg {
	return ConnectorResumeCalledMsg{name}
}

type MockConnectionCheckedMsg struct {
	Config *config.KafkaConnectConfig
}

func CheckConn(c *config.KafkaConnectConfig) tea.Msg {
	return MockConnectionCheckedMsg{c}
}

func NewMockConnChecker() ConnChecker {
	return CheckConn
}

func WithListActiveConnectorsResponse(msg tea.Msg) Option {
	return func(m *MockKcAdmin) {
		m.listActiveConnectorsFunc = func() tea.Msg {
			return msg
		}
	}
}

func WithListActiveConnectorsFunc(f func() tea.Msg) Option {
	return func(m *MockKcAdmin) {
		m.listActiveConnectorsFunc = f
	}
}

func NewMock(options ...Option) *MockKcAdmin {
	admin := MockKcAdmin{}
	for _, option := range options {
		option(&admin)
	}
	return &admin
}
