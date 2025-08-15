package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"testing"
)

type TestMsg struct{}

type AssertDeletedMsg struct {
	deleteValue string
}

func TestDeleteCmdBar(t *testing.T) {
	t.Run("When invalid do not delete", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return nil
		}
		cmdBar := NewDeleteCmdBar[string](nil,
			deleteFunc,
			WithValidateFn(func(string) (bool, tea.Cmd) {
				return false, func() tea.Msg {
					return TestMsg{}
				}
			}))

		cmdBar.Update(tests.Key(tea.KeyF2))
		active, msg, cmd := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.True(t, active)
		assert.Nil(t, msg)
		assert.IsType(t, TestMsg{}, cmd())
	})

	t.Run("deleteFunc called upon deleting", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return func() tea.Msg {
				return AssertDeletedMsg{}
			}
		}
		cmdBar := NewDeleteCmdBar[string](nil, deleteFunc)

		cmdBar.Update(tests.Key(tea.KeyF2))
		cmdBar.Update(tests.Key('d'))
		active, msg, cmd := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.True(t, active)
		assert.Nil(t, msg)
		assert.IsType(t, AssertDeletedMsg{}, cmd())
	})

	t.Run("deleteValue is passed to deleteFunc upon deleting", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return func() tea.Msg {
				return AssertDeletedMsg{deleteValue: s}
			}
		}
		cmdBar := NewDeleteCmdBar[string](nil, deleteFunc)

		cmdBar.Update(tests.Key(tea.KeyF2))
		cmdBar.Update(tests.Key('d'))
		cmdBar.Delete("deleteMe")
		_, _, cmd := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.Equal(t, AssertDeletedMsg{"deleteMe"}, cmd())
	})

	t.Run("Cancel button cancels dialog", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return nil
		}
		cmdBar := NewDeleteCmdBar[string](nil,
			deleteFunc,
			WithValidateFn(func(string) (bool, tea.Cmd) {
				return true, nil
			}))

		cmdBar.Update(tests.Key(tea.KeyF2))
		cmdBar.Update(tests.Key('c'))
		active, msg, _ := cmdBar.Update(tests.Key(tea.KeyEnter))

		assert.False(t, active)
		assert.Nil(t, msg)
	})

	t.Run("esc cancels dialog", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return nil
		}
		cmdBar := NewDeleteCmdBar[string](nil,
			deleteFunc,
			WithValidateFn(func(string) (bool, tea.Cmd) {
				return true, nil
			}))

		cmdBar.Update(tests.Key(tea.KeyF2))
		active, msg, _ := cmdBar.Update(tests.Key(tea.KeyEsc))

		assert.False(t, active)
		assert.Nil(t, msg)
	})

	t.Run("override delete key", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return nil
		}
		cmdBar := NewDeleteCmdBar[string](nil,
			deleteFunc,
			WithValidateFn(func(string) (bool, tea.Cmd) {
				return true, nil
			}),
			WithDeleteKey[string]("f4"))

		t.Run("uses overridden key as toggle", func(t *testing.T) {

			assert.False(t, cmdBar.IsFocussed())

			cmdBar.Update(tests.Key(tea.KeyF2))

			assert.False(t, cmdBar.IsFocussed())

			cmdBar.Update(tests.Key(tea.KeyF4))

			assert.True(t, cmdBar.IsFocussed())

			cmdBar.Update(tests.Key(tea.KeyF4))

			assert.False(t, cmdBar.IsFocussed())
		})

		t.Run("uses overridden key in shortcuts", func(t *testing.T) {
			// make active again
			cmdBar.Update(tests.Key(tea.KeyF4))

			assert.Contains(t, cmdBar.Shortcuts(), statusbar.Shortcut{Name: "Cancel", Keybinding: "esc/F4"})
		})
	})

	t.Run("hide bar", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return nil
		}
		cmdBar := NewDeleteCmdBar[string](nil,
			deleteFunc,
			WithValidateFn(func(string) (bool, tea.Cmd) {
				return true, nil
			}))

		cmdBar.Update(tests.Key(tea.KeyF2))
		assert.True(t, cmdBar.IsFocussed())
		cmdBar.Hide()
		assert.False(t, cmdBar.IsFocussed())
	})

	t.Run("nil shortcuts when not active", func(t *testing.T) {
		var deleteFunc DeleteFn[string] = func(s string) tea.Cmd {
			return func() tea.Msg {
				return AssertDeletedMsg{}
			}
		}
		cmdBar := NewDeleteCmdBar[string](nil, deleteFunc)

		assert.Nil(t, cmdBar.Shortcuts())
	})

	t.Run("rendering", func(t *testing.T) {
		var deleteFn DeleteFn[string] = func(s string) tea.Cmd {
			return func() tea.Msg {
				return AssertDeletedMsg{}
			}
		}
		var deleteMsgFn = func(s string) string {
			return s
		}
		cmdBar := NewDeleteCmdBar[string](deleteMsgFn, deleteFn)
		cmdBar.Update(tests.Key(tea.KeyF2))
		cmdBar.Delete("subjectX")

		render := cmdBar.View(tests.TestKontext, tests.TestRenderer)

		assert.Contains(t, render, `┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃┃ 🗑️  subjectX    Delete!     Cancel.                                                             ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`)
	})
}
