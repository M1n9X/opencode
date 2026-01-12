package sidebar

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

const SidebarWidth = 42

// SidebarComponent interface for the sidebar
type SidebarComponent interface {
	tea.Model
	tea.ViewModel
	SetVisible(visible bool)
	IsVisible() bool
	Toggle()
}

type sidebarComponent struct {
	app     *app.App
	width   int
	height  int
	visible bool

	// Expanded state for sections
	mcpExpanded  bool
	lspExpanded  bool
	diffExpanded bool
	todoExpanded bool
}

func (s *sidebarComponent) Init() tea.Cmd {
	return nil
}

func (s *sidebarComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = SidebarWidth
		s.height = msg.Height
	}
	return s, nil
}

func (s *sidebarComponent) View() string {
	if !s.visible {
		return ""
	}

	t := theme.CurrentTheme()
	width := SidebarWidth - 4 // Account for padding

	var sections []string

	// Session title
	if s.app.Session != nil && s.app.Session.ID != "" {
		titleStyle := styles.NewStyle().
			Foreground(t.Text()).
			Bold(true).
			Width(width)
		sections = append(sections, titleStyle.Render(s.app.Session.Title))

		// Share URL if shared
		if s.app.Session.Share.URL != "" {
			shareStyle := styles.NewStyle().Foreground(t.TextMuted()).Width(width)
			sections = append(sections, shareStyle.Render(s.app.Session.Share.URL))
		}
	}

	// Context section
	contextSection := s.renderContextSection(width)
	if contextSection != "" {
		sections = append(sections, contextSection)
	}

	// MCP section
	mcpSection := s.renderMCPSection(width)
	if mcpSection != "" {
		sections = append(sections, mcpSection)
	}

	// LSP section
	lspSection := s.renderLSPSection(width)
	if lspSection != "" {
		sections = append(sections, lspSection)
	}

	// Todo section
	todoSection := s.renderTodoSection(width)
	if todoSection != "" {
		sections = append(sections, todoSection)
	}

	// Modified files section
	diffSection := s.renderDiffSection(width)
	if diffSection != "" {
		sections = append(sections, diffSection)
	}

	// Footer
	footer := s.renderFooter(width)

	// Combine all sections
	content := strings.Join(sections, "\n\n")

	// Main container style
	containerStyle := styles.NewStyle().
		Background(t.BackgroundPanel()).
		Width(SidebarWidth).
		Height(s.height).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2)

	// Place content at top and footer at bottom
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		strings.Repeat("\n", max(0, s.height-lipgloss.Height(content)-lipgloss.Height(footer)-4)),
		footer,
	)

	return containerStyle.Render(mainContent)
}

func (s *sidebarComponent) renderContextSection(_ int) string {
	t := theme.CurrentTheme()
	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())

	var totalTokens float64
	var totalCost float64

	for _, msg := range s.app.Messages {
		if assistant, ok := msg.Info.(opencode.AssistantMessage); ok {
			totalCost += assistant.Cost
			totalTokens += float64(assistant.Tokens.Input + assistant.Tokens.Output +
				assistant.Tokens.Reasoning + assistant.Tokens.Cache.Read + assistant.Tokens.Cache.Write)
		}
	}

	var lines []string
	lines = append(lines, textStyle.Bold(true).Render("Context"))

	tokenStr := fmt.Sprintf("%.0f tokens", totalTokens)
	lines = append(lines, mutedStyle.Render(tokenStr))

	if s.app.Model != nil && s.app.Model.Limit.Context > 0 {
		percentage := int(totalTokens / float64(s.app.Model.Limit.Context) * 100)
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("%d%% used", percentage)))
	}

	costStr := fmt.Sprintf("$%.4f spent", totalCost)
	lines = append(lines, mutedStyle.Render(costStr))

	return strings.Join(lines, "\n")
}

