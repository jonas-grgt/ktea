package ui

import tea "charm.land/bubbletea/v2"

// Navigate to a new page
type Navigate func() tea.Cmd

// NavigateWithMsg to a new page with a message
type NavigateWithMsg[T any] func(T) tea.Cmd
