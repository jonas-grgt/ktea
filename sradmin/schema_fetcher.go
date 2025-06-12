package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
)

type SchemaFetcher interface {
	GetSchemaById(id int) tea.Msg
}

type GettingSchemaByIdMsg struct {
	SchemaChan chan Schema
	ErrChan    chan error
}

type SchemaByIdReceived struct {
	Schema Schema
}

type FailedToGetSchemaById struct {
	Err error
}

func (msg *GettingSchemaByIdMsg) AwaitCompletion() tea.Msg {
	select {
	case schema := <-msg.SchemaChan:
		return SchemaByIdReceived{
			Schema: schema,
		}
	case err := <-msg.ErrChan:
		return FailedToGetSchemaById{Err: err}
	}
}

func (s *DefaultSrAdmin) GetSchemaById(id int) tea.Msg {
	if schema, ok := s.schemaCache[id]; ok {
		return SchemaByIdReceived{Schema: schema}
	}
	schemaChan := make(chan Schema)
	errChan := make(chan error)

	go s.doGetSchemaById(id, schemaChan, errChan)
	return GettingSchemaByIdMsg{
		SchemaChan: schemaChan,
		ErrChan:    errChan,
	}
}

func (s *DefaultSrAdmin) doGetSchemaById(id int, schemaChan chan Schema, errChan chan error) {
	schema, err := s.client.GetSchema(id)
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
