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

func (s *DefaultSrClient) HardDeleteSubject(subject string) tea.Msg {
	deletedChan := make(chan bool)
	errChan := make(chan error)

	go s.doDeleteSubject(subject, true, deletedChan, errChan)

	return SubjectDeletionStartedMsg{
		subject,
		deletedChan,
		errChan,
	}
}

func (s *DefaultSrClient) SoftDeleteSubject(subject string) tea.Msg {
	deletedChan := make(chan bool)
	errChan := make(chan error)

	go s.doDeleteSubject(subject, false, deletedChan, errChan)

	return SubjectDeletionStartedMsg{
		subject,
		deletedChan,
		errChan,
	}
}

func (s *DefaultSrClient) doDeleteSubject(
	subject string,
	permanent bool,
	deletedChan chan bool,
	errChan chan error,
) {
	maybeIntroduceLatency()
	err := s.client.DeleteSubject(subject, permanent)
	if err != nil {
		errChan <- err
		return
	} else {
		deletedChan <- true
		return
	}
}
