package kadmin

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

const UnknownRecordCount = -1

type TopicLister interface {
	ListTopics() tea.Msg
}

type TopicsListedMsg struct {
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
		return TopicsListedMsg{Topics: topics}
	case err := <-m.Err:
		return TopicListedErrorMsg{Err: err}
	}
}

type ListedTopic struct {
	Name           string
	PartitionCount int
	Replicas       int
	Cleanup        string
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

	var (
		topics []ListedTopic
		wg     = sync.WaitGroup{}
		mu     = sync.Mutex{}
	)

	for topic, t := range listResult {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := ka.ListConfigs(topic).(TopicConfigListingStartedMsg)

			var configs map[string]string
			switch c := msg.AwaitCompletion().(type) {
			case TopicConfigListingErrorMsg:
				log.Errorf("error listing configs for topic %s: %v", topic, c.Err)
				errChan <- c.Err
				return
			case TopicConfigsListedMsg:
				configs = c.Configs
			}

			cleanupPolicy := "unknown"
			if policy, ok := configs["cleanup.policy"]; ok {
				cleanupPolicy = policy
			}

			mu.Lock()
			topics = append(topics, ListedTopic{
				topic,
				int(t.NumPartitions),
				int(t.ReplicationFactor),
				cleanupPolicy,
			})
			mu.Unlock()
		}()
	}

	wg.Wait()

	topicsChan <- topics
	close(topicsChan)
}
