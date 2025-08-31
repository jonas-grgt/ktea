package statusbar

import (
	"fmt"
	lg "github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
)

type Model struct {
	provider      Provider
	showShortcuts bool
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
	if ktx.Config.HasClusters() {
		activeCluster = styles.Statusbar.
			Cluster(ktx.Config.ActiveCluster().Color).
			Render(ktx.Config.ActiveCluster().Name)
	}
	indicator := styles.Statusbar.Indicator.Render(m.provider.Title())

	var shortCuts string
	if m.showShortcuts {
		shortcuts := m.provider.Shortcuts()
		shortcuts = append([]Shortcut{
			{
				Name:       "Switch Tabs",
				Keybinding: "C-←/→/h/l",
			},
		}, shortcuts...)
		rowsPerColumn := 2 // Fixed maximum rows per column
		var columns int

		if len(shortcuts) <= 4 {
			columns = len(shortcuts)
			rowsPerColumn = 1
		} else {
			columns = (len(shortcuts) + rowsPerColumn - 1) / rowsPerColumn
		}

		// Organize shortcuts into columns
		var shortcutsTable [][]Shortcut
		for i := 0; i < rowsPerColumn; i++ {
			var row []Shortcut
			for j := 0; j < columns; j++ {
				index := j*rowsPerColumn + i
				if index < len(shortcuts) {
					row = append(row, shortcuts[index])
				}
			}
			shortcutsTable = append(shortcutsTable, row)
		}

		// Calculate the maximum width for names and keybindings per column
		nameWidths := make([]int, columns)
		keyWidths := make([]int, columns)
		for _, row := range shortcutsTable {
			for col, shortcut := range row {
				nameWidth := lg.Width(shortcut.Name)
				keyWidth := lg.Width(shortcut.Keybinding)
				if nameWidth > nameWidths[col] {
					nameWidths[col] = nameWidth
				}
				if keyWidth > keyWidths[col] {
					keyWidths[col] = keyWidth
				}
			}
		}

		// Build the shortcuts display
		var rows []string
		for _, row := range shortcutsTable {
			var rowCells []string
			for col, shortcut := range row {
				paddedName := fmt.Sprintf("%-*s", nameWidths[col], shortcut.Name)
				paddedKey := fmt.Sprintf("%-*s", keyWidths[col], shortcut.Keybinding)
				shortcutCell := fmt.Sprintf("%s: ≪ %s »   ",
					styles.Statusbar.BindingName.Render(paddedName),
					styles.Statusbar.Key.Render(paddedKey),
				)
				rowCells = append(rowCells, shortcutCell)
			}
			rows = append(rows, styles.Statusbar.Text.Render(lg.JoinHorizontal(lg.Left, rowCells...)))
		}

		shortCuts = lg.JoinVertical(lg.Top, rows...)
	} else {
		shortCuts = ""
	}

	leftover := ktx.WindowWidth - (lg.Width(activeCluster)) - (lg.Width(indicator))
	barView := lg.NewStyle().MarginTop(1).Render(lg.JoinHorizontal(lg.Top,
		activeCluster,
		indicator,
		styles.Statusbar.Spacer.Width(leftover).Render(""),
	))

	if shortCuts == "" {
		return renderer.Render(barView)
	} else {
		return renderer.Render(lg.JoinVertical(lg.Top, styles.Statusbar.Shortcuts.Render(shortCuts), barView))
	}
}

func (m *Model) ToggleShortcuts() {
	m.showShortcuts = !m.showShortcuts
}

func (m *Model) SetProvider(provider Provider) {
	m.provider = provider
}

func New() *Model {
	return &Model{showShortcuts: false}
}
