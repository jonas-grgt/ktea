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
	Focused      bool
	tabs         []string
	onTabChanged OnTabChangedFunc
	textByPos    map[Position]TextFunc
	activeTabIdx int
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

func WithTitle(title string) Option {
	return func(m *Model) {
		m.textByPos[TopMiddleBorder] = func(_ *Model) string {
			return title
		}
	}
}

func WithTabs(tabs ...string) Option {
	return func(m *Model) {
		if len(tabs) == 0 {
			return
		}
		m.tabs = tabs
		m.textByPos[TopLeftBorder] = func(m *Model) string {

			var renderedTabs string
			for i, tab := range tabs {

				var padding string
				if i != 0 {
					padding = " "
				}

				if m.activeTabIdx == i {
					renderedTabs += padding + lipgloss.NewStyle().
						Background(lipgloss.Color(styles.ColorPurple)).
						Padding(0).
						Render(tab)
				} else {
					renderedTabs += padding + lipgloss.NewStyle().
						Padding(0).
						Render(tab)
				}
			}
			return fmt.Sprintf("[ %s ]", renderedTabs)
		}
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

	for _, option := range options {
		option(m)
	}

	return m
}
