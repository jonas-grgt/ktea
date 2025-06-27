package kadmin

import (
	tea "github.com/charmbracelet/bubbletea"
)

const UnknownRecordCount = -1

type TopicLister interface {
	ListTopics() tea.Msg
}

type TopicListedMsg struct {
	Topics []ListedTopic
}

type TopicRecordCount struct {
	Topic        string
	RecordCount  int64
	CountedTopic chan TopicRecordCount
}

type TopicListingStartedMsg struct {
	Err    chan error
	Topics chan []ListedTopic
}

type TopicListedErrorMsg struct {
	Err error
}

func (m *TopicListingStartedMsg) AwaitTopicListCompletion() tea.Msg {
	select {
	case topics := <-m.Topics:
		return TopicListedMsg{Topics: topics}
	case err := <-m.Err:
		return TopicListedErrorMsg{Err: err}
	}
}

type AllTopicRecordCountCalculatedMsg struct {
}

type ListedTopic struct {
	Name           string
	PartitionCount int
	Replicas       int
}

func (t *ListedTopic) Partitions() []int {
	partToConsume := make([]int, t.PartitionCount)
	for i := range t.PartitionCount {
		partToConsume[i] = i
	}
	return partToConsume
}

func (ka *SaramaKafkaAdmin) ListTopics() tea.Msg {
	errChan := make(chan error)
	topicsChan := make(chan []ListedTopic)

	go ka.doListTopics(errChan, topicsChan)

	return TopicListingStartedMsg{
		errChan,
		topicsChan,
	}
}

func (ka *SaramaKafkaAdmin) doListTopics(
	errChan chan error,
	topicsChan chan []ListedTopic,
) {
	MaybeIntroduceLatency()
	listResult, err := ka.admin.ListTopics()
	if err != nil {
		errChan <- err
		return
	}

	var topics []ListedTopic
	for name, t := range listResult {
		topics = append(topics, ListedTopic{
			name,
			int(t.NumPartitions),
			int(t.ReplicationFactor),
		})
	}
	topicsChan <- topics
	close(topicsChan)
}
