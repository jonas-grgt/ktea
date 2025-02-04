package create_cluster_page

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"strings"
)

type authSelection int

type srSelection int

type formState int

type mode int

const (
	editMode           mode          = 0
	newMode            mode          = 1
	noneSelected       authSelection = 0
	saslSelected       authSelection = 1
	nothingSelected    authSelection = 2
	none               formState     = 0
	loading            formState     = 1
	srNothingSelected  srSelection   = 0
	srDisabledSelected srSelection   = 1
	srEnabledSelected  srSelection   = 2
)

type Model struct {
	form               *huh.Form
	formValues         *FormValues
	ktx                *kontext.ProgramKtx
	clusterRegisterer  config.ClusterRegisterer
	authSelectionState authSelection
	srSelectionState   srSelection
	state              formState
	preEditName        *string
	mode               mode
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Clusters / Create"
}

type FormValues struct {
	Name             string
	Color            string
	Host             string
	AuthMethod       config.AuthMethod
	SecurityProtocol config.SecurityProtocol
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
	views = append(views, renderer.RenderWithStyle(m.form.View(), styles.Form))
	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if !m.formValues.HasSASLAuthMethodSelected() &&
		m.authSelectionState == saslSelected {
		// if SASL authentication mode was previously selected and switched back to none
		m.form = m.createForm()
		m.NextField(3)
		m.authSelectionState = noneSelected
	} else if m.formValues.HasSASLAuthMethodSelected() &&
		(m.authSelectionState == nothingSelected || m.authSelectionState == noneSelected) {
		// SASL authentication mode selected and previously nothing or none auth mode was selected
		m.form = m.createForm()
		m.NextField(3)
		m.authSelectionState = saslSelected
	}

	// Schema Registry was previously enabled and switched back to disabled
	if !m.formValues.SrEnabled && m.srSelectionState == srEnabledSelected {
		m.form = m.createForm()
		m.NextField(4)
		if m.formValues.HasSASLAuthMethodSelected() {
			m.NextField(3)
		}
		m.srSelectionState = srDisabledSelected
	} else if m.formValues.SrEnabled &&
		((m.srSelectionState == srNothingSelected) || m.srSelectionState == srDisabledSelected) {
		// Schema Registry enabled selected and previously nothing or enabled selected
		m.form = m.createForm()
		m.NextField(4)
		if m.formValues.HasSASLAuthMethodSelected() {
			m.NextField(3)
		}
		m.srSelectionState = srEnabledSelected
	}

	if m.form.State == huh.StateCompleted && m.state != loading {
		m.state = loading
		return func() tea.Msg {
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
			return m.clusterRegisterer.RegisterCluster(details)
		}
	}
	return cmd
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
	srEnabled := huh.NewSelect[bool]().
		Value(&m.formValues.SrEnabled).
		Title("Schema Registry").
		Options(
			huh.NewOption("Disabled", false),
			huh.NewOption("Enabled", true),
		)
	var fields []huh.Field
	fields = append(fields, name, color, host, auth)
	if m.formValues.HasSASLAuthMethodSelected() {
		securityProtocol := huh.NewSelect[config.SecurityProtocol]().
			Value(&m.formValues.SecurityProtocol).
			Title("Security Protocol").
			Options(
				huh.NewOption("SASL_SSL", config.SSLSecurityProtocol),
				huh.NewOption("SASL_PLAINTEXT", config.PlaintextSecurityProtocol),
			)
		username := huh.NewInput().
			Value(&m.formValues.Username).
			Title("Username")
		pwd := huh.NewInput().
			Value(&m.formValues.Password).
			EchoMode(huh.EchoModePassword).
			Title("Password")
		fields = append(fields, securityProtocol, username, pwd)
	}

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

	form := huh.NewForm(huh.NewGroup(fields...))
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func NewForm(registerer config.ClusterRegisterer, ktx *kontext.ProgramKtx) *Model {
	var formValues = &FormValues{}
	model := Model{
		formValues: formValues,
	}

	model.form = model.createForm()
	model.mode = newMode
	model.clusterRegisterer = registerer
	model.ktx = ktx

	model.authSelectionState = nothingSelected
	if formValues.SrEnabled {
		model.srSelectionState = srEnabledSelected
	} else {
		model.srSelectionState = srDisabledSelected
	}
	model.state = none
	model.mode = editMode

	if model.formValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}
	model.srSelectionState = srNothingSelected
	return &model
}

func NewEditForm(registerer config.ClusterRegisterer, ktx *kontext.ProgramKtx, formValues *FormValues) *Model {
	model := Model{
		formValues: formValues,
	}
	if formValues.Name != "" {
		// copied to prevent model.preEditedName to follow the formValues.Name pointer
		preEditedName := formValues.Name
		model.preEditName = &preEditedName
	}
	model.form = model.createForm()
	model.clusterRegisterer = registerer
	model.ktx = ktx
	model.authSelectionState = nothingSelected
	if formValues.SrEnabled {
		model.srSelectionState = srEnabledSelected
	} else {
		model.srSelectionState = srDisabledSelected
	}
	model.state = none
	model.mode = editMode

	if model.formValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}

	return &model
}
