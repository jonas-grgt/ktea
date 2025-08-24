package sradmin

import tea "github.com/charmbracelet/bubbletea"

type GlobalCompatibilityListedMsg struct {
	Compatibility string
}

type GlobalCompatibilityListingErrorMsg struct {
	Err error
}

type GlobalCompatibilityListingStartedMsg struct {
	compatibility chan string
	err           chan error
}

func (msg *GlobalCompatibilityListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case compatibility := <-msg.compatibility:
		return GlobalCompatibilityListedMsg{compatibility}
	case err := <-msg.err:
		return GlobalCompatibilityListingErrorMsg{err}
	}
}

func (s *DefaultSrClient) ListGlobalCompatibility() tea.Msg {
	maybeIntroduceLatency()

	compatibilityChan := make(chan string)
	errChan := make(chan error)

	go s.doListGlobalCompatibility(compatibilityChan, errChan)

	return GlobalCompatibilityListingStartedMsg{
		compatibility: compatibilityChan,
		err:           errChan,
	}
}

func (s *DefaultSrClient) doListGlobalCompatibility(
	compatibilityChan chan string,
	errChan chan error,
) {
	compatibility, err := s.client.GetGlobalCompatibilityLevel()
	if err != nil {
		errChan <- err
		return
	}
	compatibilityChan <- compatibility.String()
}
