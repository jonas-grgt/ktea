package create_topic_page

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/tests"
	"ktea/ui/pages/nav"
	"testing"
)

func batchUpdate(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	cmd = m.Update(msg)
	msg = cmd()
	cmd = m.Update(msg)
}

func CreateTopicSectionWithCursorAtPartitionsField() *Model {
	m := New(&MockTopicCreator{})
	cmd := m.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
		Alt:   false,
		Paste: false,
	})
	batchUpdate(m, cmd)
	cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	batchUpdate(m, cmd)
	return m
}

func CreateTopicSectionWithCursorAtReplicationFactor() *Model {
	m := New(&MockTopicCreator{})
	cmd := m.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
		Alt:   false,
		Paste: false,
	})
	cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// next field
	m.Update(cmd())
	// number of partitions
	tests.UpdateKeys(m, "2")
	cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// next field
	m.Update(cmd())
	return m
}

type MockTopicCreator struct {
	CreateTopicFunc func(details kadmin.TopicCreationDetails) tea.Msg
}

func (m *MockTopicCreator) CreateTopic(tcd kadmin.TopicCreationDetails) tea.Msg {
	if m.CreateTopicFunc != nil {
		return m.CreateTopicFunc(tcd)
	}
	return nil
}

func TestCreateTopic(t *testing.T) {

	type CapturedTopicCreationDetails struct {
		kadmin.TopicCreationDetails
	}

	t.Run("esc", func(t *testing.T) {
		mockCreator := MockTopicCreator{
			CreateTopicFunc: func(details kadmin.TopicCreationDetails) tea.Msg {
				if details.Name == "" {
					return errors.New("topic name cannot be empty")
				}
				return CapturedTopicCreationDetails{details}
			},
		}
		m := New(&mockCreator)

		t.Run("goes back to topic list page", func(t *testing.T) {
			cmd := m.Update(tests.Key(tea.KeyEsc))

			assert.Equal(t, nav.LoadTopicsPageMsg{Refresh: false}, cmd())
		})

		t.Run("after at least one created topic refreshes topics list", func(t *testing.T) {
			m = New(&mockCreator)

			m.Update(kadmin.TopicCreatedMsg{})

			cmd := m.Update(tests.Key(tea.KeyEsc))

			assert.Equal(t, nav.LoadTopicsPageMsg{Refresh: true}, cmd())
		})
	})

	t.Run("c-r resets form", func(t *testing.T) {
		m := New(&MockTopicCreator{})

		// topic name
		tests.UpdateKeys(m, "topicA")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// partition count
		tests.UpdateKeys(m, "2")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// replication factor
		tests.UpdateKeys(m, "1")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// cleanup policy
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// topic config
		tests.UpdateKeys(m, "foo=bar")
		cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		// next field
		cmd = m.Update(cmd())
		// next group
		cmd = m.Update(cmd())

		render := m.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "Custom Topic configurations:")
		assert.Contains(t, render, "2")
		assert.Contains(t, render, "1")

		m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
		render = m.View(tests.NewKontext(), tests.TestRenderer)

		assert.NotContains(t, render, "Custom Topic configurations:")
		assert.NotContains(t, render, "2")
		assert.NotContains(t, render, "1")
	})

	t.Run("creation failed", func(t *testing.T) {
		mockCreator := MockTopicCreator{
			CreateTopicFunc: func(details kadmin.TopicCreationDetails) tea.Msg {
				return nil
			},
		}
		m := New(&mockCreator)

		m.Update(kadmin.TopicCreationErrMsg{
			Err: fmt.Errorf("Topic with this name already exists - Topic 'topic-0' already exists."),
		})

		render := m.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "Failed to create Topic: Topic with this name already exists - Topic 'topic-0' already exists.")

	})

	t.Run("create topic", func(t *testing.T) {
		mockCreator := MockTopicCreator{
			CreateTopicFunc: func(details kadmin.TopicCreationDetails) tea.Msg {
				if details.Name == "" {
					return errors.New("topic name cannot be empty")
				}
				return CapturedTopicCreationDetails{details}
			},
		}
		m := New(&mockCreator)

		// topic name
		tests.UpdateKeys(m, "topicA")
		cmd := m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// partition count
		tests.UpdateKeys(m, "2")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// replication factor
		tests.UpdateKeys(m, "3")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// cleanup policy
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		m.Update(cmd())
		// config
		tests.UpdateKeys(m, "delete.retention.ms=1029394")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())

		// actual submit
		msgs := tests.Submit(m)

		var capturedDetails CapturedTopicCreationDetails
		for _, msg := range msgs {
			switch m := msg.(type) {
			case CapturedTopicCreationDetails:
				capturedDetails = m
			}
		}

		assert.NotNil(t, capturedDetails)
		assert.Equal(t, CapturedTopicCreationDetails{
			TopicCreationDetails: kadmin.TopicCreationDetails{
				"topicA",
				2,
				map[string]string{
					"cleanup.policy":      "delete-compact",
					"delete.retention.ms": "1029394",
				},
				int16(3),
			},
		}, capturedDetails)
	})
}

