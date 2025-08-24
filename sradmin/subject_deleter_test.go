package sradmin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"testing"
)

func TestHardSubjectDelete(t *testing.T) {

	sra := New(&config.SchemaRegistryConfig{
		Url: fmt.Sprintf("http://localhost:%s", schemaRegistryPort.Port()),
	})

	t.Run("Delete Subject Permanently", func(t *testing.T) {
		// given a soft deleted subject
		msg := sra.CreateSchema(SubjectCreationDetails{
			Subject: "test.hard.subject",
			Schema:  `{"type":"string"}`,
		})

		assert.IsType(t, msg, SchemaCreationStartedMsg{})
		scsMsg := msg.(SchemaCreationStartedMsg)
		msg = scsMsg.AwaitCompletion()
		assert.IsType(t, SchemaCreatedMsg{}, msg)

		msg = sra.SoftDeleteSubject("test.hard.subject")

		assert.IsType(t, msg, SubjectDeletionStartedMsg{})
		sdsMsg := msg.(SubjectDeletionStartedMsg)
		msg = sdsMsg.AwaitCompletion()

		assert.IsType(t, msg, SubjectDeletedMsg{})
		assert.Equal(t, SubjectDeletedMsg{"test.hard.subject"}, msg)

		// when
		msg = sra.HardDeleteSubject("test.hard.subject")

		// then
		assert.IsType(t, msg, SubjectDeletionStartedMsg{})
		sdsMsg = msg.(SubjectDeletionStartedMsg)
		msg = sdsMsg.AwaitCompletion()

		assert.IsType(t, msg, SubjectDeletedMsg{})
		assert.Equal(t, SubjectDeletedMsg{"test.hard.subject"}, msg)
	})
}

func TestSoftSubjectDelete(t *testing.T) {

	sra := New(&config.SchemaRegistryConfig{
		Url: fmt.Sprintf("http://localhost:%s", schemaRegistryPort.Port()),
	})

	t.Run("Delete Subject Permanently", func(t *testing.T) {
		// given
		msg := sra.CreateSchema(SubjectCreationDetails{
			Subject: "test.soft.subject",
			Schema:  `{"type":"string"}`,
		})

		assert.IsType(t, msg, SchemaCreationStartedMsg{})
		scsMsg := msg.(SchemaCreationStartedMsg)
		msg = scsMsg.AwaitCompletion()
		assert.IsType(t, SchemaCreatedMsg{}, msg)

		// when
		msg = sra.SoftDeleteSubject("test.soft.subject")

		// then
		assert.IsType(t, msg, SubjectDeletionStartedMsg{})
		sdsMsg := msg.(SubjectDeletionStartedMsg)
		msg = sdsMsg.AwaitCompletion()

		assert.IsType(t, msg, SubjectDeletedMsg{})
		assert.Equal(t, SubjectDeletedMsg{"test.soft.subject"}, msg)

		// clean-up
		msg = sra.HardDeleteSubject("test.soft.subject")

		// then
		assert.IsType(t, msg, SubjectDeletionStartedMsg{})
		sdsMsg = msg.(SubjectDeletionStartedMsg)
		msg = sdsMsg.AwaitCompletion()

		assert.IsType(t, msg, SubjectDeletedMsg{})
		assert.Equal(t, SubjectDeletedMsg{"test.soft.subject"}, msg)
	})
}
