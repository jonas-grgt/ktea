package consume_form_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"ktea/ui/tabs"
	"strconv"
)

type selectionState int

const (
	notSelected selectionState = iota
	selected
)

type Model struct {
	form                      *huh.Form
	formValues                *formValues
	windowResized             bool
	keyFilterSelectionState   selectionState
	valueFilterSelectionState selectionState
	ktx                       *kontext.ProgramKtx
	availableHeight           int
	topic                     *kadmin.ListedTopic
	navigator                 tabs.TopicsTabNavigator
}

type formValues struct {
	startPoint      kadmin.StartPoint
	limit           int
	partitions      []int
	keyFilter       kadmin.FilterType
	keyFilterTerm   string
	valueFilter     kadmin.FilterType
	valueFilterTerm string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.form == nil {
		m.availableHeight = ktx.AvailableHeight
		m.form = m.newForm(m.topic.PartitionCount, m.ktx)
	}

	if m.windowResized {
		m.windowResized = false
		m.availableHeight = ktx.AvailableHeight
		m.form = m.newForm(m.topic.PartitionCount, ktx)
	}

	return renderer.RenderWithStyle(m.form.View(), styles.Form)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.form == nil {
		return nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.formValues.keyFilter != kadmin.NoFilterType && m.keyFilterSelectionState == notSelected {
		// if key filter type is selected and previously not selected
		m.keyFilterSelectionState = selected
		m.form = m.newForm(m.topic.PartitionCount, m.ktx)
		m.NextField(3)
		m.form.NextGroup()
	} else if m.formValues.keyFilter == kadmin.NoFilterType && m.keyFilterSelectionState == selected {
		// if no key filter type is selected and previously selected
		m.keyFilterSelectionState = notSelected
		m.form = m.newForm(m.topic.PartitionCount, m.ktx)
		m.NextField(3)
		m.form.NextGroup()
	}

	if m.formValues.valueFilter != kadmin.NoFilterType && m.valueFilterSelectionState == notSelected {
		// if value filter type is selected and previously not selected
		m.valueFilterSelectionState = selected
		m.form = m.newForm(m.topic.PartitionCount, m.ktx)
		m.NextField(3)
		m.form.NextGroup()
		m.NextField(1)
	} else if m.formValues.valueFilter == kadmin.NoFilterType && m.valueFilterSelectionState == selected {
		// if no key filter type is selected and previously selected
		m.valueFilterSelectionState = notSelected
		m.form = m.newForm(m.topic.PartitionCount, m.ktx)
		m.NextField(3)
		m.form.NextGroup()
		m.NextField(1)
	}

	switch msg.(type) {
	case tea.WindowSizeMsg:
		m.windowResized = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return ui.PublishMsg(nav.LoadTopicsPageMsg{})
		}
	}

	filter := kadmin.Filter{}
	if m.formValues.keyFilter != kadmin.NoFilterType {
		filter.KeySearchTerm = m.formValues.keyFilterTerm
		filter.KeyFilter = m.formValues.keyFilter
	}
	if m.formValues.valueFilter != kadmin.NoFilterType {
		filter.ValueSearchTerm = m.formValues.valueFilterTerm
		filter.ValueFilter = m.formValues.valueFilter
	}
	if m.form.State == huh.StateCompleted {
		return m.submit(filter)
	}
	return cmd
}

func (m *Model) submit(filter kadmin.Filter) tea.Cmd {
	var partToConsume []int
	if m.noPartitionsSelected() {
		// consume from all partitions
		partToConsume = m.topic.Partitions()
	} else {
		partToConsume = m.formValues.partitions
	}

	return m.navigator.ToConsumePage(nav.ConsumePageDetails{
		Origin: nav.OriginConsumeFormPage,
		Topic:  m.topic,
		ReadDetails: kadmin.ReadDetails{
			TopicName:       m.topic.Name,
			PartitionToRead: partToConsume,
			StartPoint:      m.formValues.startPoint,
			Limit:           m.formValues.limit,
			Filter:          &filter,
		},
	})
}

func (m *Model) noPartitionsSelected() bool {
	return len(m.formValues.partitions) == 0
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Select Partition", "space"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Consumption details"
}

