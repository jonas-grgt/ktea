package kadmin

import (
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"sync"
)

type TopicLister interface {
	ListTopics() tea.Msg
}

type TopicListedMsg struct {
	Topics []ListedTopic
}

type TopicListingStartedMsg struct {
	Err    chan error
	Topics chan []ListedTopic
}

type TopicListedErrorMsg struct {
	Err error
}

func (m *TopicListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case topics := <-m.Topics:
		return TopicListedMsg{Topics: topics}
	case err := <-m.Err:
		return TopicListedErrorMsg{Err: err}
	}
}

type ListedTopic struct {
	Name           string
	PartitionCount int
	Replicas       int
	RecordCount    int64
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

func (ka *SaramaKafkaAdmin) doListTopics(errChan chan error, topicsChan chan []ListedTopic) {
	MaybeIntroduceLatency()
	listResult, err := ka.admin.ListTopics()
	if err != nil {
		errChan <- err
	}

	partByTopic := make(map[string]ListedTopic)
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for name, topic := range listResult {
		wg.Add(1)

		go func(name string, topic sarama.TopicDetail) {
			defer wg.Done()

			partitions := make([]int, topic.NumPartitions)
			for i := range topic.NumPartitions {
				partitions[i] = int(i)
			}

			offsets, err := ka.fetchOffsets(partitions, name)
			if err != nil {
				errChan <- err
				wg.Done()
				return
			}

			var recordCount int64
			for _, offset := range offsets {
				recordCount += offset.firstAvailable - offset.oldest
			}

			mu.Lock()
			partByTopic[name] = ListedTopic{
				Name:           name,
				PartitionCount: int(topic.NumPartitions),
				Replicas:       int(topic.ReplicationFactor),
				RecordCount:    offsets[0].firstAvailable - offsets[0].oldest,
			}
			mu.Unlock()

		}(name, topic)
	}
	wg.Wait()

	var topics []ListedTopic
	for _, t := range partByTopic {
		topics = append(topics, ListedTopic{
			t.Name,
			t.PartitionCount,
			t.Replicas,
			t.RecordCount,
		})
	}
	topicsChan <- topics
}
