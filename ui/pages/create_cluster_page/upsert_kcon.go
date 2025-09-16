package create_cluster_page

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/tabs"
	"reflect"
)

type UpsertKcModel struct {
	ktx             *kontext.ProgramKtx
	connectClusters []config.KafkaConnectConfig
	connChecker     kcadmin.ConnChecker
	form            *huh.Form
	table           table.Model
	rows            []table.Row
	cmdBar          *cmdbar.NotifierCmdBar
	registerer      tea.Cmd
	formValues
	state
	deleteCmdbar *cmdbar.DeleteCmdBar[string]
	navigator    tabs.ClustersTabNavigator
}

type state int

type formValues struct {
	url      string
	username string
	password string
	name     string
	prevName string
}

func (f *formValues) usernameOrNil() *string {
	var username *string
	if f.username != "" {
		username = &f.username
	}
	return username
}

func (f *formValues) passwordOrNil() *string {
	var password *string
	if f.password != "" {
		password = &f.password
	}
	return password
}

const (
	entering state = iota
	listing
	registering
)

func (m *UpsertKcModel) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.state == listing {
		m.table.SetColumns([]table.Column{
			{"Name", ktx.WindowWidth - 5},
		})
		m.table.SetHeight(ktx.AvailableHeight - 1)
		m.table.SetWidth(ktx.WindowWidth - 2)
		m.table.SetRows(m.rows)

		views = append(views, renderer.Render(m.table.View()))
		return ui.JoinVertical(lipgloss.Top, views...)
	}
	m.form.WithHeight(ktx.AvailableHeight - 3)
	return renderer.RenderWithStyle(m.form.View(), styles.Form)
}

func (m *UpsertKcModel) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.state == entering {
				if len(m.connectClusters) > 0 {
					m.state = listing
				}
			} else {
				m.navigator.ToClustersPage()
			}
			return nil
		case tea.KeyF2:
			clusterName := m.table.SelectedRow()[0]
			m.deleteCmdbar.Delete(clusterName)
			_, _, cmd := m.deleteCmdbar.Update(msg)
			return cmd
		case tea.KeyCtrlN:
			m.state = entering
			m.resetFormValues()
			m.form = m.createKcForm()
		case tea.KeyCtrlE:
			if m.state == entering {
				break
			}
			m.state = entering
			clusterName := m.table.SelectedRow()[0]
			for _, cluster := range m.connectClusters {
				if clusterName == cluster.Name {
					m.formValues.prevName = cluster.Name
					m.formValues.name = cluster.Name
					m.formValues.url = cluster.Url

					if cluster.Username == nil {
						m.formValues.username = ""
					} else {
						m.formValues.username = *cluster.Username
					}

					if cluster.Password == nil {
						m.formValues.password = ""
					} else {
						m.formValues.password = *cluster.Password
					}

					m.form = m.createKcForm()
					return nil
				}
			}
			panic("Kafka Connect cluster not found: " + clusterName)
		}
	case kcadmin.ConnCheckStartedMsg:
		return msg.AwaitCompletion
	case kcadmin.ConnCheckSucceededMsg:
		return m.registerer
	case config.ClusterRegisteredMsg:
		m.state = listing
		m.connectClusters = msg.Cluster.KafkaConnectClusters
		m.rows = m.createRows()
		return nil
	case config.ConnectClusterDeleted:
		for i, cluster := range m.connectClusters {
			if msg.Name == cluster.Name {
				m.connectClusters = append(m.connectClusters[:i], m.connectClusters[i+1:]...)
			}
		}
		m.rows = m.createRows()
		if len(m.connectClusters) == 0 {
			m.state = entering
		}
		m.deleteCmdbar.Hide()
		return nil
	}

	if m.deleteCmdbar.IsFocussed() {
		_, _, cmd := m.deleteCmdbar.Update(msg)
		return cmd
	}

	if m.state == entering {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form.State == huh.StateCompleted {
			return m.processFormSubmission()
		}
		return cmd
	}

	t, c := m.table.Update(msg)
	m.table = t
	return c
}

func (m *UpsertKcModel) processFormSubmission() tea.Cmd {
	m.state = registering

	kcConfig := config.KafkaConnectConfig{
		Name:     m.formValues.name,
		Url:      m.formValues.url,
		Username: m.formValues.usernameOrNil(),
		Password: m.formValues.passwordOrNil(),
	}

	return func() tea.Msg {
		return m.connChecker(&kcConfig)
	}
}

