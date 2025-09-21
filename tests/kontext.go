package tests

import (
	"ktea/config"
	"ktea/kontext"
	"ktea/ui"
)

var Kontext = NewKontext()

var Renderer = ui.NewRenderer(Kontext)

type ContextOption func(ktx *kontext.ProgramKtx)

func WithConfig(config *config.Config) ContextOption {
	return func(ktx *kontext.ProgramKtx) {
		ktx.RegisterConfig(config)
	}
}

func WithWindowWidth(width int) ContextOption {
	return func(ktx *kontext.ProgramKtx) {
		ktx.WindowWidth = width
	}
}

func NewKontext(options ...ContextOption) *kontext.ProgramKtx {
	model := &kontext.ProgramKtx{
		WindowWidth:     100,
		WindowHeight:    100,
		AvailableHeight: 100,
	}
	model.RegisterConfig(&config.Config{})
	for _, option := range options {
		option(model)
	}
	return model
}
