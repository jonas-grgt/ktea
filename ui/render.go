package ui

import (
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
)

type Renderer struct {
	ktx *kontext.ProgramKtx
}

func (r *Renderer) Render(view string) string {
	if view == "" {
		return view
	}
	height := lipgloss.Height(view)
	r.ktx.HeightUsed(height)
	return view
}

func (r *Renderer) RenderWithStyle(view string, style lipgloss.Style) string {
	return r.Render(style.Render(view))
}

func NewRenderer(ktx *kontext.ProgramKtx) *Renderer {
	return &Renderer{ktx}
}
