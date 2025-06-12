package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"sync"
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

func (s *DefaultSrAdmin) ListVersions(subject string, versions []int) tea.Msg {
	schemaChan := make(chan Schema, len(versions))
	var wg sync.WaitGroup

	wg.Add(len(versions))
	for _, version := range versions {
		go func(version int) {
			defer wg.Done()
			schema, err := s.client.GetSchemaByVersion(subject, version)
			if err == nil {
				schemaChan <- Schema{
					Id:      strconv.Itoa(schema.ID()),
					Value:   schema.Schema(),
					Version: version,
				}
			} else {
				schemaChan <- Schema{
					Err:     err,
					Version: version,
				}
			}
		}(version)
	}

	go func() {
		wg.Wait()
		close(schemaChan)
	}()

	return SchemaListingStarted{
		schemaChan: schemaChan, versionCount: len(versions),
	}
}
