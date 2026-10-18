package tests

import (
	"fmt"
	"ktea/ui"
	"ktea/ui/pages"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

type AKey interface{}

func KeyWithAlt(key rune) tea.Msg {
	return keyMsg(key, true)
}

func Key(key AKey) tea.Msg {
	return keyMsg(key, false)
}

// Ctrl returns a ctrl+<key> press, e.g. Ctrl('r') for ctrl+r.
func Ctrl(key rune) tea.Msg {
	return tea.KeyPressMsg{Code: key, Mod: tea.ModCtrl}
}

func keyMsg(key AKey, altKey bool) tea.Msg {
	var code rune
	switch key := key.(type) {
	case rune:
		code = key
	case int:
		code = rune(key)
	default:
		panic(fmt.Sprintf("Cannot handle %v", key))
	}
	msg := tea.KeyPressMsg{Code: code}
	if altKey {
		msg.Mod = tea.ModAlt
	} else if unicode.IsPrint(code) {
		msg.Text = string(code)
	}
	return msg
}

type Input struct {
	ui.View
}

func (i Input) Enter() {
	cmd := i.View.Update(Key(tea.KeyEnter))
	i.View.Update(cmd())
}

func UpdateKeys(m ui.View, keys string) *Input {
	for _, k := range keys {
		m.Update(Key(k))
	}
	return &Input{
		View: m,
	}
}

type KeyBoard struct {
	view ui.View
}

func (k *KeyBoard) Type(keys string) *KeyBoard {
	UpdateKeys(k.view, keys)
	return k
}

func (k *KeyBoard) Enter() {
	cmd := k.view.Update(Key(tea.KeyEnter))
	if cmd != nil {
		k.view.Update(cmd())
	}
}

func (k *KeyBoard) EnterAndGetCmd() tea.Cmd {
	return k.view.Update(Key(tea.KeyEnter))
}

func (k *KeyBoard) Submit() []tea.Msg {
	cmd := k.view.Update(Key(tea.KeyEnter))
	// next field
	cmd = k.view.Update(cmd())
	// next group and submit
	cmd = k.view.Update(cmd())
	return ExecuteBatchCmd(cmd)
}

func (k *KeyBoard) Down() *KeyBoard {
	k.view.Update(Key(tea.KeyDown))
	return k
}

func (k *KeyBoard) Up() *KeyBoard {
	k.view.Update(Key(tea.KeyUp))
	return k
}

func (k *KeyBoard) Right() *KeyBoard {
	k.view.Update(Key(tea.KeyRight))
	return k
}

func (k *KeyBoard) Backspace() *KeyBoard {
	k.view.Update(Key(tea.KeyBackspace))
	return k
}

func (k *KeyBoard) F5() *KeyBoard {
	k.view.Update(Key(tea.KeyF5))
	return k
}

func (k *KeyBoard) F2() *KeyBoard {
	k.view.Update(Key(tea.KeyF2))
	return k
}

func NewKeyboard(view ui.View) *KeyBoard {
	return &KeyBoard{
		view: view,
	}
}

func Submit(page pages.Page) []tea.Msg {
	cmd := page.Update(Key(tea.KeyEnter))
	// next field
	cmd = page.Update(cmd())
	// next group and submit
	cmd = page.Update(cmd())
	return ExecuteBatchCmd(cmd)
}

func NextGroup(page pages.Page, cmd tea.Cmd) {
	// next field
	cmd = page.Update(cmd())
	// next group
	cmd = page.Update(cmd())
}