func (m *Model) newForm(partitions int, ktx *kontext.ProgramKtx) *huh.Form {
	var partOptions []huh.Option[int]
	for i := 0; i < partitions; i++ {
		partOptions = append(partOptions, huh.NewOption[int](strconv.Itoa(i), i))
	}
	optionsHeight := 13 // 12 fixed height of form minus partitions field + padding and margins
	if len(partOptions) < 13 {
		optionsHeight = len(partOptions) + 2 // 2 for field title + padding
	} else {
		optionsHeight = m.availableHeight - optionsHeight
	}
	topicGroup := huh.NewGroup(
		huh.NewSelect[kadmin.StartPoint]().
			Value(&m.formValues.startPoint).
			Title("Start form").
			Options(
				huh.NewOption("Beginning", kadmin.Beginning),
				huh.NewOption("Most Recent", kadmin.MostRecent)),
		huh.NewMultiSelect[int]().
			Value(&m.formValues.partitions).
			Height(optionsHeight).
			Title("Partitions").
			Description(m.getPartitionDescription(ktx)).
			Options(partOptions...),
		huh.NewSelect[int]().
			Value(&m.formValues.limit).
			Title("Limit").
			Options(
				huh.NewOption("50", 50),
				huh.NewOption("500", 500),
				huh.NewOption("5000", 5000)),
	)
	filterGroup := m.createFilterGroup()
	form := huh.NewForm(
		topicGroup.WithWidth(ktx.WindowWidth/2),
		filterGroup,
	)
	form.WithLayout(huh.LayoutColumns(2))
	form.Init()
	return form
}

func (m *Model) createFilterGroup() *huh.Group {
	var fields []huh.Field

	fields = append(fields, m.keyFilterTypeField())
	if m.formValues.keyFilter != kadmin.NoFilterType {
		fields = append(fields, m.keyFilterTermField())
	}

	fields = append(fields, m.valueFilterTypeField())
	if m.formValues.valueFilter != kadmin.NoFilterType {
		fields = append(fields, m.valueFilterTermField())
	}

	return huh.NewGroup(fields...)
}

func (m *Model) valueFilterTermField() *huh.Input {
	return huh.NewInput().
		Value(&m.formValues.valueFilterTerm).
		Title("Value Filter Term")
}

func (m *Model) valueFilterTypeField() *huh.Select[kadmin.FilterType] {
	return huh.NewSelect[kadmin.FilterType]().
		Value(&m.formValues.valueFilter).
		Title("Value Filter Type").
		Options(
			huh.NewOption("None", kadmin.NoFilterType),
			huh.NewOption("Contains", kadmin.ContainsFilterType),
			huh.NewOption("Starts With", kadmin.StartsWithFilterType))
}

func (m *Model) keyFilterTermField() *huh.Input {
	return huh.NewInput().
		Value(&m.formValues.keyFilterTerm).
		Title("Key Filter Term")
}

func (m *Model) keyFilterTypeField() *huh.Select[kadmin.FilterType] {
	return huh.NewSelect[kadmin.FilterType]().
		Value(&m.formValues.keyFilter).
		Title("Key Filter Type").
		Options(
			huh.NewOption("None", kadmin.NoFilterType),
			huh.NewOption("Contains", kadmin.ContainsFilterType),
			huh.NewOption("Starts With", kadmin.StartsWithFilterType))
}

// hack until https://github.com/charmbracelet/huh/issues/525 has been resolved
func (m *Model) getPartitionDescription(ktx *kontext.ProgramKtx) string {
	partitionDescription := "Select none to consume from all available partitions"
	columnWidth := ktx.WindowWidth / 2
	extraSpaces := columnWidth - lipgloss.Width(partitionDescription)
	for i := 0; i < extraSpaces; i++ {
		partitionDescription += " "
	}
	return partitionDescription
}

func (m *Model) NextField(count int) {
	for i := 0; i < count; i++ {
		m.form.NextField()
	}
}

func NewWithDetails(
	details *kadmin.ReadDetails,
	topic *kadmin.ListedTopic,
	navigator tabs.TopicsTabNavigator,
	ktx *kontext.ProgramKtx,
) *Model {
	var partitionsToRead []int
	if topic.PartitionCount != len(details.PartitionToRead) {
		partitionsToRead = details.PartitionToRead
	}
	return &Model{
		ktx:       ktx,
		navigator: navigator,
		topic:     topic,
		formValues: &formValues{
			startPoint:      details.StartPoint,
			limit:           details.Limit,
			partitions:      partitionsToRead,
			keyFilter:       details.Filter.KeyFilter,
			keyFilterTerm:   details.Filter.KeySearchTerm,
			valueFilter:     details.Filter.ValueFilter,
			valueFilterTerm: details.Filter.ValueSearchTerm,
		}}
}

func New(
	topic *kadmin.ListedTopic,
	navigator tabs.TopicsTabNavigator,
	ktx *kontext.ProgramKtx,
) *Model {
	return &Model{
		topic:      topic,
		navigator:  navigator,
		formValues: &formValues{},
		ktx:        ktx,
	}
}
