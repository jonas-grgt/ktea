package main

import (
	"testing"

	"ktea/config"
	"ktea/kadmin"
	"ktea/styles"

	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalBackground(t *testing.T) {
	load := func(name string) styles.Scheme {
		s, err := styles.Load(name, "", true)
		require.NoError(t, err)
		return s
	}
	t.Cleanup(func() { styles.Apply(load("default")) })

	disableNerdFonts := true
	newModel := func() *Model {
		m := NewModel(&disableNerdFonts, kadmin.NewMockKadminInstantiator(), config.NewInMemoryConfigIO(&config.Config{}))
		m.termBg = lipgloss.Color("#101010")
		m.termFg = lipgloss.Color("#eeeeee")
		return m
	}

	t.Run("a theme without background leaves the terminal alone", func(t *testing.T) {
		styles.Apply(load("default"))
		v := newModel().View()
		assert.Nil(t, v.BackgroundColor)
		assert.Nil(t, v.ForegroundColor)
	})

	t.Run("switching away from a painted background restores the terminal's own", func(t *testing.T) {
		m := newModel()
		styles.Apply(load("catppuccin-latte"))
		v := m.View()
		assert.Equal(t, load("catppuccin-latte").Background, v.BackgroundColor)
		assert.Equal(t, load("catppuccin-latte").Text, v.ForegroundColor)

		styles.Apply(load("dracula"))
		v = m.View()
		assert.Equal(t, lipgloss.Color("#101010"), v.BackgroundColor)
		assert.Equal(t, lipgloss.Color("#eeeeee"), v.ForegroundColor)
	})

	t.Run("quitting restores the terminal's own background", func(t *testing.T) {
		m := newModel()
		styles.Apply(load("desert"))
		m.View()

		m.quitting = true
		v := m.View()
		assert.Equal(t, lipgloss.Color("#101010"), v.BackgroundColor)
		assert.Equal(t, lipgloss.Color("#eeeeee"), v.ForegroundColor)
	})
}

func TestThemeChangeRepaintsScreen(t *testing.T) {
	t.Cleanup(func() { styles.Apply(mustLoad(t, "default")) })
	disableNerdFonts := true
	m := NewModel(&disableNerdFonts, kadmin.NewMockKadminInstantiator(), config.NewInMemoryConfigIO(&config.Config{}))
	load := func(theme string) tea.Msg {
		_, cmd := m.Update(config.LoadedMsg{Config: &config.Config{Theme: theme}})
		if cmd == nil {
			return nil
		}
		return cmd()
	}

	assert.True(t, containsMsg(load("desert"), tea.ClearScreen()), "painting a background must repaint the screen")
	assert.False(t, containsMsg(load("desert"), tea.ClearScreen()), "same background, no repaint")
	assert.True(t, containsMsg(load("default"), tea.ClearScreen()), "removing the background must repaint the screen")
}

func mustLoad(t *testing.T, name string) styles.Scheme {
	t.Helper()
	s, err := styles.Load(name, "", true)
	require.NoError(t, err)
	return s
}

func containsMsg(msg, want tea.Msg) bool {
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil && containsMsg(c(), want) {
				return true
			}
		}
		return false
	}
	return msg == want
}
