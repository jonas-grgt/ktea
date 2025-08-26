package kadmin

import (
	"context"
	"ktea/config"
	"ktea/sradmin"

	tea "github.com/charmbracelet/bubbletea"
)

type MockKadmin struct {
}

type MockConnectionCheckedMsg struct {
	Cluster *config.Cluster
}

func MockConnChecker(cluster *config.Cluster) tea.Msg {
	return MockConnectionCheckedMsg{Cluster: cluster}
}

func (m MockKadmin) CreateTopic(tcd TopicCreationDetails) tea.Msg {
	return nil
}

func (m MockKadmin) DeleteTopic(topic string) tea.Msg {
	return nil
}

func (m MockKadmin) ListTopics() tea.Msg {
	return nil
}

func (m MockKadmin) PublishRecord(p *ProducerRecord) PublicationStartedMsg {
	return PublicationStartedMsg{}
}

func (m MockKadmin) ReadRecords(ctx context.Context, rd ReadDetails) tea.Msg {
	return ReadingStartedMsg{}
}

func (m MockKadmin) ListOffsets(group string) tea.Msg {
	return nil
}

func (m MockKadmin) ListCGroups() tea.Msg {
	return nil
}

func (m MockKadmin) DeleteCGroup(name string) tea.Msg {
	return nil
}

func (m MockKadmin) UpdateConfig(t TopicConfigToUpdate) tea.Msg {
	return nil
}

func (m MockKadmin) ListConfigs(topic string) tea.Msg {
	return nil
}

func (m MockKadmin) SetSra(sra sradmin.Client) {
}

func (m MockKadmin) GetClusterConfig() tea.Msg {
	return nil
}

func (m MockKadmin) GetBrokerConfig(brokerID int32) tea.Msg {
	return nil
}

func NewMockKadminInstantiator() Instantiator {
	return func(cd ConnectionDetails) (Kadmin, error) {
		return &MockKadmin{}, nil
	}
}

func NewMockKadmin() Kadmin {
	return &MockKadmin{}
}
