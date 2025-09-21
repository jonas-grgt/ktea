package kadmin

import (
	"github.com/IBM/sarama"
	"github.com/burdiyan/kafkautil"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type Publisher interface {
	PublishRecord(p *ProducerRecord) PublicationStartedMsg
}

type ProducerRecord struct {
	Key       string
	Value     []byte
	Topic     string
	Partition *int
	Headers   map[string]string
	Timestamp time.Time
}

type PublicationStartedMsg struct {
	Err       chan error
	Published chan bool
}

type PublicationFailed struct {
	Err error
}

type PublicationSucceeded struct {
}

func (p *PublicationStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case err := <-p.Err:
		return PublicationFailed{err}
	case <-p.Published:
		return PublicationSucceeded{}
	}
}

func (ka *SaramaKafkaAdmin) PublishRecord(p *ProducerRecord) PublicationStartedMsg {
	errChan := make(chan error)
	published := make(chan bool)

	go ka.doPublishRecord(p, errChan, published)

	return PublicationStartedMsg{
		Err:       errChan,
		Published: published,
	}
}

func (ka *SaramaKafkaAdmin) doPublishRecord(
	p *ProducerRecord,
	errChan chan error,
	published chan bool,
) {
	MaybeIntroduceLatency()
	var partition int32
	if p.Partition == nil {
		ka.config.Producer.Partitioner = kafkautil.NewJVMCompatiblePartitioner
	} else {
		partition = int32(*p.Partition)
		ka.config.Producer.Partitioner = sarama.NewManualPartitioner
	}

	var headers []sarama.RecordHeader
	for key, value := range p.Headers {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	_, _, err := ka.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     p.Topic,
		Key:       sarama.StringEncoder(p.Key),
		Value:     sarama.ByteEncoder(p.Value),
		Partition: partition,
		Headers:   headers,
		Timestamp: p.Timestamp,
	})
	if err != nil {
		errChan <- err
		return
	}
	published <- true
}
