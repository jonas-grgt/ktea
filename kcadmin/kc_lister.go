package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"net/http"
)

func (k *DefaultKcAdmin) ListActiveConnectors() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/connectors?expand=status", nil)

	if err != nil {
		log.Error("Error Listing Kafka Connectors", err)
		return ConnectorListingErrMsg{err}
	}

	var (
		cChan = make(chan Connectors)
		eChan = make(chan error)
	)

	go execReq(
		req,
		k.client,
		func(r Connectors) { cChan <- r },
		func(e error) { eChan <- e },
	)

	return ConnectorListingStartedMsg{cChan, eChan}
}
