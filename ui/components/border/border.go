package border

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"ktea/styles"
	"strings"
)

const (
	TopLeftBorder Position = iota
	TopMiddleBorder
	TopRightBorder
	BottomLeftBorder
	BottomMiddleBorder
	BottomRightBorder
)

type Model struct {
	Focused       bool
	tabs          []Tab
	onTabChanged  OnTabChangedFunc
	textByPos     map[Position]TextFunc
	activeTabIdx  int
	activeColor   lipgloss.Color
	inActiveColor lipgloss.Color
}

type Label string

type Tab struct {
	Title string
	Label
}

type Position int

type TextFunc func(m *Model) string

type OnTabChangedFunc func(newTab string, m *Model)

type Option func(m *Model)

func (m *Model) View(content string) string {
	return m.borderize(content)
}

func (m *Model) encloseText(text string) string {
	if text != "" {
		return " " + text + " "
	}
	return text
}

func (m *Model) buildBorderLine(
	style lipgloss.Style,
	maxWidth int,
	leftText, middleText, rightText, leftCorner, border, rightCorner string,
) string {
	leftText = m.encloseText(leftText)
	middleText = m.encloseText(middleText)
	rightText = m.encloseText(rightText)

	// Calculate remaining space for borders
	remaining := maxWidth - lipgloss.Width(leftText) - lipgloss.Width(middleText) - lipgloss.Width(rightText)
	if remaining < 0 {
		remaining = 0
	}

	leftBorderLen := remaining / 2
	rightBorderLen := remaining - leftBorderLen

	// Build the borderline
	borderLine := leftText +
		style.Render(strings.Repeat(border, leftBorderLen)) +
		middleText +
		style.Render(strings.Repeat(border, rightBorderLen)) +
		rightText

	// Add corners
	return style.Render(leftCorner) + borderLine + style.Render(rightCorner)
}

func (m *Model) borderize(content string) string {

	borderColor := styles.ColorFocusBorder
	if !m.Focused {
		borderColor = styles.ColorBlurBorder
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

	// Create the bordered content
	topBorder := m.buildBorderLine(
		style,
		maxWidth,
		m.getTextOrEmpty(m.textByPos[TopLeftBorder]),
		m.getTextOrEmpty(m.textByPos[TopMiddleBorder]),
		m.getTextOrEmpty(m.textByPos[TopRightBorder]),
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
	bottomBorder := m.buildBorderLine(
		style,
		maxWidth,
		m.getTextOrEmpty(m.textByPos[BottomLeftBorder]),
		m.getTextOrEmpty(m.textByPos[BottomMiddleBorder]),
		m.getTextOrEmpty(m.textByPos[BottomRightBorder]),
		"╰", "─", "╯",
	)

	// Final content with borders
	return topBorder + "\n" + borderedContent + "\n" + bottomBorder
}

func (m *Model) getTextOrEmpty(embeddedText TextFunc) string {
	if embeddedText == nil {
		return ""
	}
	return embeddedText(m)
}

func (m *Model) NextTab() {
	if m.activeTabIdx == len(m.tabs)-1 {
		m.activeTabIdx = 0
	} else {
		m.activeTabIdx++
	}
}

func (m *Model) GoTo(label Label) {
	for i, tab := range m.tabs {
		if tab.Label == label {
			m.activeTabIdx = i
			break
		}
	}
}

func (m *Model) WithInActiveColor(c lipgloss.Color) {
	m.inActiveColor = c
}

func WithTitle(title string) Option {
	return func(m *Model) {
		m.textByPos[TopMiddleBorder] = func(_ *Model) string {
			return title
		}
	}
}

func WithTitleFunc(titleFunc func() string) Option {
	return func(m *Model) {
		m.textByPos[TopMiddleBorder] = func(_ *Model) string {
			return titleFunc()
		}
	}
}

func WithTabs(tabs ...Tab) Option {
	return func(m *Model) {
		if len(tabs) == 0 {
			return
		}
		m.tabs = tabs
		m.textByPos[TopLeftBorder] = func(m *Model) string {

			var renderedTabs string
			for i, tab := range tabs {

				if m.activeTabIdx == i {
					renderedTabs += lipgloss.NewStyle().
						Bold(true).
						Background(m.activeColor).
						Padding(0, 1).
						Render(tab.Title)
				} else {
					renderedTabs += lipgloss.NewStyle().
						Padding(0, 1).
						Foreground(m.inActiveColor).
						Render(tab.Title)
				}
			}
			return fmt.Sprintf("|%s|", renderedTabs)
		}
	}
}

func WithActiveColor(c lipgloss.Color) Option {
	return func(m *Model) {
		m.activeColor = c
	}
}

func WithInactiveColor(c lipgloss.Color) Option {
	return func(m *Model) {
		m.inActiveColor = c
	}
}

func WithOnTabChanged(o OnTabChangedFunc) Option {
	return func(m *Model) {
		m.onTabChanged = o
	}
}

func New(options ...Option) *Model {
	m := &Model{}
	m.textByPos = make(map[Position]TextFunc)
	m.Focused = true
	m.activeColor = styles.ColorPurple

	for _, option := range options {
		option(m)
	}

	return m
}
