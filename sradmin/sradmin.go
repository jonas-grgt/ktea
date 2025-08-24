package sradmin

import tea "github.com/charmbracelet/bubbletea"

type SubjectLister interface {
	// ListSubjects returns a sradmin.SubjectListingStartedMsg
	ListSubjects() tea.Msg
}

type SubjectCreationDetails struct {
	Subject string
	Schema  string
}

// SchemaCreator registers a schema
type SchemaCreator interface {
	CreateSchema(details SubjectCreationDetails) tea.Msg
}

type SubjectDeleter interface {
	HardDeleteSubject(subject string) tea.Msg

	SoftDeleteSubject(subject string) tea.Msg
}

type SchemaDeleter interface {
	DeleteSchema(subject string, version int) tea.Msg
}

type VersionLister interface {
	ListVersions(subject string, versions []int) tea.Msg
}

type GlobalCompatibilityLister interface {
	ListGlobalCompatibility() tea.Msg
}