func (m *UpsertKcModel) resetFormValues() {
	m.formValues.name = ""
	m.formValues.url = ""
	m.formValues.username = ""
	m.formValues.password = ""
}

func (m *UpsertKcModel) clusterDetails() []config.KafkaConnectClusterDetails {
	var details []config.KafkaConnectClusterDetails
	var updated bool
	for _, cluster := range m.connectClusters {
		if m.formValues.prevName == cluster.Name && m.formValues.prevName != "" {
			updated = true
			details = append(details, config.KafkaConnectClusterDetails{
				Name:     m.formValues.name,
				Url:      m.formValues.url,
				Username: m.formValues.usernameOrNil(),
				Password: m.formValues.passwordOrNil(),
			})
		} else {
			details = append(details, config.KafkaConnectClusterDetails{
				Name:     cluster.Name,
				Url:      cluster.Url,
				Username: cluster.Username,
				Password: cluster.Password,
			})
		}
	}

	if !updated && m.formValues.name != "" {
		details = append(details, config.KafkaConnectClusterDetails{
			Name:     m.formValues.name,
			Url:      m.formValues.url,
			Username: m.formValues.usernameOrNil(),
			Password: m.formValues.passwordOrNil(),
		})
	}

	return details
}

func (m *UpsertKcModel) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Go back", "esc"},
		{"Edit Connect Cluster", "C-e"},
		{"New Connect Cluster", "C+n"},
		{"Delete Connect Cluster", "F2"},
	}
}

func (m *UpsertKcModel) Title() string {
	//TODO implement me
	panic("implement me")
}

func (m *UpsertKcModel) createRows() []table.Row {
	var rows []table.Row
	for _, c := range m.connectClusters {
		rows = append(rows, table.Row{c.Name})
	}
	return rows
}

func (m *UpsertKcModel) createKcForm() *huh.Form {
	var fields []huh.Field
	name := huh.NewInput().
		Value(&m.formValues.name).
		Title("Kafka Connect Name")
	url := huh.NewInput().
		Value(&m.formValues.url).
		Title("Kafka Connect URL")
	username := huh.NewInput().
		Value(&m.formValues.username).
		Title("Kafka Connect Username")
	password := huh.NewInput().
		Value(&m.formValues.password).
		EchoMode(huh.EchoModePassword).
		Title("Kafka Connect Password")
	fields = append(fields, name, url, username, password)

	form := huh.NewForm(
		huh.NewGroup(fields...).
			Title("Kafka Connect"),
	)
	form.QuitAfterSubmit = false
	form.Init()

	return form
}

type ClusterDeleter func(name string) tea.Msg

func NewUpsertKcModel(
	navigator tabs.ClustersTabNavigator,
	ktx *kontext.ProgramKtx,
	deleter ClusterDeleter,
	configs []config.KafkaConnectConfig,
	connChecker kcadmin.ConnChecker,
	cmdBar *cmdbar.NotifierCmdBar,
	registerer tea.Cmd,
) *UpsertKcModel {
	m := UpsertKcModel{}

	m.navigator = navigator
	m.ktx = ktx
	m.connectClusters = configs
	m.connChecker = connChecker
	m.cmdBar = cmdBar
	m.registerer = registerer

	m.rows = m.createRows()
	m.form = m.createKcForm()
	m.table = ktable.NewDefaultTable()

	if len(m.connectClusters) > 0 {
		m.state = listing
	} else {
		m.state = entering
	}

	deleteMsgFunc := func(c string) string {
		return c + " will be permanently deleted!"
	}
	deleteFunc := func(c string) tea.Cmd {
		return func() tea.Msg {
			return deleter(c)
		}
	}
	m.deleteCmdbar = cmdbar.NewDeleteCmdBar[string](deleteMsgFunc, deleteFunc)

	cmdbar.WithMsgHandler(m.cmdBar, func(msg kcadmin.ConnCheckStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Testing cluster connectivity")
	})

	cmdbar.WithMsgHandler(m.cmdBar, func(msg kcadmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.form = m.createKcForm()
		m.state = entering
		nm.ShowErrorMsg("Unable to reach the cluster", msg.Err)
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})

	cmdbar.WithMsgHandler(m.cmdBar, func(msg kcadmin.ConnCheckSucceededMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Connection succeeded, creating cluster")
	})

	return &m
}