func TestCreateTopic_Validation(t *testing.T) {
	t.Run("Validate ListedTopic Name", func(t *testing.T) {
		t.Run("When field is empty", func(t *testing.T) {
			m := New(&MockTopicCreator{})

			cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			batchUpdate(m, cmd)

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "* topic Name cannot be empty")
		})
	})

	t.Run("Validate Number of Partitions", func(t *testing.T) {

		t.Run("When field is empty", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtPartitionsField()

			cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			batchUpdate(m, cmd)

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "* number of Partitions cannot be empty")
		})

		t.Run("When field is zero", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtPartitionsField()

			cmd := m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'0'},
				Alt:   false,
				Paste: false,
			})
			batchUpdate(m, cmd)
			cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			batchUpdate(m, cmd)

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "value must be greater than zero")
		})

		t.Run("When field is negative", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtPartitionsField()

			cmd := m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'-'},
				Alt:   false,
				Paste: false,
			})
			batchUpdate(m, cmd)
			cmd = m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'1'},
				Alt:   false,
				Paste: false,
			})
			cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			batchUpdate(m, cmd)

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "value must be greater than zero")
		})

		t.Run("When field is not a number", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtPartitionsField()

			cmd := m.Update(tests.Key('a'))
			batchUpdate(m, cmd)
			cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			batchUpdate(m, cmd)

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "'a' is not a valid numeric partition count value")
		})
	})

	t.Run("Validate Replication Factor", func(t *testing.T) {
		t.Run("When field is empty", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtReplicationFactor()

			m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "* replication factory cannot be empty")
		})

		t.Run("When field is zero", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtReplicationFactor()

			m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'0'},
				Alt:   false,
				Paste: false,
			})
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "value must be greater than zero")
		})

		t.Run("When field is negative", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtReplicationFactor()

			cmd := m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'-'},
				Alt:   false,
				Paste: false,
			})
			batchUpdate(m, cmd)
			cmd = m.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'1'},
				Alt:   false,
				Paste: false,
			})
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "value must be greater than zero")
		})

		t.Run("When field is not a number", func(t *testing.T) {
			m := CreateTopicSectionWithCursorAtPartitionsField()

			m.Update(tests.Key('a'))
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			render := m.View(&kontext.ProgramKtx{}, tests.TestRenderer)

			assert.Contains(t, render, "'a' is not a valid numeric partition count value")
		})

	})

	t.Run("Validate configuration", func(t *testing.T) {

		t.Run("When field does not conform config=value format", func(t *testing.T) {
			m := New(&MockTopicCreator{})

			// topic name
			tests.UpdateKeys(m, "topicA")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// partition count
			tests.UpdateKeys(m, "2")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// replication factor
			tests.UpdateKeys(m, "2")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// cleanup policy
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// next field
			tests.UpdateKeys(m, "foo:bar")
			cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			render := m.View(tests.NewKontext(), tests.TestRenderer)

			assert.Contains(t, render, "please enter configurations in the format \"config=value\"")
		})

		t.Run("When field conforms config=value format", func(t *testing.T) {
			m := New(&MockTopicCreator{})

			// topic name
			tests.UpdateKeys(m, "topicA")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// partition count
			tests.UpdateKeys(m, "2")
			m.Update(cmd())
			cmd = m.Update(tests.Key(tea.KeyEnter))
			// replication factor
			tests.UpdateKeys(m, "2")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())
			// cleanup policy
			m.Update(cmd())
			cmd = m.Update(tests.Key(tea.KeyEnter))
			// next field
			tests.UpdateKeys(m, "foo=bar")
			cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			batchUpdate(m, cmd)

			render := m.View(tests.NewKontext(), tests.TestRenderer)

			assert.NotContains(t, render, "please enter configurations in the format \"config=value\"")
		})
	})
}
