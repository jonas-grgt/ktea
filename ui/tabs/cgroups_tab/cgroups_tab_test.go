package cgroups_tab

import (
	"ktea/config"
	"ktea/kadmin"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

type MockConsumerGroupOffsetLister struct{}

func (m *MockConsumerGroupOffsetLister) ListOffsets(_ string) tea.Msg {
	return nil
}

type MockConsumerGroupLister struct{}

func (m *MockConsumerGroupLister) ListCGroups() tea.Msg {
	return nil
}

type MockConsumerGroupDeleter struct{}

func (m *MockConsumerGroupDeleter) DeleteCGroup(name string) tea.Msg {
	return nil
}

func TestGroupsTab(t *testing.T) {
	t.Run("List consumer groups", func(t *testing.T) {
		groupsTab, _ := New(&MockConsumerGroupLister{}, &MockConsumerGroupDeleter{}, &MockConsumerGroupOffsetLister{}, statusbar.New())

		groupsTab.Update(kadmin.ConsumerGroupsListedMsg{
			ConsumerGroups: []*kadmin.ConsumerGroup{
				{
					Name: "Group1",
					Members: []kadmin.GroupMember{
						{
							MemberId:   "Group1Id1",
							ClientId:   "Group1ClientId1",
							ClientHost: "127.0.0.1",
						},
					},
				},
				{
					Name:    "Group2",
					Members: nil,
				},
			},
		})

		ktx := tests.NewKontext(tests.WithConfig(&config.Config{
			Clusters: []config.Cluster{
				{
					Name:             "PRD",
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig: config.SASLConfig{
						AuthMethod: config.AuthMethodNone,
					},
				},
			},
		}))
		render := ansi.Strip(groupsTab.View(ktx, tests.Renderer))

		assert.Contains(t, render, "Group1")
		assert.Contains(t, render, "Group2")

		t.Run("Refresh resets table", func(t *testing.T) {
			groupsTab.Update(kadmin.ConsumerGroupsListedMsg{
				ConsumerGroups: []*kadmin.ConsumerGroup{
					{
						Name: "Group1",
						Members: []kadmin.GroupMember{
							{
								MemberId:   "Group1Id1",
								ClientId:   "Group1ClientId1",
								ClientHost: "127.0.0.1",
							},
						},
					},
					{
						Name:    "Group2",
						Members: nil,
					},
				},
			})

			ktx := tests.NewKontext(tests.WithConfig(&config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig: config.SASLConfig{
							AuthMethod: config.AuthMethodNone,
						},
					},
				},
			}))
			render = ansi.Strip(groupsTab.View(ktx, tests.Renderer))

			g1Count := strings.Count(render, "Group1")
			g2Count := strings.Count(render, "Group2")

			assert.Equal(t, 1, g1Count)
			assert.Equal(t, 1, g2Count)
		})
	})

}
