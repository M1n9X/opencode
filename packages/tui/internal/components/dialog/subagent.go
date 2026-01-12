package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/list"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

// SubagentDialog interface for the subagent actions dialog
type SubagentDialog interface {
	layout.Modal
}

// OpenSubagentSessionMsg is sent when user wants to open a subagent's session
type OpenSubagentSessionMsg struct {
	SessionID string
}

// subagentAction represents an action in the subagent dialog
type subagentAction struct {
	title       string
	value       string
	description string
}

func (s subagentAction) Render(
	selected bool,
	width int,
	isFirstInViewport bool,
	baseStyle styles.Style,
) string {
	t := theme.CurrentTheme()

	titleStyle := baseStyle.Foreground(t.Text())
	descStyle := baseStyle.Foreground(t.TextMuted())

	if selected {
		titleStyle = baseStyle.Background(t.Primary()).Foreground(t.BackgroundElement())
		descStyle = baseStyle.Background(t.Primary()).Foreground(t.BackgroundElement())
	}

	titleText := titleStyle.Bold(true).Render(s.title)
	descText := ""
	if s.description != "" {
		descText = descStyle.Render(" " + s.description)
	}

	var itemStyle styles.Style
	if selected {
		itemStyle = baseStyle.
			Background(t.Primary()).
			Foreground(t.BackgroundElement()).
			Width(width).
			PaddingLeft(1)
	} else {
		itemStyle = baseStyle.PaddingLeft(1)
	}

	return itemStyle.Render(titleText + descText)
}

func (s subagentAction) Selectable() bool {
	return true
}

type subagentDialog struct {
	width     int
	height    int
	modal     *modal.Modal
	list      list.List[subagentAction]
	app       *app.App
	sessionID string
}

func (s *subagentDialog) Init() tea.Cmd {
	return nil
}

func (s *subagentDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.list.SetMaxWidth(layout.Current.Container.Width - 12)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, idx := s.list.GetSelectedItem(); idx >= 0 {
				switch item.value {
				case "subagent.view":
					return s, tea.Sequence(
						util.CmdHandler(modal.CloseModalMsg{}),
						util.CmdHandler(OpenSubagentSessionMsg{SessionID: s.sessionID}),
					)
				}
			}
		}
	}

	var cmd tea.Cmd
	listModel, cmd := s.list.Update(msg)
	s.list = listModel.(list.List[subagentAction])
	return s, cmd
}

func (s *subagentDialog) Render(background string) string {
	listView := s.list.View()

	t := theme.CurrentTheme()
	keyStyle := styles.NewStyle().
		Foreground(t.Text()).
		Background(t.BackgroundPanel()).
		Bold(true).
		Render
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted()).Background(t.BackgroundPanel()).Render

	helpText := keyStyle("enter") + mutedStyle(" select   ") + keyStyle("esc") + mutedStyle(" close")

	bgColor := t.BackgroundPanel()
	helpView := styles.NewStyle().
		Background(bgColor).
		Width(layout.Current.Container.Width - 14).
		PaddingLeft(1).
		PaddingTop(1).
		Render(helpText)

	content := strings.Join([]string{listView, helpView}, "\n")

	return s.modal.Render(content, background)
}

func (s *subagentDialog) Close() tea.Cmd {
	return nil
}

// NewSubagentDialog creates a new subagent actions dialog
func NewSubagentDialog(app *app.App, sessionID string) SubagentDialog {
	actions := []subagentAction{
		{
			title:       "Open",
			value:       "subagent.view",
			description: "the subagent's session",
		},
	}

	listComponent := list.NewListComponent(
		list.WithItems(actions),
		list.WithMaxVisibleHeight[subagentAction](5),
		list.WithFallbackMessage[subagentAction]("No actions available"),
		list.WithAlphaNumericKeys[subagentAction](true),
		list.WithRenderFunc(
			func(item subagentAction, selected bool, width int, baseStyle styles.Style) string {
				return item.Render(selected, width, false, baseStyle)
			},
		),
		list.WithSelectableFunc(func(item subagentAction) bool {
			return true
		}),
	)
	listComponent.SetMaxWidth(layout.Current.Container.Width - 12)

	return &subagentDialog{
		list:      listComponent,
		app:       app,
		sessionID: sessionID,
		modal: modal.New(
			modal.WithTitle("Subagent Actions"),
			modal.WithMaxWidth(layout.Current.Container.Width-8),
		),
	}
}
