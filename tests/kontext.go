package tests

import (
	"ktea/config"
	"ktea/kontext"
	"ktea/ui"
)

var TestKontext = &kontext.ProgramKtx{
	Config:          nil,
	WindowWidth:     100,
	WindowHeight:    100,
	AvailableHeight: 100,
}

var TestRenderer = ui.NewRenderer(TestKontext)

type TestContextOption func(ktx *kontext.ProgramKtx)

func WithConfig(config *config.Config) TestContextOption {
	return func(ktx *kontext.ProgramKtx) {
		ktx.Config = config
	}
}

func NewKontext(options ...TestContextOption) *kontext.ProgramKtx {
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
