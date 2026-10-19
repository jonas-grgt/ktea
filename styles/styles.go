package styles

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/table"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

var Env EnvStyle
var Tab TabStyle
var Statusbar StatusBarStyle
var Help HelpStyle
var Table TableStyle
var CmdBar lipgloss.Style
var Form = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingTop(1)
var TextViewPort = lipgloss.NewStyle().
	PaddingLeft(1).
	BorderStyle(lipgloss.RoundedBorder())
var Notifier NotifierStyle
var clusterColors = map[string]string{
	ColorRed:    ColorWhite,
	ColorYellow: ColorBlack,
	ColorGreen:  ColorBlack,
	ColorOrange: ColorBlack,
	ColorPurple: ColorWhite,
	ColorBlue:   ColorWhite,
}

// Current is the active color scheme. Set it through Apply.
var Current Scheme

// Huh is the form theme. huh reads it on every render, so forms follow Apply.
var Huh huh.Theme = huh.ThemeFunc(func(bool) *huh.Styles {
	t := *huhStyles // copy, so a form can't change the shared styles
	return &t
})

var huhStyles *huh.Styles

// Cluster environment colors. They are stored per cluster in config.yaml,
// so they are fixed and not part of a theme.
const ColorRed = "#FF0000"
const ColorGreen = "#00FF00"
const ColorBlue = "#0000FF"
const ColorOrange = "#FFA500"
const ColorPurple = "#800080"
const ColorYellow = "#FFFF00"
const ColorWhite = "#FFFFFF"
const ColorBlack = "#000000"

type NotifierStyle struct {
	Spinner lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
}

type TableStyle struct {
	Focus  lipgloss.Style
	Blur   lipgloss.Style
	Styles table.Styles
}

type StatusBarStyle struct {
	style     lipgloss.Style
	Indicator lipgloss.Style
	Spacer    lipgloss.Style
	cluster   lipgloss.Style
}

type HelpStyle struct {
	Modal  lipgloss.Style
	Title  lipgloss.Style
	Key    lipgloss.Style
	Desc   lipgloss.Style
	Footer lipgloss.Style
}

type BorderPosition int

const (
	TopLeftBorder BorderPosition = iota
	TopMiddleBorder
	TopRightBorder
	BottomLeftBorder
	BottomMiddleBorder
	BottomRightBorder
)

type EmbeddedTextFunc func(active bool) string

func EmbeddedBorderText(
	keyLabel string,
	valueLabel string,
) EmbeddedTextFunc {
	return func(active bool) string {
		colorLabel, colorCount := Current.TextMuted, Current.PrimaryMuted
		if active {
			colorLabel, colorCount = Current.Text, Current.Primary
		}

		var renderedValueLabel string
		if valueLabel == "" {
			renderedValueLabel = ""
		} else {
			renderedValueLabel = ":" + lipgloss.NewStyle().
				Foreground(colorCount).
				Bold(true).
				Render(fmt.Sprintf(" %s", valueLabel))
		}

		return lipgloss.NewStyle().
			Foreground(colorLabel).
			Bold(true).
			Render(fmt.Sprintf("[ %s", keyLabel)) + renderedValueLabel +
			lipgloss.NewStyle().
				Foreground(colorLabel).
				Bold(true).
				Render(" ]")
	}
}

func CenterText(width int, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		AlignVertical(lipgloss.Center).
		AlignHorizontal(lipgloss.Center).
		Bold(true).
		Foreground(Current.Primary)
}

func CmdBarWithWidth(width int) lipgloss.Style {
	// lipgloss v2 widths include the border; keep width as the content+padding width.
	return CmdBar.Width(width + CmdBar.GetHorizontalBorderSize())
}

