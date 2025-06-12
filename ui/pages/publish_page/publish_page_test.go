package publish_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/tests"
	"ktea/ui/components/notifier"
	"ktea/ui/pages/nav"
	"testing"
)

type MockPublisher struct {
	PublishRecordFunc func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg
}

func (m *MockPublisher) PublishRecord(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
	if m.PublishRecordFunc != nil {
		return m.PublishRecordFunc(p)
	}
	return kadmin.PublicationStartedMsg{}
}

func TestParseHeaders(t *testing.T) {
	t.Run("header format is key=value", func(t *testing.T) {
		fv := formValues{
			Headers: "key1=value1\n\nkey2=value2\n",
		}

		headers := fv.parsedHeaders()

		assert.Equal(t, map[string]string{
			"key1": "value1",
			"key2": "value2",
		}, headers)
	})

	t.Run("no headers filled in", func(t *testing.T) {
		formValues := formValues{
			Headers: "",
		}

		headers := formValues.parsedHeaders()

		assert.Equal(t, map[string]string{}, headers)
	})
}

func TestPublish(t *testing.T) {
	t.Run("esc goes back to topic list page", func(t *testing.T) {
		m := New(&MockPublisher{}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 1,
			Replicas:       1,
		})

		cmd := m.Update(tests.Key(tea.KeyEsc))

		assert.IsType(t, nav.LoadTopicsPageMsg{}, cmd())
	})

	t.Run("publish plain text", func(t *testing.T) {
		var producerRecord *kadmin.ProducerRecord
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				producerRecord = p
				return kadmin.PublicationStartedMsg{}
			},
		}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 10,
			Replicas:       1,
		})

		m.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, tests.TestRenderer)

		// Key
		tests.UpdateKeys(m, "key")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// Partition
		tests.UpdateKeys(m, "2")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// headers
		tests.UpdateKeys(m, "id=123")
		cmd = m.Update(tests.KeyWithAlt(tea.KeyEnter))
		tests.UpdateKeys(m, "user=456")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		tests.UpdateKeys(m, "payload")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		tests.Submit(m)

		assert.Equal(t, "key", producerRecord.Key)
		assert.Equal(t, "topic1", producerRecord.Topic)
		assert.Equal(t, 2, *producerRecord.Partition)
		assert.Equal(t, "payload", producerRecord.Value)
		assert.Equal(
			t,
			map[string]string{
				"id":   "123",
				"user": "456",
			},
			producerRecord.Headers,
		)
	})

	t.Run("reset form after successful publication", func(t *testing.T) {
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				return kadmin.PublicationStartedMsg{}
			},
		}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 10,
			Replicas:       1,
		})

		m.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, tests.TestRenderer)

		// Key
		tests.UpdateKeys(m, "key")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// Partition
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// headers
		tests.UpdateKeys(m, "id=123")
		cmd = m.Update(tests.KeyWithAlt(tea.KeyEnter))
		tests.UpdateKeys(m, "user=456")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		// payload
		tests.UpdateKeys(m, "payload")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		tests.Submit(m)

		m.Update(kadmin.PublicationSucceeded{})

		render := m.View(tests.TestKontext, tests.TestRenderer)

		assert.Regexp(t, "Key\\W+Payload\\W+\n.*1.*\n\\W+>\\W+\n", render)
		assert.Regexp(t, "Partition\\W+\n.*\n\\W+>\\W+\n", render)
		assert.Regexp(t, "Headers\\W+\n.*\\n\\W+1\\W+\n", render)
	})

	t.Run("publish without partition info", func(t *testing.T) {
		var producerRecord *kadmin.ProducerRecord
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				producerRecord = p
				return kadmin.PublicationStartedMsg{}
			},
		}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 10,
			Replicas:       1,
		})

		m.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, tests.TestRenderer)

		// Key
		tests.UpdateKeys(m, "key")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// Partition
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// headers
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		// payload
		tests.UpdateKeys(m, "payload")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		tests.Submit(m)

		assert.Equal(t, "key", producerRecord.Key)
		assert.Equal(t, "topic1", producerRecord.Topic)
		assert.Nil(t, producerRecord.Partition)
		assert.Equal(t, "payload", producerRecord.Value)
	})

	t.Run("upon successful publication", func(t *testing.T) {
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				return kadmin.PublicationStartedMsg{}
			},
		}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 10,
			Replicas:       1,
		})

		cmds := m.Update(kadmin.PublicationSucceeded{})
		msgs := executeBatchCmd(cmds)

		t.Run("displays success notification", func(t *testing.T) {
			render := m.View(tests.TestKontext, tests.TestRenderer)
			assert.Contains(t, render, "🎉 Record published!")
			assert.Contains(t, msgs, notifier.HideNotificationMsg{})
		})

		t.Run("hides success notification automatically", func(t *testing.T) {
			cmds := m.Update(notifier.HideNotificationMsg{})
			executeBatchCmd(cmds)

			render := m.View(tests.TestKontext, tests.TestRenderer)
			assert.NotContains(t, render, "🎉 Record published!")
		})
	})

	t.Run("ctrl+r resets the form", func(t *testing.T) {
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				return kadmin.PublicationStartedMsg{}
			},
		}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 10,
			Replicas:       1,
		})

		m.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, tests.TestRenderer)

		// Key
		tests.UpdateKeys(m, "key")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// Partition
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())

		// headers
		tests.UpdateKeys(m, "id=123")
		cmd = m.Update(tests.KeyWithAlt(tea.KeyEnter))
		tests.UpdateKeys(m, "user=456")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		// payload
		tests.UpdateKeys(m, "payload")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		tests.NextGroup(m, cmd)

		m.Update(tests.Key(tea.KeyCtrlR))

		render := m.View(tests.TestKontext, tests.TestRenderer)

		assert.Regexp(t, "Key\\W+Payload\\W+\n.*1.*\n\\W+>\\W+\n", render)
		assert.Regexp(t, "Partition\\W+\n.*\n\\W+>\\W+\n", render)
		assert.Regexp(t, "Headers\\W+\n.*\\n\\W+1\\W+\n", render)
	})

	t.Run("Validate", func(t *testing.T) {

		t.Run("When partition is not a number", func(t *testing.T) {
			m := New(&MockPublisher{}, &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 1,
				Replicas:       1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			// Key
			tests.UpdateKeys(m, "key")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			tests.UpdateKeys(m, "a1")
			m.Update(tests.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			assert.Contains(t, render, "'a1' is not a valid numeric partition value")
		})

		t.Run("When partition is negative", func(t *testing.T) {
			m := New(&MockPublisher{}, &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 1,
				Replicas:       1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			// Key
			tests.UpdateKeys(m, "key")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			tests.UpdateKeys(m, "-1")
			m.Update(tests.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			assert.Contains(t, render, "value must be at least zero")
		})

		t.Run("When partition is zero, should be allowed", func(t *testing.T) {
			m := New(&MockPublisher{}, &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 1,
				Replicas:       1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			// Key
			tests.UpdateKeys(m, "key")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			tests.UpdateKeys(m, "0")
			m.Update(tests.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)

			assert.Regexp(t, "┃ Partition.\\W+\n.*\n\\W+┃ > 0", render)
			assert.NotContains(t, render, "value must be at least zero")
		})

		t.Run("When partition exceeds number of partitions", func(t *testing.T) {
			m := New(&MockPublisher{}, &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 5,
				Replicas:       1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			// Key
			tests.UpdateKeys(m, "key")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			tests.UpdateKeys(m, "10")
			m.Update(tests.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, tests.TestRenderer)
			assert.Contains(t, render, "partition index 10 is invalid, valid range is 0-4")
		})
	})
}

func executeBatchCmd(cmd tea.Cmd) []tea.Msg {
	var msgs []tea.Msg
	if cmd == nil {
		return msgs
	}

	msg := cmd()
	if msg == nil {
		return msgs
	}

	// If the message is a batch, process its commands
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, subCmd := range batch {
			if subCmd != nil {
				msgs = append(msgs, executeBatchCmd(subCmd)...)
			}
		}
		return msgs
	}

	// Otherwise, it's a normal message
	msgs = append(msgs, msg)
	return msgs
}
