package consume_page

import (
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"ktea/ui/tabs"
	"testing"
)

func TestConsumptionPage(t *testing.T) {
	t.Run("Display empty topic message and adjusted shortcuts", func(t *testing.T) {
		m, _ := New(
			kadmin.NewMockKadmin(),
			kadmin.ReadDetails{},
			&kadmin.ListedTopic{},
			nav.OriginTopicsPage,
			tabs.NewMockTopicsTabNavigator(),
		)

		m.Update(EmptyTopicMsg{})

		render := m.View(tests.NewKontext(), tests.Renderer)

		assert.Contains(t, render, "Empty topic")

		assert.Equal(t, []statusbar.Shortcut{{"Go Back", "esc"}}, m.Shortcuts())
	})
}
