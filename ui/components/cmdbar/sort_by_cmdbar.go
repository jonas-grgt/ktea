package cmdbar

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"strings"
)

const (
	Asc       Direction = true
	Desc      Direction = false
	AscLabel            = "▲"
	DescLabel           = "▼"
)

type SortByCmdBar struct {
	sorts                []SortLabel
	selectedIdx          int
	activeIdx            int
	Active               bool
	sortSelectedCallback SortSelectedCallback
}

type Direction bool

type SortByCmdBarOption func(*SortByCmdBar)

type SortLabel struct {
	Label     string
	Direction Direction
}

type SortSelectedCallback func(label SortLabel)

func (d Direction) String() string {
	if d == Asc {
		return AscLabel
	}
	return DescLabel
}

func (m *SortByCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	builder := strings.Builder{}

	for i, sort := range m.sorts {
		var (
			style   lipgloss.Style
			bgColor lipgloss.Color
			render  string
			arrow   string
		)

		if sort.Direction == Asc {
			arrow = AscLabel
		} else {
			arrow = DescLabel
		}

		if m.activeIdx == i {
			if m.selectedIdx == i {
				bgColor = styles.ColorLightPink
			} else {
				bgColor = styles.ColorWhite
			}
			style = lipgloss.NewStyle().
				Background(bgColor).
				Foreground(lipgloss.Color(styles.ColorBlack))

			render = fmt.Sprintf(" %s %s ", sort.Label, arrow)
		} else if m.selectedIdx == i {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(styles.ColorPink)).
				Foreground(lipgloss.Color(styles.ColorBlack))
			render = fmt.Sprintf(" %s %s ", sort.Label, arrow)
		} else {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(styles.ColorDarkGrey)).
				Foreground(lipgloss.Color(styles.ColorWhite))
			render = fmt.Sprintf(" %s %s ", sort.Label, arrow)
		}

		builder.WriteString(
			style.
				Padding(0, 1).
				MarginLeft(1).
				MarginRight(0).
				Render(render),
		)
	}

	return renderer.RenderWithStyle(builder.String(), styles.CmdBarWithWidth(ktx.WindowWidth-BorderedPadding))
}

func (m *SortByCmdBar) Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f3":
			m.Active = !m.Active
		case "h", "left":
			m.prevElem()
		case "l", "right":
			m.nextElem()
		case "esc":
			m.Active = false
			return m.Active, nil, nil
		case "enter":
			if m.activeIdx == m.selectedIdx {
				m.sorts[m.selectedIdx].Direction = !m.sorts[m.selectedIdx].Direction
			} else {
				m.activeIdx = m.selectedIdx
			}
			if m.sortSelectedCallback != nil {
				m.sortSelectedCallback(m.sorts[m.selectedIdx])
			}
		}
	}
	return m.Active, nil, nil
}

func (m *SortByCmdBar) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Move", "←/→"},
		{"Select sorting", "enter"},
		{"Toggle direction", "enter"},
		{"Cancel", "esc/F3"},
	}
}

func (m *SortByCmdBar) IsFocussed() bool {
	return true
}

func (m *SortByCmdBar) prevElem() {
	if m.selectedIdx >= 1 {
		m.selectedIdx--
	}
}

func (m *SortByCmdBar) nextElem() {
	if m.selectedIdx < len(m.sorts)-1 {
		m.selectedIdx++
	}
}

func (m *SortByCmdBar) SortedBy() SortLabel {
	return m.sorts[m.activeIdx]
}

func (m *SortByCmdBar) PrefixSortIcon(title string) string {
	sb := m.SortedBy()
	if sb.Label == title {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorPink)).
			Bold(true).
			Render(sb.Direction.String()) + " " + lipgloss.NewStyle().Bold(true).Render(title)
	}
	return lipgloss.NewStyle().Bold(true).Render(title)
}

func WithSortSelectedCallback(callback SortSelectedCallback) SortByCmdBarOption {
	return func(bar *SortByCmdBar) {
		bar.sortSelectedCallback = callback
	}
}

func WithInitialSortColumn(column string, direction Direction) SortByCmdBarOption {
	return func(bar *SortByCmdBar) {
		for i, sort := range bar.sorts {
			if sort.Label == column {
				bar.selectedIdx = i
				bar.activeIdx = i
				bar.sorts[i].Direction = direction
				return
			}
		}
	}
}

func NewSortByCmdBar(
	sorts []SortLabel,
	options ...SortByCmdBarOption,
) *SortByCmdBar {
	bar := SortByCmdBar{
		sorts:  sorts,
		Active: false,
	}

	for _, option := range options {
		option(&bar)
	}

	return &bar
}
