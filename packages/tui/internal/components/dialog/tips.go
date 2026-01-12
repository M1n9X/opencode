package dialog

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
)

// TIPS contains all the tips from OpenTUI (100+ tips with {highlight} markup)
var tips = []string{
	"Type {highlight}@{/highlight} followed by a filename to fuzzy search and attach files to your prompt.",
	"Start a message with {highlight}!{/highlight} to run shell commands directly (e.g., {highlight}!ls -la{/highlight}).",
	"Press {highlight}Tab{/highlight} to cycle between Build (full access) and Plan (read-only) agents.",
	"Use {highlight}/undo{/highlight} to revert the last message and any file changes made by OpenCode.",
	"Use {highlight}/redo{/highlight} to restore previously undone messages and file changes.",
	"Run {highlight}/share{/highlight} to create a public link to your conversation at opencode.ai.",
	"Drag and drop images into the terminal to add them as context for your prompts.",
	"Press {highlight}Ctrl+V{/highlight} to paste images from your clipboard directly into the prompt.",
	"Press {highlight}Ctrl+X E{/highlight} or {highlight}/editor{/highlight} to compose messages in your external editor.",
	"Run {highlight}/init{/highlight} to auto-generate project rules based on your codebase structure.",
	"Run {highlight}/models{/highlight} or {highlight}Ctrl+X M{/highlight} to see and switch between available AI models.",
	"Use {highlight}/theme{/highlight} or {highlight}Ctrl+X T{/highlight} to preview and switch between 50+ built-in themes.",
	"Press {highlight}Ctrl+X N{/highlight} or {highlight}/new{/highlight} to start a fresh conversation session.",
	"Use {highlight}/sessions{/highlight} or {highlight}Ctrl+X L{/highlight} to list and continue previous conversations.",
	"Run {highlight}/compact{/highlight} to summarize long sessions when approaching context limits.",
	"Press {highlight}Ctrl+X X{/highlight} or {highlight}/export{/highlight} to save the conversation as Markdown.",
	"Press {highlight}Ctrl+X Y{/highlight} to copy the assistant's last message to clipboard.",
	"Press {highlight}Ctrl+P{/highlight} to see all available actions and commands.",
	"Run {highlight}/connect{/highlight} to add API keys for 75+ supported LLM providers.",
	"The default leader key is {highlight}Ctrl+X{/highlight}; combine with other keys for quick actions.",
	"Press {highlight}F2{/highlight} to quickly switch between recently used models.",
	"Press {highlight}Ctrl+X B{/highlight} to show/hide the sidebar panel.",
	"Use {highlight}PageUp{/highlight}/{highlight}PageDown{/highlight} to navigate through conversation history.",
	"Press {highlight}Ctrl+G{/highlight} or {highlight}Home{/highlight} to jump to the beginning of the conversation.",
	"Press {highlight}Ctrl+Alt+G{/highlight} or {highlight}End{/highlight} to jump to the most recent message.",
	"Press {highlight}Shift+Enter{/highlight} or {highlight}Ctrl+J{/highlight} to add newlines in your prompt.",
	"Press {highlight}Ctrl+C{/highlight} when typing to clear the input field.",
	"Press {highlight}Escape{/highlight} to stop the AI mid-response.",
	"Switch to {highlight}Plan{/highlight} agent to get suggestions without making actual changes.",
	"Use {highlight}@<agent-name>{/highlight} in prompts to invoke specialized subagents.",
	"Press {highlight}Ctrl+X Right/Left{/highlight} to cycle through parent and child sessions.",
	"Create {highlight}opencode.json{/highlight} in project root for project-specific settings.",
	"Place settings in {highlight}~/.config/opencode/opencode.json{/highlight} for global config.",
	"Add {highlight}$schema{/highlight} to your config for autocomplete in your editor.",
	"Configure {highlight}model{/highlight} in config to set your default model.",
	"Override any keybind in config via the {highlight}keybinds{/highlight} section.",
	"Set any keybind to {highlight}none{/highlight} to disable it completely.",
	"Configure local or remote MCP servers in the {highlight}mcp{/highlight} config section.",
	"OpenCode auto-handles OAuth for remote MCP servers requiring auth.",
	"Add {highlight}.md{/highlight} files to {highlight}.opencode/command/{/highlight} to define reusable custom prompts.",
	"Use {highlight}$ARGUMENTS{/highlight}, {highlight}$1{/highlight}, {highlight}$2{/highlight} in custom commands for dynamic input.",
	"Use backticks in commands to inject shell output (e.g., {highlight}`git status`{/highlight}).",
	"Add {highlight}.md{/highlight} files to {highlight}.opencode/agent/{/highlight} for specialized AI personas.",
	"Configure per-agent permissions for {highlight}edit{/highlight}, {highlight}bash{/highlight}, and {highlight}webfetch{/highlight} tools.",
	"OpenCode auto-formats files using prettier, gofmt, ruff, and more.",
	"Define custom formatter commands with file extensions in config.",
	"OpenCode uses LSP servers for intelligent code analysis.",
	"Create {highlight}.ts{/highlight} files in {highlight}.opencode/tool/{/highlight} to define new LLM tools.",
	"Tool definitions can invoke scripts written in Python, Go, etc.",
	"Add {highlight}.ts{/highlight} files to {highlight}.opencode/plugin/{/highlight} for event hooks.",
	"Use plugins to send OS notifications when sessions complete.",
	"Create a plugin to prevent OpenCode from reading sensitive files.",
	"Use {highlight}opencode run{/highlight} for non-interactive scripting.",
	"Use {highlight}opencode run --continue{/highlight} to resume the last session.",
	"Use {highlight}opencode run -f file.ts{/highlight} to attach files via CLI.",
	"Use {highlight}--format json{/highlight} for machine-readable output in scripts.",
	"Run {highlight}opencode serve{/highlight} for headless API access to OpenCode.",
	"Use {highlight}opencode run --attach{/highlight} to connect to a running server for faster runs.",
	"Run {highlight}opencode upgrade{/highlight} to update to the latest version.",
	"Run {highlight}opencode auth list{/highlight} to see all configured providers.",
	"Run {highlight}opencode agent create{/highlight} for guided agent creation.",
	"Use {highlight}/opencode{/highlight} in GitHub issues/PRs to trigger AI actions.",
	"Run {highlight}opencode github install{/highlight} to set up the GitHub workflow.",
	"Comment {highlight}/opencode fix this{/highlight} on issues to auto-create PRs.",
	"Comment {highlight}/oc{/highlight} on PR code lines for targeted code reviews.",
	"Create JSON theme files in {highlight}.opencode/themes/{/highlight} directory.",
	"Themes support dark/light variants for both modes.",
	"Reference ANSI colors 0-255 in custom themes.",
	"Use {highlight}instructions{/highlight} in config to load additional rules files.",
	"Set agent {highlight}temperature{/highlight} from 0.0 (focused) to 1.0 (creative).",
	"Configure {highlight}maxSteps{/highlight} to limit agentic iterations per request.",
	"Override global tool settings per agent configuration.",
	"Run {highlight}/unshare{/highlight} to remove a session from public access.",
	"Run {highlight}opencode debug config{/highlight} to troubleshoot configuration.",
	"Use {highlight}--print-logs{/highlight} flag to see detailed logs in stderr.",
	"Press {highlight}Ctrl+X G{/highlight} or {highlight}/timeline{/highlight} to jump to specific messages.",
	"Press {highlight}Ctrl+X H{/highlight} to toggle code block visibility in messages.",
	"Press {highlight}Ctrl+X S{/highlight} or {highlight}/status{/highlight} to see system status info.",
	"Enable {highlight}tui.scroll_acceleration{/highlight} for smooth macOS-style scrolling.",
	"Toggle username display in chat via command palette ({highlight}Ctrl+P{/highlight}).",
	"Commit your project's {highlight}AGENTS.md{/highlight} file to Git for team sharing.",
	"Use {highlight}/review{/highlight} to review uncommitted changes, branches, or PRs.",
	"Run {highlight}/help{/highlight} or {highlight}Ctrl+X H{/highlight} to show the help dialog.",
	"Use {highlight}/details{/highlight} to toggle tool execution details visibility.",
	"Use {highlight}/rename{/highlight} to rename the current session.",
	"Press {highlight}Ctrl+Z{/highlight} to suspend the terminal and return to your shell.",
	"Press {highlight}Ctrl+F{/highlight} in model dialog to toggle favorites.",
}

