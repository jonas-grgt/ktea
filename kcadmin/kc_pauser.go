package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"fmt"
	"net/http"
)

func (k *DefaultKcAdmin) Pause(name string) tea.Msg {
	req, err := k.NewRequest(http.MethodPut, fmt.Sprintf("/connectors/%s/pause", name), nil)

	if err != nil {
		log.Error("Error Pausing "+name+" Kafka Connector", err)
		return PausingErrMsg{err}
	}

	var (
		pChan = make(chan bool)
		eChan = make(chan error)
	)

	go execReq(
		req,
		k.client,
		func(_ any) { pChan <- true },
		func(e error) { eChan <- e },
	)

	return PausingStartedMsg{pChan, eChan, name}
}
