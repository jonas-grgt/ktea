package kadmin

import (
	"github.com/IBM/sarama"
)

type SaramaKafkaAdmin struct {
	client   sarama.Client
	admin    sarama.ClusterAdmin
	addrs    []string
	config   *sarama.Config
	producer sarama.SyncProducer
}

func New(cd ConnectionDetails) (*SaramaKafkaAdmin, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	if cd.SASLConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = cd.SASLConfig.Username
		config.Net.SASL.Password = cd.SASLConfig.Password
	}

	client, err := sarama.NewClient(cd.BootstrapServers, config)
	if err != nil {
		return nil, err
	}

	admin, err := sarama.NewClusterAdmin(cd.BootstrapServers, config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return &SaramaKafkaAdmin{
		client:   client,
		admin:    admin,
		addrs:    cd.BootstrapServers,
		producer: producer,
		config:   config,
	}, nil
}

func (ka *SaramaKafkaAdmin) doCreateTopic(tcd TopicCreationDetails, created chan bool, errChan chan error) {
	err := ka.admin.CreateTopic(tcd.Name, &sarama.TopicDetail{
		NumPartitions:     int32(tcd.NumPartitions),
		ReplicationFactor: 1,
		ReplicaAssignment: nil,
		ConfigEntries:     nil,
	}, false)
	if err != nil {
		errChan <- err
	}
	created <- true
}

func (ka *SaramaKafkaAdmin) doPublishRecord(p *ProducerRecord, errChan chan error, published chan bool) {
	maybeIntroduceLatency()
	var partition int32
	if p.Partition == nil {
		ka.config.Producer.Partitioner = sarama.NewHashPartitioner
	} else {
		partition = int32(*p.Partition)
		ka.config.Producer.Partitioner = sarama.NewManualPartitioner
	}
	_, _, err := ka.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     p.Topic,
		Key:       sarama.StringEncoder(p.Key),
		Value:     sarama.StringEncoder(p.Value),
		Partition: partition,
	})
	if err != nil {
		errChan <- err
	}
	published <- true
}

type TopicPartitionOffset struct {
	Topic     string
	Partition int32
	Offset    int64
}
