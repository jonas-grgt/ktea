package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"testing"
)

type TestMsg struct{}

type AssertDeletedMsg struct {
	deleteValue string
}

func TestDeleteCmdBar(t *testing.T) {
	t.Run("When invalid do not delete", func(t *testing.T) {
		var deleteFunc DeleteFunc[string] = func(s string) tea.Cmd {
			return nil
		}
		cmdBar := NewDeleteCmdBar[string](nil, deleteFunc, func(string) (bool, tea.Cmd) {
			return false, func() tea.Msg {
				return TestMsg{}
			}
		})

		cmdBar.Update(tests.Key(tea.KeyF2))
		active, msg, cmd := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.True(t, active)
		assert.Nil(t, msg)
		assert.IsType(t, TestMsg{}, cmd())
	})

	t.Run("deleteFunc called upon deleting", func(t *testing.T) {
		var deleteFunc DeleteFunc[string] = func(s string) tea.Cmd {
			return func() tea.Msg {
				return AssertDeletedMsg{}
			}
		}
		cmdBar := NewDeleteCmdBar[string](nil, deleteFunc, nil)

		cmdBar.Update(tests.Key(tea.KeyF2))
		cmdBar.Update(tests.Key('d'))
		active, msg, cmd := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.True(t, active)
		assert.Nil(t, msg)
		assert.IsType(t, AssertDeletedMsg{}, cmd())
	})

	t.Run("deleteValue is passed to deleteFunc upon deleting", func(t *testing.T) {
		var deleteFunc DeleteFunc[string] = func(s string) tea.Cmd {
			return func() tea.Msg {
				return AssertDeletedMsg{deleteValue: s}
			}
		}
		cmdBar := NewDeleteCmdBar[string](nil, deleteFunc, nil)

		cmdBar.Update(tests.Key(tea.KeyF2))
		cmdBar.Update(tests.Key('d'))
		cmdBar.Delete("deleteMe")
		_, _, cmd := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.Equal(t, AssertDeletedMsg{"deleteMe"}, cmd())
	})

}
