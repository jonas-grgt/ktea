package topics_tab

import (
	"context"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/clipper"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/configs_page"
	"ktea/ui/pages/consumption_form_page"
	"ktea/ui/pages/consumption_page"
	"ktea/ui/pages/create_topic_page"
	"ktea/ui/pages/nav"
	"ktea/ui/pages/publish_page"
	"ktea/ui/pages/record_details_page"
	"ktea/ui/pages/topics_page"
)

type Model struct {
	active            nav.Page
	topicsPage        *topics_page.Model
	statusbar         *statusbar.Model
	ka                kadmin.Kadmin
	ktx               *kontext.ProgramKtx
	consumptionPage   nav.Page
	recordDetailsPage nav.Page
	ctx               context.Context
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.statusbar != nil {
		views = append(views, m.statusbar.View(ktx, renderer))
	}

	views = append(views, m.active.View(ktx, renderer))

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case kadmin.TopicListedMsg,
		kadmin.TopicRecordCountCalculatedMsg,
		kadmin.AllTopicRecordCountCalculatedMsg,
		spinner.TickMsg:
		// Make sure TopicListedMsg is explicitly captured and
		// properly propagated in the case when cgroupsPage
		//isn't focused anymore.
		return m.topicsPage.Update(msg)

	case nav.LoadTopicsPageMsg:
		if msg.Refresh {
			cmds = append(cmds, m.topicsPage.Refresh())
		}
		m.active = m.topicsPage

	case nav.LoadConsumptionFormPageMsg:
		if msg.ReadDetails != nil {
			m.active = consumption_form_page.NewWithDetails(msg.ReadDetails, msg.Topic, m.ktx)
		} else {
			m.active = consumption_form_page.New(msg.Topic, m.ktx)
		}

	case nav.LoadRecordDetailPageMsg:
		m.active = record_details_page.New(msg.Record, msg.TopicName, clipper.New(), m.ktx)
		m.recordDetailsPage = m.active

	case nav.LoadTopicConfigPageMsg:
		page, cmd := configs_page.New(m.ka, m.ka, m.topicsPage.SelectedTopicName())
		cmds = append(cmds, cmd)
		m.active = page

	case nav.LoadCreateTopicPageMsg:
		log.Debug("Loading create topic page")
		m.active = create_topic_page.New(m.ka)

	case nav.LoadPublishPageMsg:
		m.active = publish_page.New(m.ka, msg.Topic)

	case nav.LoadCachedConsumptionPageMsg:
		m.active = m.consumptionPage

	case nav.LoadConsumptionPageMsg:
		var cmd tea.Cmd
		m.active, cmd = consumption_page.New(m.ka, msg.ReadDetails, msg.Topic)
		m.consumptionPage = m.active
		cmds = append(cmds, cmd)

	case nav.LoadLiveConsumePageMsg:
		var cmd tea.Cmd
		readDetails := kadmin.ReadDetails{
			TopicName:       msg.Topic.Name,
			PartitionToRead: msg.Topic.Partitions(),
			StartPoint:      kadmin.Live,
			Limit:           500,
			Filter:          nil,
		}
		m.active, cmd = consumption_page.New(m.ka, readDetails, msg.Topic)
		m.consumptionPage = m.active
		cmds = append(cmds, cmd)

	}

	if cmd := m.active.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.active)

	return tea.Batch(cmds...)
}

func New(ktx *kontext.ProgramKtx, ka kadmin.Kadmin) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	listTopicView, cmd := topics_page.New(ka, ka)

	model := &Model{}
	model.ka = ka
	model.ktx = ktx
	model.active = listTopicView
	model.topicsPage = listTopicView
	model.statusbar = statusbar.New(model.active)

	return model, cmd
}
