package ui

import tea "github.com/charmbracelet/bubbletea"

// Navigate to a new page
type Navigate func() tea.Cmd

// NavigateWithMsg to a new page with a message
type NavigateWithMsg[T any] func(T) tea.Cmd