// TipPart represents a part of a tip with optional highlighting
type TipPart struct {
	Text      string
	Highlight bool
}

// parseTip parses a tip string and extracts highlighted parts
func parseTip(tip string) []TipPart {
	var parts []TipPart
	re := regexp.MustCompile(`\{highlight\}(.*?)\{/highlight\}`)
	lastIndex := 0

	matches := re.FindAllStringSubmatchIndex(tip, -1)
	for _, match := range matches {
		// Add text before the match
		if match[0] > lastIndex {
			parts = append(parts, TipPart{Text: tip[lastIndex:match[0]], Highlight: false})
		}
		// Add the highlighted text (group 1)
		parts = append(parts, TipPart{Text: tip[match[2]:match[3]], Highlight: true})
		lastIndex = match[1]
	}

	// Add remaining text after last match
	if lastIndex < len(tip) {
		parts = append(parts, TipPart{Text: tip[lastIndex:], Highlight: false})
	}

	return parts
}

type DidYouKnow struct {
	app      *app.App
	visible  bool
	tipIndex int
	width    int
}

type ToggleTipsMsg struct{}

func NewDidYouKnow(app *app.App) *DidYouKnow {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &DidYouKnow{
		app:      app,
		visible:  true,
		tipIndex: rng.Intn(len(tips)),
		width:    42,
	}
}

