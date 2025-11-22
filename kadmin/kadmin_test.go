package kadmin

import (
	"context"
	"fmt"
	"ktea/config"
	"ktea/styles"
	"log"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/IBM/sarama"
	kgo "github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

var ka Kadmin
var ctx context.Context
var kc *kafka.KafkaContainer
var brokers []string

var topicSuffix = 0

func topicName() string {
	topicName := fmt.Sprintf("topic-%d", topicSuffix)
	topicSuffix++
	return topicName
}

func TestMain(m *testing.M) {
	ctx = context.Background()

	id := kafka.WithClusterID("")
	k, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", id)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to start container: %s", err))
	}
	kc = k
	brokers, _ = kc.Brokers(ctx)

	ka, err = NewSaramaKadmin(&config.Cluster{
		Name:                 "test",
		Color:                styles.ColorGreen,
		Active:               true,
		BootstrapServers:     brokers,
		SASLConfig:           config.SASLConfig{AuthMethod: config.AuthMethodNone},
		SchemaRegistry:       nil,
		TLSConfig:            config.TLSConfig{Enable: false},
		KafkaConnectClusters: nil,
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to create connection: %s", err))
	}

	exitCode := m.Run()

	if err := kc.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate Kafka container: %v", err)
	}

	os.Exit(exitCode)
}

func kafkaClient() sarama.Client {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		panic("Unable to create client")
	}
	return client
}

type testConsumer struct {
	ExpectedMsgCount int
	msgCount         int
}

func (t *testConsumer) Setup(session sarama.ConsumerGroupSession) error { return nil }

func (t *testConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (t *testConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			t.msgCount++
			session.MarkMessage(message, "")
			if t.msgCount == t.ExpectedMsgCount-1 {
				return nil
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

func createTopic(t *testing.T, topicConfigs []kgo.TopicConfig) {
	conn, err := kgo.Dial("tcp", brokers[0])
	controller, err := conn.Controller()
	kgo.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	var controllerConn *kgo.Conn
	controllerConn, err = kgo.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatal(err.Error())
	}
}
