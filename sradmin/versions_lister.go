package sradmin

import (
	"strconv"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Schema struct {
	Id      string
	Value   string
	Version int
	Err     error
}

type SchemasListed struct {
	Schemas []Schema
}

type SchemaListingStarted struct {
	schemaChan   chan Schema
	versionCount int
}

func (s *SchemaListingStarted) AwaitCompletion() tea.Msg {
	var schemas []Schema
	count := 0

	for count < s.versionCount {
		select {
		case schema, ok := <-s.schemaChan:
			if !ok {
				// Channel is closed; exit loop
				return SchemasListed{Schemas: schemas}
			}
			schemas = append(schemas, schema)
			count++
		}
	}

	return SchemasListed{Schemas: schemas}
}

func (s *DefaultSrClient) ListVersions(subject string, versions []int) tea.Msg {
	schemaChan := make(chan Schema, len(versions))
	var wg sync.WaitGroup

	for _, version := range versions {
		v := version
		wg.Go(func() {
			schema, err := s.client.GetSchemaByVersion(subject, v)
			if err == nil {
				schemaChan <- Schema{
					Id:      strconv.Itoa(schema.ID()),
					Value:   schema.Schema(),
					Version: version,
				}
			} else {
				schemaChan <- Schema{
					Err:     err,
					Version: v,
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(schemaChan)
	}()

	return SchemaListingStarted{
		schemaChan: schemaChan, versionCount: len(versions),
	}
}
