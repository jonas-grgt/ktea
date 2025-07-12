package create_cluster_page

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"strings"
)

type authSelection int

type srSelection int

type activeForm int

type formState int

type Option func(m *Model)

const (
	noneSelected       authSelection = 0
	saslSelected       authSelection = 1
	nothingSelected    authSelection = 2
	none               formState     = 0
	loading            formState     = 1
	srNothingSelected  srSelection   = 0
	srDisabledSelected srSelection   = 1
	srEnabledSelected  srSelection   = 2
	clusterActive      activeForm    = 0
	srActive           activeForm    = 1
)

type Model struct {
	form               *huh.Form
	formValues         *FormValues
	notifierCmdBar     *cmdbar.NotifierCmdBar
	ktx                *kontext.ProgramKtx
	clusterRegisterer  config.ClusterRegisterer
	connChecker        kadmin.ConnChecker
	authSelectionState authSelection
	srSelectionState   srSelection
	state              formState
	preEditName        *string
	shortcuts          []statusbar.Shortcut
	title              string
	activeForm         activeForm
}

type FormValues struct {
	Name             string
	Color            string
	Host             string
	AuthMethod       config.AuthMethod
	SecurityProtocol config.SecurityProtocol
	SSLEnabled       bool
	Username         string
	Password         string
	SrEnabled        bool
	SrUrl            string
	SrUsername       string
	SrPassword       string
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
	b := border.New(
		border.WithTabs("Cluster ≪ F1 »", "Schema Registry ≪ F2 »", "Kafka Connect ≪ F3 »"))
	formView = b.View(lipgloss.NewStyle().
		PaddingBottom(ktx.AvailableHeight - 1).
		Render(formView))
	views = append(views, notifierView, formView)

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			m.formValues = &FormValues{}
			m.form = m.createForm()
			m.authSelectionState = noneSelected
			m.srSelectionState = srNothingSelected
		case "f2":
			m.form = m.createSrForm()
			m.activeForm = srActive
		}
	case kadmin.ConnCheckStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.ConnCheckSucceededMsg:
		cmds = append(cmds, func() tea.Msg {
			details := m.getRegistrationDetails()
			return m.clusterRegisterer.RegisterCluster(details)
		})
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

	if m.activeForm == clusterActive {
		if !m.formValues.HasSASLAuthMethodSelected() &&
			m.authSelectionState == saslSelected {
			// if SASL authentication mode was previously selected and switched back to none
			m.form = m.createForm()
			m.NextField(4)
			m.authSelectionState = noneSelected
		} else if m.formValues.HasSASLAuthMethodSelected() &&
			(m.authSelectionState == nothingSelected || m.authSelectionState == noneSelected) {
			// SASL authentication mode selected and previously nothing or none auth mode was selected
			m.form = m.createForm()
			m.NextField(4)
			m.authSelectionState = saslSelected
		}
	}

	if m.activeForm == srActive {
		// Schema Registry was previously enabled and switched back to disabled
		if !m.formValues.SrEnabled && m.srSelectionState == srEnabledSelected {
			m.form = m.createForm()
			m.NextField(4)
			if m.formValues.HasSASLAuthMethodSelected() {
				m.NextField(3)
			}
			m.form.NextGroup()
			m.srSelectionState = srDisabledSelected
		} else if m.formValues.SrEnabled &&
			((m.srSelectionState == srNothingSelected) || m.srSelectionState == srDisabledSelected) {
			// Schema Registry enabled selected and previously nothing or enabled selected
			m.form = m.createSrForm()
			m.NextField(4)
			if m.formValues.HasSASLAuthMethodSelected() {
				m.NextField(3)
			}
			m.form.NextGroup()
			m.srSelectionState = srEnabledSelected
			return nil
		}
	}

	if m.form.State == huh.StateCompleted && m.state != loading {
		return m.processFormSubmission()
	}
	return tea.Batch(cmds...)
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

func (m *Model) processFormSubmission() tea.Cmd {
	m.state = loading
	details := m.getRegistrationDetails()

	cluster := config.ToCluster(details)
	return func() tea.Msg {
		return m.connChecker(&cluster)
	}
}

func (m *Model) getRegistrationDetails() config.RegistrationDetails {
	var name string
	var newName *string
	if m.preEditName == nil { // When creating a cluster
		name = m.formValues.Name
		newName = nil
	} else { // When updating a cluster.
		name = *m.preEditName
		if m.formValues.Name != *m.preEditName {
			newName = &m.formValues.Name
		}
	}

	var authMethod config.AuthMethod
	var securityProtocol config.SecurityProtocol
	if m.formValues.HasSASLAuthMethodSelected() {
		authMethod = config.SASLAuthMethod
		securityProtocol = m.formValues.SecurityProtocol
	} else {
		authMethod = config.NoneAuthMethod
	}

	details := config.RegistrationDetails{
		Name:             name,
		NewName:          newName,
		Color:            m.formValues.Color,
		Host:             m.formValues.Host,
		AuthMethod:       authMethod,
		SecurityProtocol: securityProtocol,
		SSLEnabled:       m.formValues.SSLEnabled,
		Username:         m.formValues.Username,
		Password:         m.formValues.Password,
	}
	if m.formValues.SrEnabled {
		details.SchemaRegistry = &config.SchemaRegistryDetails{
			Url:      m.formValues.SrUrl,
			Username: m.formValues.SrUsername,
			Password: m.formValues.SrPassword,
		}
	}
	return details
}

