package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

type MockSrAdmin struct {
	GetSchemaByIdFunc           func(id int) tea.Msg
	hardDeleteSubjectCallbackFn func(string) tea.Msg
	softDeleteSubjectCallbackFn func(string) tea.Msg
}

type Option func(m *MockSrAdmin)

type HardDeletedSubjectMsg struct {
	Subject string
}

type SoftDeletedSubjectMsg struct {
	Subject string
}

func (m *MockSrAdmin) DeleteSchema(subject string, version int) tea.Msg {
	panic("implement me")
}

func (m *MockSrAdmin) SoftDeleteSubject(subject string) tea.Msg {
	if m.softDeleteSubjectCallbackFn != nil {
		return m.softDeleteSubjectCallbackFn(subject)
	}
	return nil
}

type MockConnectionCheckedMsg struct {
	Config *config.SchemaRegistryConfig
}

func MockConnChecker(config *config.SchemaRegistryConfig) tea.Msg {
	return MockConnectionCheckedMsg{config}
}

func (m *MockSrAdmin) SoftDeleteSchema(subject string) tea.Msg {
	if m.softDeleteSubjectCallbackFn != nil {
		return m.softDeleteSubjectCallbackFn(subject)
	}
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

func (m *MockSrAdmin) HardDeleteSubject(subject string) tea.Msg {
	if m.hardDeleteSubjectCallbackFn != nil {
		return m.hardDeleteSubjectCallbackFn(subject)
	}
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

func (m *MockSrAdmin) ListSubjects() tea.Msg {
	return nil
}

func WithSoftDeleteSubjectCallbackFn(fn func(string) tea.Msg) Option {
	return func(m *MockSrAdmin) {
		m.softDeleteSubjectCallbackFn = fn
	}
}

func WithHardDeleteSubjectCallbackFn(fn func(string) tea.Msg) Option {
	return func(m *MockSrAdmin) {
		m.hardDeleteSubjectCallbackFn = fn
	}
}

func NewMock(options ...Option) *MockSrAdmin {
	admin := &MockSrAdmin{}
	for _, opt := range options {
		opt(admin)
	}
	return admin
}
