package kadmin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTopic(t *testing.T) {
	t.Run("Create new topic", func(t *testing.T) {
		topic := topicName()
		// when
		properties := map[string]string{}
		properties["compression.type"] = "lz4"
		topicCreatedMsg := ka.CreateTopic(TopicCreationDetails{
			topic,
			2,
			properties,
			1,
		}).(TopicCreationStartedMsg)

		select {
		case <-topicCreatedMsg.Created:
		case err := <-topicCreatedMsg.Err:
			t.Error("Unable to create Topic", err)
			return
		}

		// then
		listTopicsMsg := ka.ListTopics().(TopicListingStartedMsg)

		var topics []ListedTopic
		msg := listTopicsMsg.AwaitTopicListCompletion()
		switch msg := msg.(type) {
		case TopicListedMsg:
			topics = msg.Topics
		case TopicListedErrorMsg:
			assert.Fail(t, "Failed to list topics")
			return
		}

		assert.Contains(t, topics, ListedTopic{topic, 2, 1})

		// and
		var configs map[string]string
		configListingStartedMsg := ka.ListConfigs(topic).(TopicConfigListingStartedMsg)
		msg = configListingStartedMsg.AwaitCompletion()
		switch msg := msg.(type) {
		case TopicConfigsListedMsg:
			configs = msg.Configs
		case TopicConfigListingErrorMsg:
			assert.Fail(t, "Failed to list topic configs")
			return
		}

		assert.Equal(t, "lz4", configs["compression.type"])

		t.Run("Creation fails", func(t *testing.T) {
			// when
			topicCreatedMsg := ka.CreateTopic(TopicCreationDetails{
				topic,
				2,
				map[string]string{},
				3,
			}).(TopicCreationStartedMsg)

			msg = topicCreatedMsg.AwaitCompletion()
			switch msg := msg.(type) {
			case TopicCreatedMsg:
				t.Error("Topic created but expected to fail")
				return
			case TopicListedErrorMsg:
				assert.Equal(t, "kafka server: Topic with this name already exists - Topic 'topic-0' already exists.", msg.Err.Error())
			}

		})

		// clean up
		ka.DeleteTopic(topic)
	})
}
