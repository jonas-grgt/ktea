package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"net/http"
)

func (k *DefaultKcAdmin) DeleteConnector(name string) tea.Msg {
	req, err := k.NewRequest(http.MethodDelete, "/connectors/"+name, nil)
	if err != nil {
		log.Error("Error Deleting Kafka Connector", err)
		return ConnectorDeletionErrMsg{err}
	}

	var (
		dc = make(chan bool)
		ec = make(chan error)
	)

	go execReq(
		req,
		k.client,
		func(_ any) { dc <- true },
		func(e error) { ec <- e },
	)

	return ConnectorDeletionStartedMsg{name, dc, ec}
}