func (s *sidebarComponent) renderMCPSection(_ int) string {
	if len(s.app.MCPStatus) == 0 {
		return ""
	}

	t := theme.CurrentTheme()
	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	successStyle := styles.NewStyle().Foreground(t.Success())
	errorStyle := styles.NewStyle().Foreground(t.Error())
	warningStyle := styles.NewStyle().Foreground(t.Warning())

	var lines []string

	// Sort MCP names for consistent display
	var names []string
	for name := range s.app.MCPStatus {
		names = append(names, name)
	}
	sort.Strings(names)

	// Count connected and error
	connectedCount := 0
	errorCount := 0
	for _, status := range s.app.MCPStatus {
		if status.Status == "connected" {
			connectedCount++
		} else if status.Status == "failed" || status.Status == "needs_auth" {
			errorCount++
		}
	}

	// Header with collapse indicator
	header := textStyle.Bold(true).Render("MCP")
	if len(names) > 2 && !s.mcpExpanded {
		summary := fmt.Sprintf(" (%d active", connectedCount)
		if errorCount > 0 {
			summary += fmt.Sprintf(", %d error", errorCount)
		}
		summary += ")"
		header = "▶ " + header + mutedStyle.Render(summary)
	} else if len(names) > 2 {
		header = "▼ " + header
	}
	lines = append(lines, header)

	// Show items if expanded or few items
	if len(names) <= 2 || s.mcpExpanded {
		for _, name := range names {
			status := s.app.MCPStatus[name]
			var dot string
			var statusText string

			switch status.Status {
			case "connected":
				dot = successStyle.Render("•")
				statusText = "Connected"
			case "failed":
				dot = errorStyle.Render("•")
				statusText = status.Error
				if statusText == "" {
					statusText = "Failed"
				}
			case "disabled":
				dot = mutedStyle.Render("•")
				statusText = "Disabled"
			case "needs_auth":
				dot = warningStyle.Render("•")
				statusText = "Needs auth"
			default:
				dot = mutedStyle.Render("•")
				statusText = status.Status
			}

			line := fmt.Sprintf("%s %s %s", dot, name, mutedStyle.Render(statusText))
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

func (s *sidebarComponent) renderLSPSection(_ int) string {
	if len(s.app.LSPStatus) == 0 {
		return ""
	}

	t := theme.CurrentTheme()
	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	successStyle := styles.NewStyle().Foreground(t.Success())
	errorStyle := styles.NewStyle().Foreground(t.Error())

	var lines []string

	// Header
	header := textStyle.Bold(true).Render("LSP")
	if len(s.app.LSPStatus) > 2 && !s.lspExpanded {
		header = "▶ " + header
	} else if len(s.app.LSPStatus) > 2 {
		header = "▼ " + header
	}
	lines = append(lines, header)

	// Show items if expanded or few items
	if len(s.app.LSPStatus) <= 2 || s.lspExpanded {
		for _, lsp := range s.app.LSPStatus {
			var dot string
			if lsp.Status == "connected" {
				dot = successStyle.Render("•")
			} else {
				dot = errorStyle.Render("•")
			}

			line := fmt.Sprintf("%s %s %s", dot, lsp.ID, mutedStyle.Render(lsp.Root))
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

func (s *sidebarComponent) renderTodoSection(width int) string {
	if len(s.app.Todos) == 0 {
		return ""
	}

	// Filter to show only non-completed todos
	var activeTodos []app.Todo
	for _, todo := range s.app.Todos {
		if todo.Status != "completed" {
			activeTodos = append(activeTodos, todo)
		}
	}

	if len(activeTodos) == 0 {
		return ""
	}

	t := theme.CurrentTheme()
	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	successStyle := styles.NewStyle().Foreground(t.Success())
	warningStyle := styles.NewStyle().Foreground(t.Warning())

	var lines []string

	// Header
	header := textStyle.Bold(true).Render("Todo")
	if len(activeTodos) > 2 && !s.todoExpanded {
		header = "▶ " + header
	} else if len(activeTodos) > 2 {
		header = "▼ " + header
	}
	lines = append(lines, header)

	// Show items if expanded or few items
	if len(activeTodos) <= 2 || s.todoExpanded {
		for _, todo := range activeTodos {
			var icon string
			switch todo.Status {
			case "completed":
				icon = successStyle.Render("✓")
			case "in_progress":
				icon = warningStyle.Render("○")
			default:
				icon = mutedStyle.Render("○")
			}

			content := todo.Content
			if len(content) > width-4 {
				content = content[:width-7] + "..."
			}

			line := fmt.Sprintf("%s %s", icon, content)
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

func (s *sidebarComponent) renderDiffSection(width int) string {
	if len(s.app.SessionDiff) == 0 {
		return ""
	}

	t := theme.CurrentTheme()
	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	addedStyle := styles.NewStyle().Foreground(t.DiffAdded())
	removedStyle := styles.NewStyle().Foreground(t.DiffRemoved())

	var lines []string

	// Header
	header := textStyle.Bold(true).Render("Modified Files")
	if len(s.app.SessionDiff) > 2 && !s.diffExpanded {
		header = "▶ " + header
	} else if len(s.app.SessionDiff) > 2 {
		header = "▼ " + header
	}
	lines = append(lines, header)

	// Show items if expanded or few items
	if len(s.app.SessionDiff) <= 2 || s.diffExpanded {
		for _, diff := range s.app.SessionDiff {
			// Truncate file path
			file := diff.File
			parts := strings.Split(file, string(filepath.Separator))
			if len(parts) > 0 {
				last := parts[len(parts)-1]
				rest := strings.Join(parts[:len(parts)-1], string(filepath.Separator))
				maxRestLen := width - len(last) - 10
				if maxRestLen > 0 && len(rest) > maxRestLen {
					rest = "..." + rest[len(rest)-maxRestLen:]
				}
				if rest != "" {
					file = rest + "/" + last
				} else {
					file = last
				}
			}

			var stats string
			if diff.Additions > 0 {
				stats += addedStyle.Render(fmt.Sprintf("+%d", diff.Additions))
			}
			if diff.Deletions > 0 {
				if stats != "" {
					stats += " "
				}
				stats += removedStyle.Render(fmt.Sprintf("-%d", diff.Deletions))
			}

			line := fmt.Sprintf("%s %s", mutedStyle.Render(file), stats)
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

func (s *sidebarComponent) renderFooter(_ int) string {
	t := theme.CurrentTheme()
	textStyle := styles.NewStyle().Foreground(t.Text())
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted())
	successStyle := styles.NewStyle().Foreground(t.Success())

	var lines []string

	// Current directory
	cwd := util.CwdPath
	parts := strings.Split(cwd, string(filepath.Separator))
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		rest := strings.Join(parts[:len(parts)-1], string(filepath.Separator))
		if rest != "" {
			rest += "/"
		}
		dirLine := mutedStyle.Render(rest) + textStyle.Render(last)
		lines = append(lines, dirLine)
	}

	// Version
	version := s.app.Version
	if version == "" {
		version = "dev"
	}
	versionLine := successStyle.Render("•") + " " +
		textStyle.Bold(true).Render("Open") +
		textStyle.Bold(true).Render("Code") + " " +
		mutedStyle.Render(version)
	lines = append(lines, versionLine)

	return strings.Join(lines, "\n")
}

func (s *sidebarComponent) SetVisible(visible bool) {
	s.visible = visible
}

func (s *sidebarComponent) IsVisible() bool {
	return s.visible
}

func (s *sidebarComponent) Toggle() {
	s.visible = !s.visible
	if s.app.State != nil {
		s.app.State.SidebarVisible = &s.visible
		s.app.SaveState()
	}
}

// NewSidebarComponent creates a new sidebar component
func NewSidebarComponent(app *app.App) SidebarComponent {
	visible := false
	if app.State.SidebarVisible != nil {
		visible = *app.State.SidebarVisible
	}
	return &sidebarComponent{
		app:          app,
		width:        SidebarWidth,
		visible:      visible,
		mcpExpanded:  true,
		lspExpanded:  true,
		diffExpanded: true,
		todoExpanded: true,
	}
}
