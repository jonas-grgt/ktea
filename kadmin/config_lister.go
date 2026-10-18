package kadmin

import (
	tea "charm.land/bubbletea/v2"
	"github.com/IBM/sarama"
)

type TopicConfigLister interface {
	ListConfigs(topic string) tea.Msg
}

type TopicConfigListingStartedMsg struct {
	Err     chan error
	Configs chan map[string]string
}

type TopicConfigsListedMsg struct {
	Configs map[string]string
}

type TopicConfigListingErrorMsg struct {
	Err error
}

func (m *TopicConfigListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case e := <-m.Err:
		return TopicConfigListingErrorMsg{e}
	case c := <-m.Configs:
		return TopicConfigsListedMsg{c}
	}
}

func (ka *SaramaKafkaAdmin) ListConfigs(topic string) tea.Msg {
	errChan := make(chan error)
	configsChan := make(chan map[string]string)

	go ka.doListConfigs(topic, configsChan, errChan)

	return TopicConfigListingStartedMsg{
		errChan,
		configsChan,
	}
}

func (ka *SaramaKafkaAdmin) doListConfigs(topic string, configsChan chan map[string]string, errorChan chan error) {
	MaybeIntroduceLatency()
	configsResp, err := ka.admin.DescribeConfig(sarama.ConfigResource{
		Type: TopicResourceType,
		Name: topic,
	})
	if err != nil {
		errorChan <- err
		return
	}
	configs := make(map[string]string)
	for _, e := range configsResp {
		configs[e.Name] = e.Value
	}
	configsChan <- configs
}
