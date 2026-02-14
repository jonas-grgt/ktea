package create_cluster_page

import (
	"errors"
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/tabs"
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
	authMethodNone        authSelection = 0
	authMethodSasl        authSelection = 1
	authMethodNotSelected authSelection = 2

	none              formState       = 6
	loading           formState       = 7
	notifierCmdbarTag                 = "upsert-cluster-page"
	cTab              border.TabLabel = "f4"
	srTab             border.TabLabel = "f5"
	kcTab             border.TabLabel = "f6"
)

type Model struct {
	navigator                                 tabs.ClustersTabNavigator
	form                                      *huh.Form // the active form
	formState                                 formState
	srForm                                    *huh.Form
	cForm                                     *huh.Form
	cFormValues                               *clusterFormValues
	clusterToEdit                             *config.Cluster
	notifierCmdBar                            *cmdbar.NotifierCmdBar
	ktx                                       *kontext.ProgramKtx
	clusterRegisterer                         config.ClusterRegisterer
	kConnChecker                              kadmin.ConnChecker
	srConnChecker                             sradmin.ConnChecker
	authSelState                              authSelection
	transportOption                           transportOption
	verificationOption                        verificationOption
	preEditName                               *string
	shortcuts                                 []statusbar.Shortcut
	title                                     string
	border                                    *border.Model
	kcModel                                   *UpsertKcModel
	hasVerificationStatePreviouslyNotSelected bool
	validateCert                              kadmin.CertValidationFunc
}

type transportOption string

const (
	transportOptionNotSelected transportOption = ""
	transportOptionPlaintext   transportOption = "PLAINTEXT"
	transportOptionTLS         transportOption = "TLS"
)

type verificationOption string

const (
	verificationOptionNotSelected verificationOption = ""
	verificationOptionBroker      verificationOption = "BROKER"
	verificationOptionSkip        verificationOption = "SKIP"
)

type clusterFormValues struct {
	name               string
	color              string
	host               string
	authMethod         config.AuthMethod
	transportOption    transportOption
	verificationOption verificationOption
	brokerCACertPath   string
	username           string
	password           string
	srURL              string
	srUsername         string
	srPassword         string
}

