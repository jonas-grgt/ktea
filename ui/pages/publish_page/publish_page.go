package publish_page

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"errors"
	"fmt"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type state int

const (
	none       = 0
	publishing = 1
)

type Model struct {
	state      state
	topicForm  *huh.Form
	publisher  kadmin.Publisher
	topic      *kadmin.ListedTopic
	notifier   *notifier.Model
	formValues *formValues
}

type LoadPageMsg struct {
	Topic kadmin.ListedTopic
}

type formValues struct {
	Key       string
	Partition string
	Payload   string
	Headers   string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	notifierView := m.notifier.View(ktx, renderer)
	if notifierView != "" {
		notifierView = styles.CmdBarWithWidth(ktx.WindowWidth - cmdbar.BorderedPadding).Render(notifierView)
	}

	if m.topicForm == nil {
		m.topicForm = m.newForm(ktx)
	}

	return ui.JoinVertical(lipgloss.Top,
		notifierView,
		renderer.RenderWithStyle(m.topicForm.View(), styles.Form),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	switch msg := msg.(type) {
	case spinner.TickMsg, notifier.HideNotificationMsg:
		return m.notifier.Update(msg)
	case kadmin.PublicationStartedMsg:
		return tea.Batch(
			m.notifier.SpinWithLoadingMsg("Publishing record"),
			msg.AwaitCompletion,
		)
	case kadmin.PublicationFailed:
		m.state = none
		m.topicForm.Init()
		return m.notifier.ShowErrorMsg("Publication failed!", fmt.Errorf("TODO"))
	case kadmin.PublicationSucceeded:
		m.resetForm()
		return tea.Batch(
			m.notifier.ShowSuccessMsg("Record published!"),
			func() tea.Msg {
				time.Sleep(5 * time.Second)
				return notifier.HideNotificationMsg{}
			})
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.state == publishing {
				return nil
			}
			return ui.PublishMsg(nav.LoadTopicsPageMsg{})
		case "ctrl+r":
			m.resetForm()
		}
	}
	if m.topicForm != nil && m.state != publishing {
		form, cmd := m.topicForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.topicForm = f
		}
		if m.topicForm != nil && m.topicForm.State == huh.StateCompleted {
			m.state = publishing
			m.topicForm.State = huh.StateNormal
			return tea.Batch(
				m.notifier.SpinWithRocketMsg("Publishing record"),
				func() tea.Msg {
					var part *int
					if m.formValues.Partition != "" {
						if p, err := strconv.Atoi(m.formValues.Partition); err == nil {
							part = &p
						}
					}

					return m.publisher.PublishRecord(&kadmin.ProducerRecord{
						Key:       m.formValues.Key,
						Value:     []byte(m.formValues.Payload),
						Topic:     m.topic.Name,
						Headers:   m.formValues.parsedHeaders(),
						Partition: part,
					})
				})
		}
		return cmd
	}
	return nil
}

func (v *formValues) parsedHeaders() map[string]string {
	if v.Headers == "" {
		return map[string]string{}
	}
	headers := map[string]string{}
	for _, line := range strings.Split(v.Headers, "\n") {
		if strings.Contains(line, "=") {
			split := strings.Split(line, "=")
			key := split[0]
			value := split[1]
			headers[key] = value
		}
	}
	return headers
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Topics / " + m.topic.Name + " / Produce"
}

func (m *Model) resetForm() {
	m.state = none
	m.formValues.Key = ""
	m.formValues.Partition = ""
	m.formValues.Payload = ""
	m.formValues.Headers = ""
	m.topicForm = nil
}

func (m *Model) newForm(ktx *kontext.ProgramKtx) *huh.Form {
	payload := huh.NewText().
		ShowLineNumbers(true).
		Value(&m.formValues.Payload).
		Title("Payload").
		WithHeight(ktx.AvailableHeight - 10)
	key := huh.NewInput().
		Title("Key").
		Description("Leave empty to use a null key for the message.").
		Value(&m.formValues.Key)
	partition := huh.NewInput().
		Value(&m.formValues.Partition).
		Description("Leave empty to use murmur2 key based partitioner (identical to JVM clients).").
		Title("Partition").
		Validate(func(str string) error {
			if str == "" {
				return nil
			}
			if n, e := strconv.Atoi(str); e != nil {
				return errors.New(fmt.Sprintf("'%s' is not a valid numeric partition value", str))
			} else if n < 0 {
				return errors.New("value must be at least zero")
			} else if n > m.topic.PartitionCount-1 {
				return errors.New(fmt.Sprintf("partition index %s is invalid, valid range is 0-%d", str, m.topic.PartitionCount-1))
			}
			return nil
		})
	headers := huh.NewText().
		Description("Enter headers in the format key=value, one per line.").
		ShowLineNumbers(true).
		Value(&m.formValues.Headers).
		Title("Headers").
		WithHeight(10)

	form := huh.NewForm(
		huh.NewGroup(
			key,
			partition,
			headers,
		).WithWidth(ktx.WindowWidth/2),
		huh.NewGroup(
			payload,
		),
		huh.NewGroup(huh.NewConfirm().
			Inline(true).
			Affirmative("Produce").
			Negative(""),
		),
	)
	form.WithLayout(huh.LayoutGrid(4, 2))
	form.Init()
	return form
}

func New(p kadmin.Publisher, topic *kadmin.ListedTopic) *Model {
	return &Model{
		topic:      topic,
		publisher:  p,
		notifier:   notifier.New(),
		formValues: &formValues{},
	}
}
