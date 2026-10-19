package settings_page

import (
	"os"
	"path/filepath"
	"testing"

	"ktea/config"
	"ktea/styles"
	"ktea/tests"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsPage(t *testing.T) {
	t.Cleanup(func() { styles.Apply(mustLoad(t, styles.DefaultTheme)) })

	newPage := func(t *testing.T, dir string) (*Model, *config.Config) {
		cfg := config.New(config.NewInMemoryConfigIO(&config.Config{}))
		ktx := tests.NewKontext(tests.WithConfig(cfg))
		page := New(ktx, dir, true)
		page.View(ktx, tests.Renderer) // lays out the table
		return page, cfg
	}

	t.Run("lists built-in and custom themes, marking the active one", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "ocean.yaml"), []byte("base: dracula\n"), 0o644))
		page, _ := newPage(t, dir)
		ktx := tests.NewKontext()

		render := ansi.Strip(page.View(ktx, tests.Renderer))

		assert.Regexp(t, `●\s+default`, render)
		assert.Contains(t, render, "desert")
		assert.Contains(t, render, "ocean")
		assert.Contains(t, render, "custom")
	})

	t.Run("enter applies the selected theme and saves it", func(t *testing.T) {
		page, cfg := newPage(t, t.TempDir())

		page.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		selected := page.table.SelectedRow()[1]
		page.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Equal(t, selected, cfg.Theme)
		assert.Equal(t, mustLoad(t, selected).Primary, styles.Current.Primary)
	})

	t.Run("a broken theme shows the error and keeps the current theme", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "zz-broken.yaml"), []byte("dark:\n  primry: \"#fff\"\n"), 0o644))
		styles.Apply(mustLoad(t, styles.DefaultTheme))
		page, cfg := newPage(t, dir)

		page.table.GotoBottom()
		page.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Empty(t, cfg.Theme)
		assert.Equal(t, mustLoad(t, styles.DefaultTheme).Primary, styles.Current.Primary)
		assert.Contains(t, ansi.Strip(page.View(tests.NewKontext(), tests.Renderer)), "unknown color dark.primry")
	})
}

func mustLoad(t *testing.T, name string) styles.Scheme {
	s, err := styles.Load(name, "", true)
	require.NoError(t, err)
	return s
}