func (cv *clusterFormValues) toTLSConfig() config.TLSConfig {
	if cv.transportOption == transportOptionTLS {
		return config.TLSConfig{
			Enable:     true,
			SkipVerify: false,
			CACertPath: cv.brokerCACertPath,
		}
	}
	return config.TLSConfig{
		Enable:     false,
		SkipVerify: false,
		CACertPath: "",
	}
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if !ktx.Config().HasClusters() {
		builder := strings.Builder{}
		builder.WriteString("\n")
		builder.WriteString(lipgloss.NewStyle().PaddingLeft(1).Render("No clusters configured. Please create your first cluster!"))
		builder.WriteString("\n")
		views = append(views, renderer.Render(builder.String()))
	}

	notifierView := m.notifierCmdBar.View(ktx, renderer)

	deleteCmdbar := ""
	if m.kcModel.deleteCmdbar.IsFocussed() {
		deleteCmdbar = m.kcModel.deleteCmdbar.View(ktx, renderer)
	}

	var mainView string
	if m.border.ActiveTab() == kcTab {
		mainView = renderer.Render(lipgloss.
			NewStyle().
			Width(ktx.WindowWidth - 2).
			Render(m.kcModel.View(ktx, renderer)))
	} else {
		mainView = renderer.RenderWithStyle(m.form.View(), styles.Form)
	}

	mainView = m.border.View(lipgloss.NewStyle().
		PaddingBottom(ktx.AvailableHeight - 2).
		Render(mainView))

	views = append(views, deleteCmdbar, notifierView, mainView)

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	activeTab := m.border.ActiveTab()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if activeTab == kcTab {
				return m.kcModel.Update(msg)
			}
			m.title = "Clusters"
			return m.navigator.ToClustersPage()
		case "ctrl+r":
			m.cFormValues = &clusterFormValues{}
			if activeTab == cTab {
				m.authSelState = authMethodNone
				m.transportOption = transportOptionNotSelected
				m.verificationOption = verificationOptionNotSelected
				m.cForm = m.createCForm()
				m.form = m.cForm
			} else {
				m.srForm = m.createSrForm()
				m.form = m.srForm
			}
		case "f4":
			m.form = m.cForm
			m.border.GoTo("f4")
			return nil
		case "f5":
			m.border.GoTo("handling f5")
			if m.inEditingMode() {
				m.form = m.srForm
				m.form.State = huh.StateNormal
				m.border.GoTo("f5")
				return nil
			}

			return tea.Batch(
				m.notifierCmdBar.Notifier.ShowError(fmt.Errorf("create a cluster before adding a schema registry")),
				m.notifierCmdBar.Notifier.AutoHideCmd(notifierCmdbarTag),
			)
		case "f6":
			if m.inEditingMode() {
				m.form.State = huh.StateNormal
				m.border.GoTo("f6")
				return nil
			}

			return tea.Batch(
				m.notifierCmdBar.Notifier.ShowError(fmt.Errorf("create a cluster before adding a Kafka Connect Cluster")),
				m.notifierCmdBar.Notifier.AutoHideCmd(notifierCmdbarTag),
			)
		}
	case kadmin.ConnCheckStartedMsg:
		m.formState = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.ConnCheckSucceededMsg:
		m.formState = none
		cmds = append(cmds, m.registerCluster)
	case sradmin.ConnCheckStartedMsg:
		m.formState = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.ConnCheckSucceededMsg:
		m.formState = none
		return m.registerCluster
	case config.ClusterRegisteredMsg:
		m.preEditName = &msg.Cluster.Name
		m.clusterToEdit = msg.Cluster
		m.formState = none
		m.border.WithInActiveColor(styles.ColorGrey)
		if activeTab == cTab {
			m.cForm = m.createCForm()
			m.form = m.cForm
		} else if activeTab == srTab {
			m.srForm = m.createSrForm()
			m.form = m.srForm
		} else {
			m.kcModel.Update(msg)
		}
	}

	if activeTab == kcTab {
		cmd := m.kcModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	_, msg, cmd := m.notifierCmdBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if msg == nil {
		return tea.Batch(cmds...)
	}

	if activeTab == cTab || activeTab == srTab {
		form, cmd := m.form.Update(msg)
		cmds = append(cmds, cmd)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
	}

	if activeTab == cTab {
		t, done := m.updateClusterTab()
		if done {
			return t
		}
	}

	if activeTab == srTab {
		if m.form.State == huh.StateCompleted && m.formState != loading {
			return m.processSrSubmission()
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateClusterTab() (tea.Cmd, bool) {
	m.updateTransportOption()

	m.updateVO()

	if !m.cFormValues.selectedSASLAuthMethod() &&
		m.authSelState == authMethodSasl {
		// if SASL authentication mode was previously selected and switched back transportOption none
		m.cForm = m.createCForm()
		m.form = m.cForm
		if m.cFormValues.selectedTLSTransportOption() {
			if m.cFormValues.selectedBrokerVerificationOption() {
				m.nextField(6)
			} else {
				m.nextField(5)
			}
		} else {
			m.nextField(4)
		}
		m.authSelState = authMethodNone
	} else if m.cFormValues.selectedSASLAuthMethod() &&
		(m.authSelState == authMethodNotSelected || m.authSelState == authMethodNone) {
		// SASL authentication mode selected and previously nothing or none auth mode was selected
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.authSelState = authMethodSasl
		if m.cFormValues.selectedTLSTransportOption() {
			if m.cFormValues.selectedBrokerVerificationOption() {
				m.nextField(6)
			} else {
				m.nextField(5)
			}
		} else {
			m.nextField(4)
		}
	}

	if m.form.State == huh.StateCompleted && m.formState != loading {
		return m.processClusterSubmission(), true
	}
	return nil, false
}

func (m *Model) updateTransportOption() {
	if m.cFormValues.selectedTLSTransportOption() && m.prevSelPlaintextOrNoTO() {
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.transportOption = m.cFormValues.transportOption
		m.nextField(3)
	} else if m.cFormValues.selectedPlainTextTransportOption() && m.prevSelTlsTO() {
		m.transportOption = m.cFormValues.transportOption

		m.verificationOption = verificationOptionNotSelected
		m.cFormValues.verificationOption = verificationOptionNotSelected
		m.hasVerificationStatePreviouslyNotSelected = true

		m.cForm = m.createCForm()
		m.form = m.cForm
		m.nextField(3)
	}
}

func (m *Model) updateVO() {
	if m.cFormValues.selectedBrokerVerificationOption() && m.prevSelNoVO() {
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.verificationOption = verificationOptionBroker
		m.nextField(3)
		if m.hasVerificationStatePreviouslyNotSelected {
			m.hasVerificationStatePreviouslyNotSelected = false
		} else {
			m.nextField(1)
		}
	} else if !m.cFormValues.selectedBrokerVerificationOption() && m.verificationOption == verificationOptionBroker {
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.verificationOption = verificationOptionNotSelected
		m.nextField(4)
	}
}

func (m *Model) prevSelPlaintextOrNoTO() bool {
	return m.transportOption == transportOptionNotSelected || m.transportOption == transportOptionPlaintext
}

func (m *Model) prevSelNoVO() bool {
	return m.verificationOption == verificationOptionNotSelected
}

func (m *Model) prevSelTlsTO() bool {
	return m.transportOption == transportOptionTLS
}

func (m *Model) registerCluster() tea.Msg {
	details := m.getRegistrationDetails()
	return m.clusterRegisterer.RegisterCluster(details)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.border.ActiveTab() == kcTab {
		return m.kcModel.Shortcuts()
	}
	return m.shortcuts
}

func (m *Model) Title() string {
	if m.title == "" {
		return "Clusters / Create"
	}
	return m.title
}

func (m *Model) processSrSubmission() tea.Cmd {
	m.formState = loading
	details := m.getRegistrationDetails()

	cluster := config.ToCluster(details)
	return func() tea.Msg {
		return m.srConnChecker(cluster.SchemaRegistry)
	}
}

func (m *Model) processClusterSubmission() tea.Cmd {
	m.formState = loading
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
		name = m.cFormValues.name
		newName = nil
	} else { // When updating a cluster.
		name = *m.preEditName
		if m.cFormValues.name != *m.preEditName {
			newName = &m.cFormValues.name
		}
	}

	var authMethod config.AuthMethod
	if m.cFormValues.selectedSASLAuthMethod() {
		authMethod = m.cFormValues.authMethod
	} else {
		authMethod = config.AuthMethodNone
	}

	details := config.RegistrationDetails{
		Name:       name,
		NewName:    newName,
		Color:      m.cFormValues.color,
		Host:       m.cFormValues.host,
		AuthMethod: authMethod,
		TLSConfig:  m.cFormValues.toTLSConfig(),
		Username:   m.cFormValues.username,
		Password:   m.cFormValues.password,
	}
	if m.cFormValues.schemaRegistryEnabled() {
		details.SchemaRegistry = &config.SchemaRegistryDetails{
			Url:      m.cFormValues.srURL,
			Username: m.cFormValues.srUsername,
			Password: m.cFormValues.srPassword,
		}
	}

	details.KafkaConnectClusters = m.kcModel.clusterDetails()

	return details
}

func (cv *clusterFormValues) selectedSASLAuthMethod() bool {
	return cv.authMethod == config.AuthMethodSASLPlaintext
}

func (cv *clusterFormValues) schemaRegistryEnabled() bool {
	return len(cv.srURL) > 0
}

func (cv *clusterFormValues) selectedBrokerVerificationOption() bool {
	return cv.verificationOption == verificationOptionBroker
}

func (cv *clusterFormValues) selectedTLSTransportOption() bool {
	return cv.transportOption == transportOptionTLS
}

func (cv *clusterFormValues) selectedPlainTextTransportOption() bool {
	return cv.transportOption == transportOptionPlaintext
}

func (m *Model) nextField(count int) {
	for i := 0; i < count; i++ {
		m.form.NextField()
	}
}

func (m *Model) createCForm() *huh.Form {
	name := huh.NewInput().
		Value(&m.cFormValues.name).
		Title("Name").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("name cannot be empty")
			}
			if m.preEditName != nil {
				// When updating.
				if m.ktx.Config().FindClusterByName(v) != nil && v != *m.preEditName {
					return errors.New("cluster " + v + " already exists, name most be unique")
				}
			} else {
				// When creating a new cluster
				if m.ktx.Config().FindClusterByName(v) != nil {
					return errors.New("cluster " + v + " already exists, name most be unique")
				}
			}
			return nil
		})
	color := huh.NewSelect[string]().
		Value(&m.cFormValues.color).
		Title("Color ").
		Options(
			huh.NewOption(styles.Env.Colors.Green.Render("green"), styles.ColorGreen),
			huh.NewOption(styles.Env.Colors.Blue.Render("blue"), styles.ColorBlue),
			huh.NewOption(styles.Env.Colors.Orange.Render("orange"), styles.ColorOrange),
			huh.NewOption(styles.Env.Colors.Purple.Render("purple"), styles.ColorPurple),
			huh.NewOption(styles.Env.Colors.Yellow.Render("yellow"), styles.ColorYellow),
			huh.NewOption(styles.Env.Colors.Red.Render("red"), styles.ColorRed),
		).Inline(true)
	host := huh.NewInput().
		Value(&m.cFormValues.host).
		Title("Host").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("host cannot be empty")
			}
			return nil
		})

	transport := huh.NewSelect[transportOption]().
		Value(&m.cFormValues.transportOption).
		Title("Transport").
		Options(
			huh.NewOption("Plaintext", transportOptionPlaintext),
			huh.NewOption("TLS", transportOptionTLS),
		)

	var clusterFields []huh.Field
	clusterFields = append(clusterFields, name, color, host, transport)

	if m.cFormValues.selectedTLSTransportOption() {
		tlsVerification := huh.NewSelect[verificationOption]().
			Value(&m.cFormValues.verificationOption).
			Title("Verification").
			Options(
				huh.NewOption("Verify Broker Certificate", verificationOptionBroker),
				huh.NewOption("Skip verification (INSECURE)", verificationOptionSkip),
			)
		clusterFields = append(clusterFields, tlsVerification)
	}

	if m.cFormValues.selectedBrokerVerificationOption() {
		caCert := huh.NewInput().
			Value(&m.cFormValues.brokerCACertPath).
			Title("Path to Broker CA Certificate").
			Validate(func(certFile string) error {
				if certFile == "" {
					return errors.New("broker CA Certificate Path cannot be empty")
				}

				return m.validateCert(certFile)
			})
		clusterFields = append(clusterFields, caCert)
	}

	auth := huh.NewSelect[config.AuthMethod]().
		Value(&m.cFormValues.authMethod).
		Title("Authentication method").
		Options(
			huh.NewOption("NONE", config.AuthMethodNone),
			huh.NewOption("SASL_PLAINTEXT", config.AuthMethodSASLPlaintext),
		)

	clusterFields = append(clusterFields, auth)

	if m.cFormValues.selectedSASLAuthMethod() {
		username := huh.NewInput().
			Value(&m.cFormValues.username).
			Title("SASL username")
		pwd := huh.NewInput().
			Value(&m.cFormValues.password).
			EchoMode(huh.EchoModePassword).
			Title("SASL password")
		clusterFields = append(clusterFields, username, pwd)
	}

	form := huh.NewForm(
		huh.NewGroup(clusterFields...).
			Title("Cluster").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func (m *Model) createSrForm() *huh.Form {
	var fields []huh.Field
	srUrl := huh.NewInput().
		Value(&m.cFormValues.srURL).
		Title("Schema Registry URL")
	srUsername := huh.NewInput().
		Value(&m.cFormValues.srUsername).
		Title("Schema Registry Username")
	srPwd := huh.NewInput().
		Value(&m.cFormValues.srPassword).
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

func (m *Model) createNotifierCmdBar() {
	m.notifierCmdBar = cmdbar.NewNotifierCmdBar(notifierCmdbarTag)
	cmdbar.BindNotificationHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Testing cluster connectivity")
	})
	cmdbar.BindNotificationHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckSucceededMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Connection success creating cluster")
	})
	cmdbar.BindNotificationHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.formState = none
		nMsg := "Failed to Create Cluster"
		if m.inEditingMode() {
			nMsg = "Failed to Update Cluster"
		}
		return true, nm.ShowErrorMsg(nMsg, msg.Err)
	})
	cmdbar.BindNotificationHandler(m.notifierCmdBar, func(msg config.ClusterRegisteredMsg, nm *notifier.Model) (bool, tea.Cmd) {
		if m.form == m.srForm {
			nm.ShowSuccessMsg("Schema registry registered! <ESC> transportOption go back.")
		} else if m.form == m.cForm {
			if m.inEditingMode() {
				nm.ShowSuccessMsg("Cluster updated!")
			} else {
				nm.ShowSuccessMsg("Cluster registered! <ESC> transportOption go back or <F5> transportOption add a schema registry.")
			}
		} else {
			nm.ShowSuccessMsg("Cluster registered!")
		}
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})
	cmdbar.BindNotificationHandler(m.notifierCmdBar, func(msg sradmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.srForm = m.createSrForm()
		m.form = m.srForm
		m.formState = none
		nm.ShowErrorMsg("unable transportOption reach the schema registry", msg.Err)
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})
}

