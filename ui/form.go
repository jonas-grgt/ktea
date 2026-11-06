package ui

import (
	"charm.land/huh/v2"
	"ktea/styles"
)

// RenderForm renders f, keeping it visible after submission.
//
// huh blanks a submitted form (View returns "") while it keeps ignoring
// input. Pages submit forms to start async work, so rendering the layout
// directly keeps the form on screen yet blocked until the page builds a
// new one. layout must be the layout the form was built with.
func RenderForm(f *huh.Form, layout huh.Layout) string {
	if v := f.View(); v != "" {
		return v
	}
	return styles.Huh.Theme(false).Form.Base.Render(layout.View(f))
}
