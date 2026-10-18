package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"fmt"
	"net/http"
)

type ResumingErrMsg struct {
	Err error
}

type ResumingStartedMsg struct {
	Resumed chan bool
	Err     chan error
	Name    string
}

type ResumeRequestedMsg struct{}

func (c *ResumingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Resumed:
		return ResumeRequestedMsg{}
	case err := <-c.Err:
		return ResumingErrMsg{err}
	}
}

func (k *DefaultKcAdmin) Resume(name string) tea.Msg {
	req, err := k.NewRequest(http.MethodPut, fmt.Sprintf("/connectors/%s/resume", name), nil)

	if err != nil {
		log.Error("Error Resuming "+name+" Kafka Connector", err)
		return ResumingErrMsg{err}
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

	return ResumingStartedMsg{pChan, eChan, name}
}