func (m *Model) inEditingMode() bool {
	return m.clusterToEdit != nil
}

func WithTitle(title string) Option {
	return func(m *Model) {
		m.title = title
	}
}

func initBorder(options ...border.Option) *border.Model {
	return border.New(
		append([]border.Option{
			border.WithTabs(
				border.Tab{Title: "Cluster ≪ F4 »", TabLabel: cTab},
				border.Tab{Title: "Schema Registry ≪ F5 »", TabLabel: srTab},
				border.Tab{Title: "Kafka Connect ≪ F6 »", TabLabel: kcTab},
			),
		}, options...)...)
}

func NewCreateClusterPage(
	navigator tabs.ClustersTabNavigator,
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	registerer config.ClusterRegisterer,
	ktx *kontext.ProgramKtx,
	shortcuts []statusbar.Shortcut,
	certValidator kadmin.CertValidationFunc,
	options ...Option,
) *Model {
	formValues := &clusterFormValues{}
	model := Model{
		navigator:     navigator,
		cFormValues:   formValues,
		kConnChecker:  kConnChecker,
		srConnChecker: srConnChecker,
		shortcuts:     shortcuts,
		validateCert:  certValidator,
	}

	model.ktx = ktx

	model.border = initBorder(border.WithInactiveColor(styles.ColorDarkGrey))

	model.cForm = model.createCForm()
	model.srForm = model.createSrForm()
	model.form = model.cForm

	model.createNotifierCmdBar()

	model.kcModel = NewUpsertKcModel(
		navigator,
		ktx,
		nil,
		[]config.KafkaConnectConfig{},
		kcadmin.CheckKafkaConnectClustersConn,
		model.notifierCmdBar,
		model.registerCluster,
	)

	model.clusterRegisterer = registerer

	model.authSelState = authMethodNotSelected
	model.transportOption = transportOptionNotSelected
	model.verificationOption = verificationOptionNotSelected
	model.hasVerificationStatePreviouslyNotSelected = true
	model.formState = none

	if model.cFormValues.selectedSASLAuthMethod() {
		model.authSelState = authMethodSasl
	}

	for _, option := range options {
		option(&model)
	}

	return &model
}

