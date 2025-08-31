package pages

import (
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type Page interface {
	ui.View
	statusbar.Provider
}
