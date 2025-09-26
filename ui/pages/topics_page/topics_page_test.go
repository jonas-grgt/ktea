package topics_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests"
	"ktea/ui/tabs"
	"strings"
	"testing"
)

func TestTopicsPage(t *testing.T) {
	t.Run("Ignore KeyMsg when topics aren't loaded yet", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		cmd := page.Update(tests.Key(tea.KeyCtrlN))
		assert.NotNil(t, cmd)

		cmd = page.Update(tests.Key(tea.KeyCtrlI))
		assert.Nil(t, cmd)

		cmd = page.Update(tests.Key(tea.KeyCtrlP))
		assert.Nil(t, cmd)
	})

	t.Run("F5 refreshes topic list", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
			},
		})

		cmd := page.Update(tests.Key(tea.KeyF5))

		assert.Contains(t, tests.ExecuteBatchCmd(cmd), kadmin.ListTopicsCalledMsg{})
	})

	t.Run("When topics are loaded or refresh then the search form is reset", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
			},
		})

		page.Update(tests.Key('/'))
		tests.UpdateKeys(page, "topic2")

		render := page.View(tests.NewKontext(), tests.Renderer)
		assert.Contains(t, render, "> topic2")

		page.Update(kadmin.TopicsListedMsg{})

		render = page.View(tests.NewKontext(), tests.Renderer)
		assert.NotContains(t, render, "> topic2")
	})

	t.Run("Searching resets selected row to top row", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		var topics []kadmin.ListedTopic
		for i := range 10 {
			topics = append(topics, kadmin.ListedTopic{
				Name:           fmt.Sprintf("topic%d", i),
				PartitionCount: 1,
				Replicas:       1,
			})
		}
		_ = page.Update(kadmin.TopicsListedMsg{Topics: topics})
		page.View(tests.NewKontext(), tests.Renderer)

		page.Update(tests.Key(tea.KeyDown))
		page.Update(tests.Key(tea.KeyDown))
		page.Update(tests.Key(tea.KeyDown))

		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, "topic3", page.table.SelectedRow()[0])

		page.Update(tests.Key('/'))
		tests.UpdateKeys(page, "topic")

		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, "topic0", page.table.SelectedRow()[0])
	})

	t.Run("Default sort by Name Asc", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
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

		render := page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Name")
	})

	t.Run("Toggle sort by Name", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
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
		render := page.View(tests.NewKontext(), tests.Renderer)

		render = page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Name")

		t1Idx := strings.Index(render, "topic1")
		t2Idx := strings.Index(render, "topic2")
		t3Idx := strings.Index(render, "topic3")

		assert.Less(t, t3Idx, t1Idx)
		assert.Less(t, t3Idx, t2Idx)
		assert.Less(t, t2Idx, t1Idx)
	})

	t.Run("Toggle sort by Partitions", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
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
		render := page.View(tests.NewKontext(), tests.Renderer)

		assert.NotContains(t, render, "▲ Name")
		assert.Contains(t, render, "▼ Part")

		t1Idx := strings.Index(render, "topic1")
		t2Idx := strings.Index(render, "topic2")
		t3Idx := strings.Index(render, "topic3")

		assert.Less(t, t3Idx, t2Idx)
		assert.Less(t, t3Idx, t1Idx)
		assert.Less(t, t2Idx, t1Idx)

		page.Update(tests.Key(tea.KeyEnter))
		render = page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Part")

		t1Idx = strings.Index(render, "topic1")
		t2Idx = strings.Index(render, "topic2")
		t3Idx = strings.Index(render, "topic3")

		assert.Greater(t, t3Idx, t2Idx)
		assert.Greater(t, t3Idx, t1Idx)
		assert.Greater(t, t2Idx, t1Idx)
	})

	t.Run("Toggle sort by Replicas", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
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
		render := page.View(tests.NewKontext(), tests.Renderer)

		assert.NotContains(t, render, "▲ Name")
		assert.Contains(t, render, "▼ Repl")

		t1Idx := strings.Index(render, "b-topic1")
		t2Idx := strings.Index(render, "c-topic2")
		t3Idx := strings.Index(render, "d-topic3")

		assert.Less(t, t3Idx, t2Idx)
		assert.Less(t, t3Idx, t1Idx)
		assert.Less(t, t2Idx, t1Idx)

		page.Update(tests.Key(tea.KeyEnter))
		render = page.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Repl")

		t1Idx = strings.Index(render, "b-topic1")
		t2Idx = strings.Index(render, "c-topic2")
		t3Idx = strings.Index(render, "d-topic3")

		assert.Greater(t, t3Idx, t2Idx)
		assert.Greater(t, t3Idx, t1Idx)
		assert.Greater(t, t2Idx, t1Idx)
	})

	t.Run("C-g navigates to consume form page", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
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

		page.View(tests.NewKontext(), tests.Renderer)

		cmd := page.Update(tests.Key(tea.KeyCtrlG))

		msgs := tests.ExecuteBatchCmd(cmd)

		assert.Len(t, msgs, 1)

		assert.IsType(t, msgs[0], tabs.ToConsumeFormPageCalledMsg{})

		assert.Equal(
			t,
			tabs.ConsumeFormPageDetails{
				Topic: &kadmin.ListedTopic{
					Name:           "b-topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
				ReadDetails: nil,
			},
			msgs[0].(tabs.ToConsumeFormPageCalledMsg).Details,
		)
	})

	t.Run("enter navigates to consume page", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
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

		page.View(tests.NewKontext(), tests.Renderer)

		cmd := page.Update(tests.Key(tea.KeyEnter))

		msgs := tests.ExecuteBatchCmd(cmd)

		assert.Len(t, msgs, 1)

		assert.IsType(t, msgs[0], tabs.ToConsumePageCalledMsg{})

		assert.Equal(
			t,
			tabs.ConsumePageDetails{
				Origin: tabs.OriginTopicsPage,
				ReadDetails: kadmin.ReadDetails{
					TopicName:       "b-topic1",
					PartitionToRead: []int{0},
					StartPoint:      kadmin.MostRecent,
					Limit:           500,
					Filter: &kadmin.Filter{
						KeyFilter:       "",
						KeySearchTerm:   "",
						ValueFilter:     "",
						ValueSearchTerm: "",
					},
				},
				Topic: &kadmin.ListedTopic{
					Name:           "b-topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
			},
			msgs[0].(tabs.ToConsumePageCalledMsg).Details,
		)
	})

	t.Run("hidden internal topics", func(t *testing.T) {
		page, _ := New(
			kadmin.NewMockKadmin(),
			tabs.NewMockTopicsTabNavigator(),
		)

		_ = page.Update(kadmin.TopicsListedMsg{
			Topics: []kadmin.ListedTopic{
				{Name: "__consumer_offsets", PartitionCount: 50, Replicas: 3},
				{Name: "__schema_registry", PartitionCount: 1, Replicas: 1},
				{Name: "_schemas", PartitionCount: 1, Replicas: 1},
				{Name: "a-topics", PartitionCount: 5, Replicas: 3},
				{Name: "b-topics", PartitionCount: 3, Replicas: 2},
				{Name: "c-topics", PartitionCount: 1, Replicas: 1},
			},
		})

		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, 3, page.hiddenInternalTopicsCount)

		page.Update(tests.Key(tea.KeyF4))
		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, 0, page.hiddenInternalTopicsCount)

		page.Update(tests.Key(tea.KeyF4))
		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, 3, page.hiddenInternalTopicsCount)
	})
}
