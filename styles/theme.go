package styles

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	chroma "github.com/alecthomas/chroma/v2/styles"
	catppuccin "github.com/catppuccin/go"
	"gopkg.in/yaml.v3"
)

// Scheme holds the semantic colors the UI is drawn with.
type Scheme struct {
	Primary      color.Color // active tab, table selection, highlighted counts
	PrimaryMuted color.Color // inactive counts and chips
	Accent       color.Color // cmdbar prompts, form titles
	Highlight    color.Color // help keys
	Text         color.Color
	TextMuted    color.Color
	TextFaint    color.Color // inactive tabs
	OnPrimary    color.Color // text drawn on a Primary background
	Surface      color.Color // statusbar background
	SurfaceAlt   color.Color // inactive chips, inactive form borders
	BorderFocus  color.Color
	BorderBlur   color.Color
	Active       color.Color // active border tab
	Success      color.Color
	Error        color.Color
	// Background, when set, paints the whole terminal while ktea runs and
	// Text becomes the terminal's default text color. Nil keeps the terminal's colors.
	Background color.Color
	Syntax     string // chroma style used to highlight record payloads
	IsDark     bool
}

// ThemeFunc returns the scheme for a dark or light terminal background,
// mirroring huh.ThemeFunc.
type ThemeFunc func(isDark bool) Scheme

const DefaultTheme = "default"

var builtins = map[string]ThemeFunc{
	"default":              themeDefault,
	"dracula":              themeDracula,
	"catppuccin":           adaptive(catppuccinScheme(catppuccin.Mocha), catppuccinScheme(catppuccin.Latte)),
	"catppuccin-latte":     fixed(catppuccinScheme(catppuccin.Latte)),
	"catppuccin-frappe":    fixed(catppuccinScheme(catppuccin.Frappe)),
	"catppuccin-macchiato": fixed(catppuccinScheme(catppuccin.Macchiato)),
	"catppuccin-mocha":     fixed(catppuccinScheme(catppuccin.Mocha)),
	"desert":               themeDesert,
	"tokyonight":           adaptive(tokyoNightNight, tokyoNightDay),
	"tokyonight-night":     fixed(tokyoNightNight),
	"tokyonight-storm":     fixed(tokyoNightStorm),
	"tokyonight-moon":      fixed(tokyoNightMoon),
	"tokyonight-day":       fixed(tokyoNightDay),
	"gruvbox":              adaptive(gruvboxDark, gruvboxLight),
	"gruvbox-dark":         fixed(gruvboxDark),
	"gruvbox-light":        fixed(gruvboxLight),
	"nord":                 fixed(nord),
	"rose-pine":            adaptive(rosePineMain, rosePineDawn),
	"rose-pine-main":       fixed(rosePineMain),
	"rose-pine-moon":       fixed(rosePineMoon),
	"rose-pine-dawn":       fixed(rosePineDawn),
	"solarized":            adaptive(solarizedDark, solarizedLight),
	"solarized-dark":       fixed(solarizedDark),
	"solarized-light":      fixed(solarizedLight),
	"onedark":              fixed(oneDark),
}

// Builtins returns the names of the built-in themes.
func Builtins() []string {
	names := make([]string, 0, len(builtins))
	for n := range builtins {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Available lists the built-in themes followed by the user themes in dir.
func Available(dir string) []string {
	names := Builtins()
	entries, _ := os.ReadDir(dir) // a missing dir just means no user themes
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".yaml")
		if ok && !e.IsDir() && builtins[name] == nil {
			names = append(names, name)
		}
	}
	return names
}

// IsBuiltin reports whether name is a built-in theme.
func IsBuiltin(name string) bool { return builtins[name] != nil }

func c(s string) color.Color { return lipgloss.Color(s) }

func themeDefault(isDark bool) Scheme {
	if isDark {
		return Scheme{
			Primary:      c("205"),
			PrimaryMuted: c("132"),
			Accent:       c("#7571F9"),
			Highlight:    c("#FFFF00"),
			Text:         c("#FFFFFF"),
			TextMuted:    c("#C1C1C1"),
			TextFaint:    c("#AAAAAA"),
			OnPrimary:    c("#000000"),
			Surface:      c("#353533"),
			SurfaceAlt:   c("#343433"),
			BorderFocus:  c("#F5F5F5"),
			BorderBlur:   c("235"),
			Active:       c("#800080"),
			Success:      c("#00FF00"),
			Error:        c("#FF0000"),
			Syntax:       "github-dark",
			IsDark:       true,
		}
	}
	return Scheme{
		Primary:      c("162"),
		PrimaryMuted: c("132"),
		Accent:       c("#5A56E0"),
		Highlight:    c("#AF8700"),
		Text:         c("#1A1A1A"),
		TextMuted:    c("#555555"),
		TextFaint:    c("#777777"),
		OnPrimary:    c("#FFFFFF"),
		Surface:      c("#E4E4E4"),
		SurfaceAlt:   c("#D0D0D0"),
		BorderFocus:  c("#303030"),
		BorderBlur:   c("250"),
		Active:       c("#800080"),
		Success:      c("#008700"),
		Error:        c("#D70000"),
		Syntax:       "github",
	}
}

