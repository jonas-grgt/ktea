package create_cluster_page

import (
	"errors"
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type authSelection int

type formState int

type Option func(m *Model)

const (
	noneSelected      authSelection = 0
	saslSelected      authSelection = 1
	nothingSelected   authSelection = 2
	none              formState     = 0
	loading           formState     = 1
	notifierCmdbarTag               = "upsert-cluster-page"
)

type Model struct {
	form               *huh.Form // the active form
	state              formState
	srForm             *huh.Form
	kcForm             *huh.Form
	cForm              *huh.Form
	kForm              *huh.Form
	clusterValues      *clusterValues
	registeredCluster  *config.Cluster
	notifierCmdBar     *cmdbar.NotifierCmdBar
	ktx                *kontext.ProgramKtx
	clusterRegisterer  config.ClusterRegisterer
	kConnChecker       kadmin.ConnChecker
	srConnChecker      sradmin.ConnChecker
	authSelectionState authSelection
	preEditName        *string
	shortcuts          []statusbar.Shortcut
	title              string
	border             *border.Model
}

type clusterValues struct {
	name             string
	color            string
	host             string
	authMethod       config.AuthMethod
	securityProtocol config.SecurityProtocol
	sslEnabled       bool
	username         string
	password         string
	srUrl            string
	srUsername       string
	srPassword       string
	kcUrl            string
	kcUsername       string
	kcPassword       string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if !ktx.Config.HasClusters() {
		builder := strings.Builder{}
		builder.WriteString("\n")
		builder.WriteString(lipgloss.NewStyle().PaddingLeft(1).Render("No clusters configured. Please create your first cluster!"))
		builder.WriteString("\n")
		views = append(views, renderer.Render(builder.String()))
	}

	notifierView := m.notifierCmdBar.View(ktx, renderer)
	formView := renderer.RenderWithStyle(m.form.View(), styles.Form)
	formView = m.border.View(lipgloss.NewStyle().
		PaddingBottom(ktx.AvailableHeight - 1).
		Render(formView))
	views = append(views, notifierView, formView)

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			m.clusterValues = &clusterValues{}
			if m.form == m.cForm {
				m.cForm = m.createCForm()
				m.form = m.cForm
				m.authSelectionState = noneSelected
			} else {
				m.srForm = m.createSrForm()
				m.form = m.srForm
			}
		case "f1":
			m.form = m.cForm
			m.border.GoTo("f1")
		case "f2":
			if m.inEditingMode() {
				m.form = m.srForm
				m.form.State = huh.StateNormal
				m.border.GoTo("f2")
			} else {
				return tea.Batch(
					m.notifierCmdBar.Notifier.ShowError(fmt.Errorf("create a cluster before adding a schema registry")),
					m.notifierCmdBar.Notifier.AutoHideCmd(notifierCmdbarTag),
				)
			}
		case "f3":
			m.form = m.kcForm
			m.form.State = huh.StateNormal
			m.border.GoTo("f3")
		}
	case kadmin.ConnCheckStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.ConnCheckSucceededMsg:
		m.state = none
		cmds = append(cmds, func() tea.Msg {
			details := m.getRegistrationDetails()
			return m.clusterRegisterer.RegisterCluster(details)
		})
	case sradmin.ConnCheckStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.ConnCheckSucceededMsg:
		m.state = none
		return func() tea.Msg {
			details := m.getRegistrationDetails()
			return m.clusterRegisterer.RegisterCluster(details)
		}
	case config.ClusterRegisteredMsg:
		m.preEditName = &msg.Cluster.Name
		m.registeredCluster = msg.Cluster
		m.state = none
		m.border.WithInActiveColor(styles.ColorGrey)
		if m.form == m.cForm {
			m.cForm = m.createCForm()
			m.form = m.cForm
		} else {
			m.srForm = m.createSrForm()
			m.form = m.srForm
		}
	}

	_, msg, cmd := m.notifierCmdBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if msg == nil {
		return tea.Batch(cmds...)
	}

	form, cmd := m.form.Update(msg)
	cmds = append(cmds, cmd)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form == m.cForm {
		if !m.clusterValues.HasSASLAuthMethodSelected() &&
			m.authSelectionState == saslSelected {
			// if SASL authentication mode was previously selected and switched back to none
			m.cForm = m.createCForm()
			m.form = m.cForm
			m.NextField(4)
			m.authSelectionState = noneSelected
		} else if m.clusterValues.HasSASLAuthMethodSelected() &&
			(m.authSelectionState == nothingSelected || m.authSelectionState == noneSelected) {
			// SASL authentication mode selected and previously nothing or none auth mode was selected
			m.cForm = m.createCForm()
			m.form = m.cForm
			m.NextField(4)
			m.authSelectionState = saslSelected
		}

		if m.form.State == huh.StateCompleted && m.state != loading {
			return m.processClusterSubmission()
		}
	}

	if m.form == m.srForm {
		if m.form.State == huh.StateCompleted && m.state != loading {
			return m.processSrSubmission()
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) processSrSubmission() tea.Cmd {
	m.state = loading
	details := m.getRegistrationDetails()

	cluster := config.ToCluster(details)
	return func() tea.Msg {
		return m.srConnChecker(cluster.SchemaRegistry)
	}
}

func (m *Model) processClusterSubmission() tea.Cmd {
	m.state = loading
	details := m.getRegistrationDetails()

	cluster := config.ToCluster(details)
	return func() tea.Msg {
		return m.kConnChecker(&cluster)
	}
}

func (m *Model) getRegistrationDetails() config.RegistrationDetails {
	var name string
	var newName *string
	if m.preEditName == nil { // When creating a cluster
		name = m.clusterValues.name
		newName = nil
	} else { // When updating a cluster.
		name = *m.preEditName
		if m.clusterValues.name != *m.preEditName {
			newName = &m.clusterValues.name
		}
	}

	var authMethod config.AuthMethod
	var securityProtocol config.SecurityProtocol
	if m.clusterValues.HasSASLAuthMethodSelected() {
		authMethod = config.SASLAuthMethod
		securityProtocol = m.clusterValues.securityProtocol
	} else {
		authMethod = config.NoneAuthMethod
	}

	details := config.RegistrationDetails{
		Name:             name,
		NewName:          newName,
		Color:            m.clusterValues.color,
		Host:             m.clusterValues.host,
		AuthMethod:       authMethod,
		SecurityProtocol: securityProtocol,
		SSLEnabled:       m.clusterValues.sslEnabled,
		Username:         m.clusterValues.username,
		Password:         m.clusterValues.password,
	}
	if m.clusterValues.SrEnabled() {
		details.SchemaRegistry = &config.SchemaRegistryDetails{
			Url:      m.clusterValues.srUrl,
			Username: m.clusterValues.srUsername,
			Password: m.clusterValues.srPassword,
		}
	}
	return details
}

func (f *clusterValues) HasSASLAuthMethodSelected() bool {
	return f.authMethod == config.SASLAuthMethod
}

func (f *clusterValues) SrEnabled() bool {
	return len(f.srUrl) > 0
}

func (m *Model) NextField(count int) {
	for i := 0; i < count; i++ {
		m.form.NextField()
	}
}

func (m *Model) createCForm() *huh.Form {
	name := huh.NewInput().
		Value(&m.clusterValues.name).
		Title("Name").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("name cannot be empty")
			}
			if m.preEditName != nil {
				// When updating.
				if m.ktx.Config.FindClusterByName(v) != nil && v != *m.preEditName {
					return errors.New("cluster " + v + " already exists, name most be unique")
				}
			} else {
				// When creating a new cluster
				if m.ktx.Config.FindClusterByName(v) != nil {
					return errors.New("cluster " + v + " already exists, name most be unique")
				}
			}
			return nil
		})
	color := huh.NewSelect[string]().
		Value(&m.clusterValues.color).
		Title("Color").
		Options(
			huh.NewOption(styles.Env.Colors.Green.Render("green"), styles.ColorGreen),
			huh.NewOption(styles.Env.Colors.Blue.Render("blue"), styles.ColorBlue),
			huh.NewOption(styles.Env.Colors.Orange.Render("orange"), styles.ColorOrange),
			huh.NewOption(styles.Env.Colors.Purple.Render("purple"), styles.ColorPurple),
			huh.NewOption(styles.Env.Colors.Yellow.Render("yellow"), styles.ColorYellow),
			huh.NewOption(styles.Env.Colors.Red.Render("red"), styles.ColorRed),
		)
	host := huh.NewInput().
		Value(&m.clusterValues.host).
		Title("Host").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("host cannot be empty")
			}
			return nil
		})
	auth := huh.NewSelect[config.AuthMethod]().
		Value(&m.clusterValues.authMethod).
		Title("Authentication method").
		Options(
			huh.NewOption("NONE", config.NoneAuthMethod),
			huh.NewOption("SASL", config.SASLAuthMethod),
		)

	sslEnabled := huh.NewSelect[bool]().
		Value(&m.clusterValues.sslEnabled).
		Title("SSL").
		Options(
			huh.NewOption("Disable SSL", false),
			huh.NewOption("Enable SSL", true),
		)

	var clusterFields []huh.Field
	clusterFields = append(clusterFields, name, color, host, sslEnabled, auth)

	if m.clusterValues.HasSASLAuthMethodSelected() {
		securityProtocol := huh.NewSelect[config.SecurityProtocol]().
			Value(&m.clusterValues.securityProtocol).
			Title("Security Protocol").
			Options(
				huh.NewOption("SASL_PLAINTEXT", config.SASLPlaintextSecurityProtocol),
			)
		username := huh.NewInput().
			Value(&m.clusterValues.username).
			Title("Username")
		pwd := huh.NewInput().
			Value(&m.clusterValues.password).
			EchoMode(huh.EchoModePassword).
			Title("Password")
		clusterFields = append(clusterFields, securityProtocol, username, pwd)
	}

	form := huh.NewForm(
		huh.NewGroup(clusterFields...).
			Title("Cluster").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.WithLayout(huh.LayoutColumns(1))
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func (m *Model) createSrForm() *huh.Form {
	var fields []huh.Field
	srUrl := huh.NewInput().
		Value(&m.clusterValues.srUrl).
		Title("Schema Registry URL")
	srUsername := huh.NewInput().
		Value(&m.clusterValues.srUsername).
		Title("Schema Registry Username")
	srPwd := huh.NewInput().
		Value(&m.clusterValues.srPassword).
		EchoMode(huh.EchoModePassword).
		Title("Schema Registry Password")
	fields = append(fields, srUrl, srUsername, srPwd)

	form := huh.NewForm(
		huh.NewGroup(fields...).
			Title("Schema Registry").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.QuitAfterSubmit = false
	form.Init()

	return form
}

func (m *Model) createKcForm() *huh.Form {
	var fields []huh.Field
	srUrl := huh.NewInput().
		Value(&m.clusterValues.kcUrl).
		Title("Kafka Connect URL")
	srUsername := huh.NewInput().
		Value(&m.clusterValues.kcUsername).
		Title("Kafka Connect Username")
	srPwd := huh.NewInput().
		Value(&m.clusterValues.kcPassword).
		EchoMode(huh.EchoModePassword).
		Title("Kafka Connect Password")
	fields = append(fields, srUrl, srUsername, srPwd)

	form := huh.NewForm(
		huh.NewGroup(fields...).
			Title("Kafka Connect").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.QuitAfterSubmit = false
	form.Init()

	return form
}

func (m *Model) createNotifierCmdBar() {
	m.notifierCmdBar = cmdbar.NewNotifierCmdBar(notifierCmdbarTag)
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Testing cluster connectivity")
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckSucceededMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Connection success creating cluster")
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.state = none
		nMsg := "Cluster not crated"
		if m.inEditingMode() {
			nMsg = "Cluster not updated"
		}
		return true, nm.ShowErrorMsg(nMsg, msg.Err)
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg config.ClusterRegisteredMsg, nm *notifier.Model) (bool, tea.Cmd) {
		if m.form == m.srForm {
			nm.ShowSuccessMsg("Schema registry registered! <ESC> to go back.")
		} else {
			if m.inEditingMode() {
				nm.ShowSuccessMsg("Cluster updated!")
			} else {
				nm.ShowSuccessMsg("Cluster registered! <ESC> to go back or <F2> to add a schema registry.")
			}
		}
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg sradmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.srForm = m.createSrForm()
		m.form = m.srForm
		m.state = none
		nm.ShowErrorMsg("unable to reach the schema registry", msg.Err)
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return m.shortcuts
}

func (m *Model) Title() string {
	if m.title == "" {
		return "Clusters / Create"
	}
	return m.title
}

func (m *Model) inEditingMode() bool {
	return m.registeredCluster != nil
}

func WithTitle(title string) Option {
	return func(m *Model) {
		m.title = title
	}
}

func initBorder() *border.Model {
	return border.New(
		border.WithInactiveColor(styles.ColorDarkGrey),
		border.WithTabs(
			border.Tab{Title: "Cluster ≪ F1 »", Label: "f1"},
			border.Tab{Title: "Schema Registry ≪ F2 »", Label: "f2"},
			border.Tab{Title: "Kafka Connect ≪ f3 »", Label: "f3"},
		),
	)
}

func NewCreateClusterPage(
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	registerer config.ClusterRegisterer,
	ktx *kontext.ProgramKtx,
	shortcuts []statusbar.Shortcut,
	options ...Option,
) *Model {
	formValues := &clusterValues{}
	model := Model{
		clusterValues: formValues,
		kConnChecker:  kConnChecker,
		srConnChecker: srConnChecker,
		shortcuts:     shortcuts,
	}

	model.ktx = ktx

	model.border = initBorder()

	model.cForm = model.createCForm()
	model.srForm = model.createSrForm()
	model.kcForm = model.createKcForm()
	model.form = model.cForm

	model.clusterRegisterer = registerer

	model.authSelectionState = nothingSelected
	model.state = none

	if model.clusterValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}

	model.createNotifierCmdBar()

	for _, option := range options {
		option(&model)
	}

	return &model
}

func NewEditClusterPage(
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	registerer config.ClusterRegisterer,
	ktx *kontext.ProgramKtx,
	cluster *config.Cluster,
	options ...Option,
) *Model {
	formValues := &clusterValues{
		name:  cluster.Name,
		color: cluster.Color,
		host:  cluster.BootstrapServers[0],
	}
	if cluster.SASLConfig != nil {
		formValues.securityProtocol = cluster.SASLConfig.SecurityProtocol
		formValues.username = cluster.SASLConfig.Username
		formValues.password = cluster.SASLConfig.Password
		formValues.authMethod = config.SASLAuthMethod
		formValues.sslEnabled = cluster.SSLEnabled
	}
	if cluster.SchemaRegistry != nil {
		formValues.srUrl = cluster.SchemaRegistry.Url
		formValues.srUsername = cluster.SchemaRegistry.Username
		formValues.srPassword = cluster.SchemaRegistry.Password
	}
	model := Model{
		registeredCluster: cluster,
		clusterValues:     formValues,
		kConnChecker:      kConnChecker,
		srConnChecker:     srConnChecker,
		shortcuts: []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Next Field", "tab"},
			{"Prev. Field", "s-tab"},
			{"Reset Form", "C-r"},
			{"Go Back", "esc"},
		},
	}
	if cluster.Name != "" {
		// copied to prevent model.preEditedName to follow the formValues.Name pointer
		preEditedName := cluster.Name
		model.preEditName = &preEditedName
	}
	model.ktx = ktx

	model.border = initBorder()

	model.cForm = model.createCForm()
	model.srForm = model.createSrForm()
	model.kcForm = model.createKcForm()
	model.form = model.cForm

	model.clusterRegisterer = registerer
	model.authSelectionState = nothingSelected
	model.state = none

	if model.clusterValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}

	model.createNotifierCmdBar()

	for _, o := range options {
		o(&model)
	}

	return &model
}
