package dialog

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/components/toast"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

type ExportDialog interface {
	layout.Modal
}

type ExportOptions struct {
	Filename                 string
	IncludeThinking          bool
	IncludeToolDetails       bool
	IncludeAssistantMetadata bool
	OpenWithoutSave          bool
}

type exportDialog struct {
	app           *app.App
	modal         *modal.Modal
	width         int
	height        int
	options       ExportOptions
	focusIndex    int
	filenameInput textinput.Model
}

func (e *exportDialog) Init() tea.Cmd {
	return e.filenameInput.Focus()
}

func (e *exportDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, searchKeys.Escape):
			return e, util.CmdHandler(modal.CloseModalMsg{})

		case msg.String() == "tab":
			e.focusIndex = (e.focusIndex + 1) % 5 // 4 options + 1 input
			if e.focusIndex == 0 {
				e.filenameInput.Focus()
			} else {
				e.filenameInput.Blur()
			}
			return e, nil

		case msg.String() == "shift+tab":
			e.focusIndex = (e.focusIndex - 1 + 5) % 5
			if e.focusIndex == 0 {
				e.filenameInput.Focus()
			} else {
				e.filenameInput.Blur()
			}
			return e, nil

		case msg.String() == "enter":
			e.options.Filename = e.filenameInput.Value()
			return e, e.export()

		case msg.String() == " ":
			if e.focusIndex != 0 {
				e.toggleOption()
				return e, nil
			}
		}
	}

	if e.focusIndex == 0 {
		var cmd tea.Cmd
		e.filenameInput, cmd = e.filenameInput.Update(msg)
		return e, cmd
	}

	return e, nil
}

func (e *exportDialog) toggleOption() {
	switch e.focusIndex {
	case 1:
		e.options.IncludeThinking = !e.options.IncludeThinking
	case 2:
		e.options.IncludeToolDetails = !e.options.IncludeToolDetails
	case 3:
		e.options.IncludeAssistantMetadata = !e.options.IncludeAssistantMetadata
	case 4:
		e.options.OpenWithoutSave = !e.options.OpenWithoutSave
	}
}

func (e *exportDialog) export() tea.Cmd {
	return func() tea.Msg {
		if e.app.Session == nil || e.app.Session.ID == "" {
			return toast.NewErrorToast("No session to export")()
		}

		// Generate the transcript
		transcript := formatTranscript(e.app.Session, e.app.Messages, TranscriptOptions{
			Thinking:          e.options.IncludeThinking,
			ToolDetails:       e.options.IncludeToolDetails,
			AssistantMetadata: e.options.IncludeAssistantMetadata,
		})

		if e.options.OpenWithoutSave {
			// Open in editor without saving
			return e.openInEditor(transcript)
		}

		// Save to file
		filename := strings.TrimSpace(e.options.Filename)
		if filename == "" {
			filename = fmt.Sprintf("session-%s.md", e.app.Session.ID[:8])
		}

		filepath := filepath.Join(util.CwdPath, filename)
		if err := os.WriteFile(filepath, []byte(transcript), 0644); err != nil {
			return toast.NewErrorToast("Failed to save: " + err.Error())()
		}

		// Try to open in editor
		e.openInEditor(transcript)

		return tea.Batch(
			util.CmdHandler(modal.CloseModalMsg{}),
			toast.NewSuccessToast("Session exported to "+filename),
		)()
	}
}

func (e *exportDialog) openInEditor(content string) tea.Msg {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		return nil // No editor configured
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "opencode-export-*.md")
	if err != nil {
		return nil
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return nil
	}

	// Open editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

func (e *exportDialog) View() string {
	t := theme.CurrentTheme()
	baseStyle := styles.NewStyle().Foreground(t.Text())

	var content strings.Builder

	// Filename input
	content.WriteString("Filename\n")
	content.WriteString(e.filenameInput.View())
	content.WriteString("\n\n")

	// Helper to render checkbox
	renderCheckbox := func(label string, checked bool, index int) {
		cursor := "  "
		if e.focusIndex == index {
			cursor = "> "
		}

		icon := "[ ]"
		if checked {
			icon = "[x]"
		}

		style := baseStyle
		if e.focusIndex == index {
			style = style.Foreground(t.Primary())
		}

		content.WriteString(style.Render(cursor + icon + " " + label + "\n"))
	}

	renderCheckbox("Include thinking process", e.options.IncludeThinking, 1)
	renderCheckbox("Include tool details", e.options.IncludeToolDetails, 2)
	renderCheckbox("Include assistant metadata", e.options.IncludeAssistantMetadata, 3)
	renderCheckbox("Open without save", e.options.OpenWithoutSave, 4)

	content.WriteString("\n")

	// Help text
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	if e.focusIndex == 0 {
		content.WriteString(mutedStyle.Render("Press enter to confirm, tab for options\n"))
	} else {
		content.WriteString(mutedStyle.Render("Press space to toggle, enter to confirm\n"))
	}

	return e.modal.Render(content.String(), "")
}

