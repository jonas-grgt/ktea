package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kadmin"
)

type TopicsTabNavigator interface {
	ToConsumeFormPage(ConsumeFormPageDetails) tea.Cmd

	ToConsumePage(ConsumePageDetails) tea.Cmd

	ToTopicsPage() tea.Cmd

	ToRecordDetailsPage(LoadRecordDetailPageMsg) tea.Cmd
}

type ConsumeFormPageDetails struct {
	Topic *kadmin.ListedTopic
	// ReadDetails is used to pre-fill the consume form with the provided - previous - details.
	ReadDetails *kadmin.ReadDetails
}

type Origin int

const (
	OriginTopicsPage Origin = iota
	OriginConsumeFormPage
)

type ConsumePageDetails struct {
	Origin      Origin
	ReadDetails kadmin.ReadDetails
	Topic       *kadmin.ListedTopic
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
