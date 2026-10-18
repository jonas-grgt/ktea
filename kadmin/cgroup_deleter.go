package kadmin

import tea "charm.land/bubbletea/v2"

type CGroupDeleter interface {
	DeleteCGroup(name string) tea.Msg
}

type CGroupDeletionStartedMsg struct {
	GroupName string
	Deleted   chan bool
	Err       chan error
}

func (c *CGroupDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Deleted:
		return CGroupDeletedMsg{GroupName: c.GroupName}
	case err := <-c.Err:
		return CGroupDeletionErrMsg{Err: err}
	}
}

type CGroupDeletionErrMsg struct {
	Err error
}

type CGroupDeletedMsg struct {
	GroupName string
}

func (ka *SaramaKafkaAdmin) DeleteCGroup(name string) tea.Msg {
	errChan := make(chan error)
	deletedChan := make(chan bool)

	go ka.doDeleteCGroup(name, deletedChan, errChan)

	return CGroupDeletionStartedMsg{
		GroupName: name,
		Deleted:   deletedChan,
		Err:       errChan,
	}
}

func (ka *SaramaKafkaAdmin) doDeleteCGroup(
	name string,
	deletedChan chan bool,
	errChan chan error,
) {
	err := ka.admin.DeleteConsumerGroup(name)
	if err != nil {
		errChan <- err
		return
	}
	deletedChan <- true
}