func NewEditClusterPage(
	navigator tabs.ClustersTabNavigator,
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	registerer config.ClusterRegisterer,
	connectClusterDeleter config.ConnectClusterDeleter,
	ktx *kontext.ProgramKtx,
	cluster config.Cluster,
	certValidator kadmin.CertValidationFunc,
	options ...Option,
) *Model {
	formValues := &clusterFormValues{
		name:  cluster.Name,
		color: cluster.Color,
		host:  cluster.BootstrapServers[0],
	}
	formValues.authMethod = cluster.SASLConfig.AuthMethod
	formValues.username = cluster.SASLConfig.Username
	formValues.password = cluster.SASLConfig.Password
	formValues.authMethod = cluster.SASLConfig.AuthMethod
	if cluster.TLSConfig.Enable {
		formValues.transportOption = transportOptionTLS
		if cluster.TLSConfig.SkipVerify {
			formValues.verificationOption = verificationOptionNotSelected
		} else {
			formValues.verificationOption = verificationOptionBroker
			formValues.brokerCACertPath = cluster.TLSConfig.CACertPath
		}
	} else {
		formValues.transportOption = transportOptionPlaintext
	}

	if cluster.SchemaRegistry != nil {
		formValues.srURL = cluster.SchemaRegistry.Url
		formValues.srUsername = cluster.SchemaRegistry.Username
		formValues.srPassword = cluster.SchemaRegistry.Password
	}
	model := Model{
		transportOption:    formValues.transportOption,
		verificationOption: formValues.verificationOption,
		navigator:          navigator,
		clusterToEdit:      &cluster,
		cFormValues:        formValues,
		kConnChecker:       kConnChecker,
		srConnChecker:      srConnChecker,
		shortcuts: []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Next Field", "tab"},
			{"Prev. Field", "s-tab"},
			{"Reset Form", "C-r"},
			{"Go Back", "esc"},
		},
		validateCert: certValidator,
	}
	if cluster.Name != "" {
		// copied transportOption prevent model.preEditedName transportOption follow the formValues.Name pointer
		preEditedName := cluster.Name
		model.preEditName = &preEditedName
	}
	model.ktx = ktx

	model.border = initBorder(border.WithInactiveColor(styles.ColorGrey))

	model.cForm = model.createCForm()
	model.srForm = model.createSrForm()
	model.form = model.cForm

	model.createNotifierCmdBar()

	model.kcModel = NewUpsertKcModel(
		navigator,
		ktx,
		func(name string) tea.Msg {
			return connectClusterDeleter.DeleteKafkaConnectCluster(cluster.Name, name)
		},
		cluster.KafkaConnectClusters,
		kcadmin.CheckKafkaConnectClustersConn,
		model.notifierCmdBar,
		model.registerCluster,
	)

	model.clusterRegisterer = registerer
	model.authSelState = authMethodNotSelected
	model.formState = none

	if model.cFormValues.selectedSASLAuthMethod() {
		model.authSelState = authMethodSasl
	}

	for _, o := range options {
		o(&model)
	}

	return &model
}
