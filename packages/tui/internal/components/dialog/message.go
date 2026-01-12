package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/list"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

// MessageDialog interface for message actions dialog
type MessageDialog interface {
	layout.Modal
}

// MessageRevertMsg is sent when user wants to revert to a message
type MessageRevertMsg struct {
	MessageID string
	SessionID string
	Prompt    string
}

// MessageCopyMsg is sent when user wants to copy message content
type MessageCopyMsg struct {
	Content string
}

// messageActionItem represents an action in the message dialog
type messageActionItem struct {
	title       string
	value       string
	description string
}

func (m messageActionItem) Render(
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

	// Render title and description
	titleText := titleStyle.Bold(true).Render(m.title)
	descText := descStyle.Render(" - " + m.description)

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

func (m messageActionItem) Selectable() bool {
	return true
}

type messageDialog struct {
	width     int
	height    int
	modal     *modal.Modal
	list      list.List[messageActionItem]
	app       *app.App
	messageID string
	sessionID string
	parts     []opencode.PartUnion
}

func (m *messageDialog) Init() tea.Cmd {
	return nil
}

func (m *messageDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetMaxWidth(layout.Current.Container.Width - 12)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, idx := m.list.GetSelectedItem(); idx >= 0 {
				return m.handleAction(item.value)
			}
		}
	}

	var cmd tea.Cmd
	listModel, cmd := m.list.Update(msg)
	m.list = listModel.(list.List[messageActionItem])
	return m, cmd
}

func (m *messageDialog) handleAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "session.revert":
		// Extract text content from parts and send message to trigger revert
		prompt := m.extractTextContent()
		return m, tea.Sequence(
			util.CmdHandler(MessageRevertMsg{
				MessageID: m.messageID,
				SessionID: m.sessionID,
				Prompt:    prompt,
			}),
			util.CmdHandler(modal.CloseModalMsg{}),
		)

	case "message.copy":
		content := m.extractTextContent()
		return m, tea.Sequence(
			util.CmdHandler(MessageCopyMsg{Content: content}),
			util.CmdHandler(modal.CloseModalMsg{}),
		)
	}

	return m, nil
}

func (m *messageDialog) extractTextContent() string {
	var content strings.Builder
	for _, part := range m.parts {
		switch casted := part.(type) {
		case opencode.TextPart:
			if !casted.Synthetic {
				content.WriteString(casted.Text)
			}
		}
	}
	return content.String()
}

func (m *messageDialog) Render(background string) string {
	listView := m.list.View()

	t := theme.CurrentTheme()
	keyStyle := styles.NewStyle().
		Foreground(t.Text()).
		Background(t.BackgroundPanel()).
		Bold(true).
		Render
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted()).Background(t.BackgroundPanel()).Render

	helpText := keyStyle("↑/↓") + mutedStyle(" select   ") + keyStyle("enter") + mutedStyle(" confirm")

	bgColor := t.BackgroundPanel()
	helpView := styles.NewStyle().
		Background(bgColor).
		Width(layout.Current.Container.Width - 14).
		PaddingLeft(1).
		PaddingTop(1).
		Render(helpText)

	content := strings.Join([]string{listView, helpView}, "\n")

	return m.modal.Render(content, background)
}

func (m *messageDialog) Close() tea.Cmd {
	return nil
}

// NewMessageDialog creates a new message actions dialog
func NewMessageDialog(app *app.App, messageID, sessionID string, parts []opencode.PartUnion) MessageDialog {
	items := []messageActionItem{
		{
			title:       "Revert",
			value:       "session.revert",
			description: "undo messages and file changes",
		},
		{
			title:       "Copy",
			value:       "message.copy",
			description: "message text to clipboard",
		},
	}

	listComponent := list.NewListComponent(
		list.WithItems(items),
		list.WithMaxVisibleHeight[messageActionItem](5),
		list.WithFallbackMessage[messageActionItem]("No actions available"),
		list.WithAlphaNumericKeys[messageActionItem](true),
		list.WithRenderFunc(
			func(item messageActionItem, selected bool, width int, baseStyle styles.Style) string {
				return item.Render(selected, width, false, baseStyle)
			},
		),
		list.WithSelectableFunc(func(item messageActionItem) bool {
			return true
		}),
	)
	listComponent.SetMaxWidth(layout.Current.Container.Width - 12)

	return &messageDialog{
		list:      listComponent,
		app:       app,
		messageID: messageID,
		sessionID: sessionID,
		parts:     parts,
		modal: modal.New(
			modal.WithTitle("Message Actions"),
			modal.WithMaxWidth(layout.Current.Container.Width-8),
		),
	}
}
