package consume_page

import (
	"fmt"
	"ktea/kadmin"
	"ktea/styles"
	"ktea/ui/components/statusbar"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lg "charm.land/lipgloss/v2"
)

type startPoint int

const (
	mostRecent startPoint = iota
	beginning
	live
	relativeDate
	absoluteDate
)

const maxModalWidth = 72

// searchValues holds what the user entered; the huh fields write into it.
type searchValues struct {
	term               string
	keyFilter          kadmin.FilterType
	keyTerm            string
	valueFilter        kadmin.FilterType
	valueTerm          string
	startFrom          startPoint
	relativeStartPoint kadmin.StartPoint
	absoluteStartPoint string
	partitions         []int
	limit              int
}

// searchTab groups the fields shown under one tab header.
type searchTab struct {
	name   string
	fields []huh.Field
}

// searchModal searches a topic. In simple mode it only asks for a term matched
// against key or value; expanded, it offers tabs to refine filters, start point,
// partitions and limit. Every tab is a huh group, so tab/shift+tab walk across them.
type searchModal struct {
	topic    *kadmin.ListedTopic
	values   searchValues
	expanded bool
	form     *huh.Form
	tabs     []searchTab
	err      string
}

func newSearchModal(topic *kadmin.ListedTopic, details kadmin.ReadDetails, expanded bool) *searchModal {
	m := &searchModal{
		topic:    topic,
		values:   toSearchValues(topic, details),
		expanded: expanded,
	}
	m.buildForm()
	return m
}

func toSearchValues(topic *kadmin.ListedTopic, details kadmin.ReadDetails) searchValues {
	v := searchValues{
		keyFilter:          kadmin.ContainsFilterType,
		valueFilter:        kadmin.ContainsFilterType,
		startFrom:          toFormStartPoint(details.StartPoint),
		relativeStartPoint: kadmin.Today,
		absoluteStartPoint: toAbsoluteStartPoint(details.StartPoint),
		limit:              details.Limit,
	}
	if v.startFrom == relativeDate {
		v.relativeStartPoint = details.StartPoint
	}
	// all partitions is shown as none selected
	if len(details.PartitionToRead) != topic.PartitionCount {
		v.partitions = details.PartitionToRead
	}
	if f := details.Filter; f != nil {
		v.term = f.Term
		if f.KeyFilter == kadmin.ContainsFilterType || f.KeyFilter == kadmin.StartsWithFilterType {
			v.keyFilter, v.keyTerm = f.KeyFilter, f.KeySearchTerm
		}
		if f.ValueFilter == kadmin.ContainsFilterType || f.ValueFilter == kadmin.StartsWithFilterType {
			v.valueFilter, v.valueTerm = f.ValueFilter, f.ValueSearchTerm
		}
	}
	return v
}

