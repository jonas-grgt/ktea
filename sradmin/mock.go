package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

type MockSrAdmin struct {
	GetSchemaByIdFunc func(id int) tea.Msg
}

func (m *MockSrAdmin) DeleteSchema(subject string, version int) tea.Msg {
	panic("implement me")
}

func (m *MockSrAdmin) SoftDeleteSubject(subject string) tea.Msg {
	panic("implement me")
}

type MockConnectionCheckedMsg struct {
	Config *config.SchemaRegistryConfig
}

func MockConnChecker(config *config.SchemaRegistryConfig) tea.Msg {
	return MockConnectionCheckedMsg{config}
}

func (m *MockSrAdmin) SoftDeleteSchema(string, int) tea.Msg {
	return nil
}

func (m *MockSrAdmin) ListGlobalCompatibility() tea.Msg {
	return nil
}

func (m *MockSrAdmin) GetSchemaById(id int) tea.Msg {
	if m.GetSchemaByIdFunc != nil {
		return m.GetSchemaByIdFunc(id)
	}
	return nil
}

func (m *MockSrAdmin) HardDeleteSubject(string) tea.Msg {
	return nil
}

func (m *MockSrAdmin) ListSubjects() tea.Msg {
	return nil
}

func (m *MockSrAdmin) CreateSchema(SubjectCreationDetails) tea.Msg {
	return nil
}

func (m *MockSrAdmin) ListVersions(string, []int) tea.Msg {
	return nil
}

func (m *MockSrAdmin) GetLatestSchemaBySubject(string) tea.Msg {
	return nil
}

func NewMock() *MockSrAdmin {
	return &MockSrAdmin{}
}
