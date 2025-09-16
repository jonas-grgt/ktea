package nav

import (
	"ktea/kadmin"
	"ktea/sradmin"
)

type LoadTopicsPageMsg struct {
	Refresh bool
}

type LoadCreateTopicPageMsg struct{}

type LoadTopicConfigPageMsg struct{}

type LoadPublishPageMsg struct {
	Topic *kadmin.ListedTopic
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

type LoadLiveConsumePageMsg struct {
	Topic *kadmin.ListedTopic
}

type LoadCachedConsumptionPageMsg struct {
}

type ConsumeFormPageDetails struct {
	Topic *kadmin.ListedTopic
	// ReadDetails is used to pre-fill the consume form with the provided - previous - details.
	ReadDetails *kadmin.ReadDetails
}

type LoadRecordDetailPageMsg struct {
	Record    *kadmin.ConsumerRecord
	TopicName string
}

type LoadCGroupsPageMsg struct {
}

type LoadCGroupTopicsPageMsg struct {
	GroupName string
}

type LoadCreateSubjectPageMsg struct{}

type LoadSubjectsPageMsg struct {
	Refresh bool
}

type LoadSchemaDetailsPageMsg struct {
	Subject sradmin.Subject
}
