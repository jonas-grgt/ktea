package kadmin

import tea "charm.land/bubbletea/v2"

type TopicDeleter interface {
	DeleteTopic(topic string) tea.Msg
}

type TopicDeletedMsg struct {
	TopicName string
}

type TopicDeletionStartedMsg struct {
	Deleted chan bool
	Err     chan error
	Topic   string
}

type TopicDeletionErrorMsg struct {
	Err error
}

func (m *TopicDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-m.Deleted:
		return TopicDeletedMsg{TopicName: m.Topic}
	case err := <-m.Err:
		return TopicDeletionErrorMsg{Err: err}
	}
}

func (ka *SaramaKafkaAdmin) DeleteTopic(topic string) tea.Msg {
	errChan := make(chan error)
	deletedChan := make(chan bool)

	go ka.doDeleteTopic(topic, deletedChan, errChan)
	return TopicDeletionStartedMsg{
		Topic:   topic,
		Deleted: deletedChan,
		Err:     errChan,
	}
}

func (ka *SaramaKafkaAdmin) doDeleteTopic(
	topic string,
	deletedChan chan bool,
	errChan chan error,
) {
	MaybeIntroduceLatency()
	err := ka.admin.DeleteTopic(topic)
	if err != nil {
		errChan <- err
	}
	deletedChan <- true
}
