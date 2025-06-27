package topics_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests"
	"strings"
	"testing"
)

type MockTopicLister struct {
}

type ListTopicsCalledMsg struct{}

func (m *MockTopicLister) ListTopics() tea.Msg {
	return ListTopicsCalledMsg{}
}

type MockTopicDeleter struct {
}

func (m *MockTopicDeleter) DeleteTopic(_ string) tea.Msg {
	return nil
}

func TestTopicsPage(t *testing.T) {
	t.Run("Ignore KeyMsg when topics aren't loaded yet", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		cmd := page.Update(tests.Key(tea.KeyCtrlN))
		assert.NotNil(t, cmd)

		cmd = page.Update(tests.Key(tea.KeyCtrlI))
		assert.Nil(t, cmd)

		cmd = page.Update(tests.Key(tea.KeyCtrlP))
		assert.Nil(t, cmd)
	})

	t.Run("F5 refreshes topic list", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		_ = page.Update(kadmin.TopicListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
			},
		})

		cmd := page.Update(tests.Key(tea.KeyF5))

		assert.Contains(t, tests.ExecuteBatchCmd(cmd), ListTopicsCalledMsg{})
	})

	t.Run("Default sort by Name Asc", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		_ = page.Update(kadmin.TopicListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
				{
					Name:           "topic2",
					PartitionCount: 2,
					Replicas:       1,
				},
				{
					Name:           "topic3",
					PartitionCount: 3,
					Replicas:       1,
				},
			},
		})

		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▲ Name")
	})

	t.Run("Toggle sort by Name", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		_ = page.Update(kadmin.TopicListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
				{
					Name:           "topic2",
					PartitionCount: 2,
					Replicas:       1,
				},
				{
					Name:           "topic3",
					PartitionCount: 3,
					Replicas:       1,
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.TestRenderer)

		render = page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▼ Name")

		t1Idx := strings.Index(render, "topic1")
		t2Idx := strings.Index(render, "topic2")
		t3Idx := strings.Index(render, "topic3")

		assert.Less(t, t3Idx, t1Idx)
		assert.Less(t, t3Idx, t2Idx)
		assert.Less(t, t2Idx, t1Idx)
	})

	t.Run("Toggle sort by Partitions", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		_ = page.Update(kadmin.TopicListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
				{
					Name:           "topic2",
					PartitionCount: 2,
					Replicas:       1,
				},
				{
					Name:           "topic3",
					PartitionCount: 3,
					Replicas:       1,
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyRight))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.NotContains(t, render, "▲ Name")
		assert.Contains(t, render, "▼ Part")

		t1Idx := strings.Index(render, "topic1")
		t2Idx := strings.Index(render, "topic2")
		t3Idx := strings.Index(render, "topic3")

		assert.Less(t, t3Idx, t2Idx)
		assert.Less(t, t3Idx, t1Idx)
		assert.Less(t, t2Idx, t1Idx)

		page.Update(tests.Key(tea.KeyEnter))
		render = page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▲ Part")

		t1Idx = strings.Index(render, "topic1")
		t2Idx = strings.Index(render, "topic2")
		t3Idx = strings.Index(render, "topic3")

		assert.Greater(t, t3Idx, t2Idx)
		assert.Greater(t, t3Idx, t1Idx)
		assert.Greater(t, t2Idx, t1Idx)
	})

	t.Run("Toggle sort by Replicas", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		_ = page.Update(kadmin.TopicListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "b-topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
				{
					Name:           "c-topic2",
					PartitionCount: 2,
					Replicas:       2,
				},
				{
					Name:           "d-topic3",
					PartitionCount: 3,
					Replicas:       3,
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyRight))
		page.Update(tests.Key(tea.KeyRight))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.NotContains(t, render, "▲ Name")
		assert.Contains(t, render, "▼ Repl")

		t1Idx := strings.Index(render, "b-topic1")
		t2Idx := strings.Index(render, "c-topic2")
		t3Idx := strings.Index(render, "d-topic3")

		assert.Less(t, t3Idx, t2Idx)
		assert.Less(t, t3Idx, t1Idx)
		assert.Less(t, t2Idx, t1Idx)

		page.Update(tests.Key(tea.KeyEnter))
		render = page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▲ Repl")

		t1Idx = strings.Index(render, "b-topic1")
		t2Idx = strings.Index(render, "c-topic2")
		t3Idx = strings.Index(render, "d-topic3")

		assert.Greater(t, t3Idx, t2Idx)
		assert.Greater(t, t3Idx, t1Idx)
		assert.Greater(t, t2Idx, t1Idx)
	})

}
