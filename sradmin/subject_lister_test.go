package sradmin

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"sort"
	"testing"
)

func TestListSubjectIncludingDeleted(t *testing.T) {
	var msg tea.Msg

	sra := New(&config.SchemaRegistryConfig{
		Url: fmt.Sprintf("http://localhost:%s", schemaRegistryPort.Port()),
	})

	// given
	schemas := map[string]string{
		"subject6":  `{"type":"string"}`,
		"subject7":  `{"type":"string"}`,
		"subject8":  `{"type":"string"}`,
		"subject9":  `{"type":"string"}`,
		"subject10": `{"type":"string"}`,
	}
	createSchemas(t, sra, schemas)
	// and subject 1 and 4 are deleted softly
	for _, subject := range []string{"subject6", "subject9"} {
		softDeleteSubject(t, sra, subject)
	}

	// when
	msg = sra.ListSubjects()

	// then: soft deleted are included
	assert.IsType(t, SubjectListingStartedMsg{}, msg)
	sdsMsg := msg.(SubjectListingStartedMsg)
	msg = sdsMsg.AwaitCompletion()

	assert.IsType(t, msg, SubjectsListedMsg{})
	subjects := msg.(SubjectsListedMsg).Subjects
	expected := []Subject{
		{Name: "subject6", Versions: nil, Compatibility: "BACKWARD", Deleted: true},
		{Name: "subject7", Versions: []int{1}, Compatibility: "BACKWARD", Deleted: false},
		{Name: "subject8", Versions: []int{1}, Compatibility: "BACKWARD", Deleted: false},
		{Name: "subject9", Versions: nil, Compatibility: "BACKWARD", Deleted: true},
		{Name: "subject10", Versions: []int{1}, Compatibility: "BACKWARD", Deleted: false},
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})
	sort.Slice(subjects, func(i, j int) bool {
		return subjects[i].Name < subjects[j].Name
	})
	assert.Equal(t, expected, subjects)
}

func softDeleteSubject(
	t *testing.T,
	sra *DefaultSrClient,
	subject string,
) {
	msg := sra.SoftDeleteSubject(subject)
	assert.IsType(t, msg, SubjectDeletionStartedMsg{})
	sdsMsg := msg.(SubjectDeletionStartedMsg)
	msg = sdsMsg.AwaitCompletion()

	assert.IsType(t, msg, SubjectDeletedMsg{})
	assert.Equal(t, SubjectDeletedMsg{subject}, msg)
}

func createSchemas(
	t *testing.T,
	sra *DefaultSrClient,
	schemas map[string]string,
) {
	for subject, schema := range schemas {
		msg := sra.CreateSchema(SubjectCreationDetails{
			Subject: subject,
			Schema:  schema,
		})

		assert.IsType(t, msg, SchemaCreationStartedMsg{})
		scsMsg := msg.(SchemaCreationStartedMsg)
		msg = scsMsg.AwaitCompletion()
		assert.IsType(t, SchemaCreatedMsg{}, msg)
	}
}
