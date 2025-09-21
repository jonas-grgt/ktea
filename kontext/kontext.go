package kontext

import (
	"ktea/config"
)

type ProgramKtx struct {
	config *config.Config
	cmdLineFlags
	WindowWidth     int
	WindowHeight    int
	AvailableHeight int
}

type cmdLineFlags struct {
	disableNerdFonts *bool
}

func (k *ProgramKtx) HeightUsed(height int) {
	if k.AvailableHeight < height {
		k.AvailableHeight -= k.AvailableHeight
	} else {
		k.AvailableHeight -= height
	}
}

func (k *ProgramKtx) AvailableTableHeight() int {
	// 2 for top and bottom border + 1 for top extra padding
	return k.AvailableHeight - 3
}

func (k *ProgramKtx) Config() *config.Config {
	return k.config
}

func (k *ProgramKtx) RegisterConfig(c *config.Config) {
	k.config = c
	if k.disableNerdFonts != nil {
		k.config.PlainFonts = *k.disableNerdFonts
	}
}

func New(disableNerdFonts *bool) *ProgramKtx {
	return &ProgramKtx{
		cmdLineFlags: cmdLineFlags{
			disableNerdFonts: disableNerdFonts,
		},
	}
}

func WithNewAvailableDimensions(ktx *ProgramKtx) *ProgramKtx {
	ktx.AvailableHeight = ktx.WindowHeight
	return ktx
}
