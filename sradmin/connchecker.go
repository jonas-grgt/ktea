package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

type ConnCheckStartedMsg struct {
	ConnOk chan bool
	Err    chan error
}

func (c *ConnCheckStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.ConnOk:
		return ConnCheckSucceededMsg{}
	case err := <-c.Err:
		return ConnCheckErrMsg{err}
	}
}

func (s *DefaultSrAdmin) CheckConnection() tea.Msg {
	connOkChan := make(chan bool)
	errChan := make(chan error)

	go s.doCheckConnection(connOkChan, errChan)

	return ConnCheckStartedMsg{connOkChan, errChan}
}

func (s *DefaultSrAdmin) doCheckConnection(connOkChan chan bool, errChan chan error) {
	maybeIntroduceLatency()

	_, err := s.client.GetSubjects()
	if err != nil {
		errChan <- err
		return
	}

	connOkChan <- true
}

// CheckSchemaRegistryConn implements ConnChecker.
// CheckSchemaRegistryConn checks if the Schema Registry is reachable and returns a tea.Msg to report status.
func CheckSchemaRegistryConn(c *config.SchemaRegistryConfig) tea.Msg {
	client := New(c)
	return client.CheckConnection()
}
