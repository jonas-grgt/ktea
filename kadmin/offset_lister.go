package kadmin

import (
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
)

type OffsetLister interface {
	ListOffsets(group string) tea.Msg
}

type TopicPartitionOffset struct {
	Topic         string
	Partition     int32
	Offset        int64
	HighWaterMark int64
	Lag           int64
}

type OffsetListingStartedMsg struct {
	Err     chan error
	Offsets chan []TopicPartitionOffset
}

func (msg *OffsetListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case offsets := <-msg.Offsets:
		return OffsetListedMsg{offsets}
	case err := <-msg.Err:
		return OffsetListingErrorMsg{err}
	}
}

type OffsetListedMsg struct {
	Offsets []TopicPartitionOffset
}

type OffsetListingErrorMsg struct {
	Err error
}

func (ka *SaramaKafkaAdmin) ListOffsets(group string) tea.Msg {
	errChan := make(chan error)
	offsets := make(chan []TopicPartitionOffset)

	go ka.doListOffsets(group, offsets, errChan)

	return OffsetListingStartedMsg{
		errChan,
		offsets,
	}
}

func (ka *SaramaKafkaAdmin) doListOffsets(group string, offsetsChan chan []TopicPartitionOffset, errChan chan error) {
	MaybeIntroduceLatency()
	listResult, err := ka.admin.ListConsumerGroupOffsets(group, nil)
	if err != nil {
		errChan <- err
	}

	var topicPartitionOffsets []TopicPartitionOffset
	for t, m := range listResult.Blocks {
		for p, block := range m {
			hwm, _ := ka.client.GetOffset(t, p, sarama.OffsetNewest)
			topicPartitionOffsets = append(topicPartitionOffsets, TopicPartitionOffset{
				Topic:         t,
				Partition:     p,
				Offset:        block.Offset,
				HighWaterMark: hwm,
				Lag:           hwm - block.Offset,
			})
		}
	}

	offsetsChan <- topicPartitionOffsets
}
