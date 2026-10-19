package statusbar

import (
	lg "charm.land/lipgloss/v2"
	"image/color"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
)

type Model struct {
	provider Provider
}

type Provider interface {
	Shortcuts() []Shortcut
	Title() string
}

type UpdateMsg struct {
	StatusBar Provider
}

type Shortcut struct {
	Name       string
	Keybinding string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var activeCluster string
	if ktx.Config().HasClusters() {
		activeCluster = styles.Statusbar.
			Cluster(ktx.Config().ActiveCluster().Color).
			Render(ktx.Config().ActiveCluster().Name)
		clusterColor := ktx.Config().ActiveCluster().Color

		rightSeparator := "\uE0B0"
		if ktx.Config().PlainFonts {
			rightSeparator = ""
		}
		leftSeparator := "\uE0B6"
		if ktx.Config().PlainFonts {
			leftSeparator = ""
		}
		clusterRight := renderRune(rightSeparator, lg.Color(clusterColor), styles.Current.Primary)
		clusterLeft := renderRune(leftSeparator, lg.Color(clusterColor), lg.NoColor{})

		activeCluster = clusterLeft + activeCluster + clusterRight
	}
	indicator := styles.Statusbar.Indicator.Render(m.provider.Title())

	endSeparator := "\uE0B4"
	if ktx.Config().PlainFonts {
		endSeparator = ""
	}

	indicator += renderRune(endSeparator, styles.Current.Primary, styles.Current.Surface)

	leftover := ktx.WindowWidth - lg.Width(activeCluster) - lg.Width(indicator)

	if !ktx.Config().PlainFonts {
		leftover--
	}

	barView := lg.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		Render(lg.JoinHorizontal(lg.Top,
			activeCluster,
			indicator,
			styles.Statusbar.Spacer.Width(leftover).Render(""),
			renderText(endSeparator, styles.Current.Surface),
		))

	return renderer.Render(barView)
}

func renderRune(symbol string, fg, bg color.Color) string {
	return lg.NewStyle().
		Foreground(fg).
		Background(bg).
		Render(symbol)
}

func renderText(text string, fg color.Color) string {
	return lg.NewStyle().
		Foreground(fg).
		Render(text)
}

// Shortcuts returns the global shortcuts followed by those of the active provider.
func (m *Model) Shortcuts() []Shortcut {
	shortcuts := []Shortcut{{Name: "Switch Tabs", Keybinding: "C-←/→/h/l"}}
	if m.provider != nil {
		shortcuts = append(shortcuts, m.provider.Shortcuts()...)
	}
	return shortcuts
}

func (m *Model) Title() string {
	if m.provider == nil {
		return ""
	}
	return m.provider.Title()
}

func (m *Model) SetProvider(provider Provider) {
	m.provider = provider
}

func New() *Model {
	return &Model{}
}
