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
		view := model.View(tests.NewKontext(), tests.Renderer)

		lines := strings.Split(view, "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " ")
		}
		normalizedView := strings.Join(lines, "\n")

		assert.Contains(t, normalizedView,
			`┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⣾ ⏳ Loading Offsets                                                                            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
╭──────────────────────────── [ Total Topics:  0 ] ╮╭─────────────────────────────── [ Total Lag:  0 ] ╮
│                                                  ││                                                  │
│ Topic Name                                       ││ Partition   Offset       High Water…  Lag        │
│──────────────────────────────────────────────────││──────────────────────────────────────────────────│`)
	})

	t.Run("List consumer groups", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListedMsg{
			Offsets: []kadmin.TopicPartitionOffset{
				{
					Topic:         "topic-1",
					Partition:     0,
					Offset:        10,
					HighWaterMark: 18,
					Lag:           8,
				},
				{
					Topic:         "topic-1",
					Partition:     1,
					Offset:        11,
					HighWaterMark: 17,
					Lag:           6,
				},
				{
					Topic:         "topic-2",
					Partition:     0,
					Offset:        30,
					HighWaterMark: 49,
					Lag:           19,
				},
				{
					Topic:         "topic-2",
					Partition:     1,
					Offset:        40,
					HighWaterMark: 69,
					Lag:           29,
				},
			},
		})

		view := model.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, view, "topic-1")
		assert.Contains(t, view, "topic-2")
		assert.Contains(t, view, "0")
		assert.Contains(t, view, "10")
		assert.Contains(t, view, "18")
		assert.Contains(t, view, "8")
		assert.Contains(t, view, "1")
		assert.Contains(t, view, "11")
		assert.Contains(t, view, "17")
		assert.Contains(t, view, "6")
		assert.Contains(t, view, "14")
		assert.NotContains(t, view, "30")
		assert.NotContains(t, view, "49")
		assert.NotContains(t, view, "19")
		assert.NotContains(t, view, "40")
		assert.NotContains(t, view, "69")
		assert.NotContains(t, view, "29")
	})

	t.Run("List consumer groups when hwm and lag values are not available", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListedMsg{
			Offsets: []kadmin.TopicPartitionOffset{
				{
					Topic:         "topic-1",
					Partition:     0,
					Offset:        10,
					HighWaterMark: kadmin.ErrorValue,
					Lag:           kadmin.ErrorValue,
				},
				{
					Topic:         "topic-1",
					Partition:     1,
					Offset:        11,
					HighWaterMark: kadmin.ErrorValue,
					Lag:           kadmin.ErrorValue,
				},
			},
		})

		view := model.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, view, "topic-1")
		assert.Contains(t, view, "0")
		assert.Contains(t, view, "10")
		assert.Contains(t, view, "1")
		assert.Contains(t, view, "11")
		assert.Equal(t, 4, strings.Count(view, na))
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

		view := model.View(tests.NewKontext(), tests.Renderer)

		model.Update(tests.Key('/'))

		view = model.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, view, "┃ >")

		t.Run("only displays matching topics", func(t *testing.T) {
			model.Update(tests.Key('2'))
			model.Update(tests.Key('2'))

			view = model.View(tests.NewKontext(), tests.Renderer)

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
				Topic:         "topic-1",
				Partition:     int32(i),
				Offset:        int64(10),
				HighWaterMark: int64(10),
				Lag:           0,
			})
		}

		model.Update(kadmin.OffsetListedMsg{
			Offsets: topicPartOffsets,
		})

		view := model.View(tests.NewKontext(), tests.Renderer)

		idx10 := strings.Index(view, "│ 10          10           10           0")
		assert.Greater(t, idx10, 0, "Expected partition 10 to be present")

		idx2 := strings.Index(view, "│ 2           10           10           0")
		assert.Greater(t, idx2, 0, "Expected partition 2 to be present")

		idx5 := strings.Index(view, "│ 5           10           10           0")
		assert.Greater(t, idx5, 0, "Expected partition 5 to be present")

		idx20 := strings.Index(view, "│ 20          10           10           0")
		assert.Greater(t, idx20, 0, "Expected partition 20 to be present")

		idx0 := strings.Index(view, "│ 0           10           10           0")
		assert.Greater(t, idx0, 0, "Expected partition 0 to be present")

		idx9 := strings.LastIndex(view, "│ 9           10           10           0")
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

		view := model.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, view, "👀 No Committed Offsets Found")
	})

}
