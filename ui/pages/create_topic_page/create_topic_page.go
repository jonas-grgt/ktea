package create_topic_page

import (
	"errors"
	"fmt"
	bsp "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"regexp"
	"strconv"
	"strings"
)

type formState int

const (
	initial       formState = 0
	configEntered formState = 1
	loading       formState = 2
)

type Model struct {
	shortcuts              []statusbar.Shortcut
	form                   *huh.Form
	notifier               *cmdbar.NotifierCmdBar
	topicCreator           kadmin.TopicCreator
	formValues             topicFormValues
	formState              formState
	createdAtLeastOneTopic bool
}

type config struct {
	key   string
	value string
}

type topicFormValues struct {
	name              string
	numPartitions     string
	config            string
	configs           []config
	cleanupPolicy     string
	replicationFactor string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	notifierView := m.notifier.View(ktx, renderer)
	formView := renderer.RenderWithStyle(m.form.View(), styles.Form)
	views = append(views, notifierView, formView)

	if len(m.formValues.configs) > 0 {
		views = append(views, renderer.Render("Custom Topic configurations:\n\n"))
		for _, c := range m.formValues.configs {
			views = append(views, renderer.Render(fmt.Sprintf("%s: %s\n", c.key, c.value)))
		}
	}

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	_, _, cmd := m.notifier.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case kadmin.TopicCreationStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
		return tea.Batch(cmds...)
	case kadmin.TopicCreationErrMsg:
		m.initForm(initial)
		return tea.Batch(cmds...)
	case bsp.TickMsg:
		return tea.Batch(cmds...)
	case tea.KeyMsg:
		if msg.String() == "esc" && m.formState != loading {
			return ui.PublishMsg(nav.LoadTopicsPageMsg{Refresh: m.createdAtLeastOneTopic})
		} else if msg.String() == "ctrl+r" {
			m.formValues.name = ""
			m.formValues.cleanupPolicy = ""
			m.formValues.config = ""
			m.formValues.numPartitions = ""
			m.formValues.replicationFactor = ""
			m.formValues.configs = []config{}
			m.initForm(initial)
			return propagateMsgToForm(m, msg)
		} else {
			return propagateMsgToForm(m, msg)
		}
	case kadmin.TopicCreatedMsg:
		m.formValues.name = ""
		m.formValues.cleanupPolicy = ""
		m.formValues.config = ""
		m.formValues.numPartitions = ""
		m.formValues.replicationFactor = ""
		m.formValues.configs = []config{}
		m.createdAtLeastOneTopic = true
		m.initForm(initial)
		return nil
	default:
		return propagateMsgToForm(m, msg)
	}

}

func propagateMsgToForm(m *Model, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	form, c := m.form.Update(msg)
	cmd = c
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted && m.formState != loading {
		if m.formValues.config == "" {
			m.formState = loading
			return tea.Batch(
				//m.notifier.SpinWithRocketMsg("Creating topic"),
				func() tea.Msg {
					numPartitions, _ := strconv.Atoi(m.formValues.numPartitions)
					configs := map[string]string{
						"cleanup.policy": m.formValues.cleanupPolicy,
					}
					for _, c := range m.formValues.configs {
						configs[c.key] = c.value
					}
					replicationFactor, _ := strconv.Atoi(m.formValues.replicationFactor)
					return m.topicCreator.CreateTopic(
						kadmin.TopicCreationDetails{
							Name:              m.formValues.name,
							NumPartitions:     numPartitions,
							Properties:        configs,
							ReplicationFactor: int16(replicationFactor),
						})
				})
		} else {
			m.formState = configEntered
			split := strings.Split(m.formValues.config, "=")
			m.formValues.configs = append(m.formValues.configs, config{split[0], split[1]})
			m.formValues.config = ""
			m.initForm(configEntered)
			return cmd
		}
	}
	return cmd
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return m.shortcuts
}

func (m *Model) Title() string {
	return "Topics / Create"
}

