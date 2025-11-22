package tests

import (
	"fmt"
	"ktea/ui"
	"ktea/ui/pages"

	tea "github.com/charmbracelet/bubbletea"
)

type AKey interface{}

func KeyWithAlt(key tea.KeyType) tea.Msg {
	return keyMsg(key, true)
}

func Key(key AKey) tea.Msg {
	return keyMsg(key, false)
}

func keyMsg(key AKey, altKey bool) tea.Msg {
	switch key := key.(type) {
	case rune:
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{key},
			Alt:   altKey,
			Paste: false,
		}
	case int:
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{rune(key)},
			Alt:   altKey,
			Paste: false,
		}
	case tea.KeyType:
		return tea.KeyMsg{
			Type:  key,
			Runes: []rune{},
			Alt:   altKey,
			Paste: false,
		}
	default:
		panic(fmt.Sprintf("Cannot handle %v", key))
	}
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
