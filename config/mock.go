package config

import tea "github.com/charmbracelet/bubbletea"

type MockClusterRegisterer struct {
}

type CapturedRegistrationDetails struct {
	RegistrationDetails
}

func (m MockClusterRegisterer) RegisterCluster(d RegistrationDetails) tea.Msg {
	return CapturedRegistrationDetails{d}
}
