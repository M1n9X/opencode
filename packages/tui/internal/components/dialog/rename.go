package dialog

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/components/toast"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

// RenameDialog interface for the session rename dialog
type RenameDialog interface {
	layout.Modal
}

// SessionRenamedMsg is sent when a session has been renamed
type SessionRenamedMsg struct {
	SessionID string
	NewTitle  string
}

type renameDialog struct {
	width  int
	height int
	modal  *modal.Modal
	app    *app.App
	input  textinput.Model
}

func (r *renameDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (r *renameDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
		r.input.SetWidth(layout.Current.Container.Width - 20)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			newTitle := strings.TrimSpace(r.input.Value())
			if newTitle == "" {
				return r, nil
			}
			return r, r.renameSession(newTitle)
		case "esc":
			return r, util.CmdHandler(modal.CloseModalMsg{})
		default:
			var cmd tea.Cmd
			r.input, cmd = r.input.Update(msg)
			return r, cmd
		}
	}

	var cmd tea.Cmd
	r.input, cmd = r.input.Update(msg)
	return r, cmd
}

func (r *renameDialog) renameSession(newTitle string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := r.app.UpdateSession(ctx, r.app.Session.ID, newTitle)
		if err != nil {
			return toast.NewErrorToast("Failed to rename session: " + err.Error())()
		}

		r.app.Session.Title = newTitle

		return tea.Batch(
			util.CmdHandler(modal.CloseModalMsg{}),
			toast.NewSuccessToast("Session renamed successfully"),
			util.CmdHandler(SessionRenamedMsg{
				SessionID: r.app.Session.ID,
				NewTitle:  newTitle,
			}),
		)()
	}
}

func (r *renameDialog) View() string {
	t := theme.CurrentTheme()

	inputView := r.input.View()

	mutedStyle := styles.NewStyle().
		Foreground(t.TextMuted()).
		Background(t.BackgroundPanel()).
		Render
	helpText := mutedStyle("Enter to confirm, Esc to cancel")
	helpText = styles.NewStyle().PaddingLeft(1).PaddingTop(1).Render(helpText)

	content := strings.Join([]string{inputView, helpText}, "\n")

	return content
}

func (r *renameDialog) Render(background string) string {
	return r.modal.Render(r.View(), background)
}

func (r *renameDialog) Close() tea.Cmd {
	return nil
}

// NewRenameDialog creates a new session rename dialog
func NewRenameDialog(app *app.App) RenameDialog {
	t := theme.CurrentTheme()
	bgColor := t.BackgroundPanel()
	textColor := t.Text()
	textMutedColor := t.TextMuted()

	ti := textinput.New()
	ti.Placeholder = "Enter new session title..."
	ti.Focus()
	ti.CharLimit = 100
	ti.SetWidth(layout.Current.Container.Width - 20)

	// Set current title as initial value
	if app.Session != nil {
		ti.SetValue(app.Session.Title)
	}

	ti.Styles.Blurred.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	ti.Styles.Blurred.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	ti.Styles.Focused.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	ti.Styles.Focused.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	ti.Styles.Focused.Prompt = styles.NewStyle().
		Background(bgColor).
		Lipgloss()

	return &renameDialog{
		app:   app,
		input: ti,
		modal: modal.New(
			modal.WithTitle("Rename Session"),
			modal.WithMaxWidth(layout.Current.Container.Width-8),
		),
	}
}
