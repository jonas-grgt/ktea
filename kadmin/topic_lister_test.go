package kadmin

import (
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestListTopics(t *testing.T) {
	t.Run("List all topics", func(t *testing.T) {
		topic1 := topicName()
		topic2 := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic1,
				NumPartitions:     2,
				ReplicationFactor: 1,
			},
			{
				Topic:             topic2,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		})

		// when

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			listTopicsMsg := ka.ListTopics().(TopicListingStartedMsg)
			var topics []ListedTopic
			select {
			case topics = <-listTopicsMsg.Topics:
			case err := <-listTopicsMsg.Err:
				t.Error(t, "Failed to list topics", err)
				return
			}
			assert.Contains(t, topics, ListedTopic{topic1, 2, 1})
			assert.Contains(t, topics, ListedTopic{topic2, 1, 1})
		}, 2*time.Second, 10*time.Millisecond)

		// clean up
		ka.DeleteTopic(topic1)
		ka.DeleteTopic(topic2)
	})
}
