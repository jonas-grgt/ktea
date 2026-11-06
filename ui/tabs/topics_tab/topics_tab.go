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
	"ktea/ui/pages/consume_page"
	"ktea/ui/pages/create_topic_page"
	"ktea/ui/pages/publish_page"
	"ktea/ui/pages/record_details_page"
	"ktea/ui/pages/topics_page"
	"ktea/ui/tabs"
	"reflect"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
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

	if msg, ok := msg.(kadmin.TopicsListedMsg); ok {
		// Make sure TopicsListedMsg is explicitly captured and
		// properly propagated in the case when the topics page
		// isn't focused anymore.
		return m.topicsPage.Update(msg)
	}

	cmd := m.active.Update(msg)

	// always recreate the statusbar in case the active page might have changed
	m.statusbar.SetProvider(m.active)

	return cmd
}

func (m *Model) ToTopicsPage(refresh bool) tea.Cmd {
	m.active = m.topicsPage
	if refresh {
		return m.topicsPage.Refresh()
	}
	return nil
}

func (m *Model) ToCreateTopicPage() tea.Cmd {
	m.active = create_topic_page.New(m.ka, m)
	return nil
}

func (m *Model) ToTopicConfigPage(topic string) tea.Cmd {
	page, cmd := configs_page.New(m.ka, m.ka, topic, m)
	m.active = page
	return cmd
}

func (m *Model) ToPublishPage(topic *kadmin.ListedTopic) tea.Cmd {
	m.active = publish_page.New(m.ka, topic, m)
	return nil
}

func (m *Model) ToConsumePage(msg tabs.ConsumePageDetails) tea.Cmd {
	var cmd tea.Cmd
	m.active, cmd = consume_page.New(
		m.ka,
		msg.ReadDetails,
		msg.Topic,
		msg.OpenSearch,
		m,
	)
	m.consumptionPage = m.active
	return cmd
}

func (m *Model) ToCachedConsumptionPage() tea.Cmd {
	m.active = m.consumptionPage
	// lets the page resume its loading indicator while it was not receiving msgs
	return ui.PublishMsg(ui.RegainedFocusMsg{})
}

func (m *Model) ToRecordDetailsPage(d tabs.RecordDetailsPageDetails) tea.Cmd {
	m.active = record_details_page.New(d.Record, d.TopicName, d.Records, d.Index, clipper.New(), m.ktx, m)
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
