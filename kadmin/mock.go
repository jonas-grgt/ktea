package kadmin

import (
	"context"
	"ktea/sradmin"

	tea "github.com/charmbracelet/bubbletea"
)

type MockKadmin struct {
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

func (m MockKadmin) SetSra(sra sradmin.SrAdmin) {
}

func (m MockKadmin) GetClusterConfig() (ClusterConfig, error) {
	return ClusterConfig{}, nil
}

func NewMockKadminInstantiator() Instantiator {
	return func(cd ConnectionDetails) (Kadmin, error) {
		return &MockKadmin{}, nil
	}
}

func NewMockKadmin() Kadmin {
	return &MockKadmin{}
}
