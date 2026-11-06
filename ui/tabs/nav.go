package tabs

import (
	tea "charm.land/bubbletea/v2"
	"ktea/kadmin"
	"ktea/sradmin"
)

type TopicsTabNavigator interface {
	// ToTopicsPage activates the topics page, reloading the topics when refresh is true.
	ToTopicsPage(refresh bool) tea.Cmd

	ToCreateTopicPage() tea.Cmd

	ToTopicConfigPage(topic string) tea.Cmd

	ToPublishPage(topic *kadmin.ListedTopic) tea.Cmd

	ToConsumePage(ConsumePageDetails) tea.Cmd

	// ToCachedConsumptionPage returns to the last opened consume page without re-reading.
	ToCachedConsumptionPage() tea.Cmd

	ToRecordDetailsPage(RecordDetailsPageDetails) tea.Cmd
}

type ConsumePageDetails struct {
	ReadDetails kadmin.ReadDetails
	Topic       *kadmin.ListedTopic
	// OpenSearch opens the expanded search and only starts reading once it is submitted.
	OpenSearch bool
}

type RecordDetailsPageDetails struct {
	Record    *kadmin.ConsumerRecord
	TopicName string
	Records   []kadmin.ConsumerRecord
	Index     int
}

type CGroupsTabNavigator interface {
	ToCGroupsPage() tea.Cmd

	ToCGroupTopicsPage(group string) tea.Cmd
}

type SrTabNavigator interface {
	// ToSubjectsPage activates the subjects page, reloading the subjects when refresh is true.
	ToSubjectsPage(refresh bool) tea.Cmd

	ToCreateSubjectPage() tea.Cmd

	ToSchemaDetailsPage(subject sradmin.Subject) tea.Cmd
}

type ClustersTabNavigator interface {
	ToClustersPage() tea.Cmd
}

type KConTabNavigator interface {
	ToKConsPage() tea.Cmd
}
