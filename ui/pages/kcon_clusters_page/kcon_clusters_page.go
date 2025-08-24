package kcon_clusters_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/kcon_page"
	"reflect"
)

type Model struct {
	table        table.Model
	loadKConPage kcon_page.LoadPage
	cluster      *config.Cluster
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {

	var rows []table.Row
	for _, cluster := range ktx.Config.ActiveCluster().KafkaConnectClusters {
		rows = append(rows, table.Row{cluster.Name})
	}

	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{"Name", ktx.WindowWidth - 4},
	})
	m.table.SetRows(rows)
	m.table.SetHeight(ktx.AvailableHeight - 2)
	b := border.New(border.WithTitleFn(func() string {
		return border.KeyValueTitle(
			"Total Kafka Connect Clusters",
			fmt.Sprintf("%d/%d", len(rows), len(m.cluster.KafkaConnectClusters)),
			true,
		)
	}))
	return renderer.Render(b.View(m.table.View()))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			var name = m.table.SelectedRow()[0]
			for _, cluster := range m.cluster.KafkaConnectClusters {
				if cluster.Name == name {
					return m.loadKConPage(cluster)
				}
			}
			panic("Kafka Connect cluster " + name + " not found in active Kafka Cluster config.")
		}
	}

	t, c := m.table.Update(msg)
	m.table = t

	return c
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"View", "enter"},
	}
}

func (m *Model) Title() string {
	return "Kafka Connect Clusters"
}

func New(cluster *config.Cluster, loadKConPage kcon_page.LoadPage) (*Model, tea.Cmd) {
	m := Model{
		loadKConPage: loadKConPage,
		cluster:      cluster,
	}
	m.table = ktable.NewDefaultTable()
	return &m, nil
}
