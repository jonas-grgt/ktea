package kadmin

import (
	"context"
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPublish(t *testing.T) {
	t.Run("Publish a text record", func(t *testing.T) {
		topic := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		})

		// when
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			psm := ka.PublishRecord(&ProducerRecord{
				Topic: topic,
				Key:   "123",
				Value: "{\"id\":\"123\"}",
			})

			select {
			case err := <-psm.Err:
				t.Fatal(c, "Unable to publish", err)
			case p := <-psm.Published:
				assert.True(c, p)
			}
		}, 10*time.Second, 10*time.Millisecond)

		// then
		ctx, cancel := context.WithCancel(context.Background())
		rsm := ka.ReadRecords(ctx, ReadDetails{
			TopicName:       topic,
			PartitionToRead: []int{0},
			StartPoint:      Beginning,
			Limit:           1,
		}).(ReadingStartedMsg)
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.Equal(c, "{\"id\":\"123\"}", (<-rsm.ConsumerRecord).Payload)
		}, 2*time.Second, 10*time.Millisecond)

		// clean up
		cancel()
		ka.DeleteTopic(topic)
	})

	t.Run("Publish with headers", func(t *testing.T) {
		topic := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		})

		// when
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			psm := ka.PublishRecord(&ProducerRecord{
				Topic: topic,
				Key:   "123",
				Value: "{\"id\":\"123\"}",
				Headers: map[string]string{
					"id":   "123",
					"user": "456",
				},
			})

			select {
			case err := <-psm.Err:
				t.Fatal(c, "Unable to publish", err)
			case p := <-psm.Published:
				assert.True(c, p)
			}
		}, 2*time.Second, 10*time.Millisecond)

		// then
		ctx, cancel := context.WithCancel(context.Background())
		rsm := ka.ReadRecords(ctx, ReadDetails{
			TopicName:       topic,
			PartitionToRead: []int{0},
			StartPoint:      0,
			Limit:           1,
		}).(ReadingStartedMsg)

		var receivedRecords []ConsumerRecord
		for {
			select {
			case r, ok := <-rsm.ConsumerRecord:
				if !ok {
					goto assertRecords
				}
				receivedRecords = append(receivedRecords, r)
			}
		}

	assertRecords:
		assert.Equal(t, "{\"id\":\"123\"}", receivedRecords[0].Payload)
		assert.Contains(t, receivedRecords[0].Headers, Header{
			"id", NewHeaderValue("123"),
		})
		assert.Contains(t, receivedRecords[0].Headers, Header{
			"user", NewHeaderValue("456"),
		})

		// clean up
		cancel()
		ka.DeleteTopic(topic)
	})

	t.Run("Publish to specific partition", func(t *testing.T) {
		topic := topicName()
		// given
		msg := ka.CreateTopic(TopicCreationDetails{
			Name:              topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		}).(TopicCreationStartedMsg)

		switch msg.AwaitCompletion().(type) {
		case TopicCreatedMsg:
		case TopicCreationErrMsg:
			t.Fatal("Unable to create topic", msg.Err)
		}

		// when
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			var partition = 2
			psm := ka.PublishRecord(&ProducerRecord{
				Topic:     topic,
				Key:       "123",
				Value:     "{\"id\":\"123\"}",
				Partition: &partition,
			})

			select {
			case err := <-psm.Err:
				t.Fatal(c, "Unable to publish", err)
			case p := <-psm.Published:
				assert.True(c, p)
			}
		}, 10*time.Second, 10*time.Millisecond)

		// then
		rsm := ka.ReadRecords(context.Background(), ReadDetails{
			TopicName:       topic,
			PartitionToRead: []int{2},
		}).(ReadingStartedMsg)
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			record := <-rsm.ConsumerRecord
			assert.Equal(c, "{\"id\":\"123\"}", record.Payload)
			assert.Equal(c, int64(2), record.Partition)
		}, 5*time.Second, 10*time.Millisecond)

		// clean up
		ka.DeleteTopic(topic)
	})
}
