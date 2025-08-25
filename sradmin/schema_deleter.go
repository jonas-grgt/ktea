package sradmin

import tea "github.com/charmbracelet/bubbletea"

type SchemaDeletionStartedMsg struct {
	Subject string
	Version int
	Deleted chan bool
	Err     chan error
}

type SchemaDeletedMsg struct {
	Version int
}

type SchemaDeletionErrMsg struct {
	Err error
}

func (m *SchemaDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-m.Deleted:
		return SchemaDeletedMsg{m.Version}
	case err := <-m.Err:
		return SchemaDeletionErrMsg{Err: err}
	}
}

func (s *DefaultSrClient) DeleteSchema(subject string, version int) tea.Msg {
	deletedChan := make(chan bool)
	errChan := make(chan error)

	go s.doDeleteSchema(subject, version, deletedChan, errChan)

	return SchemaDeletionStartedMsg{
		subject,
		version,
		deletedChan,
		errChan,
	}
}

func (s *DefaultSrClient) doDeleteSchema(
	subject string,
	version int,
	deletedChan chan bool,
	errChan chan error,
) {
	err := s.client.DeleteSubjectByVersion(subject, version, true)
	if err != nil {
		errChan <- err
	} else {
		deletedChan <- true
	}
}