func FG(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

func (s *StatusBarStyle) Cluster(color string) lipgloss.Style {
	return s.cluster.
		Background(lipgloss.Color(color)).
		Foreground(lipgloss.Color(clusterColors[color]))
}

type TabStyle struct {
	Tab    lipgloss.Style
	Active lipgloss.Style
	Help   lipgloss.Style
}

type EnvStyle struct {
	Colors struct {
		Red    lipgloss.Style
		Green  lipgloss.Style
		Blue   lipgloss.Style
		Orange lipgloss.Style
		Yellow lipgloss.Style
		Purple lipgloss.Style
	}
}

func (s *StatusBarStyle) Render(strs ...string) string {
	return s.style.Render(strs...)
}

func init() {
	Apply(themeDefault(true))
}

// Apply makes s the active scheme and rebuilds all shared styles from it.
// Call it before building UI components; styles already rendered keep their colors.
func Apply(s Scheme) {
	Current = s
	huhStyles = huhTheme(s)

	{
		tabStyle := TabStyle{}
		activeTabBorder := lipgloss.Border{
			Top:         "─",
			Bottom:      " ",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┘",
			BottomRight: "└",
		}
		tabBorder := lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┴",
			BottomRight: "┴",
		}
		helpBorder := lipgloss.Border{
			Top:         " ",
			Bottom:      "─",
			Left:        " ",
			Right:       " ",
			TopLeft:     " ",
			TopRight:    " ",
			BottomLeft:  "─",
			BottomRight: "─",
		}
		tabStyle.Tab = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(s.TextFaint).
			Border(tabBorder)
		tabStyle.Active = lipgloss.NewStyle().
			Padding(0, 1).
			Border(activeTabBorder).
			Foreground(s.Primary).
			Bold(true)
		tabStyle.Help = lipgloss.NewStyle().
			Padding(0, 1).
			Border(helpBorder).
			Foreground(s.Text).
			Bold(true)
		Tab = tabStyle
	}

	{
		statusBarStyle := StatusBarStyle{}
		statusBarStyle.style = lipgloss.NewStyle().
			Background(s.Surface)

		statusBarStyle.Indicator = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			Foreground(s.OnPrimary).
			Background(s.Primary).
			Bold(true).
			Padding(0, 2)

		statusBarStyle.cluster = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorRed)).
			Bold(true).
			Padding(0, 4)

		statusBarStyle.Spacer = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			PaddingLeft(1)

		Statusbar = statusBarStyle
	}

	{
		Help = HelpStyle{
			Modal: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(s.Primary).
				Padding(1, 2),
			Title: lipgloss.NewStyle().
				Foreground(s.Primary).
				Bold(true),
			Key: lipgloss.NewStyle().
				Foreground(s.Highlight).
				Bold(true),
			Desc: lipgloss.NewStyle().
				Foreground(s.Text),
			Footer: lipgloss.NewStyle().
				Foreground(s.TextMuted).
				Faint(true),
		}
	}

	{
		CmdBar = lipgloss.NewStyle().
			MarginTop(0).
			MarginBottom(0).
			BorderStyle(lipgloss.ThickBorder())
	}

	{
		Notifier.Spinner = lipgloss.NewStyle().
			MarginTop(0).
			MarginBottom(0).
			MarginLeft(2).
			Height(1)
		Notifier.Success = lipgloss.NewStyle().
			MarginTop(0).
			MarginBottom(0).
			MarginLeft(2).
			Height(1).
			Foreground(s.Success).
			Bold(true)
		Notifier.Error = lipgloss.NewStyle().
			MarginTop(0).
			MarginBottom(0).
			MarginLeft(2).
			Height(1).
			Bold(true)

	}

	{
		envStyle := EnvStyle{}
		envStyle.Colors.Red = lipgloss.NewStyle().
			Foreground(lipgloss.Color(clusterColors[ColorRed])).
			Background(lipgloss.Color(ColorRed)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Green = lipgloss.NewStyle().
			Foreground(lipgloss.Color(clusterColors[ColorGreen])).
			Background(lipgloss.Color(ColorGreen)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Blue = lipgloss.NewStyle().
			Foreground(lipgloss.Color(clusterColors[ColorBlue])).
			Background(lipgloss.Color(ColorBlue)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Orange = lipgloss.NewStyle().
			Foreground(lipgloss.Color(clusterColors[ColorOrange])).
			Background(lipgloss.Color(ColorOrange)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Purple = lipgloss.NewStyle().
			Foreground(lipgloss.Color(clusterColors[ColorPurple])).
			Background(lipgloss.Color(ColorPurple)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Yellow = lipgloss.NewStyle().
			Foreground(lipgloss.Color(clusterColors[ColorYellow])).
			Background(lipgloss.Color(ColorYellow)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)

		Env = envStyle
	}

	{
		blur := lipgloss.NewStyle().
			Padding(0).
			Margin(0)
		focus := lipgloss.NewStyle().
			Padding(0).
			Margin(0).
			Inherit(blur)
		tableStyles := table.DefaultStyles()
		tableStyles.Header = tableStyles.Header.
			BorderBottom(true).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false).
			BorderStyle(lipgloss.Border{
				Bottom: "─",
			}).
			Bold(false)
		tableStyles.Selected = tableStyles.Selected.
			Foreground(s.OnPrimary).
			Background(s.Primary).
			Bold(true)
		Table = TableStyle{
			Focus:  focus,
			Blur:   blur,
			Styles: tableStyles,
		}
	}
}

// huhTheme starts from huh's Charm theme (ktea's look so far) and swaps in
// the scheme's colors.
func huhTheme(s Scheme) *huh.Styles {
	{
		t := huh.ThemeCharm(s.IsDark)
		f := &t.Focused
		f.Title = f.Title.Foreground(s.Accent)
		f.NoteTitle = f.NoteTitle.Foreground(s.Accent)
		f.Directory = f.Directory.Foreground(s.Accent)
		f.Description = f.Description.Foreground(s.TextMuted)
		f.ErrorIndicator = f.ErrorIndicator.Foreground(s.Error)
		f.ErrorMessage = f.ErrorMessage.Foreground(s.Error)
		f.SelectSelector = f.SelectSelector.Foreground(s.Primary)
		f.NextIndicator = f.NextIndicator.Foreground(s.Primary)
		f.PrevIndicator = f.PrevIndicator.Foreground(s.Primary)
		f.Option = f.Option.Foreground(s.Text)
		f.MultiSelectSelector = f.MultiSelectSelector.Foreground(s.Primary)
		f.SelectedOption = f.SelectedOption.Foreground(s.Success)
		f.SelectedPrefix = f.SelectedPrefix.Foreground(s.Success)
		f.UnselectedOption = f.UnselectedOption.Foreground(s.Text)
		f.UnselectedPrefix = f.UnselectedPrefix.Foreground(s.TextFaint)
		f.FocusedButton = f.FocusedButton.Foreground(s.OnPrimary).Background(s.Primary)
		f.Next = f.FocusedButton
		f.BlurredButton = f.BlurredButton.Foreground(s.Text).Background(s.SurfaceAlt)
		f.TextInput.Cursor = f.TextInput.Cursor.Foreground(s.Success)
		f.TextInput.Placeholder = f.TextInput.Placeholder.Foreground(s.TextFaint)
		f.TextInput.Prompt = f.TextInput.Prompt.Foreground(s.Primary)

		t.Blurred = t.Focused
		t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
		t.Blurred.Card = t.Blurred.Base
		t.Blurred.NextIndicator = lipgloss.NewStyle()
		t.Blurred.PrevIndicator = lipgloss.NewStyle()
		t.Group.Title = t.Focused.Title
		t.Group.Description = t.Focused.Description
		return t
	}
}
