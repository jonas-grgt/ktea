package styles

import (
	"ktea/kontext"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var Env EnvStyle
var Tab TabStyle
var Statusbar StatusBarStyle
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

const ColorRed = "#FF0000"
const ColorGreen = "#00FF00"
const ColorBlue = "#0000FF"
const ColorOrange = "#FFA500"
const ColorPurple = "#800080"
const ColorYellow = "#FFFF00"
const ColorWhite = "#FFFFFF"
const ColorBlack = "#000000"
const ColorPink = "205"
const ColorGrey = "#C1C1C1"
const ColorFocusBorder = "#F5F5F5"
const ColorBlurBorder = "235"

var Bold = lipgloss.NewStyle().
	Bold(true)

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
	style       lipgloss.Style
	Indicator   lipgloss.Style
	Text        lipgloss.Style
	BindingName lipgloss.Style
	Shortcuts   lipgloss.Style
	Key         lipgloss.Style
	Spacer      lipgloss.Style
	cluster     lipgloss.Style
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

// Borderize creates a border around content with optional embedded text in different positions
func Borderize(content string, active bool, embeddedText map[BorderPosition]string) string {
	if embeddedText == nil {
		embeddedText = make(map[BorderPosition]string)
	}

	borderColor := ColorFocusBorder
	if !active {
		borderColor = ColorBlurBorder
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))

	// Split content into lines to get the maximum width
	lines := strings.Split(content, "\n")
	maxWidth := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}

	encloseText := func(text string) string {
		if text != "" {
			return style.Render(" " + text + " ")
		}
		return text
	}

	buildBorderLine := func(leftText, middleText, rightText, leftCorner, border, rightCorner string) string {
		leftText = encloseText(leftText)
		middleText = encloseText(middleText)
		rightText = encloseText(rightText)

		// Calculate remaining space for borders
		remaining := maxWidth - lipgloss.Width(leftText) - lipgloss.Width(middleText) - lipgloss.Width(rightText)
		if remaining < 0 {
			remaining = 0
		}

		leftBorderLen := (remaining / 2)
		rightBorderLen := remaining - leftBorderLen

		// Build the border line
		borderLine := leftText +
			style.Render(strings.Repeat(border, leftBorderLen)) +
			middleText +
			style.Render(strings.Repeat(border, rightBorderLen)) +
			rightText

		// Add corners
		return style.Render(leftCorner) + borderLine + style.Render(rightCorner)
	}

	// Create the bordered content
	topBorder := buildBorderLine(
		embeddedText[TopLeftBorder],
		embeddedText[TopMiddleBorder],
		embeddedText[TopRightBorder],
		"╭", "─", "╮",
	)

	// Create side borders for content
	borderedLines := make([]string, len(lines))
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		var paddedLine string
		if lineWidth < maxWidth {
			paddedLine = line + strings.Repeat(" ", maxWidth-lineWidth)
		} else if lineWidth > maxWidth {
			paddedLine = lipgloss.NewStyle().MaxWidth(maxWidth).Render(line)
		} else {
			paddedLine = line
		}
		borderedLines[i] = style.Render("│") + paddedLine + style.Render("│")
	}
	borderedContent := strings.Join(borderedLines, "\n")

	// Create bottom border
	bottomBorder := buildBorderLine(
		embeddedText[BottomLeftBorder],
		embeddedText[BottomMiddleBorder],
		embeddedText[BottomRightBorder],
		"╰", "─", "╯",
	)

	// Final content with borders
	return topBorder + "\n" + borderedContent + "\n" + bottomBorder
}

func CenterText(width int, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		AlignVertical(lipgloss.Center).
		AlignHorizontal(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color(ColorPink))
}

func CmdBarWithWidth(width int) lipgloss.Style {
	return CmdBar.Width(width)
}

func FG(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(color)
}

func (s *StatusBarStyle) Cluster(color string) lipgloss.Style {
	return s.cluster.
		Background(lipgloss.Color(color)).
		Foreground(lipgloss.Color(clusterColors[color]))
}

type TabStyle struct {
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style
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

func (s *StatusBarStyle) Render(c *kontext.ProgramKtx, strs ...string) string {
	return s.style.Render(strs...)
}

func init() {
	{
		clusterColors = map[string]string{
			ColorRed:    ColorWhite,
			ColorGreen:  ColorBlack,
			ColorBlue:   ColorWhite,
			ColorOrange: ColorBlack,
			ColorPurple: ColorWhite,
			ColorYellow: ColorBlack,
		}
	}

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
		tabStyle.Tab = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#AAAAAA")).
			Border(tabBorder)
		tabStyle.ActiveTab = lipgloss.NewStyle().
			Padding(0, 1).
			Border(activeTabBorder).
			Foreground(lipgloss.Color(ColorPink)).
			Bold(true)
		Tab = tabStyle
	}

	{
		statusBarStyle := StatusBarStyle{}
		statusBarStyle.style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

		statusBarStyle.Indicator = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			Foreground(lipgloss.Color(ColorBlack)).
			Background(lipgloss.Color(ColorPink)).
			Bold(true).
			Padding(0, 2)

		statusBarStyle.cluster = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorRed)).
			Bold(true).
			Padding(0, 4)

		statusBarStyle.Text = lipgloss.NewStyle().
			MarginTop(1).
			PaddingLeft(0).
			PaddingRight(5)

		statusBarStyle.BindingName = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow))

		statusBarStyle.Key = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite))

		statusBarStyle.Spacer = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			PaddingLeft(1)

		statusBarStyle.Shortcuts = lipgloss.NewStyle().
			PaddingLeft(2)

		Statusbar = statusBarStyle
	}

	{
		CmdBar = lipgloss.NewStyle().
			MarginTop(0).
			MarginBottom(0).
			BorderStyle(lipgloss.RoundedBorder())
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
			Foreground(lipgloss.Color(ColorGreen)).
			Bold(true)
		Notifier.Error = lipgloss.NewStyle().
			MarginTop(0).
			MarginBottom(0).
			MarginLeft(2).
			Height(1).
			Foreground(lipgloss.Color(ColorRed)).
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
		styles := table.DefaultStyles()
		styles.Header = styles.Header.
			BorderBottom(true).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false).
			BorderStyle(lipgloss.Border{
				Bottom: "─",
			}).
			Bold(false)
		styles.Selected = styles.Selected.
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorPink)).
			Bold(true)
		Table = TableStyle{
			Focus:  focus,
			Blur:   blur,
			Styles: styles,
		}
	}
}
