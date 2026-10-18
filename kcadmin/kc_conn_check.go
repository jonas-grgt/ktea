package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"ktea/config"
	"net/http"
)

type ConnCheckStartedMsg struct {
	ConnOk chan bool
	Err    chan error
}

type ConnCheckSucceededMsg struct{}

type ConnCheckErrMsg struct {
	Err error
}

func (c *ConnCheckStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.ConnOk:
		return ConnCheckSucceededMsg{}
	case err := <-c.Err:
		return ConnCheckErrMsg{err}
	}
}

func (k *DefaultKcAdmin) CheckConnection() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		return ConnectorListingErrMsg{err}
	}

	connOkChan := make(chan bool)
	errChan := make(chan error)

	go k.doCheckConnection(connOkChan, errChan, req)

	return ConnCheckStartedMsg{connOkChan, errChan}
}

func (k *DefaultKcAdmin) doCheckConnection(connOkChan chan bool, conErrChan chan error, req *http.Request) {
	versionChan := make(chan KafkaConnectVersion)
	versionErrChan := make(chan error)

	go execReq(
		req,
		k.client,
		func(v KafkaConnectVersion) { versionChan <- v },
		func(e error) { versionErrChan <- e },
	)

	select {
	case <-versionChan:
		connOkChan <- true
	case err := <-versionErrChan:
		conErrChan <- err
	}
}

// CheckKafkaConnectClustersConn implements ConnChecker.
// CheckKafkaConnectClustersConn checks if the Kafka Connect is reachable and returns a tea.Msg to report status.
func CheckKafkaConnectClustersConn(c *config.KafkaConnectConfig) tea.Msg {
	client := New(http.DefaultClient, c)
	return client.CheckConnection()
}
