package config

import tea "charm.land/bubbletea/v2"

type MockClusterRegisterer struct {
}

type CapturedRegistrationDetails struct {
	RegistrationDetails
}

func (m MockClusterRegisterer) RegisterCluster(d RegistrationDetails) tea.Msg {
	return CapturedRegistrationDetails{d}
}
