package schema_details_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"strconv"
)

type CmdBar struct {
	notifierWidget cmdbar.CmdBar
	deleteCmdBar   *cmdbar.DeleteCmdBar[int]
	active         cmdbar.CmdBar
}

func (c *CmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if c.active != nil {
		return c.active.View(ktx, renderer)
	}
	return ""
}

func (c *CmdBar) Update(msg tea.Msg, selection string) (tea.Msg, tea.Cmd) {
	// when the notifier is active and has priority (because of a loading spinner) it should handle all msgs
	if c.active == c.notifierWidget {
		if c.notifierWidget.(*cmdbar.NotifierCmdBar).Notifier.HasPriority() {
			active, pmsg, cmd := c.active.Update(msg)
			if !active {
				c.active = nil
			}
			return pmsg, cmd
		}
	}

	// notifier was not actively spinning
	// if it is able to handle the msg it will return nil and the processing can stop
	active, pmsg, cmd := c.notifierWidget.Update(msg)
	if active && pmsg == nil {
		c.active = c.notifierWidget
		return msg, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f2":
			active, pmsg, cmd := c.deleteCmdBar.Update(msg)
			if active {
				c.active = c.deleteCmdBar
				version, err := strconv.Atoi(selection)
				if err != nil {
					panic(fmt.Sprintf("Version (%s) cannot be converted to int", selection))
				}
				c.deleteCmdBar.Delete(version)
			} else {
				c.active = nil
			}
			return pmsg, cmd
		}
	}

	if c.active != nil {
		active, pmsg, cmd := c.active.Update(msg)
		if !active {
			c.active = nil
		}
		return pmsg, cmd
	}
	return msg, nil
}

func NewCmdBar(deleteFunc cmdbar.DeleteFn[int], schemaDeletedNotifier cmdbar.NotificationHandler[sradmin.SchemaDeletedMsg]) *CmdBar {
	schemaListingStartedNotifier := func(msg sradmin.SchemaListingStarted, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Loading schema")
		return true, cmd
	}
	schemaListedNotifier := func(msg sradmin.SchemasListed, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	schemaDeletionStartedNotifier := func(msg sradmin.SchemaDeletionStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Deleting schema version " + strconv.Itoa(msg.Version))
		return true, cmd
	}

	notifierCmdBar := cmdbar.NewNotifierCmdBar("schema-details-cmd-bar")
	cmdbar.BindNotificationHandler(notifierCmdBar, schemaListingStartedNotifier)
	cmdbar.BindNotificationHandler(notifierCmdBar, schemaDeletionStartedNotifier)
	cmdbar.BindNotificationHandler(notifierCmdBar, schemaDeletedNotifier)
	cmdbar.BindNotificationHandler(notifierCmdBar, schemaListedNotifier)
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg SchemaCopiedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Schema copied")
		return true, m.AutoHideCmd("schema-details-cmd-bar")
	})
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg CopyErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Copy failed", msg.Err)
		return true, m.AutoHideCmd("schema-details-cmd-bar")
	})

	deleteMsgFunc := func(schemaId int) string {
		return renderFG("Delete version ", styles.ColorIndigo) +
			renderFG(fmt.Sprintf("%d", schemaId), styles.ColorWhite) +
			renderFG(" of schema?", styles.ColorIndigo)
	}

	return &CmdBar{
		notifierWidget: notifierCmdBar,
		active:         notifierCmdBar,
		deleteCmdBar:   cmdbar.NewDeleteCmdBar[int](deleteMsgFunc, deleteFunc),
	}
}

func renderFG(value string, color string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render(value)
}
