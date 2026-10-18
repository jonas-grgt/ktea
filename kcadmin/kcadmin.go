package kcadmin

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"encoding/json"
	"fmt"
	"io"
	"ktea/config"
	"ktea/kadmin"
	"net/http"
)

type Admin interface {
	ConnectorLister
	ConnectorDeleter
	VersionLister
	Pauser
	Resumer
}

// ConnectorLister defines the behavior of listing active Kafka connectors.
// returns a tea.Msg that can be either a ConnectorListingStartedMsg or a ConnectorListingErrMsg
type ConnectorLister interface {
	ListActiveConnectors() tea.Msg
}

type ConnectorDeleter interface {
	DeleteConnector(name string) tea.Msg
}

// VersionLister defines the behavior of listing the kafka connect version
// return a tea.Msg that can either be a VersionListingStartedMsg or a VersionListingErrMsg
type VersionLister interface {
	ListVersion() tea.Msg
}

// Pauser Pauses the connector and its tasks by its name
// return a tea.Msg that can either be a PausingStartedMsg or a PausingErrMsg
type Pauser interface {
	Pause(name string) tea.Msg
}

type Resumer interface {
	Resume(name string) tea.Msg
}

// ConnChecker is a function that checks a Kafka Connect Cluster connection and returns a tea.Msg.
type ConnChecker func(c *config.KafkaConnectConfig) tea.Msg

type ConnectorStatus struct {
	Name      string         `json:"name"`
	Connector ConnectorState `json:"connector"`
	Tasks     []TaskState    `json:"tasks"`
	Type      string         `json:"type"`
}

type ConnectorState struct {
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
}

type TaskState struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
}

type Connectors map[string]struct {
	Status ConnectorStatus `json:"status"`
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultKcAdmin struct {
	client   Client
	baseUrl  string
	username *string
	password *string
}

type ConnectorListingStartedMsg struct {
	Connectors chan Connectors
	Err        chan error
}

type ConnectorsListedMsg struct {
	Connectors
}

type KafkaConnectVersion struct {
	Version   string `json:"version"`
	ClusterId string `json:"kafka_cluster_id"`
}

type VersionListingStartedMsg struct {
	Version chan KafkaConnectVersion
	Err     chan error
}

type VersionListingErrMsg struct {
	Err error
}

type PausingStartedMsg struct {
	Paused chan bool
	Err    chan error
	Name   string
}

type PauseRequestedMsg struct {
	Name string
}

func (c *PausingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Paused:
		return PauseRequestedMsg{c.Name}
	case err := <-c.Err:
		return PausingErrMsg{err}
	}
}

type PausingErrMsg struct {
	Err error
}

func (c *ConnectorListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case con := <-c.Connectors:
		return ConnectorsListedMsg{con}
	case err := <-c.Err:
		return ConnectorListingErrMsg{err}
	}
}

type ConnectorListingErrMsg struct {
	Err error
}

type ConnectorDeletionStartedMsg struct {
	Name    string
	Deleted chan bool
	Err     chan error
}

func (m *ConnectorDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-m.Deleted:
		return ConnectorDeletedMsg{m.Name}
	case err := <-m.Err:
		return ConnectorDeletionErrMsg{err}
	}
}

type ConnectorDeletedMsg struct {
	Name string
}

type ConnectorDeletionErrMsg struct {
	Err error
}

func (k *DefaultKcAdmin) url(path string) string {
	return k.baseUrl + path
}

func (k *DefaultKcAdmin) NewRequest(
	method string,
	path string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequest(method, k.url(path), body)

	if err != nil {
		return nil, err
	}

	if k.password != nil && k.username != nil {
		req.SetBasicAuth(*k.username, *k.password)
	}

	return req, nil
}

type successFunc[T any] func(body T)

type errorFunc func(err error)

func execReq[T any](
	req *http.Request,
	client Client,
	sf successFunc[T],
	ef errorFunc,
) {

	kadmin.MaybeIntroduceLatency()

	log.Debug("Executing request", "request", req)

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error during request", "error", err)
		ef(err)
		return
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				ef(err)
			}
		}(resp.Body)

		if resp.ContentLength == 0 || resp.StatusCode == http.StatusNoContent {
			log.Info("Executed Request Successfully without content")
			var res T
			sf(res)
			return
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error Reading Response Body", "error", err)
			ef(err)
			return
		}

		var res T
		if err := json.Unmarshal(b, &res); err != nil {
			log.Error("Error Unmarshalling", "error", err)
			ef(err)
			return
		}

		log.Info("Request executed successfully", "statusCode", resp.StatusCode)

		sf(res)
	} else {
		log.Error("Error", "statusCode", resp.StatusCode)
		ef(fmt.Errorf("Error unexpected response code (%d)", resp.StatusCode))
	}
}

func New(c Client, config *config.KafkaConnectConfig) *DefaultKcAdmin {
	return &DefaultKcAdmin{client: c, baseUrl: config.Url, username: config.Username, password: config.Password}
}