func (m *Model) initForm(fs formState) {
	topicNameInput := huh.NewInput().
		Title("Topic name").
		Value(&m.formValues.name).
		Validate(func(str string) error {
			if str == "" {
				return errors.New("topic Name cannot be empty")
			}
			return nil
		})

	numPartField := huh.NewInput().
		Title("Number of Partitions").
		Value(&m.formValues.numPartitions).
		Validate(func(str string) error {
			if str == "" {
				return errors.New("number of Partitions cannot be empty")
			}
			if n, e := strconv.Atoi(str); e != nil {
				return errors.New(fmt.Sprintf("'%s' is not a valid numeric partition count value", str))
			} else if n <= 0 {
				return errors.New("value must be greater than zero")
			}
			return nil
		})

	replicationFactorField := huh.NewInput().
		Title("Replication Factor").
		Validate(func(r string) error {
			if r == "" {
				return errors.New("replication factory cannot be empty")
			}
			if n, e := strconv.Atoi(r); e != nil {
				return errors.New(fmt.Sprintf("'%s' is not a valid numeric replication factor value", r))
			} else if n <= 0 {
				return errors.New("value must be greater than zero")
			}
			return nil
		}).
		Value(&m.formValues.replicationFactor)

	cleanupPolicySelect := huh.NewSelect[string]().
		Title("Cleanup Policy").
		Value(&m.formValues.cleanupPolicy).
		Options(
			huh.NewOption("Retention (delete)", "delete"),
			huh.NewOption("Compaction (compact)", "compact"),
			huh.NewOption("Retention + Compaction", "delete-compact"))

	configInput := huh.NewInput().
		Description("Enter custom topic configurations in the format config=value. Leave blank to create the topic.").
		Title("Config").
		Key("configKey").
		Suggestions([]string{
			"cleanup.policy",
			"compression.type",
			"delete.retention.ms",
			"file.delete.delay.ms",
			"flush.messages",
			"flush.ms",
			"follower.replication.throttled.replicas",
			"index.interval.bytes",
			"leader.replication.throttled.replicas",
			"local.retention.bytes",
			"local.retention.ms",
			"max.compaction.lag.ms",
			"max.message.bytes",
			"message.format.version",
			"message.timestamp.difference.max.ms",
			"message.timestamp.type",
			"min.cleanable.dirty.ratio",
			"min.compaction.lag.ms",
			"min.insync.replicas",
			"preallocate",
			"remote.storage.enable",
			"retention.bytes",
			"retention.ms",
			"segment.bytes",
			"segment.index.bytes",
			"segment.jitter.ms",
			"segment.ms",
			"unclean.leader.election.enable",
			"message.downconversion.enable",
		}).
		Validate(func(str string) error {
			if str == "" {
				return nil
			}

			r, _ := regexp.Compile(`^[\w.]+=\w+$`)
			if r.MatchString(str) {
				return nil
			}

			return errors.New("please enter configurations in the format \"config=value\"")
		}).
		Value(&m.formValues.config)

	form := huh.NewForm(huh.NewGroup(
		topicNameInput,
		numPartField,
		replicationFactorField,
		cleanupPolicySelect,
		configInput,
	))
	form.QuitAfterSubmit = false
	if m.formState == configEntered {
		form.NextField()
		form.NextField()
		form.NextField()
		form.NextField()
	}
	form.Init()
	m.formState = fs
	m.form = form
}

func New(tc kadmin.TopicCreator) *Model {
	var t = Model{}
	t.topicCreator = tc
	t.shortcuts = []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
	t.initForm(initial)
	notifierCmdBar := cmdbar.NewNotifierCmdBar("create-topic-page")
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg kadmin.TopicCreationStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Creating Topic")
		return true, cmd
	})
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg kadmin.TopicCreationErrMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Failed to create Topic", msg.Err)
		return true, nil
	})
	cmdbar.BindNotificationHandler(notifierCmdBar, func(msg kadmin.TopicCreatedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Topic created!")
		return true, m.AutoHideCmd("create-topic-page")
	})
	t.notifier = notifierCmdBar

	return &t
}
