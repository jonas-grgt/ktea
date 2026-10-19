package settings_page

import (
	"fmt"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model lists the available themes. Enter applies the selected theme
// immediately and saves it to the config.
type Model struct {
	ktx       *kontext.ProgramKtx
	themesDir string
	// isDark is the terminal's background, not the active theme's: a fixed
	// theme like catppuccin-latte must not make the next theme load its light variant.
	isDark bool
	themes []string
	table  table.Model
	border *border.Model
	err    string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	current := m.currentTheme()
	rows := make([]table.Row, len(m.themes))
	for i, name := range m.themes {
		marker, kind := "", "custom"
		if name == current {
			marker = "●"
		}
		if styles.IsBuiltin(name) {
			kind = "built-in"
		}
		rows[i] = table.Row{marker, name, kind}
	}

	height := ktx.AvailableHeight - 2
	if m.err != "" {
		height--
	}
	m.table.SetColumns([]table.Column{
		{Title: "", Width: 2},
		{Title: "Theme", Width: max(ktx.WindowWidth-20, 10)},
		{Title: "Type", Width: 10},
	})
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetHeight(height)
	m.table.SetRows(rows)

	view := m.border.View(ktable.View(&m.table))
	if m.err != "" {
		view = lipgloss.JoinVertical(lipgloss.Left, view, styles.FG(styles.Current.Error).Render(" "+m.err))
	}
	return renderer.Render(view)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ui.RegainedFocusMsg:
		// pick up theme files added while ktea runs
		m.themes = styles.Available(m.themesDir)
		return nil
	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEnter {
			m.apply()
			return nil
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

func (m *Model) apply() {
	row := m.table.SelectedRow()
	if row == nil {
		return
	}
	name := row[1]
	s, err := styles.Load(name, m.themesDir, m.isDark)
	if err != nil {
		m.err = err.Error()
		return
	}
	m.err = ""
	styles.Apply(s)
	m.ktx.Config().SetTheme(name)
}

func (m *Model) currentTheme() string {
	if t := m.ktx.Config().Theme; t != "" {
		return t
	}
	return styles.DefaultTheme
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{Name: "Apply Theme", Keybinding: "enter"},
	}
}

func (m *Model) Title() string {
	return "Settings"
}

func New(ktx *kontext.ProgramKtx, themesDir string, isDark bool) *Model {
	m := &Model{
		ktx:       ktx,
		themesDir: themesDir,
		isDark:    isDark,
		themes:    styles.Available(themesDir),
		table:     ktable.NewDefaultTable(),
	}
	m.border = border.New(border.WithTitleFn(func() string {
		return border.KeyValueTitle("Themes", fmt.Sprintf("%d", len(m.themes)), true)
	}))
	// start the cursor on the active theme
	for i, name := range m.themes {
		if name == m.currentTheme() {
			m.table.SetCursor(i)
		}
	}
	return m
}
