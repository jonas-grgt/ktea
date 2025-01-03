package keys

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/ui"
)

type AKey interface{}

func Key(key AKey) tea.Msg {
	switch key := key.(type) {
	case rune:
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{key},
			Alt:   false,
			Paste: false,
		}
	case int:
		return tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{rune(key)},
			Alt:   false,
			Paste: false,
		}
	case tea.KeyType:
		return tea.KeyMsg{
			Type:  key,
			Runes: []rune{},
			Alt:   false,
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
