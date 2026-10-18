package create_cluster_page

import (
	tea "charm.land/bubbletea/v2"
)

type mockKafkaConnectRegistered struct {
}

func mockKafkaConnectRegisterer() tea.Msg {
	return mockKafkaConnectRegistered{}
}
