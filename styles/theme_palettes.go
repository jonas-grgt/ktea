package styles

// Popular editor themes, using each project's official palette. Every variant
// paints its own background and highlights payloads with the chroma style of
// the same name.

// fixed always uses s, whatever the terminal background.
func fixed(s Scheme) ThemeFunc { return func(bool) Scheme { return s } }

// adaptive picks dark or light from the terminal background.
func adaptive(dark, light Scheme) ThemeFunc {
	return func(isDark bool) Scheme {
		if isDark {
			return dark
		}
		return light
	}
}

// tokyoNight holds a Tokyo Night style (github.com/folke/tokyonight.nvim).
type tokyoNight struct {
	bg, bgDark, bgHighlight, fg, fgDark, comment, fgGutter string
	blue, cyan, magenta, orange, green, red                string
}

func (p tokyoNight) scheme(style string, isDark bool) Scheme {
	return Scheme{
		Primary:      c(p.blue),
		PrimaryMuted: c(p.comment),
		Accent:       c(p.magenta),
		Highlight:    c(p.orange),
		Text:         c(p.fg),
		TextMuted:    c(p.fgDark),
		TextFaint:    c(p.comment),
		OnPrimary:    c(p.bgDark),
		Surface:      c(p.bgDark),
		SurfaceAlt:   c(p.bgHighlight),
		BorderFocus:  c(p.cyan),
		BorderBlur:   c(p.fgGutter),
		Active:       c(p.blue),
		Success:      c(p.green),
		Error:        c(p.red),
		Background:   c(p.bg),
		Syntax:       "tokyonight-" + style,
		IsDark:       isDark,
	}
}

var (
	tokyoNightNight = tokyoNight{
		bg: "#1a1b26", bgDark: "#16161e", bgHighlight: "#292e42", fg: "#c0caf5", fgDark: "#a9b1d6",
		comment: "#565f89", fgGutter: "#3b4261", blue: "#7aa2f7", cyan: "#7dcfff", magenta: "#bb9af7",
		orange: "#ff9e64", green: "#9ece6a", red: "#f7768e",
	}.scheme("night", true)
	tokyoNightStorm = tokyoNight{
		bg: "#24283b", bgDark: "#1f2335", bgHighlight: "#292e42", fg: "#c0caf5", fgDark: "#a9b1d6",
		comment: "#565f89", fgGutter: "#3b4261", blue: "#7aa2f7", cyan: "#7dcfff", magenta: "#bb9af7",
		orange: "#ff9e64", green: "#9ece6a", red: "#f7768e",
	}.scheme("storm", true)
	tokyoNightMoon = tokyoNight{
		bg: "#222436", bgDark: "#1e2030", bgHighlight: "#2f334d", fg: "#c8d3f5", fgDark: "#828bb8",
		comment: "#636da6", fgGutter: "#3b4261", blue: "#82aaff", cyan: "#86e1fc", magenta: "#c099ff",
		orange: "#ff966c", green: "#c3e88d", red: "#ff757f",
	}.scheme("moon", true)
	tokyoNightDay = tokyoNight{
		bg: "#e1e2e7", bgDark: "#d0d5e3", bgHighlight: "#c4c8da", fg: "#3760bf", fgDark: "#6172b0",
		comment: "#848cb5", fgGutter: "#a8aecb", blue: "#2e7de9", cyan: "#007197", magenta: "#9854f1",
		orange: "#b15c00", green: "#587539", red: "#f52a65",
	}.scheme("day", false)
)

// gruvbox follows github.com/morhetz/gruvbox: orange and yellow on warm browns.
type gruvbox struct {
	bg0, bg1, bg2, bg3, fg, fg4, gray      string
	red, green, yellow, aqua, orange, blue string
}

func (p gruvbox) scheme(syntax string, isDark bool) Scheme {
	return Scheme{
		Primary:      c(p.orange),
		PrimaryMuted: c(p.gray),
		Accent:       c(p.aqua),
		Highlight:    c(p.yellow),
		Text:         c(p.fg),
		TextMuted:    c(p.fg4),
		TextFaint:    c(p.gray),
		OnPrimary:    c(p.bg0),
		Surface:      c(p.bg1),
		SurfaceAlt:   c(p.bg2),
		BorderFocus:  c(p.yellow),
		BorderBlur:   c(p.bg3),
		Active:       c(p.blue),
		Success:      c(p.green),
		Error:        c(p.red),
		Background:   c(p.bg0),
		Syntax:       syntax,
		IsDark:       isDark,
	}
}

var (
	gruvboxDark = gruvbox{
		bg0: "#282828", bg1: "#3c3836", bg2: "#504945", bg3: "#665c54", fg: "#ebdbb2", fg4: "#a89984",
		gray: "#928374", red: "#fb4934", green: "#b8bb26", yellow: "#fabd2f", aqua: "#8ec07c",
		orange: "#fe8019", blue: "#83a598",
	}.scheme("gruvbox", true)
	gruvboxLight = gruvbox{
		bg0: "#fbf1c7", bg1: "#ebdbb2", bg2: "#d5c4a1", bg3: "#bdae93", fg: "#3c3836", fg4: "#7c6f64",
		gray: "#928374", red: "#9d0006", green: "#79740e", yellow: "#b57614", aqua: "#427b58",
		orange: "#af3a03", blue: "#076678",
	}.scheme("gruvbox-light", false)
)

