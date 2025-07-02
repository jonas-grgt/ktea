package kadmin

import (
	"fmt"
	"sort"

	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type ClusterConfigLister interface {
	GetClusterConfig() tea.Msg
}

type BrokerConfigLister interface {
	GetBrokerConfig(brokerID int32) tea.Msg
}

type Broker struct {
	ID      int32
	Address string
}

type ClusterConfig struct {
	Brokers []Broker
}

type ClusterConfigStartedMsg struct {
	Err     chan error
	Configs chan ClusterConfig
}

func (m *ClusterConfigStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case e := <-m.Err:
		return ClusterConfigErrorMsg{e}
	case c := <-m.Configs:
		return ClusterConfigMsg{c}
	}
}

type ClusterConfigMsg struct {
	Config ClusterConfig
}
type ClusterConfigErrorMsg struct {
	Err error
}

// GetClusterConfig retrieves the configuration of the Kafka cluster
func (ka *SaramaKafkaAdmin) GetClusterConfig() tea.Msg {
	MaybeIntroduceLatency()
	errChan := make(chan error)
	configsChan := make(chan ClusterConfig)
	go ka.doGetClusterConfig(errChan, configsChan)
	log.Debug("Cluster configuration retrieval started")
	return ClusterConfigStartedMsg{
		Err:     errChan,
		Configs: configsChan,
	}
}

func (ka *SaramaKafkaAdmin) doGetClusterConfig(errChan chan error, configsChan chan ClusterConfig) {
	log.Debug("Fetching cluster configuration...")
	saramaBrokers, _, err := ka.admin.DescribeCluster()
	if err != nil {
		log.Error("Failed to describe cluster", "error", err)
		errChan <- err
		return
	}
	sort.Slice(saramaBrokers, func(i, j int) bool {
		return saramaBrokers[i].ID() < saramaBrokers[j].ID()
	})
	brokers := make([]Broker, 0)
	for _, saramaBroker := range saramaBrokers {
		brokers = append(brokers, Broker{
			ID:      saramaBroker.ID(),
			Address: saramaBroker.Addr(),
		})
	}
	if err != nil {
		errChan <- err
		return
	}
	log.Debug("Cluster configuration fetched successfully", "brokers", len(brokers))
	configsChan <- ClusterConfig{Brokers: brokers}
	close(configsChan)
}

type BrokerConfigListingStartedMsg struct {
	Err     chan error
	Configs chan BrokerConfig
}

type BrokerConfig struct {
	ID      int32
	Configs map[string]string
}

type BrokerConfigListedMsg struct {
	Config BrokerConfig
}

type BrokerConfigErrorMsg struct {
	Err error
}

func (m *BrokerConfigListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case e := <-m.Err:
		return BrokerConfigErrorMsg{e}
	case c := <-m.Configs:
		return BrokerConfigListedMsg{c}
	}
}

// GetBrokerConfig retrieves the configuration for a specific broker in the Kafka cluster
func (ka *SaramaKafkaAdmin) GetBrokerConfig(brokerID int32) tea.Msg {
	MaybeIntroduceLatency()
	log.Debug("Fetching config for broker", "brokerID", brokerID)
	errChan := make(chan error)
	configsChan := make(chan BrokerConfig)
	go ka.doGetBrokerConfig(brokerID, errChan, configsChan)
	log.Debug("Broker configuration retrieval started", "brokerID", brokerID)
	return BrokerConfigListingStartedMsg{
		Err:     errChan,
		Configs: configsChan,
	}
}

func (ka *SaramaKafkaAdmin) doGetBrokerConfig(brokerID int32, errChan chan error, configsChan chan BrokerConfig) {
	log.Debug("Fetching broker config", "brokerID", brokerID)
	resource := sarama.ConfigResource{
		Type: sarama.BrokerResource,
		Name: fmt.Sprintf("%d", brokerID),
	}

	entries, err := ka.admin.DescribeConfig(resource)
	if err != nil {
		log.Error("Failed to describe broker config", "brokerID", brokerID, "error", err)
		errChan <- err
		return
	}

	configMap := make(map[string]string)
	for _, entry := range entries {
		configMap[entry.Name] = entry.Value
	}

	configsChan <- BrokerConfig{ID: brokerID, Configs: configMap}
	close(configsChan)
}
