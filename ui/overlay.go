package ui

import "charm.land/lipgloss/v2"

// Overlay draws modal centered on top of base, within an area of width x height.
func Overlay(base, modal string, width, height int) string {
	x := max(0, (width-lipgloss.Width(modal))/2)
	y := max(0, (height-lipgloss.Height(modal))/2)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(modal).X(x).Y(y).Z(1),
	).Render()
}
