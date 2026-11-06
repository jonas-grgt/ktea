package kadmin

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/IBM/sarama"
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

// doListTopics fetches the topics and their cleanup.policy from the controller
// in two requests. The controller connection already exists, so no other
// broker is dialed: on clusters with many brokers, connecting to each of them
// at once floods DNS (see issue #177).
func (ka *SaramaKafkaAdmin) doListTopics(
	errChan chan error,
	topicsChan chan []ListedTopic,
) {
	MaybeIntroduceLatency()
	topics, err := ka.fetchTopics()
	if err != nil {
		errChan <- err
		return
	}
	topicsChan <- topics
	close(topicsChan)
}

func (ka *SaramaKafkaAdmin) fetchTopics() ([]ListedTopic, error) {
	broker, err := ka.client.Controller()
	if err != nil {
		return nil, err
	}

	metadata, err := broker.GetMetadata(sarama.NewMetadataRequest(ka.config.Version, nil))
	if err != nil {
		return nil, err
	}

	configsReq := &sarama.DescribeConfigsRequest{}
	if ka.config.Version.IsAtLeast(sarama.V2_0_0_0) {
		configsReq.Version = 2
	} else if ka.config.Version.IsAtLeast(sarama.V1_1_0_0) {
		configsReq.Version = 1
	}
	for _, t := range metadata.Topics {
		configsReq.Resources = append(configsReq.Resources, &sarama.ConfigResource{
			Type:        TopicResourceType,
			Name:        t.Name,
			ConfigNames: []string{"cleanup.policy"},
		})
	}
	configs, err := broker.DescribeConfigs(configsReq)
	if err != nil {
		return nil, err
	}

	cleanupPolicies := make(map[string]string, len(configs.Resources))
	for _, r := range configs.Resources {
		if r.ErrorCode != 0 {
			// e.g. no permission to describe this topic's configs
			log.Errorf("error listing configs for topic %s: %v %s", r.Name, sarama.KError(r.ErrorCode), r.ErrorMsg)
			continue
		}
		for _, c := range r.Configs {
			if c.Name == "cleanup.policy" {
				cleanupPolicies[r.Name] = c.Value
			}
		}
	}

	topics := make([]ListedTopic, 0, len(metadata.Topics))
	for _, t := range metadata.Topics {
		replicas := 0
		if len(t.Partitions) > 0 {
			replicas = len(t.Partitions[0].Replicas)
		}
		cleanupPolicy, ok := cleanupPolicies[t.Name]
		if !ok {
			cleanupPolicy = "unknown"
		}
		topics = append(topics, ListedTopic{
			Name:           t.Name,
			PartitionCount: len(t.Partitions),
			Replicas:       replicas,
			Cleanup:        cleanupPolicy,
		})
	}
	return topics, nil
}
