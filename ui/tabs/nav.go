package tabs

import (
	tea "charm.land/bubbletea/v2"
	"ktea/kadmin"
)

type TopicsTabNavigator interface {
	ToConsumePage(ConsumePageDetails) tea.Cmd

	ToTopicsPage() tea.Cmd

	ToRecordDetailsPage(LoadRecordDetailPageMsg) tea.Cmd
}

type ConsumePageDetails struct {
	ReadDetails kadmin.ReadDetails
	Topic       *kadmin.ListedTopic
	// OpenSearch opens the expanded search and only starts reading once it is submitted.
	OpenSearch bool
}

type LoadRecordDetailPageMsg struct {
	Record    *kadmin.ConsumerRecord
	TopicName string
	Records   []kadmin.ConsumerRecord
	Index     int
}

type ClustersTabNavigator interface {
	ToClustersPage() tea.Cmd
}

type KConTabNavigator interface {
	ToKConsPage() tea.Cmd
}
