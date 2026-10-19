package table

import (
	"charm.land/bubbles/v2/table"
	"ktea/styles"
)

func NewDefaultTable() table.Model {
	return table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
}

// View renders t with the current theme's table styles.
// Tables copy their styles when created, so re-apply them to follow theme changes.
func View(t *table.Model) string {
	t.SetStyles(styles.Table.Styles)
	return t.View()
}

// RowWidth is the rendered width of a row with the given columns.
// bubbles v2 tables clip rows to their width, which defaults to 0.
func RowWidth(cols []table.Column) int {
	cellPadding := styles.Table.Styles.Cell.GetHorizontalFrameSize()
	w := 0
	for _, c := range cols {
		w += c.Width + cellPadding
	}
	return w
}
