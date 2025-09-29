package consume_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/serdes"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"ktea/ui/tabs"
	"strings"
	"testing"
	"time"
)

func TestConsumptionPage(t *testing.T) {
	t.Run("Display empty topic message and adjusted shortcuts", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		m.Update(kadmin.EmptyTopicMsg{})

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "Empty topic")

		assert.Equal(t, []statusbar.Shortcut{{"Go Back", "esc"}}, m.Shortcuts())
	})

	t.Run("Display adjusted shortcuts when consuming", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		m.Update(&kadmin.ReadingStartedMsg{})

		assert.Equal(t, []statusbar.Shortcut{
			{"View Record", "enter"},
			{"Stop consuming", "F2"},
			{"Go Back", "esc"},
		}, m.Shortcuts())
	})

	t.Run("Title contains topic", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{
				TopicName:       "topic-name",
				PartitionToRead: nil,
				StartPoint:      0,
				Limit:           0,
				Filter:          nil,
			},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		m.Update(&kadmin.ReadingStartedMsg{})

		assert.Equal(t, "Topics / topic-name / Records", m.Title())
	})

	t.Run("Display msg when no records found for given criteria", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		m.Update(kadmin.NoRecordsFound{})

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "No records found for the given criteria")

		assert.Equal(t, []statusbar.Shortcut{{"Go Back", "esc"}}, m.Shortcuts())
	})

	t.Run("null keys are rendered as <null>", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var (
			records []kadmin.ConsumerRecord
			now     = time.Now()
		)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       "",
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		render := m.View(tests.NewKontext(), tests.Renderer)

		count := strings.Count(render, "<null>")
		assert.Equal(t, 10, count, "expected <null> to appear 10 times")
	})

	t.Run("Default sort by Timestamp Desc", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Timestamp")

		t1Idx := strings.Index(render, "2024-01-01 00:00:09")
		t2Idx := strings.Index(render, "2024-01-01 00:00:08")
		t3Idx := strings.Index(render, "2024-01-01 00:00:07")
		t4Idx := strings.Index(render, "2024-01-01 00:00:06")
		t5Idx := strings.Index(render, "2024-01-01 00:00:00")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t2Idx, t3Idx)
		assert.Less(t, t3Idx, t4Idx)
		assert.Less(t, t4Idx, t5Idx)
	})

	t.Run("Default sort by Timestamp Desc", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Timestamp")

		t1Idx := strings.Index(render, "2024-01-01 00:00:09")
		t2Idx := strings.Index(render, "2024-01-01 00:00:08")
		t3Idx := strings.Index(render, "2024-01-01 00:00:07")
		t4Idx := strings.Index(render, "2024-01-01 00:00:06")
		t5Idx := strings.Index(render, "2024-01-01 00:00:00")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t2Idx, t3Idx)
		assert.Less(t, t3Idx, t4Idx)
		assert.Less(t, t4Idx, t5Idx)
	})

	t.Run("F2 Cancels consumption", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		cmd := m.Update(tests.Key(tea.KeyF2))

		msgs := tests.ExecuteBatchCmd(cmd)

		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.ConsumptionEndedMsg{}, msgs[0])
	})

	t.Run("F2 Cancels consumption", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		originalCancel := m.(*Model).cancelConsumption
		var cancelCalled bool
		m.(*Model).cancelConsumption = func() {
			cancelCalled = true
			originalCancel()
		}

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		cmd := m.Update(tests.Key(tea.KeyF2))

		msgs := tests.ExecuteBatchCmd(cmd)

		assert.True(t, cancelCalled, "expected cancelConsumption to be called")
		assert.Len(t, msgs, 1)
		assert.IsType(t, kadmin.ConsumptionEndedMsg{}, msgs[0])
	})

	t.Run("esc", func(t *testing.T) {
		t.Run("goes back to topics page when live consuming", func(t *testing.T) {
			m, _ := New(
				kadmin.NewMockKadmin(),
				kadmin.ReadDetails{
					TopicName:       "topic1",
					PartitionToRead: []int{0},
					StartPoint:      kadmin.Live,
					Limit:           500,
					Filter:          &kadmin.Filter{},
				},
				&kadmin.ListedTopic{},
				tabs.OriginTopicsPage,
				tabs.NewMockTopicsTabNavigator(),
			)

			originalCancel := m.(*Model).cancelConsumption
			var cancelCalled bool
			m.(*Model).cancelConsumption = func() {
				cancelCalled = true
				originalCancel()
			}

			var records []kadmin.ConsumerRecord
			now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
			for i := 0; i < 10; i++ {
				records = append(records, kadmin.ConsumerRecord{
					Key:       fmt.Sprintf("key-%d", i),
					Payload:   serdes.DesData{},
					Err:       nil,
					Partition: 0,
					Offset:    int64(i),
					Headers:   nil,
					Timestamp: now.Add(time.Duration(i) * time.Second),
				})
			}
			m.Update(kadmin.ConsumerRecordReceived{
				Records: records,
			})

			cmd := m.Update(tests.Key(tea.KeyEsc))

			msgs := tests.ExecuteBatchCmd(cmd)

			assert.True(t, cancelCalled, "expected cancelConsumption to be called")
			assert.Len(t, msgs, 1)
			assert.IsType(t, tabs.ToTopicsPageCalledMsg{}, msgs[0])
		})

		t.Run("goes back to topics page when origin was topics page", func(t *testing.T) {
			m, _ := New(
				kadmin.NewMockKadmin(),
				kadmin.ReadDetails{
					TopicName:       "topic1",
					PartitionToRead: []int{0},
					StartPoint:      kadmin.Yesterday,
					Limit:           500,
					Filter:          &kadmin.Filter{},
				},
				&kadmin.ListedTopic{},
				tabs.OriginTopicsPage,
				tabs.NewMockTopicsTabNavigator(),
			)

			originalCancel := m.(*Model).cancelConsumption
			var cancelCalled bool
			m.(*Model).cancelConsumption = func() {
				cancelCalled = true
				originalCancel()
			}

			var records []kadmin.ConsumerRecord
			now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
			for i := 0; i < 10; i++ {
				records = append(records, kadmin.ConsumerRecord{
					Key:       fmt.Sprintf("key-%d", i),
					Payload:   serdes.DesData{},
					Err:       nil,
					Partition: 0,
					Offset:    int64(i),
					Headers:   nil,
					Timestamp: now.Add(time.Duration(i) * time.Second),
				})
			}
			m.Update(kadmin.ConsumerRecordReceived{
				Records: records,
			})

			cmd := m.Update(tests.Key(tea.KeyEsc))

			msgs := tests.ExecuteBatchCmd(cmd)

			assert.True(t, cancelCalled, "expected cancelConsumption to be called")
			assert.Len(t, msgs, 1)
			assert.IsType(t, tabs.ToTopicsPageCalledMsg{}, msgs[0])
		})

		t.Run("goes back to consume form page when origin was consume form page", func(t *testing.T) {
			m, _ := New(
				kadmin.NewMockKadmin(),
				kadmin.ReadDetails{
					TopicName:       "topic1",
					PartitionToRead: []int{0},
					StartPoint:      kadmin.Yesterday,
					Limit:           500,
					Filter:          &kadmin.Filter{},
				},
				&kadmin.ListedTopic{},
				tabs.OriginConsumeFormPage,
				tabs.NewMockTopicsTabNavigator(),
			)

			originalCancel := m.(*Model).cancelConsumption
			var cancelCalled bool
			m.(*Model).cancelConsumption = func() {
				cancelCalled = true
				originalCancel()
			}

			var records []kadmin.ConsumerRecord
			now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
			for i := 0; i < 10; i++ {
				records = append(records, kadmin.ConsumerRecord{
					Key:       fmt.Sprintf("key-%d", i),
					Payload:   serdes.DesData{},
					Err:       nil,
					Partition: 0,
					Offset:    int64(i),
					Headers:   nil,
					Timestamp: now.Add(time.Duration(i) * time.Second),
				})
			}
			m.Update(kadmin.ConsumerRecordReceived{
				Records: records,
			})

			cmd := m.Update(tests.Key(tea.KeyEsc))

			msgs := tests.ExecuteBatchCmd(cmd)

			assert.True(t, cancelCalled, "expected cancelConsumption to be called")
			assert.Len(t, msgs, 1)
			assert.IsType(t, tabs.ToConsumeFormPageCalledMsg{}, msgs[0])
		})
	})

	t.Run("F3 shows sort bar", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		m.Update(tests.Key(tea.KeyF3))

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Timestamp")
	})

	t.Run("Sort by Key", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: 0,
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		m.Update(tests.Key(tea.KeyF3))
		m.Update(tests.Key(tea.KeyLeft))
		m.Update(tests.Key(tea.KeyEnter))

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Key")

		t1Idx := strings.Index(render, "key-0")
		t2Idx := strings.Index(render, "key-1")
		t3Idx := strings.Index(render, "key-2")
		t4Idx := strings.Index(render, "key-3")
		t5Idx := strings.Index(render, "key-4")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t2Idx, t3Idx)
		assert.Less(t, t3Idx, t4Idx)
		assert.Less(t, t4Idx, t5Idx)

		m.Update(tests.Key(tea.KeyEnter))

		render = m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Key")

		t1Idx = strings.Index(render, "key-0")
		t2Idx = strings.Index(render, "key-1")
		t3Idx = strings.Index(render, "key-2")
		t4Idx = strings.Index(render, "key-3")
		t5Idx = strings.Index(render, "key-4")

		assert.Greater(t, t1Idx, t2Idx)
		assert.Greater(t, t2Idx, t3Idx)
		assert.Greater(t, t3Idx, t4Idx)
		assert.Greater(t, t4Idx, t5Idx)
	})

	t.Run("Sort by Partitions", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: int64(i),
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		m.Update(tests.Key(tea.KeyF3))
		m.Update(tests.Key(tea.KeyRight))
		m.Update(tests.Key(tea.KeyEnter))

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Partition")

		t1Idx := strings.Index(render, "key-0")
		t2Idx := strings.Index(render, "key-1")
		t3Idx := strings.Index(render, "key-2")
		t4Idx := strings.Index(render, "key-3")
		t5Idx := strings.Index(render, "key-4")

		assert.Greater(t, t1Idx, t2Idx)
		assert.Greater(t, t2Idx, t3Idx)
		assert.Greater(t, t3Idx, t4Idx)
		assert.Greater(t, t4Idx, t5Idx)

		m.Update(tests.Key(tea.KeyEnter))

		render = m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Partition")

		t1Idx = strings.Index(render, "key-0")
		t2Idx = strings.Index(render, "key-1")
		t3Idx = strings.Index(render, "key-2")
		t4Idx = strings.Index(render, "key-3")
		t5Idx = strings.Index(render, "key-4")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t2Idx, t3Idx)
		assert.Less(t, t3Idx, t4Idx)
		assert.Less(t, t4Idx, t5Idx)
	})

	t.Run("Sort by Offset", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: int64(i),
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		m.Update(tests.Key(tea.KeyF3))
		m.Update(tests.Key(tea.KeyRight))
		m.Update(tests.Key(tea.KeyRight))
		m.Update(tests.Key(tea.KeyEnter))

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▼ Offset")

		t1Idx := strings.Index(render, "key-0")
		t2Idx := strings.Index(render, "key-1")
		t3Idx := strings.Index(render, "key-2")
		t4Idx := strings.Index(render, "key-3")
		t5Idx := strings.Index(render, "key-4")

		assert.Greater(t, t1Idx, t2Idx)
		assert.Greater(t, t2Idx, t3Idx)
		assert.Greater(t, t3Idx, t4Idx)
		assert.Greater(t, t4Idx, t5Idx)

		m.Update(tests.Key(tea.KeyEnter))

		render = m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Offset")

		t1Idx = strings.Index(render, "key-0")
		t2Idx = strings.Index(render, "key-1")
		t3Idx = strings.Index(render, "key-2")
		t4Idx = strings.Index(render, "key-3")
		t5Idx = strings.Index(render, "key-4")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t2Idx, t3Idx)
		assert.Less(t, t3Idx, t4Idx)
		assert.Less(t, t4Idx, t5Idx)
	})

	t.Run("Sorting toggles shortcuts", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: int64(i),
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		assert.Equal(t, []statusbar.Shortcut{
			{"View Record", "enter"},
			{"Sort", "F3"},
			{"Go Back", "esc"},
		}, m.Shortcuts())

		m.Update(tests.Key(tea.KeyF3))

		assert.Equal(t, []statusbar.Shortcut{
			{"Cancel Sorting", "F3"},
			{"Select Sorting Column", "←/→/h/l"},
			{"Apply Sorting Column", "enter"},
		}, m.Shortcuts())
	})

	t.Run("Enter load consume details page", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{
				TopicName: "topic1",
			},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: int64(i),
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		m.View(tests.NewKontext(), tests.Renderer)

		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		cmd := m.Update(tests.Key(tea.KeyEnter))

		msgs := tests.ExecuteBatchCmd(cmd)
		assert.Len(t, msgs, 1)
		assert.IsType(t, tabs.ToRecordDetailsPageCalledMsg{}, msgs[0])
		assert.Equal(t, tabs.LoadRecordDetailPageMsg{
			Record: &kadmin.ConsumerRecord{
				Key:       "key-6",
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: int64(6),
				Offset:    int64(6),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(6) * time.Second),
			},
			TopicName: "topic1",
		}, msgs[0].(tabs.ToRecordDetailsPageCalledMsg).Msg)
	})

	t.Run("Search by key", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{
				TopicName: "topic1",
			},
			&kadmin.ListedTopic{},
			tabs.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		var records []kadmin.ConsumerRecord
		now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 50; i++ {
			records = append(records, kadmin.ConsumerRecord{
				Key:       fmt.Sprintf("key-%d", i),
				Payload:   serdes.DesData{},
				Err:       nil,
				Partition: int64(i),
				Offset:    int64(i),
				Headers:   nil,
				Timestamp: now.Add(time.Duration(i) * time.Second),
			})
		}
		m.Update(kadmin.ConsumerRecordReceived{
			Records: records,
		})

		m.View(tests.NewKontext(), tests.Renderer)

		m.Update(tests.Key('/'))
		m.Update(tests.Key('0'))

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "key-0")
		for i := 1; i < 10; i++ {
			assert.NotContains(t, render, fmt.Sprintf("key-%d ", i))
		}
		assert.Contains(t, render, "key-10")
		assert.Contains(t, render, "key-20")
		assert.Contains(t, render, "key-30")
		assert.Contains(t, render, "key-40")
	})
}
