package styles

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/lipgloss/v2"
	chroma "github.com/alecthomas/chroma/v2/styles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTheme(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0o644))
	return dir
}

func TestLoad(t *testing.T) {
	t.Run("default dark keeps the original ktea colors", func(t *testing.T) {
		s, err := Load("", "", true)
		require.NoError(t, err)
		assert.Equal(t, lipgloss.Color("205"), s.Primary)
		assert.Equal(t, lipgloss.Color("132"), s.PrimaryMuted)
		assert.Equal(t, lipgloss.Color("#7571F9"), s.Accent)
		assert.Equal(t, lipgloss.Color("#353533"), s.Surface)
		assert.Equal(t, lipgloss.Color("#F5F5F5"), s.BorderFocus)
		assert.Equal(t, lipgloss.Color("235"), s.BorderBlur)
		assert.Equal(t, "github-dark", s.Syntax)
	})

	t.Run("every built-in sets every color for both backgrounds", func(t *testing.T) {
		for _, name := range Builtins() {
			for _, dark := range []bool{true, false} {
				s, err := Load(name, "", dark)
				require.NoError(t, err)
				for role, c := range s.roles() {
					if role == "background" { // optional
						continue
					}
					assert.NotNil(t, *c, "%s dark=%v: %s", name, dark, role)
					assert.NotEqual(t, lipgloss.NoColor{}, *c, "%s dark=%v: %s", name, dark, role)
				}
				assert.Contains(t, chroma.Registry, s.Syntax, "%s: unknown syntax style", name)
				// building the styles must not panic
				Apply(s)
			}
		}
		Apply(themeDefault(true))
	})

	t.Run("catppuccin flavors use the official palette with a fixed background", func(t *testing.T) {
		// catppuccin.com/palette
		for name, base := range map[string]string{
			"catppuccin-latte":     "#eff1f5",
			"catppuccin-frappe":    "#303446",
			"catppuccin-macchiato": "#24273a",
			"catppuccin-mocha":     "#1e1e2e",
		} {
			for _, dark := range []bool{true, false} {
				s, err := Load(name, "", dark)
				require.NoError(t, err)
				assert.Equal(t, lipgloss.Color(base), s.Background, name)
				assert.Equal(t, name, s.Syntax, name)
				assert.Equal(t, name != "catppuccin-latte", s.IsDark, name)
			}
		}
		mocha, _ := Load("catppuccin-mocha", "", true)
		assert.Equal(t, lipgloss.Color("#cba6f7"), mocha.Primary) // mauve
		dark, _ := Load("catppuccin", "", true)
		light, _ := Load("catppuccin", "", false)
		assert.Equal(t, "catppuccin-mocha", dark.Syntax)
		assert.Equal(t, "catppuccin-latte", light.Syntax)
	})

	t.Run("popular themes paint their official background", func(t *testing.T) {
		for name, want := range map[string]struct {
			base   string
			isDark bool
		}{
			"tokyonight-night": {"#1a1b26", true},
			"tokyonight-storm": {"#24283b", true},
			"tokyonight-moon":  {"#222436", true},
			"tokyonight-day":   {"#e1e2e7", false},
			"gruvbox-dark":     {"#282828", true},
			"gruvbox-light":    {"#fbf1c7", false},
			"nord":             {"#2e3440", true},
			"rose-pine-main":   {"#191724", true},
			"rose-pine-moon":   {"#232136", true},
			"rose-pine-dawn":   {"#faf4ed", false},
			"solarized-dark":   {"#002b36", true},
			"solarized-light":  {"#fdf6e3", false},
			"onedark":          {"#282c34", true},
		} {
			for _, dark := range []bool{true, false} {
				s, err := Load(name, "", dark)
				require.NoError(t, err)
				assert.Equal(t, lipgloss.Color(want.base), s.Background, name)
				assert.Equal(t, want.isDark, s.IsDark, name)
			}
		}
		for name, variants := range map[string][2]string{
			"tokyonight": {"tokyonight-night", "tokyonight-day"},
			"gruvbox":    {"gruvbox-dark", "gruvbox-light"},
			"rose-pine":  {"rose-pine-main", "rose-pine-dawn"},
			"solarized":  {"solarized-dark", "solarized-light"},
		} {
			dark, _ := Load(name, "", true)
			light, _ := Load(name, "", false)
			wantDark, _ := Load(variants[0], "", true)
			wantLight, _ := Load(variants[1], "", true)
			assert.Equal(t, wantDark, dark, name)
			assert.Equal(t, wantLight, light, name)
		}
	})

	t.Run("user theme overrides its base", func(t *testing.T) {
		dir := writeTheme(t, "mine", `
base: dracula
syntax: monokai
dark:
  primary: "#123456"
  accent: "99"
light:
  primary: "#abcdef"
`)
		dark, err := Load("mine", dir, true)
		require.NoError(t, err)
		assert.Equal(t, lipgloss.Color("#123456"), dark.Primary)
		assert.Equal(t, lipgloss.Color("99"), dark.Accent)
		assert.Equal(t, themeDracula(true).Text, dark.Text)
		assert.Equal(t, "monokai", dark.Syntax)

		light, err := Load("mine", dir, false)
		require.NoError(t, err)
		assert.Equal(t, lipgloss.Color("#abcdef"), light.Primary)
		assert.Equal(t, themeDracula(false).Accent, light.Accent)
	})

	t.Run("missing light section falls back to dark overrides", func(t *testing.T) {
		dir := writeTheme(t, "mine", "dark:\n  primary: \"#123456\"\n")
		s, err := Load("mine", dir, false)
		require.NoError(t, err)
		assert.Equal(t, lipgloss.Color("#123456"), s.Primary)
		assert.Equal(t, themeDefault(false).Text, s.Text)
	})

	t.Run("empty file is the default theme", func(t *testing.T) {
		dir := writeTheme(t, "empty", "")
		s, err := Load("empty", dir, true)
		require.NoError(t, err)
		assert.Equal(t, themeDefault(true), s)
	})

	errCases := map[string]struct{ content, msg string }{
		"unknown color role": {"dark:\n  primry: \"#fff\"\n", "unknown color dark.primry"},
		"invalid color":      {"light:\n  primary: pink\n", "invalid color light.primary"},
		"unknown base":       {"base: nope\n", `unknown base "nope"`},
		"unknown key":        {"colours: {}\n", "colours"},
		"unknown syntax":     {"syntax: nope\n", `unknown syntax style "nope"`},
	}
	for name, tc := range errCases {
		t.Run(name, func(t *testing.T) {
			dir := writeTheme(t, "bad", tc.content)
			_, err := Load("bad", dir, true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.msg)
		})
	}

	t.Run("missing theme file", func(t *testing.T) {
		_, err := Load("ghost", t.TempDir(), true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `theme "ghost"`)
	})
}
