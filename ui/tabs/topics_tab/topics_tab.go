package topics_tab

import (
	"context"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/clipper"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"ktea/ui/pages/configs_page"
	"ktea/ui/pages/consume_form_page"
	"ktea/ui/pages/consume_page"
	"ktea/ui/pages/create_topic_page"
	"ktea/ui/pages/nav"
	"ktea/ui/pages/publish_page"
	"ktea/ui/pages/record_details_page"
	"ktea/ui/pages/topics_page"
	"ktea/ui/tabs"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Model struct {
	active            pages.Page
	topicsPage        *topics_page.Model
	statusbar         *statusbar.Model
	ka                kadmin.Kadmin
	ktx               *kontext.ProgramKtx
	consumptionPage   pages.Page
	recordDetailsPage pages.Page
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

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case kadmin.TopicsListedMsg:
		// Make sure TopicsListedMsg is explicitly captured and
		// properly propagated in the case when cgroupsPage
		//isn't focused anymore.
		return m.topicsPage.Update(msg)

	case nav.LoadTopicsPageMsg:
		if msg.Refresh {
			cmds = append(cmds, m.topicsPage.Refresh())
		}
		m.active = m.topicsPage

	case nav.LoadTopicConfigPageMsg:
		page, cmd := configs_page.New(m.ka, m.ka, msg.Topic)
		cmds = append(cmds, cmd)
		m.active = page

	case nav.LoadCreateTopicPageMsg:
		log.Debug("Loading create topic page")
		m.active = create_topic_page.New(m.ka)

	case nav.LoadPublishPageMsg:
		m.active = publish_page.New(m.ka, msg.Topic)

	case nav.LoadCachedConsumptionPageMsg:
		m.active = m.consumptionPage

	case nav.LoadLiveConsumePageMsg:
		var cmd tea.Cmd
		readDetails := kadmin.ReadDetails{
			TopicName:       msg.Topic.Name,
			PartitionToRead: msg.Topic.Partitions(),
			StartPoint:      kadmin.Live,
			Limit:           500,
			Filter:          nil,
		}
		m.active, cmd = consume_page.New(
			m.ka,
			readDetails,
			msg.Topic,
			tabs.OriginTopicsPage,
			m,
		)
		m.consumptionPage = m.active
		cmds = append(cmds, cmd)

	}

	if cmd := m.active.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// always recreate the statusbar in case the active page might have changed
	m.statusbar.SetProvider(m.active)

	return tea.Batch(cmds...)
}

func (m *Model) ToTopicsPage() tea.Cmd {
	m.active = m.topicsPage
	return nil
}

func (m *Model) ToConsumePage(msg tabs.ConsumePageDetails) tea.Cmd {
	var cmd tea.Cmd
	m.active, cmd = consume_page.New(
		m.ka,
		msg.ReadDetails,
		msg.Topic,
		msg.Origin,
		m,
	)
	m.consumptionPage = m.active
	return cmd
}

func (m *Model) ToConsumeFormPage(d tabs.ConsumeFormPageDetails) tea.Cmd {
	if d.ReadDetails != nil {
		m.active = consume_form_page.NewWithDetails(
			d.ReadDetails,
			d.Topic,
			m,
			m.ktx,
		)
	} else {
		m.active = consume_form_page.New(
			d.Topic,
			m,
			m.ktx,
		)
	}
	return nil
}

func (m *Model) ToRecordDetailsPage(msg tabs.LoadRecordDetailPageMsg) tea.Cmd {
	m.active = record_details_page.New(msg.Record, msg.TopicName, msg.Records, msg.Index, clipper.New(), m.ktx)
	m.recordDetailsPage = m.active
	return nil
}

func New(ktx *kontext.ProgramKtx, ka kadmin.Kadmin, stsBar *statusbar.Model) (*Model, tea.Cmd) {
	var cmd tea.Cmd

	model := &Model{}
	model.ka = ka
	model.ktx = ktx
	model.statusbar = stsBar
	model.statusbar.SetProvider(model.active)

	listTopicView, cmd := topics_page.New(
		ka,
		model,
	)
	model.active = listTopicView
	model.topicsPage = listTopicView

	return model, cmd
}
