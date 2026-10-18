package ui

import (
	tea "charm.land/bubbletea/v2"
)

func PublishMsg(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