func (m *searchModal) buildForm() {
	v := &m.values
	m.tabs = []searchTab{{name: "Key or Value", fields: []huh.Field{
		huh.NewInput().
			Value(&v.term).
			Title("Search").
			Description("Matches records whose key or value contains the term (ignoring case)"),
	}}}

	if m.expanded {
		var partOptions []huh.Option[int]
		for i := range m.topic.PartitionCount {
			partOptions = append(partOptions, huh.NewOption(fmt.Sprintf("Partition %d", i), i))
		}

		m.tabs = append(m.tabs,
			searchTab{name: "Key", fields: []huh.Field{
				filterTypeField(&v.keyFilter, "Key filter"),
				huh.NewInput().Value(&v.keyTerm).Title("Key term").Description("Leave empty to not filter on key"),
			}},
			searchTab{name: "Value", fields: []huh.Field{
				filterTypeField(&v.valueFilter, "Value filter"),
				huh.NewInput().Value(&v.valueTerm).Title("Value term").Description("Leave empty to not filter on value"),
			}},
			searchTab{name: "Start", fields: []huh.Field{
				huh.NewSelect[startPoint]().
					Value(&v.startFrom).
					Title("Start from").
					Options(
						huh.NewOption("Most Recent", mostRecent),
						huh.NewOption("Beginning", beginning),
						huh.NewOption("Live", live),
						huh.NewOption("Relative Date", relativeDate),
						huh.NewOption("Absolute Date", absoluteDate)),
				huh.NewSelect[kadmin.StartPoint]().
					Value(&v.relativeStartPoint).
					Title("Relatively start from").
					Options(
						huh.NewOption("Today", kadmin.Today),
						huh.NewOption("Yesterday", kadmin.Yesterday),
						huh.NewOption("Week ago", kadmin.Last7Days)),
				huh.NewInput().
					Value(&v.absoluteStartPoint).
					Title("Absolutely start from").
					Description("format(RFC3339): 1986-01-16T23:20:50.52Z").
					Validate(validateAbsoluteDate),
			}},
			searchTab{name: "Partitions", fields: []huh.Field{
				huh.NewMultiSelect[int]().
					Value(&v.partitions).
					Title("Partitions").
					Description("Select none to consume from all partitions").
					Height(min(len(partOptions)+2, 12)).
					Options(partOptions...),
			}},
			searchTab{name: "Limit", fields: []huh.Field{
				huh.NewSelect[int]().
					Value(&v.limit).
					Title("Limit").
					Options(
						huh.NewOption("50", 50),
						huh.NewOption("500", 500),
						huh.NewOption("5000", 5000)),
			}},
		)
	}

	var groups []*huh.Group
	for _, t := range m.tabs {
		if t.name != "Start" {
			groups = append(groups, huh.NewGroup(t.fields...))
			continue
		}
		// the date sub-fields get their own group so they only show when relevant
		groups = append(groups,
			huh.NewGroup(t.fields[0]),
			huh.NewGroup(t.fields[1]).WithHideFunc(func() bool { return v.startFrom != relativeDate }),
			huh.NewGroup(t.fields[2]).WithHideFunc(func() bool { return v.startFrom != absoluteDate }),
		)
	}

	m.form = huh.NewForm(groups...).WithShowHelp(false)
	m.form.Init()
}

func filterTypeField(value *kadmin.FilterType, title string) *huh.Select[kadmin.FilterType] {
	return huh.NewSelect[kadmin.FilterType]().
		Value(value).
		Title(title).
		Options(
			huh.NewOption("Contains", kadmin.ContainsFilterType),
			huh.NewOption("Starts With", kadmin.StartsWithFilterType))
}

func validateAbsoluteDate(v string) error {
	if _, err := time.Parse(time.RFC3339, v); err != nil {
		return fmt.Errorf("invalid date time format")
	}
	return nil
}

// searchSubmittedMsg is returned by the modal when the user confirms the search.
type searchSubmittedMsg struct {
	details kadmin.ReadDetails
}

// searchClosedMsg is returned by the modal when the user cancels.
type searchClosedMsg struct{}

