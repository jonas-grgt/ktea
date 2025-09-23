package consume_form_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"ktea/ui/tabs"
	"testing"
	"time"
)

func TestConsumeForm_Navigation(t *testing.T) {

	t.Run("esc goes back to topic list page", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				Replicas:       1,
				PartitionCount: 10,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		cmd := m.Update(tests.Key(tea.KeyEsc))

		assert.IsType(t, nav.LoadTopicsPageMsg{}, cmd())
	})

	t.Run("renders all available partitions when there is height enough", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				Replicas:       1,
				PartitionCount: 10,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())

		// make sure form has been initialized
		m.View(tests.Kontext, tests.Renderer)

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "> • 0")
		for i := 1; i < 10; i++ {
			assert.Regexp(t, fmt.Sprintf("• %d", i), render)
		}
		assert.NotContains(t, render, "• 10")
	})

	t.Run("renders subset of partitions when there is not enough height", func(t *testing.T) {
		ktx := &kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    23,
			AvailableHeight: 23,
		}
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				Replicas:       1,
				PartitionCount: 100,
			},
			tabs.NewMockTopicsTabNavigator(),
			ktx)
		// make sure form has been initialized
		m.View(ktx, tests.Renderer)

		render := m.View(ktx, tests.Renderer)

		assert.Contains(t, render, `> • 0`)
		assert.Contains(t, render, `• 1`)
		assert.Contains(t, render, `• 2`)
		assert.Contains(t, render, `• 3`)
		assert.Contains(t, render, `• 4`)
		assert.NotContains(t, render, "• 5")
	})

	t.Run("load form based on previous ReadDetails", func(t *testing.T) {
		m := NewWithDetails(&kadmin.ReadDetails{
			TopicName:       "topic1",
			PartitionToRead: []int{3, 6},
			StartPoint:      kadmin.MostRecent,
			Filter: &kadmin.Filter{
				KeyFilter:       kadmin.StartsWithFilterType,
				KeySearchTerm:   "starts-with-key-term",
				ValueFilter:     kadmin.ContainsFilterType,
				ValueSearchTerm: "contains-value-term",
			},
			Limit: 500,
		}, &kadmin.ListedTopic{
			Name:           "topic1",
			PartitionCount: 10,
			Replicas:       3,
		}, nil, tests.NewKontext())

		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		render := m.View(tests.Kontext, tests.Renderer)

		assert.Contains(t, render, "> Most Recent")
		assert.NotContains(t, render, "> Beginning")
		assert.Contains(t, render, "✓ 3")
		assert.Contains(t, render, "✓ 6")
		assert.Contains(t, render, "> ")
		assert.Contains(t, render, "starts-with-key-term")
		assert.Contains(t, render, "contains-value-term")

		t.Run("no partitions selected (read from all)", func(t *testing.T) {
			m := NewWithDetails(&kadmin.ReadDetails{
				TopicName:       "topic1",
				PartitionToRead: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				StartPoint:      kadmin.MostRecent,
				Filter: &kadmin.Filter{
					KeyFilter:       kadmin.StartsWithFilterType,
					KeySearchTerm:   "starts-with-key-term",
					ValueFilter:     kadmin.ContainsFilterType,
					ValueSearchTerm: "contains-value-term",
				},
				Limit: 500,
			}, &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       3,
			}, nil, tests.NewKontext())

			// make sure form has been initialized
			m.View(tests.NewKontext(), tests.Renderer)

			render := m.View(tests.Kontext, tests.Renderer)

			for i := 0; i < 10; i++ {
				assert.NotContains(t, render, fmt.Sprintf("✓ %d", i))
			}
		})

		t.Run("Start from Absolute date was selected", func(t *testing.T) {
			now := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)
			m := NewWithDetails(&kadmin.ReadDetails{
				TopicName:       "topic1",
				PartitionToRead: []int{3, 6},
				StartPoint:      kadmin.StartPoint(now.UnixMilli()),
				Filter:          &kadmin.Filter{},
				Limit:           500,
			}, &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       3,
			}, nil, tests.NewKontext())

			// make sure form has been initialized
			m.View(tests.NewKontext(), tests.Renderer)

			render := m.View(tests.Kontext, tests.Renderer)

			assert.Contains(t, render, "> Absolute Date")
			assert.NotContains(t, render, "> Relative Date")
			assert.NotContains(t, render, "> Most Recent")
			assert.NotContains(t, render, "> Beginning")
			assert.Contains(t, render, "> 2024-06-01T")
		})

		t.Run("Start from Relative ", func(t *testing.T) {
			t.Run("Today", func(t *testing.T) {
				m := NewWithDetails(&kadmin.ReadDetails{
					TopicName:       "topic1",
					PartitionToRead: []int{3, 6},
					StartPoint:      kadmin.Today,
					Filter:          &kadmin.Filter{},
					Limit:           500,
				}, &kadmin.ListedTopic{
					Name:           "topic1",
					PartitionCount: 10,
					Replicas:       3,
				}, nil, tests.NewKontext())

				// make sure form has been initialized
				m.View(tests.NewKontext(), tests.Renderer)

				render := m.View(tests.Kontext, tests.Renderer)

				assert.Contains(t, render, "> Relative Date")
				assert.NotContains(t, render, "> Absolute Date")

				assert.Contains(t, render, "> Today")
				assert.NotContains(t, render, "> Yesterday")
				assert.NotContains(t, render, "> Week ago")

				assert.NotContains(t, render, "Absolutely Start form")
			})

			t.Run("Yesterday", func(t *testing.T) {
				m := NewWithDetails(&kadmin.ReadDetails{
					TopicName:       "topic1",
					PartitionToRead: []int{3, 6},
					StartPoint:      kadmin.Yesterday,
					Filter:          &kadmin.Filter{},
					Limit:           500,
				}, &kadmin.ListedTopic{
					Name:           "topic1",
					PartitionCount: 10,
					Replicas:       3,
				}, nil, tests.NewKontext())

				// make sure form has been initialized
				m.View(tests.NewKontext(), tests.Renderer)

				render := m.View(tests.Kontext, tests.Renderer)

				assert.Contains(t, render, "> Relative Date")
				assert.NotContains(t, render, "> Absolute Date")

				assert.NotContains(t, render, "> Today")
				assert.Contains(t, render, "> Yesterday")
				assert.NotContains(t, render, "> Week ago")

				assert.NotContains(t, render, "Absolutely Start form")
			})

			t.Run("Last 7 Days", func(t *testing.T) {
				m := NewWithDetails(&kadmin.ReadDetails{
					TopicName:       "topic1",
					PartitionToRead: []int{3, 6},
					StartPoint:      kadmin.Last7Days,
					Filter:          &kadmin.Filter{},
					Limit:           500,
				}, &kadmin.ListedTopic{
					Name:           "topic1",
					PartitionCount: 10,
					Replicas:       3,
				}, nil, tests.NewKontext())

				// make sure form has been initialized
				m.View(tests.NewKontext(), tests.Renderer)

				render := m.View(tests.Kontext, tests.Renderer)

				assert.Contains(t, render, "> Relative Date")
				assert.NotContains(t, render, "> Absolute Date")

				assert.NotContains(t, render, "> Today")
				assert.NotContains(t, render, "> Yesterday")
				assert.Contains(t, render, "> Week ago")

				assert.NotContains(t, render, "Absolutely Start form")
			})
		})

	})

	t.Run("submitting form loads consumption page with consumption information", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from most recent
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select partition 3 and 5
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeySpace))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeySpace))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())
		// no key filter
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// no value filter
		msgs := tests.Submit(m)

		assert.IsType(t, msgs[0], tabs.ToConsumePageCalledMsg{})

		assert.Equal(t, tabs.ConsumePageDetails{
			ReadDetails: kadmin.ReadDetails{
				TopicName: "topic1",
				Filter: &kadmin.Filter{
					KeySearchTerm:   "",
					ValueSearchTerm: "",
				},
				Limit:           500,
				PartitionToRead: []int{3, 5},
				StartPoint:      kadmin.MostRecent,
			},
			Topic: &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			Origin: tabs.OriginConsumeFormPage,
		}, msgs[0].(tabs.ToConsumePageCalledMsg).Details)
	})

	t.Run("selecting partitions is optional", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from most recent
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select no partitions
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())
		// no key filter
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// no value filter
		msgs := tests.Submit(m)

		assert.IsType(t, msgs[0], tabs.ToConsumePageCalledMsg{})

		assert.Equal(t, tabs.ConsumePageDetails{
			ReadDetails: kadmin.ReadDetails{
				TopicName: "topic1",
				Filter: &kadmin.Filter{
					KeySearchTerm:   "",
					ValueSearchTerm: "",
				},
				Limit:           500,
				PartitionToRead: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				StartPoint:      kadmin.MostRecent,
			},
			Topic: &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			Origin: tabs.OriginConsumeFormPage,
		}, msgs[0].(tabs.ToConsumePageCalledMsg).Details)
	})

	t.Run("selecting key filter type starts-with displays key filter value field", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from most recent
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select no partitions
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())
		// starts-with key filter
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		render := m.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, tests.Renderer)

		assert.Contains(t, render, "Key Filter Term")

		t.Run("selecting no key filter type hides key value field again", func(t *testing.T) {
			m.Update(tests.Key(tea.KeyUp))
			m.Update(tests.Key(tea.KeyUp))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:     100,
				WindowHeight:    20,
				AvailableHeight: 20,
			}, tests.Renderer)

			assert.NotContains(t, render, "Key Filter Value")
		})

		t.Run("selecting no key filter after filling in key filter term does not search for entered value", func(t *testing.T) {
			// select starts-with
			m.Update(tests.Key(tea.KeyDown))
			m.Update(tests.Key(tea.KeyDown))

			tests.UpdateKeys(m, "search-term")

			// selects none
			m.Update(tests.Key(tea.KeyUp))
			m.Update(tests.Key(tea.KeyUp))

			cmd = m.Update(tests.Key(tea.KeyEnter))
			// next field
			cmd = m.Update(cmd())
			// no value filter
			msgs := tests.Submit(m)

			assert.IsType(t, msgs[0], tabs.ToConsumePageCalledMsg{})

			assert.Equal(t, tabs.ConsumePageDetails{
				ReadDetails: kadmin.ReadDetails{
					TopicName: "topic1",
					Filter: &kadmin.Filter{
						KeySearchTerm:   "",
						ValueSearchTerm: "",
					},
					Limit:           500,
					PartitionToRead: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
					StartPoint:      kadmin.MostRecent,
				},
				Topic: &kadmin.ListedTopic{
					Name:           "topic1",
					PartitionCount: 10,
					Replicas:       1,
				},
				Origin: tabs.OriginConsumeFormPage,
			}, msgs[0].(tabs.ToConsumePageCalledMsg).Details)
		})
	})

	t.Run("filter on key value", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from most recent
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select no partitions
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())
		// starts-with key filter
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// filter on key value search-term
		tests.UpdateKeys(m, "search-term")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// no value filter
		msgs := tests.Submit(m)

		assert.IsType(t, msgs[0], tabs.ToConsumePageCalledMsg{})

		assert.EqualValues(t, tabs.ConsumePageDetails{
			ReadDetails: kadmin.ReadDetails{
				TopicName: "topic1",
				Filter: &kadmin.Filter{
					KeyFilter:       kadmin.StartsWithFilterType,
					KeySearchTerm:   "search-term",
					ValueSearchTerm: "",
				},
				Limit:           500,
				PartitionToRead: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				StartPoint:      kadmin.MostRecent,
			},
			Topic: &kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			Origin: tabs.OriginConsumeFormPage,
		}, msgs[0].(tabs.ToConsumePageCalledMsg).Details)
	})

	t.Run("selecting value filter type starts-with displays filter value field", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from most recent
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select no partitions
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())
		// no key filter
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// starts-with value filter
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		render := m.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, tests.Renderer)

		assert.Contains(t, render, "Value Filter Term")

		// make sure the value filter term field is focussed
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		m.Update(cmd())

		field := m.form.GetFocusedField()
		assert.IsType(t, &huh.Input{}, field)
		assert.Contains(t, field.View(), "Value Filter Term")

		t.Run("selecting no value filter type hides key filter value field again", func(t *testing.T) {
			cmd = m.Update(tests.Key(tea.KeyShiftTab))
			// prev field
			cmd = m.Update(cmd())

			m.Update(tests.Key(tea.KeyUp))
			m.Update(tests.Key(tea.KeyUp))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:     100,
				WindowHeight:    20,
				AvailableHeight: 20,
			}, tests.Renderer)

			assert.NotContains(t, render, "Value Filter Term")
		})

		t.Run("selecting no value filter after filling in a value filter term does not search for entered value", func(t *testing.T) {
			// select starts-with
			m.Update(tests.Key(tea.KeyDown))
			m.Update(tests.Key(tea.KeyDown))

			tests.UpdateKeys(m, "search-term")

			// selects none
			m.Update(tests.Key(tea.KeyUp))
			m.Update(tests.Key(tea.KeyUp))

			cmd = m.Update(tests.Key(tea.KeyEnter))
			// next field
			cmd = m.Update(cmd())
			// no value filter
			msgs := tests.Submit(m)

			assert.IsType(t, msgs[0], tabs.ToConsumePageCalledMsg{})

			assert.Equal(t, tabs.ConsumePageDetails{
				ReadDetails: kadmin.ReadDetails{
					TopicName: "topic1",
					Filter: &kadmin.Filter{
						KeySearchTerm:   "",
						ValueSearchTerm: "",
					},
					Limit:           500,
					PartitionToRead: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
					StartPoint:      kadmin.MostRecent,
				},
				Topic: &kadmin.ListedTopic{
					Name:           "topic1",
					PartitionCount: 10,
					Replicas:       1,
				},
				Origin: tabs.OriginConsumeFormPage,
			}, msgs[0].(tabs.ToConsumePageCalledMsg).Details)
		})
	})

	t.Run("invalidate invalid absolute start from date", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from absolute date
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		tests.UpdateKeys(m, "2024-06- 01F 12:34:56Z")
		cmd = m.Update(tests.Key(tea.KeyEnter))

		render := m.View(tests.Kontext, tests.Renderer)

		assert.Contains(t, render, "invalid date time format")
	})

	t.Run("validate valid absolute start from date", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		// select start from absolute date
		cmd := m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyDown))
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		tests.UpdateKeys(m, "2024-06-01T12:34:56Z")
		cmd = m.Update(tests.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())

		render := m.View(tests.Kontext, tests.Renderer)

		assert.NotContains(t, render, "invalid date time format")
		assert.Contains(t, render, "┃ Partitions")
	})

	t.Run("shortcuts", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())

		sc := m.Shortcuts()

		assert.Equal(t, []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Next Field", "tab"},
			{"Prev. Field", "s-tab"},
			{"Select Partition", "space"},
			{"Go Back", "esc"},
		}, sc)
	})

	t.Run("title contains topic name", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())

		assert.Equal(t, "Consume from topic1", m.Title())
	})

	t.Run("selecting Relative Date displays Relative Start from options", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		render := m.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, tests.Renderer)

		assert.NotContains(t, render, "Relatively Start from")

		// select start from Relative Date
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		render = m.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, tests.Renderer)

		assert.Contains(t, render, "Relatively Start from")

		t.Run("Deselecting hides it", func(t *testing.T) {
			m.Update(tests.Key(tea.KeyDown))

			render = m.View(&kontext.ProgramKtx{
				WindowWidth:     100,
				WindowHeight:    20,
				AvailableHeight: 20,
			}, tests.Renderer)

			assert.NotContains(t, render, "Relatively Start from")
		})
	})

	t.Run("selecting Absolute Date displays Absolute Date TextField", func(t *testing.T) {
		m := New(
			&kadmin.ListedTopic{
				Name:           "topic1",
				PartitionCount: 10,
				Replicas:       1,
			},
			tabs.NewMockTopicsTabNavigator(),
			tests.NewKontext())
		// make sure form has been initialized
		m.View(tests.NewKontext(), tests.Renderer)

		render := m.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, tests.Renderer)

		assert.NotContains(t, render, "Absolutely Start from")

		// select start from Relative Date
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))
		m.Update(tests.Key(tea.KeyDown))

		render = m.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, tests.Renderer)

		assert.Contains(t, render, "Absolutely Start from")

		t.Run("Deselecting hides it", func(t *testing.T) {
			m.Update(tests.Key(tea.KeyUp))

			render = m.View(&kontext.ProgramKtx{
				WindowWidth:     100,
				WindowHeight:    20,
				AvailableHeight: 20,
			}, tests.Renderer)

			assert.NotContains(t, render, "Absolutely Start from")
		})
	})
}