func (e *exportDialog) Render(background string) string {
	return e.modal.Render(e.View(), background)
}

func (e *exportDialog) Close() tea.Cmd {
	return nil
}

func (e *exportDialog) SetSize(width, height int) {
	e.width = width
	e.height = height
}

func NewExportDialog(app *app.App) ExportDialog {
	ti := textinput.New()

	// Generate default filename
	defaultFilename := "session-export.md"
	if app.Session != nil && app.Session.ID != "" {
		defaultFilename = fmt.Sprintf("session-%s.md", app.Session.ID[:8])
	}

	ti.Placeholder = defaultFilename
	ti.SetValue(defaultFilename)
	ti.Focus()

	e := &exportDialog{
		app:           app,
		filenameInput: ti,
		options: ExportOptions{
			IncludeThinking:    true,
			IncludeToolDetails: true,
		},
	}

	e.modal = modal.New(
		modal.WithTitle("Export Session"),
		modal.WithMaxWidth(50),
	)

	return e
}

// TranscriptOptions controls what to include in the transcript
type TranscriptOptions struct {
	Thinking          bool
	ToolDetails       bool
	AssistantMetadata bool
}

// formatTranscript generates a Markdown transcript of the session
func formatTranscript(session *opencode.Session, messages []app.Message, options TranscriptOptions) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s\n\n", session.Title))
	sb.WriteString(fmt.Sprintf("**Session ID:** %s\n", session.ID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n", formatTimestamp(session.Time.Created)))
	sb.WriteString(fmt.Sprintf("**Updated:** %s\n\n", formatTimestamp(session.Time.Updated)))
	sb.WriteString("---\n\n")

	// Messages
	for _, msg := range messages {
		sb.WriteString(formatMessage(msg, options))
		sb.WriteString("---\n\n")
	}

	return sb.String()
}

func formatTimestamp(ts float64) string {
	t := time.UnixMilli(int64(ts))
	return t.Format("2006-01-02 15:04:05")
}

func formatMessage(msg app.Message, options TranscriptOptions) string {
	var sb strings.Builder

	switch m := msg.Info.(type) {
	case opencode.UserMessage:
		sb.WriteString("## User\n\n")
		for _, part := range msg.Parts {
			sb.WriteString(formatPart(part, options))
		}

	case opencode.AssistantMessage:
		sb.WriteString(formatAssistantHeader(m, options.AssistantMetadata))
		for _, part := range msg.Parts {
			sb.WriteString(formatPart(part, options))
		}
	}

	return sb.String()
}

func formatAssistantHeader(msg opencode.AssistantMessage, includeMetadata bool) string {
	if !includeMetadata {
		return "## Assistant\n\n"
	}

	var duration string
	if msg.Time.Completed > 0 && msg.Time.Created > 0 {
		d := (msg.Time.Completed - msg.Time.Created) / 1000
		duration = fmt.Sprintf(" · %.1fs", d)
	}

	mode := strings.Title(msg.Mode)
	return fmt.Sprintf("## Assistant (%s · %s%s)\n\n", mode, msg.ModelID, duration)
}

func formatPart(part opencode.PartUnion, options TranscriptOptions) string {
	var sb strings.Builder

	switch p := part.(type) {
	case opencode.TextPart:
		if !p.Synthetic {
			sb.WriteString(p.Text + "\n\n")
		}

	case opencode.ReasoningPart:
		if options.Thinking {
			sb.WriteString("_Thinking:_\n\n")
			sb.WriteString(p.Text + "\n\n")
		}

	case opencode.ToolPart:
		sb.WriteString(fmt.Sprintf("```\nTool: %s\n", p.Tool))

		if options.ToolDetails && p.State.Input != nil {
			inputJSON, _ := json.MarshalIndent(p.State.Input, "", "  ")
			sb.WriteString(fmt.Sprintf("\n**Input:**\n```json\n%s\n```", string(inputJSON)))
		}

		if options.ToolDetails && p.State.Status == opencode.ToolPartStateStatusCompleted && p.State.Output != "" {
			sb.WriteString(fmt.Sprintf("\n**Output:**\n```\n%s\n```", p.State.Output))
		}

		if options.ToolDetails && p.State.Status == opencode.ToolPartStateStatusError && p.State.Error != "" {
			sb.WriteString(fmt.Sprintf("\n**Error:**\n```\n%s\n```", p.State.Error))
		}

		sb.WriteString("\n```\n\n")

	case opencode.FilePart:
		sb.WriteString(fmt.Sprintf("📎 File: %s\n\n", p.Filename))
	}

	return sb.String()
}