// themeDracula follows the Dracula spec (draculatheme.com/spec): purple is the
// main accent, pink and yellow highlight, comment is muted text. The light
// variant is Dracula's official light palette, Alucard.
func themeDracula(isDark bool) Scheme {
	if !isDark {
		return Scheme{
			Primary:      c("#644AC9"), // purple
			PrimaryMuted: c("#6C664B"), // comment
			Accent:       c("#A3144D"), // pink
			Highlight:    c("#846E15"), // yellow
			Text:         c("#1F1F1F"), // foreground
			TextMuted:    c("#6C664B"),
			TextFaint:    c("#6C664B"),
			OnPrimary:    c("#FFFBEB"), // background
			Surface:      c("#CFCFDE"), // selection
			SurfaceAlt:   c("#CFCFDE"),
			BorderFocus:  c("#644AC9"),
			BorderBlur:   c("#CFCFDE"),
			Active:       c("#A3144D"),
			Success:      c("#14710A"), // green
			Error:        c("#CB3A2A"), // red
			Syntax:       "github",     // chroma has no Alucard style
		}
	}
	return Scheme{
		Primary:      c("#BD93F9"), // purple
		PrimaryMuted: c("#6272A4"), // comment
		Accent:       c("#FF79C6"), // pink
		Highlight:    c("#F1FA8C"), // yellow
		Text:         c("#F8F8F2"), // foreground
		TextMuted:    c("#6272A4"),
		TextFaint:    c("#6272A4"),
		OnPrimary:    c("#282A36"), // background
		Surface:      c("#44475A"), // current line / selection
		SurfaceAlt:   c("#44475A"),
		BorderFocus:  c("#BD93F9"),
		BorderBlur:   c("#44475A"),
		Active:       c("#FF79C6"),
		Success:      c("#50FA7B"), // green
		Error:        c("#FF5555"), // red
		Syntax:       "dracula",
		IsDark:       true,
	}
}

// catppuccinScheme maps a Catppuccin flavor (catppuccin.com/palette) to the
// roles from its style guide (catppuccin.com/docs/style-guide): Base background,
// Mauve accent, Lavender active border, Overlay0 inactive border, Mantle bars.
func catppuccinScheme(f catppuccin.Flavour) Scheme {
	h := func(col catppuccin.Color) color.Color { return c(col.Hex) }
	return Scheme{
		Primary:      h(f.Mauve()),
		PrimaryMuted: h(f.Overlay2()),
		Accent:       h(f.Blue()),
		Highlight:    h(f.Peach()),
		Text:         h(f.Text()),
		TextMuted:    h(f.Subtext0()),
		TextFaint:    h(f.Overlay1()),
		OnPrimary:    h(f.Crust()),
		Surface:      h(f.Mantle()),
		SurfaceAlt:   h(f.Surface0()),
		BorderFocus:  h(f.Lavender()),
		BorderBlur:   h(f.Overlay0()),
		Active:       h(f.Mauve()),
		Success:      h(f.Green()),
		Error:        h(f.Red()),
		Background:   h(f.Base()),
		Syntax:       "catppuccin-" + f.Name(),
		IsDark:       f.Name() != catppuccin.Latte.Name(),
	}
}

