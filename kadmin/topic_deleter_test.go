package kadmin

import (
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeleteTopic(t *testing.T) {
	t.Run("Delete", func(t *testing.T) {
		t.Run("existing topic", func(t *testing.T) {
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
			msg := ka.DeleteTopic(topic2).(TopicDeletionStartedMsg)

			switch msg := msg.AwaitCompletion().(type) {
			case TopicDeletionErrorMsg:
				t.Fatal("Unable to delete topic", msg.Err)
			case TopicDeletedMsg:
			}

			// then
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				listTopicsMsg := ka.ListTopics().(TopicListingStartedMsg)
				var topics []ListedTopic
				select {
				case topics = <-listTopicsMsg.Topics:
				case err := <-listTopicsMsg.Err:
					t.Error(t, "Failed to list topics", err)
				}
				assert.Contains(c, topics, ListedTopic{topic1, 2, 1})
				assert.NotContains(c, topics, ListedTopic{topic2, 2, 1})
			}, 2*time.Second, 10*time.Millisecond)
			// clean up
			ka.DeleteTopic(topic1)
		})

		t.Run("none existing topic", func(t *testing.T) {
			topic1 := topicName()
			topic2 := topicName()
			// given
			createTopic(t, []kgo.TopicConfig{
				{
					Topic:             topic1,
					NumPartitions:     2,
					ReplicationFactor: 1,
				},
			})

			// when
			noneExistingTopic := topicName()
			msg := ka.DeleteTopic(topic2).(TopicDeletionStartedMsg)

			// then
			switch msg := msg.AwaitCompletion().(type) {
			case TopicDeletionErrorMsg:
				assert.Equal(t, "kafka server: Request was for a topic or partition that does not exist on this broker", msg.Err.Error())
			case TopicDeletedMsg:
				t.Fatal("Expected topic to not be deleted but was")
			}

			// clean up
			ka.DeleteTopic(topic1)
			ka.DeleteTopic(topic2)
			ka.DeleteTopic(noneExistingTopic)
		})

	})
}
