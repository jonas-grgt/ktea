package consume_page

import (
	"ktea/kadmin"
	"ktea/tests"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestSearchModalReadDetails(t *testing.T) {
	topic := &kadmin.ListedTopic{Name: "topic1", PartitionCount: 3}

	t.Run("prefills from and returns the current read details", func(t *testing.T) {
		details := kadmin.ReadDetails{
			TopicName:       "topic1",
			PartitionToRead: []int{1},
			StartPoint:      kadmin.Yesterday,
			Limit:           50,
			Filter: &kadmin.Filter{
				Term:            "abc",
				KeyFilter:       kadmin.StartsWithFilterType,
				KeySearchTerm:   "order",
				ValueFilter:     kadmin.ContainsFilterType,
				ValueSearchTerm: "paid",
			},
		}
		got, err := newSearchModal(topic, details, true).readDetails()

		assert.NoError(t, err)
		assert.Equal(t, details, got)
	})

	t.Run("no selected partitions reads all partitions", func(t *testing.T) {
		got, err := newSearchModal(topic, kadmin.NewDefaultReadDetails(topic), false).readDetails()

		assert.NoError(t, err)
		assert.Equal(t, []int{0, 1, 2}, got.PartitionToRead)
		assert.Equal(t, &kadmin.Filter{}, got.Filter)
	})

	t.Run("absolute start point", func(t *testing.T) {
		start := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		details := kadmin.NewDefaultReadDetails(topic)
		details.StartPoint = kadmin.StartPoint(start.UnixMilli())

		got, err := newSearchModal(topic, details, true).readDetails()

		assert.NoError(t, err)
		assert.Equal(t, details.StartPoint, got.StartPoint)
	})

	t.Run("invalid absolute start point blocks the search", func(t *testing.T) {
		m := newSearchModal(topic, kadmin.NewDefaultReadDetails(topic), true)
		m.values.startFrom = absoluteDate
		m.values.absoluteStartPoint = "yesterday-ish"

		_, err := m.readDetails()

		assert.Error(t, err)
	})

	t.Run("live start point is kept", func(t *testing.T) {
		details := kadmin.NewDefaultReadDetails(topic)
		details.StartPoint = kadmin.Live

		got, err := newSearchModal(topic, details, false).readDetails()

		assert.NoError(t, err)
		assert.Equal(t, kadmin.Live, got.StartPoint)
	})
}

func TestSearchModalTabs(t *testing.T) {
	topic := &kadmin.ListedTopic{Name: "topic1", PartitionCount: 3}
	m := newSearchModal(topic, kadmin.NewDefaultReadDetails(topic), true)

	// huh moves focus through a chain of commands; feed them back until settled
	press := func(msg tea.Msg) {
		msgs := []tea.Msg{msg}
		for len(msgs) > 0 {
			var next []tea.Msg
			for _, msg := range msgs {
				next = append(next, runQuick(m.Update(msg))...)
			}
			msgs = next
		}
	}

	press(tests.Key(tea.KeyTab))
	assert.Equal(t, "Key", m.tabs[m.activeTab()].name)
	assert.Contains(t, ansi.Strip(m.View(100)), "[Key]")

	press(tests.Key(tea.KeyTab))
	press(tests.Key(tea.KeyTab))
	assert.Equal(t, "Value", m.tabs[m.activeTab()].name)

	press(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.Equal(t, "Key", m.tabs[m.activeTab()].name)
}

// runQuick executes cmd, dropping commands that wait, such as cursor blinks.
func runQuick(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		if batch, ok := msg.(tea.BatchMsg); ok {
			var msgs []tea.Msg
			for _, c := range batch {
				msgs = append(msgs, runQuick(c)...)
			}
			return msgs
		}
		if msg == nil {
			return nil
		}
		return []tea.Msg{msg}
	case <-time.After(20 * time.Millisecond):
		return nil
	}
}

func TestSearchModalChangedTabs(t *testing.T) {
	topic := &kadmin.ListedTopic{Name: "topic1", PartitionCount: 3}

	t.Run("defaults mark nothing", func(t *testing.T) {
		render := ansi.Strip(newSearchModal(topic, kadmin.NewDefaultReadDetails(topic), true).View(100))
		assert.NotContains(t, render, "●")
	})

	t.Run("marks tabs that differ from the default", func(t *testing.T) {
		details := kadmin.NewDefaultReadDetails(topic)
		details.StartPoint = kadmin.Beginning
		details.PartitionToRead = []int{1}
		details.Filter = &kadmin.Filter{KeyFilter: kadmin.ContainsFilterType, KeySearchTerm: "abc"}

		render := ansi.Strip(newSearchModal(topic, details, true).View(100))
		for _, tab := range []string{"Key ●", "Start ●", "Partitions ●"} {
			assert.Contains(t, render, tab)
		}
		for _, tab := range []string{"Value ●", "Limit ●", "Key or Value ●"} {
			assert.NotContains(t, render, tab)
		}
	})

	t.Run("collapsed modal counts hidden changes", func(t *testing.T) {
		details := kadmin.NewDefaultReadDetails(topic)
		details.Limit = 50
		details.StartPoint = kadmin.Beginning

		render := ansi.Strip(newSearchModal(topic, details, false).View(100))
		assert.Contains(t, render, "ctrl+f more options (2 set)")
	})
}
