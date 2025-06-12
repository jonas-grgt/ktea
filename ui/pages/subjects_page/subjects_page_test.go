package subjects_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/sradmin"
	"ktea/tests"
	"ktea/ui/pages/nav"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type MockSubjectsLister struct {
	subjectListingStartedMsg sradmin.SubjectListingStartedMsg
}

func (m *MockSubjectsLister) ListSubjects() tea.Msg {
	return nil
}

type MockSubjectsDeleter struct {
	deletionResultMsg tea.Msg
}

type DeletedSubjectMsg struct {
	Subject string
	Version int
}

type MockGlobalCompatibilityLister struct {
}

func (m MockGlobalCompatibilityLister) ListGlobalCompatibility() tea.Msg {
	return nil
}

func (m *MockSubjectsDeleter) DeleteSubject(subject string) tea.Msg {
	return m.deletionResultMsg
}

func TestSubjectsPage(t *testing.T) {

	t.Run("No subjects found", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&MockSubjectsDeleter{},
		)

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: []sradmin.Subject{}})

		render := subjectsPage.View(tests.TestKontext, tests.TestRenderer)

		assert.Contains(t, render, "No Subjects Found")

		t.Run("enter is ignored", func(t *testing.T) {

			cmd := subjectsPage.Update(tests.Key(tea.KeyEnter))

			assert.Nil(t, cmd)
		})
	})

	t.Run("Render listed subjects and number of versions", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&MockSubjectsDeleter{},
		)

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

		render := subjectsPage.View(tests.TestKontext, tests.TestRenderer)

		for i := 0; i < 10; i++ {
			assert.Regexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
		}
	})

	t.Run("Enter opens schema detail page", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&MockSubjectsDeleter{},
		)

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
		subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

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

	t.Run("Remove delete subject from table", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&MockSubjectsDeleter{},
		)

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

		subjectsPage.Update(sradmin.SubjectDeletedMsg{SubjectName: subjects[4].Name})

		render := subjectsPage.View(tests.TestKontext, tests.TestRenderer)

		assert.NotRegexp(t, "subject4\\W+5", render)
	})

	t.Run("When deletion spinner active do not allow other cmdbars to activate", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

		subjectsPage.Update(tests.Key('/'))

		render := subjectsPage.View(tests.TestKontext, tests.TestRenderer)

		assert.Contains(t, render, " ⏳ Deleting Subject")
	})

	t.Run("Order subjects default by Subject Name Asc", func(t *testing.T) {
		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&deleter,
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
		shuffle(subjects)

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

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
		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&deleter,
		)

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
		subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

		t.Run("by Subject Name", func(t *testing.T) {

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render := subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

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

			render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

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

			render := subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

			assert.Contains(t, render, "▲ Comp")

			lastBackwardIdx := strings.LastIndex(render, "BACKWARD")
			lastForwardIdx := strings.Index(render, "FORWARD")

			assert.Less(t, lastBackwardIdx, lastForwardIdx, "FORWARD came before BACKWARD")

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

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

			render := subjectsPage.View(tests.NewKontext(), tests.TestRenderer)
			assert.Contains(t, render, "▼ V")

			subject1Idx := strings.Index(render, "subject1")
			subject88Idx := strings.Index(render, "subject2")

			assert.Greater(t, subject1Idx, subject88Idx, "subject1 came not before subject2")

			subjectsPage.Update(tests.Key(tea.KeyF3))
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyEsc))

			render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)
			assert.Contains(t, render, "▲ V")

			subject1Idx = strings.Index(render, "subject1")
			subject88Idx = strings.Index(render, "subject2")

			assert.Less(t, subject1Idx, subject88Idx, "subject2 came not before subject1")
		})
	})

	t.Run("Loading details page (enter) after searching from selective list", func(t *testing.T) {
		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&deleter,
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
		subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

		subjectsPage.Update(tests.Key('/'))
		tests.UpdateKeys(subjectsPage, "1")
		subjectsPage.Update(tests.Key(tea.KeyEnter))
		subjectsPage.Update(tests.Key(tea.KeyDown))
		subjectsPage.Update(tests.Key(tea.KeyDown))
		cmd := subjectsPage.Update(tests.Key(tea.KeyEnter))

		msgs := tests.ExecuteBatchCmd(cmd)

		assert.Contains(t, msgs, nav.LoadSchemaDetailsPageMsg{
			Subject: sradmin.Subject{
				Name:     "subject11",
				Versions: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			},
		})
	})

	t.Run("Delete subject", func(t *testing.T) {
		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&deleter,
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
		render := subjectsPage.View(tests.NewKontext(), tests.TestRenderer)
		assert.NotRegexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)

		t.Run("F2 triggers subject delete", func(t *testing.T) {
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

			assert.Regexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)
		})

		t.Run("Delete after searching from selective list", func(t *testing.T) {
			subjectsPage.Update(tests.Key('/'))
			tests.UpdateKeys(subjectsPage, "1")
			subjectsPage.Update(tests.Key(tea.KeyEnter))
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyDown))
			subjectsPage.Update(tests.Key(tea.KeyF2))

			render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

			assert.Regexp(t, "┃ 🗑️  subject11 will be deleted permanently\\W+Delete!\\W+Cancel.", render)

			// reset search
			subjectsPage.Update(tests.Key('/'))
			subjectsPage.Update(tests.Key(tea.KeyEsc))
		})

		t.Run("Enter effectively deletes the subject", func(t *testing.T) {
			deleter.deletionResultMsg = DeletedSubjectMsg{"subject1", 1}

			subjectsPage.Update(tests.Key(tea.KeyF2))
			subjectsPage.Update(tests.Key('d'))
			cmds := subjectsPage.Update(tests.Key(tea.KeyEnter))
			msgs := tests.ExecuteBatchCmd(cmds)

			assert.Contains(t, msgs, DeletedSubjectMsg{"subject1", 1})
		})

		t.Run("Display error when deletion fails", func(t *testing.T) {
			deleter.deletionResultMsg = sradmin.SubjectDeletionStartedMsg{}

			subjectsPage.Update(tests.Key(tea.KeyF2))
			subjectsPage.Update(tests.Key('d'))
			cmds := subjectsPage.Update(tests.Key(tea.KeyEnter))

			for _, msg := range tests.ExecuteBatchCmd(cmds) {
				subjectsPage.Update(msg)
			}
			subjectsPage.Update(sradmin.SubjectDeletionErrorMsg{
				Err: fmt.Errorf("unable to delete subject"),
			})

			render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

			assert.Regexp(t, "unable to delete subject", render)

			t.Run("When deletion failure msg visible do allow other cmdbars to activate", func(t *testing.T) {
				subjectsPage.Update(tests.Key('/'))

				render = subjectsPage.View(tests.NewKontext(), tests.TestRenderer)

				assert.NotContains(t, render, "Failed to delete subject: unable to delete subject")
				assert.Contains(t, render, "> ")
			})
		})

		t.Run("When deletion started show spinning indicator", func(t *testing.T) {

			subjectsPage, _ := New(
				&MockSubjectsLister{},
				&MockGlobalCompatibilityLister{},
				&MockSubjectsDeleter{},
			)
			subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

			render := subjectsPage.View(tests.TestKontext, tests.TestRenderer)

			assert.Contains(t, render, " ⏳ Deleting Subject")
		})

		t.Run("Disable delete when no subject selected", func(t *testing.T) {
			emptyPage, _ := New(
				&MockSubjectsLister{},
				&MockGlobalCompatibilityLister{},
				&deleter,
			)

			var subjects []sradmin.Subject
			emptyPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

			render := emptyPage.View(tests.NewKontext(), tests.TestRenderer)

			cmd := emptyPage.Update(tests.Key(tea.KeyF2))

			render = emptyPage.View(tests.NewKontext(), tests.TestRenderer)

			assert.Contains(t, render, "No Subjects Found")
			assert.Nil(t, cmd)
		})
	})

	t.Run("When listing started show spinning indicator", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockGlobalCompatibilityLister{},
			&MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectListingStartedMsg{})

		render := subjectsPage.View(tests.TestKontext, tests.TestRenderer)

		assert.Contains(t, render, " ⏳ Loading subjects")
	})
}

func shuffle[T any](slice []T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
	}
}
