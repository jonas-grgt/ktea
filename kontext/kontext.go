package kontext

import (
	"ktea/config"
)

type ProgramKtx struct {
	Config          *config.Config
	WindowWidth     int
	WindowHeight    int
	AvailableHeight int
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

func New() *ProgramKtx {
	return &ProgramKtx{}
}

func WithNewAvailableDimensions(ktx *ProgramKtx) *ProgramKtx {
	ktx.AvailableHeight = ktx.WindowHeight
	return ktx
}
