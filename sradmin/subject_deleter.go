package sradmin

import tea "github.com/charmbracelet/bubbletea"

type SubjectDeletedMsg struct {
	SubjectName string
}

type SubjectDeletionErrorMsg struct {
	Err error
}

type SubjectDeletionStartedMsg struct {
	Subject string
	Deleted chan bool
	Err     chan error
}

func (msg *SubjectDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-msg.Deleted:
		return SubjectDeletedMsg{msg.Subject}
	case err := <-msg.Err:
		return SubjectDeletionErrorMsg{err}
	}
}

func (s *DefaultSrAdmin) DeleteSubject(subject string) tea.Msg {
	deletedChan := make(chan bool)
	errChan := make(chan error)

	go s.doDeleteSubject(subject, deletedChan, errChan)

	return SubjectDeletionStartedMsg{
		subject,
		deletedChan,
		errChan,
	}
}

func (s *DefaultSrAdmin) doDeleteSubject(
	subject string,
	deletedChan chan bool,
	errChan chan error,
) {
	maybeIntroduceLatency()
	err := s.client.DeleteSubject(subject, true)
	if err != nil {
		errChan <- err
		return
	} else {
		deletedChan <- true
		return
	}
}
