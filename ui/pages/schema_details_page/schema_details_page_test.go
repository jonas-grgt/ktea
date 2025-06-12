package schema_details_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/tests"
	"ktea/ui/pages/nav"
	"testing"
)

type MockSchemaLister struct{}

func (m *MockSchemaLister) ListVersions(subject string, versions []int) tea.Msg {
	return nil
}

func TestSchemaDetailsPage(t *testing.T) {

	t.Run("When schemas not loaded yet", func(t *testing.T) {

		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		t.Run("viewport ignores msgs", func(t *testing.T) {
			assert.Nil(t, page.vp)
			assert.Nil(t, page.schemas)

			page.Update(nav.LoadSchemaDetailsPageMsg{})

			assert.Nil(t, page.vp)
		})

		t.Run("Title returns empty string", func(t *testing.T) {
			assert.Nil(t, page.vp)
			assert.Nil(t, page.schemas)

			title := page.Title()

			assert.Empty(t, title)
		})
	})

	t.Run("Title contains subject and version", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{
			Name:     "subject-name",
			Versions: nil,
		})

		page.Update(sradmin.SchemasListed{
			Schemas: []sradmin.Schema{
				{
					Id:      "123",
					Value:   "{\"type\":\"string\"}",
					Version: 1,
					Err:     nil,
				},
			},
		})

		title := page.Title()

		assert.Equal(t, "Subjects / subject-name / Versions / 1", title)
	})

	t.Run("Loading indicator", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		t.Run("visible when fetching schemas", func(t *testing.T) {
			page.Update(sradmin.SchemaListingStarted{})

			render := page.View(tests.NewKontext(), tests.TestRenderer)

			assert.Contains(t, render, "Loading schema")
		})

		t.Run("hidden when schemas are fetched", func(t *testing.T) {
			page.Update(sradmin.SchemasListed{
				Schemas: []sradmin.Schema{
					{
						Id:      "123",
						Value:   "{\"type\":\"string\"}",
						Version: 1,
						Err:     nil,
					},
				},
			})

			render := page.View(tests.NewKontext(), tests.TestRenderer)

			assert.NotContains(t, render, "Loading schema")
		})
	})

	t.Run("esc goes back to subjects list", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		cmds := page.Update(tests.Key(tea.KeyEsc))

		msgs := tests.ExecuteBatchCmd(cmds)

		assert.Contains(t, msgs, nav.LoadSubjectsPageMsg{})
	})

	t.Run("Render single schema formatted", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		page.Update(sradmin.SchemasListed{
			Schemas: []sradmin.Schema{
				{
					Id:      "123",
					Value:   "{\"type\":\"string\"}",
					Version: 1,
					Err:     nil,
				},
			},
		})

		render := ansi.Strip(page.View(tests.NewKontext(), tests.TestRenderer))

		assert.Regexp(t, "│\\W+\"type\": \"string\"\\W+│\n│ }", render)
	})

	t.Run("Multiple versions", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		page.Update(sradmin.SchemasListed{
			Schemas: []sradmin.Schema{
				{
					Id:      "111",
					Value:   "{\"type\":\"string\"}",
					Version: 1,
					Err:     nil,
				},
				{
					Id:      "222",
					Value:   "{\"type\":\"string\"}",
					Version: 2,
					Err:     nil,
				},
				{
					Id:      "333",
					Value:   "{\"type\":\"string\"}",
					Version: 3,
					Err:     nil,
				},
			},
		})

		render := ansi.Strip(page.View(tests.NewKontext(), tests.TestRenderer))

		assert.Regexp(t, "1\\W+2\\W+«3»", render)
	})

	t.Run("Renders version IDs", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		page.Update(sradmin.SchemasListed{
			Schemas: []sradmin.Schema{
				{
					Id:      "111",
					Value:   "{\"type\":\"string\"}",
					Version: 1,
					Err:     nil,
				},
				{
					Id:      "222",
					Value:   "{\"type\":\"string\"}",
					Version: 2,
					Err:     nil,
				},
				{
					Id:      "333",
					Value:   "{\"type\":\"string\"}",
					Version: 3,
					Err:     nil,
				},
			},
		})

		render := ansi.Strip(page.View(tests.NewKontext(), tests.TestRenderer))

		assert.Regexp(t, "ID\\W+: 333", render)

		page.Update(tests.Key(tea.KeyLeft))
		page.Update(tests.Key(tea.KeyLeft))
		page.Update(tests.Key(tea.KeyEnter))

		render = ansi.Strip(page.View(tests.NewKontext(), tests.TestRenderer))

		assert.Regexp(t, "ID\\W+: 111", render)
	})

	t.Run("schema view is scrollable", func(t *testing.T) {
		page, _ := New(&MockSchemaLister{}, sradmin.Subject{})

		page.Update(sradmin.SchemasListed{
			Schemas: []sradmin.Schema{
				{
					Id:      "123",
					Value:   "{\n  \"type\": \"record\",\n  \"name\": \"UserProfile\",\n  \"namespace\": \"com.example.avro\",\n  \"fields\": [\n    {\n      \"name\": \"userId\",\n      \"type\": \"string\",\n      \"doc\": \"Unique identifier for the user\"\n    },\n    {\n      \"name\": \"firstName\",\n      \"type\": [\"null\", \"string\"],\n      \"default\": null,\n      \"doc\": \"The user's first name, optional\"\n    },\n    {\n      \"name\": \"lastName\",\n      \"type\": [\"null\", \"string\"],\n      \"default\": null,\n      \"doc\": \"The user's last name, optional\"\n    },\n    {\n      \"name\": \"email\",\n      \"type\": \"string\",\n      \"doc\": \"Email address of the user\"\n    },\n    {\n      \"name\": \"age\",\n      \"type\": [\"null\", \"int\"],\n      \"default\": null,\n      \"doc\": \"Age of the user, optional\"\n    },\n    {\n      \"name\": \"isActive\",\n      \"type\": \"boolean\",\n      \"doc\": \"Indicates if the user is active\"\n    },\n    {\n      \"name\": \"signupDate\",\n      \"type\": {\n        \"type\": \"long\",\n        \"logicalType\": \"timestamp-millis\"\n      },\n      \"doc\": \"Timestamp of when the user signed up\"\n    }\n  ],\n  \"doc\": \"Schema for storing user profile data\"\n}\n",
					Version: 1,
					Err:     nil,
				},
			},
		})

		render := ansi.Strip(page.View(&kontext.ProgramKtx{
			Config:          nil,
			WindowWidth:     100,
			WindowHeight:    25,
			AvailableHeight: 8,
		}, tests.TestRenderer))

		assert.Regexp(t, "│ {\\W+│\n│\\W+\"type\": \"record\",", render)
		assert.NotContains(t, render, "userId")

		page.Update(tests.Key(tea.KeyDown))
		page.Update(tests.Key(tea.KeyDown))
		page.Update(tests.Key(tea.KeyDown))

		render = ansi.Strip(page.View(&kontext.ProgramKtx{
			Config:          nil,
			WindowWidth:     100,
			WindowHeight:    30,
			AvailableHeight: 9,
		}, tests.TestRenderer))

		assert.NotRegexp(t, "│ {\\W+│\n│\\W+\"type\": \"record\",", render)
		assert.Contains(t, render, "userId")
	})
}
