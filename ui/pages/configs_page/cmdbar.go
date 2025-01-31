package configs_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/pages/nav"
)

type state int

const (
	HIDDEN           state = 0
	SEARCHING        state = 1
	SEARCHED         state = 2
	EDITING          state = 3
	UPDATING         state = 4
	UPDATE_FAILED    state = 5
	UPDATE_SUCCEEDED state = 6
	LOADING          state = 7
)

type CmdBarModel struct {
	state             state
	searchInput       *huh.Input
	editInput         *huh.Input
	notifier          *notifier.Model
	configUpdater     kadmin.ConfigUpdater
	topicConfigLister kadmin.TopicConfigLister
	topic             string
	updated           bool
}

type SelectedTopicConfig struct {
	Topic       string
	ConfigKey   string
	ConfigValue string
}

type HideBarMsg struct{}

func (m *CmdBarModel) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string

	if m.state == SEARCHING || m.state == SEARCHED {
		views = append(views, renderer.RenderWithStyle(m.searchInput.View(), styles.CmdBar.Width(ktx.WindowWidth-2)))
	} else if m.state == EDITING {
		views = append(views, renderer.RenderWithStyle(m.editInput.View(), styles.CmdBar.Width(ktx.WindowWidth-2)))
	} else if m.state == UPDATING || m.state == UPDATE_FAILED || m.state == UPDATE_SUCCEEDED || m.state == LOADING {
		views = append(views, m.notifier.View(ktx, renderer))
	}

	return ui.JoinVertical(lipgloss.Top, views...)
}

// Update returns the tea.Msg if it is not being handled or nil if it is
func (m *CmdBarModel) Update(msg tea.Msg, stc SelectedTopicConfig) (tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		return nil, m.notifier.Update(msg)
	case tea.KeyMsg:
		if msg.String() == "/" {
			m.state = SEARCHING
			m.searchInput.Focus()
			return nil, nil
		} else if msg.String() == "esc" {
			if m.state == SEARCHING {
				m.searchInput = newSearchInput()
			}
			if m.IsFocused() {
				m.state = HIDDEN
			} else {
				return nil, ui.PublishMsg(nav.LoadTopicsPageMsg{})
			}
			return nil, nil
		} else if msg.String() == "e" && isEditable(m) {
			m.state = EDITING
			m.editInput = newEditInput(stc.ConfigValue)
			m.editInput.Focus()
			return nil, nil
		} else if msg.String() == "enter" {
			if m.state == SEARCHING {
				if m.GetSearchTerm() == "" {
					m.state = HIDDEN
				} else {
					m.state = SEARCHED
				}
			} else if m.state == EDITING {
				m.state = UPDATING
				return nil, tea.Batch(
					m.notifier.SpinWithRocketMsg("Updating Topic Config"),
					func() tea.Msg {
						return m.configUpdater.UpdateConfig(kadmin.TopicConfigToUpdate{
							Topic: stc.Topic,
							Key:   stc.ConfigKey,
							Value: m.editInput.GetValue().(string),
						})
					},
				)
			}
			m.searchInput.Blur()
		} else if m.state == SEARCHING {
			confirm, _ := m.searchInput.Update(msg)
			if c, ok := confirm.(*huh.Input); ok {
				m.searchInput = c
			}
			return nil, nil
		} else if m.state == EDITING {
			confirm, _ := m.editInput.Update(msg)
			if c, ok := confirm.(*huh.Input); ok {
				m.editInput = c
			}
			return nil, nil
		}
	case kadmin.TopicConfigListingStartedMsg:
		m.state = LOADING
		var cmd tea.Cmd
		if m.updated {
			cmd = nil
		} else {
			cmd = m.notifier.SpinWithLoadingMsg("Loading " + m.topic + " Topic Configs")
		}
		return msg, cmd
	case kadmin.TopicConfigsListedMsg:
		if m.updated {
			m.updated = false
			m.state = UPDATE_SUCCEEDED
		} else {
			m.state = HIDDEN
			m.notifier.Idle()
		}
		return nil, nil
	case kadmin.UpdateTopicConfigErrorMsg:
		m.state = UPDATE_FAILED
		m.notifier.ShowErrorMsg(msg.Reason, fmt.Errorf("TODO"))
		return nil, nil
	case kadmin.TopicConfigUpdatedMsg:
		m.state = UPDATE_SUCCEEDED
		m.notifier.ShowSuccessMsg("Update succeeded")
		m.updated = true
		return nil, func() tea.Msg { return m.topicConfigLister.ListConfigs(stc.Topic) }
	case HideBarMsg:
		m.state = HIDDEN
	}
	return msg, nil
}

func isEditable(m *CmdBarModel) bool {
	return m.state == HIDDEN ||
		m.state == SEARCHED ||
		m.state == UPDATE_FAILED ||
		m.state == UPDATE_SUCCEEDED
}

func (m *CmdBarModel) IsFocused() bool {
	return !(m.state == HIDDEN ||
		m.state == SEARCHED ||
		m.state == UPDATE_FAILED ||
		m.state == UPDATE_SUCCEEDED)
}

func (m *CmdBarModel) GetSearchTerm() string {
	return m.searchInput.GetValue().(string)
}

func (m *CmdBarModel) IsLoading() bool {
	return m.state == LOADING || m.state == UPDATING
}

func newSearchInput() *huh.Input {
	searchInput := huh.NewInput().
		Placeholder("Search for Config")
	searchInput.Init()
	return searchInput
}

func newEditInput(v string) *huh.Input {
	searchInput := huh.NewInput().
		Value(&v).
		Placeholder("New config value")
	searchInput.Init()
	return searchInput
}

func NewCmdBar(cu kadmin.ConfigUpdater, tcl kadmin.TopicConfigLister, topic string) *CmdBarModel {
	return &CmdBarModel{
		topic:             topic,
		searchInput:       newSearchInput(),
		notifier:          notifier.New(),
		state:             HIDDEN,
		configUpdater:     cu,
		topicConfigLister: tcl,
	}
}
