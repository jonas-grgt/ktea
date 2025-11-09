package chips

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"strings"
)

type Model struct {
	label        string
	elems        []string
	activateElem int
	selectedIdx  int
}

func (m *Model) View(_ *kontext.ProgramKtx, _ *ui.Renderer) string {
	builder := strings.Builder{}
	builder.WriteString(m.label + ":")

	for i, elem := range m.elems {
		var (
			style   lipgloss.Style
			bgColor lipgloss.Color
		)

		if m.activateElem == i {
			if m.selectedIdx == i {
				bgColor = styles.ColorPink
			} else {
				bgColor = styles.ColorWhite
			}
			style = lipgloss.NewStyle().
				Background(bgColor).
				Foreground(lipgloss.Color(styles.ColorBlack)).
				Bold(true)
			elem = fmt.Sprintf("«%s»", elem)
		} else if m.selectedIdx == i {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(styles.ColorPink)).
				Foreground(lipgloss.Color(styles.ColorBlack))
			elem = fmt.Sprintf(" %s ", elem)
		} else {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(styles.ColorGrey)).
				Foreground(lipgloss.Color(styles.ColorBlack))
			elem = fmt.Sprintf(" %s ", elem)
		}

		builder.WriteString(
			style.
				Padding(0, 1).
				MarginLeft(1).
				MarginRight(0).
				Render(elem),
		)
	}

	return builder.String()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			m.prevElem()
		case "l", "right":
			m.nextElem()
		case "enter":
			m.activateElem = m.selectedIdx
		}
	}
	return nil
}

func (m *Model) ActivateByLabel(label string) {
	for i, elem := range m.elems {
		if elem == label {
			m.activateElem = i
			m.selectedIdx = i
		}
	}
}

func (m *Model) prevElem() {
	if m.selectedIdx >= 1 {
		m.selectedIdx--
	}
}

func (m *Model) nextElem() {
	if m.selectedIdx < len(m.elems)-1 {
		m.selectedIdx++
	}
}

func (m *Model) SelectedLabel() string {
	if m == nil || len(m.elems) == 0 {
		return ""
	}
	return m.elems[m.selectedIdx]
}

func New(
	label string,
	elems ...string,
) *Model {
	return &Model{
		label,
		elems,
		0,
		0,
	}
}
