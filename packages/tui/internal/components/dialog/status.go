package dialog

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

type StatusDialog interface {
	layout.Modal
}

type statusDialog struct {
	app    *app.App
	modal  *modal.Modal
	width  int
	height int
}

func (s *statusDialog) Init() tea.Cmd {
	return nil
}

func (s *statusDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			return s, util.CmdHandler(modal.CloseModalMsg{})
		}
	}
	return s, nil
}

func (s *statusDialog) View() string {
	t := theme.CurrentTheme()
	var content strings.Builder

	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	successStyle := styles.NewStyle().Foreground(t.Success())
	warningStyle := styles.NewStyle().Foreground(t.Warning())

	// Session info
	if s.app.Session != nil && s.app.Session.ID != "" {
		content.WriteString(textStyle.Render("Session") + "\n")
		content.WriteString(mutedStyle.Render("  ID: "+s.app.Session.ID) + "\n")
		content.WriteString(mutedStyle.Render("  Title: "+s.app.Session.Title) + "\n\n")
	}

	// Model info
	if s.app.Model != nil && s.app.Provider != nil {
		content.WriteString(textStyle.Render("Model") + "\n")
		modelName := s.app.Model.Name
		if modelName == "" {
			modelName = s.app.Model.ID
		}
		content.WriteString(mutedStyle.Render("  "+modelName) + "\n")
		content.WriteString(mutedStyle.Render("  Provider: "+s.app.Provider.Name) + "\n\n")
	}

	// Agent info
	content.WriteString(textStyle.Render("Agent") + "\n")
	agentName := s.app.Agent().Name
	if len(agentName) > 0 {
		agentName = strings.ToUpper(agentName[:1]) + agentName[1:]
	}
	content.WriteString(mutedStyle.Render("  "+agentName) + "\n")
	content.WriteString(mutedStyle.Render("  Mode: "+string(s.app.Agent().Mode)) + "\n\n")

	// Messages count
	content.WriteString(textStyle.Render("Messages") + "\n")
	content.WriteString(mutedStyle.Render(fmt.Sprintf("  Total: %d", len(s.app.Messages))) + "\n")

	// Count user vs assistant messages
	userCount := 0
	assistantCount := 0
	for _, msg := range s.app.Messages {
		switch msg.Info.(type) {
		case opencode.UserMessage:
			userCount++
		case opencode.AssistantMessage:
			assistantCount++
		}
	}
	content.WriteString(mutedStyle.Render(fmt.Sprintf("  User: %d", userCount)) + "\n")
	content.WriteString(mutedStyle.Render(fmt.Sprintf("  Assistant: %d", assistantCount)) + "\n\n")

	// Providers
	content.WriteString(textStyle.Render("Providers") + "\n")
	if len(s.app.Providers) == 0 {
		content.WriteString(mutedStyle.Render("  No providers configured") + "\n\n")
	} else {
		content.WriteString(mutedStyle.Render(fmt.Sprintf("  Total: %d", len(s.app.Providers))) + "\n")
		for _, provider := range s.app.Providers {
			modelCount := len(provider.Models)
			content.WriteString(mutedStyle.Render(fmt.Sprintf("  • %s (%d models)", provider.Name, modelCount)) + "\n")
		}
		content.WriteString("\n")
	}

	// Status indicators
	content.WriteString(textStyle.Render("Status") + "\n")
	if s.app.IsBusy() {
		content.WriteString(warningStyle.Render("  • Busy") + "\n")
	} else {
		content.WriteString(successStyle.Render("  • Ready") + "\n")
	}

	if s.app.IsCompacting() {
		content.WriteString(warningStyle.Render("  • Compacting session") + "\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(mutedStyle.Render("Press esc to close") + "\n")

	return s.modal.Render(content.String(), "")
}

func (s *statusDialog) Render(background string) string {
	return s.modal.Render(s.View(), background)
}

func (s *statusDialog) Close() tea.Cmd {
	return nil
}

func (s *statusDialog) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func NewStatusDialog(app *app.App) StatusDialog {
	return &statusDialog{
		app: app,
		modal: modal.New(
			modal.WithTitle("System Status"),
			modal.WithMaxWidth(60),
		),
	}
}
