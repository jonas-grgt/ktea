package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

type MockSrAdmin struct {
	GetSchemaByIdFunc func(id int) tea.Msg
}

func MockConnChecker(config *config.SchemaRegistryConfig) tea.Msg {
	return config
}

func (m *MockSrAdmin) DeleteSchema(string, int) tea.Msg {
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

func (m *MockSrAdmin) DeleteSubject(string) tea.Msg {
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
