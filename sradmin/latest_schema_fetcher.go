package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
)

type LatestSchemaBySubjectFetcher interface {
	GetLatestSchemaBySubject(subject string) tea.Msg
}

type FetchingLatestSchemaBySubjectMsg struct {
	SchemaChan chan Schema
	ErrChan    chan error
}

type LatestSchemaBySubjectReceived struct {
	Schema Schema
}

type FailedToGetLatestSchemaBySubject struct {
	Err error
}

func (msg *FetchingLatestSchemaBySubjectMsg) AwaitCompletion() tea.Msg {
	select {
	case schema := <-msg.SchemaChan:
		return LatestSchemaBySubjectReceived{
			Schema: schema,
		}
	case err := <-msg.ErrChan:
		return FailedToFetchLatestSchemaBySubject{Err: err}
	}
}

func (s *DefaultSrClient) GetLatestSchemaBySubject(subject string) tea.Msg {
	schemaChan := make(chan Schema)
	errChan := make(chan error)

	go s.doGetLatestSchema(subject, schemaChan, errChan)
	return FetchingLatestSchemaBySubjectMsg{
		SchemaChan: schemaChan,
		ErrChan:    errChan,
	}
}

func (s *DefaultSrClient) doGetLatestSchema(subject string, schemaChan chan Schema, errChan chan error) {
	schema, err := s.client.GetLatestSchema(subject)
	if err != nil {
		errChan <- err
		return
	}
	schemaChan <- Schema{
		Id:      strconv.Itoa(schema.ID()),
		Value:   schema.Schema(),
		Version: schema.Version(),
		Err:     nil,
	}
}
