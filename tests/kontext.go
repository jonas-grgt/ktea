package tests

import (
	"ktea/config"
	"ktea/kontext"
	"ktea/ui"
)

var Kontext = &kontext.ProgramKtx{
	Config:          nil,
	WindowWidth:     100,
	WindowHeight:    100,
	AvailableHeight: 100,
}

var Renderer = ui.NewRenderer(Kontext)

type ContextOption func(ktx *kontext.ProgramKtx)

func WithConfig(config *config.Config) ContextOption {
	return func(ktx *kontext.ProgramKtx) {
		ktx.Config = config
	}
}

func NewKontext(options ...ContextOption) *kontext.ProgramKtx {
	model := &kontext.ProgramKtx{
		Config:          nil,
		WindowWidth:     100,
		WindowHeight:    100,
		AvailableHeight: 100,
	}
	for _, option := range options {
		option(model)
	}
	return model
}
