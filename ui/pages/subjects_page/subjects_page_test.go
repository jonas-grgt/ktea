package subjects_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/sradmin"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestSubjectsPage(t *testing.T) {

	t.Run("No subjects found", func(t *testing.T) {

		subjectsPage, _ := New(sradmin.NewMock())

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: []sradmin.Subject{}})

		render := subjectsPage.View(tests.Kontext, tests.Renderer)

		assert.Contains(t, render, "No Subjects Found")

		t.Run("enter is ignored", func(t *testing.T) {

			cmd := subjectsPage.Update(tests.Key(tea.KeyEnter))

			assert.Nil(t, cmd)
		})
	})

	t.Run("List subjects and number of versions", func(t *testing.T) {

		subjectsPage, _ := New(sradmin.NewMock())

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 20; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
					Deleted:  i%2 != 0,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(tests.Kontext, tests.Renderer)

		for i := 0; i < 20; i++ {
			if i%2 == 0 {
				assert.Regexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
			} else {
				assert.NotRegexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
			}
		}
	})

	t.Run("When subjects are loaded or refresh then the search form is reset", func(t *testing.T) {

		subjectsPage, _ := New(sradmin.NewMock())

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		subjectsPage.Update(tests.Key('/'))
		tests.UpdateKeys(subjectsPage, "subject2")

		render := subjectsPage.View(tests.NewKontext(), tests.Renderer)
		assert.Contains(t, render, "> subject2")

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render = subjectsPage.View(tests.NewKontext(), tests.Renderer)
		assert.NotContains(t, render, "> subject")
	})

	t.Run("Searching resets selected row to top row", func(t *testing.T) {
		page, _ := New(sradmin.NewMock())

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:          fmt.Sprintf("subject%d", i),
					Versions:      versions,
					Deleted:       false,
					Compatibility: "BACKWARD",
				})
		}
		page.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		page.View(tests.NewKontext(), tests.Renderer)

		page.Update(tests.Key(tea.KeyDown))
		page.Update(tests.Key(tea.KeyDown))

		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, "subject2", page.table.SelectedRow()[0])

		page.Update(tests.Key('/'))
		tests.UpdateKeys(page, "subject")

		page.View(tests.NewKontext(), tests.Renderer)
		assert.Equal(t, "subject0", page.table.SelectedRow()[0])
	})

	t.Run("Enter opens schema detail page", func(t *testing.T) {

		subjectsPage, _ := New(sradmin.NewMock())

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})
		// init table
		subjectsPage.View(tests.NewKontext(), tests.Renderer)

		subjectsPage.Update(tests.Key(tea.KeyDown))
		subjectsPage.Update(tests.Key(tea.KeyDown))
		cmd := subjectsPage.Update(tests.Key(tea.KeyEnter))

		assert.Equal(t, nav.LoadSchemaDetailsPageMsg{
			Subject: sradmin.Subject{
				Name:     "subject2",
				Versions: []int{0, 1, 2},
			},
		}, cmd())
	})

	t.Run("Order subjects default by Subject Name Asc", func(t *testing.T) {
		subjectsPage, _ := New(sradmin.NewMock())

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 100; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		shuffle(subjects)

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "▲ Subject Name")

		subject1Idx := strings.Index(render, "subject1")
		subject2Idx := strings.Index(render, "subject2")
		subject50Idx := strings.Index(render, "subject50")
		subject88Idx := strings.Index(render, "subject88")
		assert.Less(t, subject1Idx, subject2Idx, "subject2 came before subject1")
		assert.Less(t, subject2Idx, subject50Idx, "subject50 came before subject2")
		assert.Less(t, subject50Idx, subject88Idx, "subject88 came before subject50")
	})

	t.Run("Sorting", func(t *testing.T) {
		subjectsPage, _ := New(sradmin.NewMock())

		var (
			subjects      []sradmin.Subject
			versions      []int
			compatibility string
		)
		for i := 0; i < 100; i++ {
			versions = append(versions, i)
			if i%2 == 0 {
				compatibility = "BACKWARD"
			} else {
				compatibility = "FORWARD_TRANSITIVE"
			}
			subjects = append(subjects,
				sradmin.Subject{
					Name:          fmt.Sprintf("subject%d", i),
					Versions:      versions,
					Compatibility: compatibility,
				})
		}
		shuffle(subjects)

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})
		subjectsPage.View(tests.NewKontext(), tests.Renderer)

		t.Run("by Subject Name", func(t *testing.T) {

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render := subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Contains(t, render, "▼ Subject Name")

			subject1Idx := strings.Index(render, "subject1")
			subject2Idx := strings.Index(render, "subject2")
			subject50Idx := strings.Index(render, "subject50")
			subject88Idx := strings.Index(render, "subject88")

			assert.Less(t, subject88Idx, subject1Idx, "subject1 came before subject88")
			assert.Less(t, subject88Idx, subject2Idx, "subject2 came before subject88")
			assert.Less(t, subject88Idx, subject50Idx, "subject50 came before subject88")

			assert.Less(t, subject50Idx, subject1Idx, "subject1 came before subject50")
			assert.Less(t, subject50Idx, subject2Idx, "subject2 came before subject50")

			assert.Less(t, subject2Idx, subject1Idx, "subject1 came before subject2")

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Contains(t, render, "▲ Subject Name")

			subject1Idx = strings.Index(render, "subject1")
			subject2Idx = strings.Index(render, "subject2")
			subject50Idx = strings.Index(render, "subject50")
			subject88Idx = strings.Index(render, "subject88")

			assert.Greater(t, subject88Idx, subject1Idx, "subject1 came before subject88")
			assert.Greater(t, subject88Idx, subject2Idx, "subject2 came before subject88")
			assert.Greater(t, subject88Idx, subject50Idx, "subject50 came before subject88")

			assert.Greater(t, subject50Idx, subject1Idx, "subject1 came before subject50")
			assert.Greater(t, subject50Idx, subject2Idx, "subject2 came before subject50")

			assert.Greater(t, subject2Idx, subject1Idx, "subject1 came before subject2")

		})

		t.Run("by Compatibility", func(t *testing.T) {

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyRight))
			subjectsPage.Update(tests.Key(tea.KeyRight))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render := subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Contains(t, render, "▲ Comp")

			lastBackwardIdx := strings.LastIndex(render, "BACKWARD")
			lastForwardIdx := strings.Index(render, "FORWARD")

			assert.Less(t, lastBackwardIdx, lastForwardIdx, "FORWARD came before BACKWARD")

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Contains(t, render, "▼ Comp")

			lastBackwardIdx = strings.LastIndex(render, "BACKWARD")
			lastForwardIdx = strings.Index(render, "FORWARD")

			assert.Greater(t, lastBackwardIdx, lastForwardIdx, "BACKWARD came before FORWARD")
		})

		t.Run("by Versions", func(t *testing.T) {
			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyLeft))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render := subjectsPage.View(tests.NewKontext(), tests.Renderer)
			assert.Contains(t, render, "▼ V")

			subject1Idx := strings.Index(render, "subject1")
			subject88Idx := strings.Index(render, "subject2")

			assert.Greater(t, subject1Idx, subject88Idx, "subject1 came not before subject2")

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)
			assert.Contains(t, render, "▲ V")

			subject1Idx = strings.Index(render, "subject1")
			subject88Idx = strings.Index(render, "subject2")

			assert.Less(t, subject1Idx, subject88Idx, "subject2 came not before subject1")
		})
	})

	t.Run("Loading details page (enter) after searching from selective list", func(t *testing.T) {
		subjectsPage, _ := New(sradmin.NewMock())

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 100; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:          fmt.Sprintf("subject%d", i),
					Versions:      versions,
					Deleted:       false,
					Compatibility: "BACKWARD",
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})
		subjectsPage.View(tests.NewKontext(), tests.Renderer)

		subjectsPage.Update(tests.Key('/'))
		tests.UpdateKeys(subjectsPage, "1")
		subjectsPage.Update(tests.Key(tea.KeyEnter))
		subjectsPage.Update(tests.Key(tea.KeyDown))
		subjectsPage.Update(tests.Key(tea.KeyDown))
		cmd := subjectsPage.Update(tests.Key(tea.KeyEnter))

		msgs := tests.ExecuteBatchCmd(cmd)

		assert.Contains(t, msgs, nav.LoadSchemaDetailsPageMsg{
			Subject: sradmin.Subject{
				Name:          "subject10",
				Versions:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				Deleted:       false,
				Compatibility: "BACKWARD",
			},
		})
	})

	t.Run("Delete subject", func(t *testing.T) {
		subjectsPage, _ := New(
			sradmin.NewMock(
				sradmin.WithHardDeleteSubjectCallbackFn(
					func(subject string) tea.Msg {
						return sradmin.HardDeletedSubjectMsg{Subject: subject}
					},
				),
				sradmin.WithSoftDeleteSubjectCallbackFn(
					func(subject string) tea.Msg {
						return sradmin.SoftDeletedSubjectMsg{Subject: subject}

					},
				),
			),
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 100; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		// render so the table's first row is selected
		render := subjectsPage.View(tests.NewKontext(), tests.Renderer)
		assert.NotRegexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)

		t.Run("F2 triggers subject soft delete", func(t *testing.T) {
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Regexp(t, "┃ 🗑️  subject1 will be deleted \\(soft\\)\\W+Delete!\\W+Cancel.", render)

			t.Run("Enter soft deletes the subject", func(t *testing.T) {
				subjectsPage.Update(tests.Key('d'))
				cmds := subjectsPage.Update(tests.Key(tea.KeyEnter))
				msgs := tests.ExecuteBatchCmd(cmds)

				assert.Contains(t, msgs, sradmin.SoftDeletedSubjectMsg{Subject: "subject1"})
			})
		})

		t.Run("Delete after sorting", func(t *testing.T) {
			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			subjectsPage.View(tests.NewKontext(), tests.Renderer)

			subjectsPage.Update(tests.Key(tea.KeyUp))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Regexp(t, "┃ 🗑️  subject99 will be deleted \\(soft\\)\\W+Delete!\\W+Cancel.", render)

			t.Run("Enter soft deletes the subject", func(t *testing.T) {
				subjectsPage.Update(tests.Key('d'))
				cmds := subjectsPage.Update(tests.Key(tea.KeyEnter))
				msgs := tests.ExecuteBatchCmd(cmds)

				assert.Contains(t, msgs, sradmin.SoftDeletedSubjectMsg{Subject: "subject99"})
			})

		})

		t.Run("Delete after searching from selective list", func(t *testing.T) {
			subjectsPage.Update(tests.Key('/'))
			tests.UpdateKeys(subjectsPage, "2")
			subjectsPage.Update(tests.Key(tea.KeyEnter))

			subjectsPage.View(tests.NewKontext(), tests.Renderer)

			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Regexp(t, "┃ 🗑️  subject72 will be deleted \\(soft\\)\\W+Delete!\\W+Cancel.", render)

			// reset search
			subjectsPage.Update(tests.Key('/'))
			subjectsPage.Update(tests.Key(tea.KeyEsc))
		})

		t.Run("Display error when deletion fails", func(t *testing.T) {
			subjectsPage.Update(tests.Key(tea.KeyF2))
			subjectsPage.Update(tests.Key('d'))
			cmds := subjectsPage.Update(tests.Key(tea.KeyEnter))

			for _, msg := range tests.ExecuteBatchCmd(cmds) {
				subjectsPage.Update(msg)
			}
			subjectsPage.Update(sradmin.SubjectDeletionErrorMsg{
				Err: fmt.Errorf("unable to delete subject"),
			})

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Regexp(t, "unable to delete subject", render)

			t.Run("When deletion failure msg visible do allow other cmdbars to activate", func(t *testing.T) {
				subjectsPage.Update(tests.Key('/'))

				render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

				assert.NotContains(t, render, "Failed to delete subject: unable to delete subject")
				assert.Contains(t, render, "> ")
			})
		})

		t.Run("When deletion started show spinning indicator", func(t *testing.T) {

			subjectsPage, _ := New(sradmin.NewMock())

			addSubjects(subjectsPage)

			subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

			render := subjectsPage.View(tests.Kontext, tests.Renderer)

			assert.Contains(t, render, " ⏳ Deleting Subject")
		})

		t.Run("Disable delete when no subject selected", func(t *testing.T) {
			emptyPage, _ := New(sradmin.NewMock())

			var subjects []sradmin.Subject
			emptyPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

			render := emptyPage.View(tests.NewKontext(), tests.Renderer)

			cmd := emptyPage.Update(tests.Key(tea.KeyF2))

			render = emptyPage.View(tests.NewKontext(), tests.Renderer)

			assert.Contains(t, render, "No Subjects Found")
			assert.Nil(t, cmd)
		})

		t.Run("When deletion spinner active do not allow other cmdbars to activate", func(t *testing.T) {
			subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

			subjectsPage.Update(tests.Key('/'))

			render := subjectsPage.View(tests.Kontext, tests.Renderer)

			assert.Contains(t, render, " ⏳ Deleting Subject")
		})

		t.Run("Remove delete subject from table", func(t *testing.T) {
			subjectsPage, _ := New(sradmin.NewMock())

			subjects := addSubjects(subjectsPage)

			subjectsPage.Update(sradmin.SubjectDeletedMsg{SubjectName: subjects[4].Name})

			render := subjectsPage.View(tests.Kontext, tests.Renderer)

			assert.NotRegexp(t, "subject4\\W+5", render)
		})
	})

	t.Run("List soft deleted subjects", func(t *testing.T) {

		srAdminMock := sradmin.NewMock(
			sradmin.WithHardDeleteSubjectCallbackFn(
				func(subject string) tea.Msg {
					return sradmin.HardDeletedSubjectMsg{Subject: subject}
				},
			),
		)
		subjectsPage, _ := New(srAdminMock)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 20; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
					Deleted:  i%2 != 0,
				})
		}

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(tests.Kontext, tests.Renderer)

		for i := 0; i < 20; i++ {
			if i%2 == 0 {
				assert.Regexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
			} else {
				assert.NotRegexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
			}
		}

		t.Run("Tab switches to soft deleted subjects", func(t *testing.T) {
			// move down first to later assert cursor went up after switching
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyDown))
			assert.Equal(t, "subject12", subjectsPage.table.SelectedRow()[0])
			// switch to soft deleted subjects
			subjectsPage.Update(tests.Key(tea.KeyTab))

			render := subjectsPage.View(tests.Kontext, tests.Renderer)

			assert.Contains(t, render, "Total Subjects:  10/20")

			for i := 0; i < 20; i++ {
				if i%2 != 0 {
					assert.Regexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
				} else {
					assert.NotRegexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
				}
			}

			// assert cursor is at top
			assert.Equal(t, "subject1", subjectsPage.table.SelectedRow()[0])
		})

		t.Run("Shortcuts list hard delete option", func(t *testing.T) {
			assert.Contains(t, subjectsPage.Shortcuts(), statusbar.Shortcut{
				Name:       "Delete (hard)",
				Keybinding: "F2",
			})
		})

		t.Run("F2 triggers subject hard delete", func(t *testing.T) {
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			render = subjectsPage.View(tests.NewKontext(), tests.Renderer)

			assert.Regexp(t, "┃ 🗑️  subject11 will be deleted \\(hard\\)\\W+Delete!\\W+Cancel.", render)

		})

		t.Run("Enter hard deletes the subject", func(t *testing.T) {
			subjectsPage.Update(tests.Key('d'))
			cmds := subjectsPage.Update(tests.Key(tea.KeyEnter))
			msgs := tests.ExecuteBatchCmd(cmds)

			assert.Contains(t, msgs, sradmin.HardDeletedSubjectMsg{Subject: "subject11"})
		})

	})

	t.Run("When listing started show spinning indicator", func(t *testing.T) {

		subjectsPage, _ := New(sradmin.NewMock())

		addSubjects(subjectsPage)

		subjectsPage.Update(sradmin.SubjectListingStartedMsg{})

		render := subjectsPage.View(tests.Kontext, tests.Renderer)

		assert.Contains(t, render, " ⏳ Loading subjects")
	})

	t.Run("Shortcuts", func(t *testing.T) {
		subjectsPage, _ := New(sradmin.NewMock())

		t.Run("when there are no subjects", func(t *testing.T) {
			subjectsPage.subjects = nil
			shortcuts := subjectsPage.Shortcuts()

			assert.Equal(t,
				[]statusbar.Shortcut{
					{Name: "Register New Schema", Keybinding: "C-n"},
					{Name: "Refresh", Keybinding: "F5"},
				}, shortcuts)
		})

		t.Run("when there are subjects", func(t *testing.T) {
			addSubjects(subjectsPage)
			subjectsPage.View(tests.Kontext, tests.Renderer)

			shortcuts := subjectsPage.Shortcuts()

			assert.Equal(t,
				[]statusbar.Shortcut{
					{Name: "Register New Schema", Keybinding: "C-n"},
					{Name: "Refresh", Keybinding: "F5"},
					{Name: "Deleted Subjects", Keybinding: "tab"},
					{Name: "Search", Keybinding: "/"},
					{Name: "Delete (soft)", Keybinding: "F2"},
				}, shortcuts)
		})

		t.Run("cmdbar shortcuts have priority", func(t *testing.T) {

			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			shortcuts := subjectsPage.Shortcuts()

			assert.Equal(t,
				[]statusbar.Shortcut{
					{Name: "Confirm", Keybinding: "enter"},
					{Name: "Select Cancel", Keybinding: "c"},
					{Name: "Select Delete", Keybinding: "d"},
					{Name: "Cancel", Keybinding: "esc/F2"},
				}, shortcuts)
		})

	})
}

func addSubjects(subjectsPage *Model) []sradmin.Subject {
	var subjects []sradmin.Subject
	var versions []int
	for i := 0; i < 10; i++ {
		versions = append(versions, i)
		subjects = append(subjects,
			sradmin.Subject{
				Name:     fmt.Sprintf("subject%d", i),
				Versions: versions,
			})
	}
	subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})
	return subjects
}

func shuffle[T any](slice []T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
	}
}