func (d *DidYouKnow) Init() tea.Cmd {
	return nil
}

func (d *DidYouKnow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ToggleTipsMsg:
		d.visible = !d.visible
		return d, nil
	case tea.KeyMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("?"))) {
			// Could toggle tips visibility here
		}
	}
	return d, nil
}

// RandomizeTip selects a new random tip
func (d *DidYouKnow) RandomizeTip() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	d.tipIndex = rng.Intn(len(tips))
}

func (d *DidYouKnow) View() string {
	if !d.visible {
		return ""
	}
	t := theme.CurrentTheme()

	// Parse the current tip
	tipParts := parseTip(tips[d.tipIndex])

	// Build the tip text with highlighting
	var tipText strings.Builder
	for _, part := range tipParts {
		if part.Highlight {
			tipText.WriteString(styles.NewStyle().Foreground(t.Text()).Bold(true).Render(part.Text))
		} else {
			tipText.WriteString(styles.NewStyle().Foreground(t.TextMuted()).Render(part.Text))
		}
	}

	// Build the box
	title := " 🅘 Did you know? "
	boxWidth := d.width
	dashes := boxWidth - 2 - len(title) - 1
	if dashes < 0 {
		dashes = 0
	}

	borderStyle := styles.NewStyle().Foreground(t.Border())
	titleStyle := styles.NewStyle().Foreground(t.Text())

	// Top border with title
	topBorder := borderStyle.Render("╭─") + titleStyle.Render(title) + borderStyle.Render(strings.Repeat("─", dashes)+"╮")

	// Content with side borders
	contentWidth := boxWidth - 4 // Account for borders and padding
	wrappedTip := wrapText(tipText.String(), contentWidth)

	var contentLines []string
	for _, line := range strings.Split(wrappedTip, "\n") {
		paddedLine := line + strings.Repeat(" ", contentWidth-lipgloss.Width(line))
		contentLines = append(contentLines, borderStyle.Render("│ ")+paddedLine+borderStyle.Render(" │"))
	}

	// Bottom border
	bottomBorder := borderStyle.Render("╰" + strings.Repeat("─", boxWidth-2) + "╯")

	// Combine all parts
	result := topBorder + "\n"
	result += borderStyle.Render("│") + strings.Repeat(" ", boxWidth-2) + borderStyle.Render("│") + "\n"
	result += strings.Join(contentLines, "\n") + "\n"
	result += borderStyle.Render("│") + strings.Repeat(" ", boxWidth-2) + borderStyle.Render("│") + "\n"
	result += bottomBorder

	return result
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	var currentLine strings.Builder
	currentWidth := 0

	words := strings.Fields(text)
	for i, word := range words {
		wordWidth := lipgloss.Width(word)

		if currentWidth+wordWidth+1 > width && currentWidth > 0 {
			result.WriteString(currentLine.String())
			result.WriteString("\n")
			currentLine.Reset()
			currentWidth = 0
		}

		if currentWidth > 0 {
			currentLine.WriteString(" ")
			currentWidth++
		}
		currentLine.WriteString(word)
		currentWidth += wordWidth

		if i == len(words)-1 {
			result.WriteString(currentLine.String())
		}
	}

	return result.String()
}