func (f *FormValues) HasSASLAuthMethodSelected() bool {
	return f.AuthMethod == config.SASLAuthMethod
}

func (m *Model) NextField(count int) {
	for i := 0; i < count; i++ {
		m.form.NextField()
	}
}

func (m *Model) createForm() *huh.Form {
	name := huh.NewInput().
		Value(&m.formValues.Name).
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
		Value(&m.formValues.Color).
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
		Value(&m.formValues.Host).
		Title("Host").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("Host cannot be empty")
			}
			return nil
		})
	auth := huh.NewSelect[config.AuthMethod]().
		Value(&m.formValues.AuthMethod).
		Title("Authentication method").
		Options(
			huh.NewOption("NONE", config.NoneAuthMethod),
			huh.NewOption("SASL", config.SASLAuthMethod),
		)

	sslEnabled := huh.NewSelect[bool]().
		Value(&m.formValues.SSLEnabled).
		Title("SSL").
		Options(
			huh.NewOption("Disable SSL", false),
			huh.NewOption("Enable SSL", true),
		)

	var clusterFields []huh.Field
	clusterFields = append(clusterFields, name, color, host, sslEnabled, auth)

	if m.formValues.HasSASLAuthMethodSelected() {
		securityProtocol := huh.NewSelect[config.SecurityProtocol]().
			Value(&m.formValues.SecurityProtocol).
			Title("Security Protocol").
			Options(
				huh.NewOption("SASL_PLAINTEXT", config.SASLPlaintextSecurityProtocol),
			)
		username := huh.NewInput().
			Value(&m.formValues.Username).
			Title("Username")
		pwd := huh.NewInput().
			Value(&m.formValues.Password).
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
	srEnabled := huh.NewSelect[bool]().
		Value(&m.formValues.SrEnabled).
		Title("Schema Registry").
		Options(
			huh.NewOption("Disabled", false),
			huh.NewOption("Enabled", true),
		)
	var fields []huh.Field
	fields = append(fields, srEnabled)
	if m.formValues.SrEnabled {
		srUrl := huh.NewInput().
			Value(&m.formValues.SrUrl).
			Title("Schema Registry URL")
		srUsername := huh.NewInput().
			Value(&m.formValues.SrUsername).
			Title("Schema Registry Username")
		srPwd := huh.NewInput().
			Value(&m.formValues.SrPassword).
			EchoMode(huh.EchoModePassword).
			Title("Schema Registry Password")
		fields = append(fields, srUrl, srUsername, srPwd)
	}
	form := huh.NewForm(
		huh.NewGroup(fields...).
			Title("Cluster").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.WithLayout(huh.LayoutColumns(1))
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func (m *Model) createNotifierCmdBar() {
	m.notifierCmdBar = cmdbar.NewNotifierCmdBar("upsert-cluster-page")
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Testing cluster connectivity")
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckSucceededMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Connection success creating cluster")
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.form = m.createForm()
		m.state = none
		return true, nm.ShowErrorMsg("Cluster not created", msg.Err)
	})
}

func WithTitle(title string) Option {
	return func(m *Model) {
		m.title = title
	}
}

func NewForm(
	connChecker kadmin.ConnChecker,
	registerer config.ClusterRegisterer,
	ktx *kontext.ProgramKtx,
	shortcuts []statusbar.Shortcut,
	options ...Option,
) *Model {
	var formValues = &FormValues{}
	model := Model{
		formValues:  formValues,
		connChecker: connChecker,
		shortcuts:   shortcuts,
	}

	model.ktx = ktx
	model.form = model.createForm()
	model.clusterRegisterer = registerer

	model.authSelectionState = nothingSelected
	if formValues.SrEnabled {
		model.srSelectionState = srEnabledSelected
	} else {
		model.srSelectionState = srDisabledSelected
	}
	model.state = none

	if model.formValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}
	model.srSelectionState = srNothingSelected

	model.createNotifierCmdBar()

	for _, option := range options {
		option(&model)
	}

	return &model
}

func NewEditForm(
	connChecker kadmin.ConnChecker,
	registerer config.ClusterRegisterer,
	ktx *kontext.ProgramKtx,
	formValues *FormValues,
) *Model {
	model := Model{
		formValues:  formValues,
		connChecker: connChecker,
		shortcuts: []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Next Field", "tab"},
			{"Prev. Field", "s-tab"},
			{"Reset Form", "C-r"},
			{"Go Back", "esc"},
		},
	}
	if formValues.Name != "" {
		// copied to prevent model.preEditedName to follow the formValues.Name pointer
		preEditedName := formValues.Name
		model.preEditName = &preEditedName
	}
	model.ktx = ktx
	model.form = model.createForm()
	model.clusterRegisterer = registerer
	model.authSelectionState = nothingSelected
	if formValues.SrEnabled {
		model.srSelectionState = srEnabledSelected
	} else {
		model.srSelectionState = srDisabledSelected
	}
	model.state = none

	if model.formValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}

	model.createNotifierCmdBar()

	return &model
}