// nord follows nordtheme.com: Polar Night backgrounds, Frost accents.
var nord = Scheme{
	Primary:      c("#88c0d0"), // nord8
	PrimaryMuted: c("#81a1c1"), // nord9
	Accent:       c("#8fbcbb"), // nord7
	Highlight:    c("#ebcb8b"), // nord13
	Text:         c("#eceff4"), // nord6
	TextMuted:    c("#d8dee9"), // nord4
	TextFaint:    c("#616e88"), // comments
	OnPrimary:    c("#2e3440"), // nord0
	Surface:      c("#3b4252"), // nord1
	SurfaceAlt:   c("#434c5e"), // nord2
	BorderFocus:  c("#88c0d0"),
	BorderBlur:   c("#4c566a"), // nord3
	Active:       c("#5e81ac"), // nord10
	Success:      c("#a3be8c"), // nord14
	Error:        c("#bf616a"), // nord11
	Background:   c("#2e3440"),
	Syntax:       "nord",
	IsDark:       true,
}

// rosePine follows rosepinetheme.com/palette.
type rosePine struct {
	base, surface, overlay, muted, subtle, text string
	love, gold, rose, pine, foam, iris, hlMed   string
}

func (p rosePine) scheme(syntax string, isDark bool) Scheme {
	return Scheme{
		Primary:      c(p.rose),
		PrimaryMuted: c(p.subtle),
		Accent:       c(p.iris),
		Highlight:    c(p.gold),
		Text:         c(p.text),
		TextMuted:    c(p.subtle),
		TextFaint:    c(p.muted),
		OnPrimary:    c(p.base),
		Surface:      c(p.surface),
		SurfaceAlt:   c(p.overlay),
		BorderFocus:  c(p.iris),
		BorderBlur:   c(p.hlMed),
		Active:       c(p.pine),
		Success:      c(p.foam),
		Error:        c(p.love),
		Background:   c(p.base),
		Syntax:       syntax,
		IsDark:       isDark,
	}
}

var (
	rosePineMain = rosePine{
		base: "#191724", surface: "#1f1d2e", overlay: "#26233a", muted: "#6e6a86", subtle: "#908caa",
		text: "#e0def4", love: "#eb6f92", gold: "#f6c177", rose: "#ebbcba", pine: "#31748f",
		foam: "#9ccfd8", iris: "#c4a7e7", hlMed: "#403d52",
	}.scheme("rose-pine", true)
	rosePineMoon = rosePine{
		base: "#232136", surface: "#2a273f", overlay: "#393552", muted: "#6e6a86", subtle: "#908caa",
		text: "#e0def4", love: "#eb6f92", gold: "#f6c177", rose: "#ea9a97", pine: "#3e8fb0",
		foam: "#9ccfd8", iris: "#c4a7e7", hlMed: "#44415a",
	}.scheme("rose-pine-moon", true)
	rosePineDawn = rosePine{
		base: "#faf4ed", surface: "#fffaf3", overlay: "#f2e9e1", muted: "#9893a5", subtle: "#797593",
		text: "#575279", love: "#b4637a", gold: "#ea9d34", rose: "#d7827e", pine: "#286983",
		foam: "#56949f", iris: "#907aa9", hlMed: "#dfdad9",
	}.scheme("rose-pine-dawn", false)
)

// Solarized follows ethanschoonover.com/solarized: the dark and light modes
// swap the base tones and share the accents.
var (
	solarizedDark = Scheme{
		Primary:      c("#268bd2"), // blue
		PrimaryMuted: c("#6c71c4"), // violet
		Accent:       c("#2aa198"), // cyan
		Highlight:    c("#b58900"), // yellow
		Text:         c("#93a1a1"), // base1
		TextMuted:    c("#839496"), // base0
		TextFaint:    c("#586e75"), // base01
		OnPrimary:    c("#002b36"), // base03
		Surface:      c("#073642"), // base02
		SurfaceAlt:   c("#073642"),
		BorderFocus:  c("#93a1a1"),
		BorderBlur:   c("#586e75"),
		Active:       c("#268bd2"),
		Success:      c("#859900"), // green
		Error:        c("#dc322f"), // red
		Background:   c("#002b36"),
		Syntax:       "solarized-dark",
		IsDark:       true,
	}
	solarizedLight = Scheme{
		Primary:      c("#268bd2"),
		PrimaryMuted: c("#6c71c4"),
		Accent:       c("#2aa198"),
		Highlight:    c("#b58900"),
		Text:         c("#586e75"), // base01
		TextMuted:    c("#657b83"), // base00
		TextFaint:    c("#93a1a1"), // base1
		OnPrimary:    c("#fdf6e3"), // base3
		Surface:      c("#eee8d5"), // base2
		SurfaceAlt:   c("#eee8d5"),
		BorderFocus:  c("#586e75"),
		BorderBlur:   c("#93a1a1"),
		Active:       c("#268bd2"),
		Success:      c("#859900"),
		Error:        c("#dc322f"),
		Background:   c("#fdf6e3"),
		Syntax:       "solarized-light",
	}
)

// oneDark is Atom's One Dark (github.com/atom/one-dark-syntax).
var oneDark = Scheme{
	Primary:      c("#61afef"), // blue
	PrimaryMuted: c("#5c6370"),
	Accent:       c("#c678dd"), // purple
	Highlight:    c("#e5c07b"), // yellow
	Text:         c("#abb2bf"), // mono-1
	TextMuted:    c("#828997"), // mono-2
	TextFaint:    c("#5c6370"), // mono-3
	OnPrimary:    c("#282c34"),
	Surface:      c("#21252b"),
	SurfaceAlt:   c("#3e4451"),
	BorderFocus:  c("#56b6c2"), // cyan
	BorderBlur:   c("#3e4451"),
	Active:       c("#61afef"),
	Success:      c("#98c379"), // green
	Error:        c("#e06c75"), // red
	Background:   c("#282c34"),
	Syntax:       "onedark",
	IsDark:       true,
}
