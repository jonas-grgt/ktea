package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"net/http"
)

func (k *DefaultKcAdmin) ListVersion() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		return VersionListingErrMsg{err}
	}

	versionChan := make(chan KafkaConnectVersion)
	errChan := make(chan error)

	go execReq(
		req,
		k.client,
		func(r KafkaConnectVersion) { versionChan <- r },
		func(e error) { errChan <- e },
	)

	return VersionListingStartedMsg{versionChan, errChan}
}