func (m *searchModal) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		m.err = ""
		switch key.String() {
		case "esc":
			return func() tea.Msg { return searchClosedMsg{} }
		case "ctrl+f":
			m.expanded = !m.expanded
			m.buildForm()
			return nil
		case "enter":
			details, err := m.readDetails()
			if err != nil {
				m.err = err.Error()
				return nil
			}
			return func() tea.Msg { return searchSubmittedMsg{details} }
		case "tab":
			// moving past the last field would complete the form
			if m.isLastField() {
				return nil
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return cmd
}

func (m *searchModal) isLastField() bool {
	last := m.tabs[len(m.tabs)-1].fields
	return m.form.GetFocusedField() == last[len(last)-1]
}

func (m *searchModal) activeTab() int {
	focused := m.form.GetFocusedField()
	for i, t := range m.tabs {
		for _, f := range t.fields {
			if f == focused {
				return i
			}
		}
	}
	return 0
}

func (m *searchModal) readDetails() (kadmin.ReadDetails, error) {
	v := m.values
	filter := &kadmin.Filter{Term: strings.TrimSpace(v.term)}
	if v.keyTerm != "" {
		filter.KeyFilter, filter.KeySearchTerm = v.keyFilter, v.keyTerm
	}
	if v.valueTerm != "" {
		filter.ValueFilter, filter.ValueSearchTerm = v.valueFilter, v.valueTerm
	}

	sp, err := m.startPoint()
	if err != nil {
		return kadmin.ReadDetails{}, err
	}

	partitions := v.partitions
	if len(partitions) == 0 {
		partitions = m.topic.Partitions()
	}

	return kadmin.ReadDetails{
		TopicName:       m.topic.Name,
		PartitionToRead: partitions,
		StartPoint:      sp,
		Limit:           v.limit,
		Filter:          filter,
	}, nil
}

func (m *searchModal) startPoint() (kadmin.StartPoint, error) {
	switch m.values.startFrom {
	case beginning:
		return kadmin.Beginning, nil
	case live:
		return kadmin.Live, nil
	case relativeDate:
		return m.values.relativeStartPoint, nil
	case absoluteDate:
		t, err := time.Parse(time.RFC3339, m.values.absoluteStartPoint)
		if err != nil {
			return 0, fmt.Errorf("invalid absolute start date, expected RFC3339")
		}
		return kadmin.StartPoint(t.UnixMilli()), nil
	default:
		return kadmin.MostRecent, nil
	}
}

func toFormStartPoint(sp kadmin.StartPoint) startPoint {
	switch sp {
	case kadmin.Beginning:
		return beginning
	case kadmin.MostRecent:
		return mostRecent
	case kadmin.Live:
		return live
	case kadmin.Today, kadmin.Yesterday, kadmin.Last7Days:
		return relativeDate
	default:
		return absoluteDate
	}
}

func toAbsoluteStartPoint(sp kadmin.StartPoint) string {
	if toFormStartPoint(sp) != absoluteDate {
		return ""
	}
	return time.UnixMilli(int64(sp)).Format(time.RFC3339)
}

// View renders the modal sized to fit within the given window width.
func (m *searchModal) View(windowWidth int) string {
	width := min(maxModalWidth, max(20, windowWidth-8))
	m.form.WithWidth(width)
	var header []string
	active := m.activeTab()
	for i, t := range m.tabs {
		name := t.name
		if m.isChanged(t.name) {
			name += changedMarker
		}
		if i == active {
			header = append(header, styles.Help.Key.Render("["+name+"]"))
		} else {
			header = append(header, styles.Help.Footer.Render(" "+name+" "))
		}
	}

	footer := "enter search · esc close · ctrl+f more options"
	if n := m.hiddenChanges(); n > 0 {
		footer += fmt.Sprintf(" (%d set)", n)
	}
	if m.expanded {
		footer = "enter search · tab next · esc close · ctrl+f fewer options"
	}

	parts := []string{
		styles.Help.Title.Render("Search " + m.topic.Name),
		"",
		lg.NewStyle().Width(width).Render(strings.Join(header, " ")),
		"",
		m.form.View(),
	}
	if m.err != "" {
		parts = append(parts, styles.Notifier.Error.UnsetMarginLeft().Render(m.err))
	}
	parts = append(parts, "", styles.Help.Footer.Width(width).Render(footer))
	return styles.Help.Modal.Render(lg.JoinVertical(lg.Left, parts...))
}

const changedMarker = " ●"

// advancedTabs are the tabs only shown when the modal is expanded.
var advancedTabs = []string{"Key", "Value", "Start", "Partitions", "Limit"}

// isChanged reports whether the tab differs from the default search.
func (m *searchModal) isChanged(tab string) bool {
	v := m.values
	def := kadmin.NewDefaultReadDetails(m.topic)
	switch tab {
	case "Key or Value":
		return strings.TrimSpace(v.term) != ""
	case "Key":
		return v.keyTerm != ""
	case "Value":
		return v.valueTerm != ""
	case "Start":
		return v.startFrom != toFormStartPoint(def.StartPoint)
	case "Partitions":
		return len(v.partitions) > 0
	case "Limit":
		return v.limit != def.Limit
	}
	return false
}

// hiddenChanges counts the advanced tabs that differ from the default search
// while the modal is collapsed.
func (m *searchModal) hiddenChanges() int {
	if m.expanded {
		return 0
	}
	n := 0
	for _, t := range advancedTabs {
		if m.isChanged(t) {
			n++
		}
	}
	return n
}

func (m *searchModal) Shortcuts() []statusbar.Shortcut {
	toggle := statusbar.Shortcut{Name: "More Options", Keybinding: "C-f"}
	if m.expanded {
		toggle.Name = "Fewer Options"
	}
	shortcuts := []statusbar.Shortcut{
		{Name: "Search", Keybinding: "enter"},
		{Name: "Close", Keybinding: "esc"},
		toggle,
	}
	if m.expanded {
		shortcuts = append(shortcuts, statusbar.Shortcut{Name: "Next/Prev Field", Keybinding: "tab/S-tab"})
	}
	return shortcuts
}
