package kadmin

import (
	tea "charm.land/bubbletea/v2"
	"github.com/IBM/sarama"
)

type TopicCreator interface {
	CreateTopic(tcd TopicCreationDetails) tea.Msg
}

type TopicCreationDetails struct {
	Name              string
	NumPartitions     int
	Properties        map[string]string
	ReplicationFactor int16
}

type TopicCreatedMsg struct {
}

type TopicCreationErrMsg struct {
	Err error
}

type TopicCreationStartedMsg struct {
	Created chan bool
	Err     chan error
}

func (msg *TopicCreationStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-msg.Created:
		return TopicCreatedMsg{}
	case err := <-msg.Err:
		return TopicCreationErrMsg{Err: err}
	}
}

func (ka *SaramaKafkaAdmin) CreateTopic(tcd TopicCreationDetails) tea.Msg {
	created := make(chan bool)
	err := make(chan error)

	go ka.doCreateTopic(tcd, created, err)

	return TopicCreationStartedMsg{
		Created: created,
		Err:     err,
	}
}

func (ka *SaramaKafkaAdmin) doCreateTopic(tcd TopicCreationDetails, created chan bool, errChan chan error) {
	MaybeIntroduceLatency()
	properties := make(map[string]*string)
	for k, v := range tcd.Properties {
		properties[k] = &v
	}
	err := ka.admin.CreateTopic(tcd.Name, &sarama.TopicDetail{
		NumPartitions:     int32(tcd.NumPartitions),
		ReplicationFactor: tcd.ReplicationFactor,
		ReplicaAssignment: nil,
		ConfigEntries:     properties,
	}, false)
	if err != nil {
		errChan <- err
	}
	created <- true
}
