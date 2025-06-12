package tests

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/ui"
	"ktea/ui/pages/nav"
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

func UpdateKeys(m ui.View, keys string) {
	for _, k := range keys {
		m.Update(Key(k))
	}
}

func Submit(page nav.Page) []tea.Msg {
	cmd := page.Update(Key(tea.KeyEnter))
	// next field
	cmd = page.Update(cmd())
	// next group and submit
	cmd = page.Update(cmd())
	return ExecuteBatchCmd(cmd)
}

func NextGroup(page nav.Page, cmd tea.Cmd) {
	// next field
	cmd = page.Update(cmd())
	// next group
	cmd = page.Update(cmd())
}
