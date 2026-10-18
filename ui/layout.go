package ui

import "charm.land/lipgloss/v2"

func JoinVertical(position lipgloss.Position, views ...string) string {
	return join(views, position)
}

func JoinHorizontal(views ...string) string {
	return join(views, lipgloss.Center)
}

func join(views []string, position lipgloss.Position) string {
	var renderViews []string
	for _, view := range views {
		if view != "" {
			renderViews = append(renderViews, view)
		}
	}
	return lipgloss.JoinVertical(position, renderViews...)
}
