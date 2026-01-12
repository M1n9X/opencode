package dialog

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/muesli/reflow/truncate"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/list"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/components/toast"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

// ForkDialog interface for the fork from timeline dialog
type ForkDialog interface {
	layout.Modal
}

// ForkFromMessageMsg is sent when a session should be forked from a specific message
type ForkFromMessageMsg struct {
	SessionID string
	MessageID string
}

// forkItem represents a user message in the fork list
type forkItem struct {
	messageID string
	content   string
	timestamp time.Time
	index     int
}

func (f forkItem) Render(
	selected bool,
	width int,
	isFirstInViewport bool,
	baseStyle styles.Style,
) string {
	t := theme.CurrentTheme()
	infoStyle := baseStyle.Background(t.BackgroundPanel()).Foreground(t.Info()).Render
	textStyle := baseStyle.Background(t.BackgroundPanel()).Foreground(t.Text()).Render

	// Format timestamp
	var timeStr string
	if selected {
		timeStr = f.timestamp.Format("15:04") + " "
	} else {
		timeStr = infoStyle(f.timestamp.Format("15:04") + " ")
	}
	timeVisualLen := len(f.timestamp.Format("15:04") + " ")

	// Calculate available space for content
	reservedSpace := timeVisualLen + 4
	contentWidth := max(width-reservedSpace, 8)

	truncatedContent := truncate.StringWithTail(
		strings.Split(f.content, "\n")[0],
		uint(contentWidth),
		"...",
	)

	var styledContent string
	if selected {
		styledContent = truncatedContent
	} else {
		styledContent = textStyle(truncatedContent)
	}

	text := timeStr + styledContent

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

	return itemStyle.Render(text)
}

func (f forkItem) Selectable() bool {
	return true
}

type forkDialog struct {
	width  int
	height int
	modal  *modal.Modal
	list   list.List[forkItem]
	app    *app.App
}

func (f *forkDialog) Init() tea.Cmd {
	return nil
}

func (f *forkDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
		f.list.SetMaxWidth(layout.Current.Container.Width - 12)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "down", "k", "j":
			var cmd tea.Cmd
			listModel, cmd := f.list.Update(msg)
			f.list = listModel.(list.List[forkItem])

			// Get the newly selected item and scroll to it
			if item, idx := f.list.GetSelectedItem(); idx >= 0 {
				return f, tea.Sequence(
					cmd,
					util.CmdHandler(ScrollToMessageMsg{MessageID: item.messageID}),
				)
			}
			return f, cmd
		case "enter":
			if item, idx := f.list.GetSelectedItem(); idx >= 0 {
				return f, tea.Sequence(
					f.forkSession(item.messageID),
					util.CmdHandler(modal.CloseModalMsg{}),
				)
			}
		}
	}

	var cmd tea.Cmd
	listModel, cmd := f.list.Update(msg)
	f.list = listModel.(list.List[forkItem])
	return f, cmd
}

func (f *forkDialog) forkSession(messageID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		// Fork is implemented by reverting to the target message
		// The revert API will restore the session to the state at that message
		revertedSession, err := f.app.Client.Session.Revert(ctx, f.app.Session.ID, opencode.SessionRevertParams{
			MessageID: opencode.F(messageID),
		})
		if err != nil {
			return toast.NewErrorToast("Failed to fork session: " + err.Error())()
		}

		return app.SessionSelectedMsg(revertedSession)
	}
}

func (f *forkDialog) Render(background string) string {
	listView := f.list.View()

	t := theme.CurrentTheme()
	keyStyle := styles.NewStyle().
		Foreground(t.Text()).
		Background(t.BackgroundPanel()).
		Bold(true).
		Render
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted()).Background(t.BackgroundPanel()).Render

	helpText := keyStyle("↑/↓") + mutedStyle(" select   ") + keyStyle("enter") + mutedStyle(" fork")

	bgColor := t.BackgroundPanel()
	helpView := styles.NewStyle().
		Background(bgColor).
		Width(layout.Current.Container.Width - 14).
		PaddingLeft(1).
		PaddingTop(1).
		Render(helpText)

	content := strings.Join([]string{listView, helpView}, "\n")

	return f.modal.Render(content, background)
}

func (f *forkDialog) Close() tea.Cmd {
	return nil
}

// extractForkMessagePreview extracts a preview from message parts
func extractForkMessagePreview(parts []opencode.PartUnion) string {
	for _, part := range parts {
		switch casted := part.(type) {
		case opencode.TextPart:
			text := strings.TrimSpace(casted.Text)
			if text != "" && !casted.Synthetic {
				return text
			}
		}
	}
	return "No text content"
}

// NewForkDialog creates a new fork from timeline dialog
func NewForkDialog(app *app.App) ForkDialog {
	var items []forkItem

	// Filter to only user messages and extract relevant info (in reverse order - newest first)
	for i := len(app.Messages) - 1; i >= 0; i-- {
		message := app.Messages[i]
		if userMsg, ok := message.Info.(opencode.UserMessage); ok {
			preview := extractForkMessagePreview(message.Parts)

			items = append(items, forkItem{
				messageID: userMsg.ID,
				content:   preview,
				timestamp: time.UnixMilli(int64(userMsg.Time.Created)),
				index:     i,
			})
		}
	}

	listComponent := list.NewListComponent(
		list.WithItems(items),
		list.WithMaxVisibleHeight[forkItem](12),
		list.WithFallbackMessage[forkItem]("No user messages in this session"),
		list.WithAlphaNumericKeys[forkItem](true),
		list.WithRenderFunc(
			func(item forkItem, selected bool, width int, baseStyle styles.Style) string {
				return item.Render(selected, width, false, baseStyle)
			},
		),
		list.WithSelectableFunc(func(item forkItem) bool {
			return true
		}),
	)
	listComponent.SetMaxWidth(layout.Current.Container.Width - 12)

	return &forkDialog{
		list: listComponent,
		app:  app,
		modal: modal.New(
			modal.WithTitle("Fork from Message"),
			modal.WithMaxWidth(layout.Current.Container.Width-8),
		),
	}
}