// themeDesert is an earthy theme: muted sand gold and sage/olive green on
// warm brown (dark) or sand (light) backgrounds.
func themeDesert(isDark bool) Scheme {
	if !isDark {
		return Scheme{
			Primary:      c("#8A6D1F"), // mustard
			PrimaryMuted: c("#A89466"),
			Accent:       c("#5B7A3A"), // olive
			Highlight:    c("#7A6A1E"),
			Text:         c("#3B3326"),
			TextMuted:    c("#6B604B"),
			TextFaint:    c("#8E8470"),
			OnPrimary:    c("#F5EFDF"),
			Surface:      c("#E6DCC3"),
			SurfaceAlt:   c("#D9CDB0"),
			BorderFocus:  c("#5B7A3A"),
			BorderBlur:   c("#D9CDB0"),
			Active:       c("#5B7A3A"),
			Success:      c("#4F7A2E"),
			Error:        c("#A3452F"), // terracotta
			Syntax:       "gruvbox-light",
		}
	}
	return Scheme{
		Primary:      c("#C8A951"), // sand gold
		PrimaryMuted: c("#8C7A4B"),
		Accent:       c("#8FA876"), // sage
		Highlight:    c("#B5B868"), // olive yellow
		Text:         c("#E8DFC8"),
		TextMuted:    c("#B3A88C"),
		TextFaint:    c("#857C66"),
		OnPrimary:    c("#2B2620"),
		Surface:      c("#3A342A"),
		SurfaceAlt:   c("#4A4335"),
		BorderFocus:  c("#8FA876"),
		BorderBlur:   c("#4A4335"),
		Active:       c("#6E8B4E"), // olive
		Success:      c("#7E9F5C"),
		Error:        c("#C0614A"), // terracotta
		Background:   c("#1C1813"), // dark warm brown
		Syntax:       "gruvbox",
		IsDark:       true,
	}
}

// roles maps the YAML key of every color role to its field.
func (s *Scheme) roles() map[string]*color.Color {
	return map[string]*color.Color{
		"primary":      &s.Primary,
		"primaryMuted": &s.PrimaryMuted,
		"accent":       &s.Accent,
		"highlight":    &s.Highlight,
		"text":         &s.Text,
		"textMuted":    &s.TextMuted,
		"textFaint":    &s.TextFaint,
		"onPrimary":    &s.OnPrimary,
		"surface":      &s.Surface,
		"surfaceAlt":   &s.SurfaceAlt,
		"borderFocus":  &s.BorderFocus,
		"borderBlur":   &s.BorderBlur,
		"active":       &s.Active,
		"success":      &s.Success,
		"error":        &s.Error,
		"background":   &s.Background,
	}
}

// themeFile is the format of ~/.config/ktea/themes/<name>.yaml.
type themeFile struct {
	Base   string            `yaml:"base"`
	Syntax string            `yaml:"syntax"`
	Dark   map[string]string `yaml:"dark"`
	Light  map[string]string `yaml:"light"`
}

// Load resolves a theme by name: a built-in, or a user theme file
// <dir>/<name>.yaml. An empty name means the default theme.
func Load(name, dir string, isDark bool) (Scheme, error) {
	if name == "" {
		name = DefaultTheme
	}
	if fn, ok := builtins[name]; ok {
		return fn(isDark), nil
	}

	path := filepath.Join(dir, name+".yaml")
	f, err := os.Open(path)
	if err != nil {
		return Scheme{}, fmt.Errorf("theme %q: %w", name, err)
	}
	defer f.Close()

	var tf themeFile
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&tf); err != nil && !errors.Is(err, io.EOF) {
		return Scheme{}, fmt.Errorf("theme %q (%s): %w", name, path, err)
	}
	return tf.resolve(name, isDark)
}

func (tf themeFile) resolve(name string, isDark bool) (Scheme, error) {
	base := tf.Base
	if base == "" {
		base = DefaultTheme
	}
	fn, ok := builtins[base]
	if !ok {
		return Scheme{}, fmt.Errorf("theme %q: unknown base %q (built-ins: %s)",
			name, base, strings.Join(Builtins(), ", "))
	}
	s := fn(isDark)

	overrides := tf.Dark
	if !isDark && tf.Light != nil {
		overrides = tf.Light
	}
	// Validate both sections so a typo is reported whatever the background.
	for sec, m := range map[string]map[string]string{"dark": tf.Dark, "light": tf.Light} {
		if err := validate(name, sec, m); err != nil {
			return Scheme{}, err
		}
	}
	roles := s.roles()
	for k, v := range overrides {
		*roles[k] = lipgloss.Color(v)
	}

	if tf.Syntax != "" {
		if _, ok := chroma.Registry[tf.Syntax]; !ok {
			return Scheme{}, fmt.Errorf("theme %q: unknown syntax style %q", name, tf.Syntax)
		}
		s.Syntax = tf.Syntax
	}
	return s, nil
}

func validate(name, section string, m map[string]string) error {
	roles := (&Scheme{}).roles()
	for k, v := range m {
		if _, ok := roles[k]; !ok {
			return fmt.Errorf("theme %q: unknown color %s.%s", name, section, k)
		}
		if lipgloss.Color(v) == (lipgloss.NoColor{}) {
			return fmt.Errorf("theme %q: invalid color %s.%s: %q (use #RRGGBB or 0-255)", name, section, k, v)
		}
	}
	return nil
}
