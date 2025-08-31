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

type LoadConsumptionPageMsg struct {
	ReadDetails kadmin.ReadDetails
	Topic       *kadmin.ListedTopic
}

type LoadLiveConsumePageMsg struct {
	Topic *kadmin.ListedTopic
}

type LoadCachedConsumptionPageMsg struct {
}

type LoadConsumptionFormPageMsg struct {
	Topic *kadmin.ListedTopic
	// ReadDetails is used to pre-fill the form with the provided details.
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
