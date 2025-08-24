package sradmin

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"slices"
	"sync"
)

type SubjectsListedMsg struct {
	Subjects []Subject
}

type SubjectListingErrorMsg struct {
	Err error
}

type SubjectListingStartedMsg struct {
	subjects chan []Subject
	err      chan error
}

// AwaitCompletion return
// a SubjectsListedMsg upon success
// or SubjectListingErrorMsg upon failure.
func (msg *SubjectListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case subjects := <-msg.subjects:
		return SubjectsListedMsg{subjects}
	case err := <-msg.err:
		log.Error("Failed to fetch subjects", "err", err)
		return SubjectListingErrorMsg{err}
	}
}

func (s *DefaultSrClient) ListSubjects() tea.Msg {
	subjectsChan := make(chan []Subject)
	errChan := make(chan error)

	go s.doListSubject(subjectsChan, errChan)

	return SubjectListingStartedMsg{subjectsChan, errChan}
}

type Subject struct {
	Name          string
	Versions      []int
	Compatibility string
	Deleted       bool
}

func (s *Subject) LatestVersion() int {
	return slices.Max(s.Versions)
}

func (s *DefaultSrClient) doListSubject(
	subjectsChan chan []Subject,
	errChan chan error,
) {
	maybeIntroduceLatency()

	var (
		subjectNames []string
		err          error
	)

	subjectNames, err = s.client.GetSubjectsIncludingDeleted()
	if err != nil {
		errChan <- err
		return
	}

	subjects := make([]Subject, len(subjectNames))
	active, _ := s.client.GetSubjects()
	for i, name := range subjectNames {
		deleted := !slices.Contains(active, name)
		subjects[i] = Subject{Name: name, Deleted: deleted}
	}

	versionResults := make([][]int, len(subjects))
	compResults := make([]string, len(subjects))

	var wg sync.WaitGroup
	errs := make(chan error, len(subjects)*2) // buffer for errors

	for i, subj := range subjects {
		// Version can only be fetched if the subject is not deleted
		if !subj.Deleted {
			wg.Add(1)
			go func(i int, name string) {
				defer wg.Done()
				versions, err := s.client.GetSchemaVersions(name)
				if err != nil {
					errs <- fmt.Errorf("get versions %s: %w", name, err)
					return
				}
				versionResults[i] = versions
			}(i, subj.Name)
		}

		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			comp, err := s.client.GetCompatibilityLevel(name, true)
			if err != nil {
				errs <- fmt.Errorf("get compatibility %s: %w", name, err)
				return
			}
			compResults[i] = comp.String()
		}(i, subj.Name)
	}

	wg.Wait()
	close(errs)

	// return first error if any
	for e := range errs {
		if e != nil {
			errChan <- e
			return
		}
	}

	for i := range subjects {
		subjects[i].Versions = versionResults[i]
		subjects[i].Compatibility = compResults[i]
	}

	s.mu.Lock()
	s.subjects = subjects
	s.mu.Unlock()

	subjectsChan <- subjects
}
