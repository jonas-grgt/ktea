package cgroups_topics_page

import (
	"fmt"
	"ktea/kadmin"
	"ktea/tests"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCgroupPartsOffsetsPage(t *testing.T) {

	t.Run("Show empty page and loading indicator when listing started", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListingStartedMsg{})
		view := model.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, view,
			`┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⣾ ⏳ Loading Offsets                                                                            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
╭─────────────── Total Topics: 0 ────────────────╮╭───────────── Total Partitions: 0 ──────────────╮
│ Topic Name                                     ││ Partition               Offset                 │
│────────────────────────────────────────────────││────────────────────────────────────────────────│
`)
	})

	t.Run("List consumer groups", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListedMsg{
			Offsets: []kadmin.TopicPartitionOffset{
				{
					Topic:     "topic-1",
					Partition: 0,
					Offset:    10,
				},
				{
					Topic:     "topic-1",
					Partition: 1,
					Offset:    11,
				},
				{
					Topic:     "topic-2",
					Partition: 0,
					Offset:    20,
				},
				{
					Topic:     "topic-2",
					Partition: 1,
					Offset:    21,
				},
			},
		})

		view := model.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, view, "topic-1")
		assert.Contains(t, view, "topic-2")
		assert.Contains(t, view, "10")
		assert.Contains(t, view, "11")
		assert.NotContains(t, view, "20")
		assert.NotContains(t, view, "21")
	})

	t.Run("Searching", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		var topicPartOffsets []kadmin.TopicPartitionOffset
		for i := 0; i < 25; i++ {
			topicPartOffsets = append(topicPartOffsets, kadmin.TopicPartitionOffset{
				Topic:     fmt.Sprintf("topic-%d", i),
				Partition: int32(0),
				Offset:    int64(10),
			})
		}

		model.Update(kadmin.OffsetListedMsg{
			Offsets: topicPartOffsets,
		})

		view := model.View(tests.NewKontext(), tests.TestRenderer)

		model.Update(tests.Key('/'))

		view = model.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, view, "┃ >")

		t.Run("only displays matching topics", func(t *testing.T) {
			model.Update(tests.Key('2'))
			model.Update(tests.Key('2'))

			view = model.View(tests.NewKontext(), tests.TestRenderer)

			assert.Contains(t, view, "┃ > 22")

			assert.Contains(t, view, "topic-22")
			assert.NotContains(t, view, "topic-1")
		})
	})

	t.Run("Order partitions ascending", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		var topicPartOffsets []kadmin.TopicPartitionOffset
		for i := 0; i < 25; i++ {
			topicPartOffsets = append(topicPartOffsets, kadmin.TopicPartitionOffset{
				Topic:     "topic-1",
				Partition: int32(i),
				Offset:    int64(10),
			})
		}

		model.Update(kadmin.OffsetListedMsg{
			Offsets: topicPartOffsets,
		})

		view := model.View(tests.NewKontext(), tests.TestRenderer)

		idx10 := strings.Index(view, "│ 10                      10")
		assert.Greater(t, idx10, 0, "Expected partition 10 to be present")

		idx2 := strings.Index(view, "│ 2                       10")
		assert.Greater(t, idx2, 0, "Expected partition 2 to be present")

		idx5 := strings.Index(view, "│ 5                       10")
		assert.Greater(t, idx5, 0, "Expected partition 5 to be present")

		idx20 := strings.Index(view, "│ 20                      10")
		assert.Greater(t, idx20, 0, "Expected partition 20 to be present")

		idx0 := strings.Index(view, "│ 0                       10")
		assert.Greater(t, idx0, 0, "Expected partition 0 to be present")

		idx9 := strings.LastIndex(view, "│ 9                       10")
		assert.Greater(t, idx9, 0, "Expected partition 9 to be present")

		assert.Less(t, idx2, idx10, "Expected partition 2 to be before partition 10")
		assert.Less(t, idx5, idx20, "Expected partition 5 to be before partition 20")
		assert.Less(t, idx0, idx9, "Expected partition 0 to be before partition 9")
	})

	t.Run("Render empty page when no offsets found", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListedMsg{
			Offsets: nil,
		})

		view := model.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, view, "👀 No Committed Offsets Found")
	})

}
