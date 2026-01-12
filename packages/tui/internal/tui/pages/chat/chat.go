package chat

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"

	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/completions"
	chatcomp "github.com/sst/opencode/internal/components/chat"
	cmdcomp "github.com/sst/opencode/internal/components/commands"
	"github.com/sst/opencode/internal/components/dialog"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

type Page struct {
	width, height int
	app           *app.App

	// Components
	Editor      chatcomp.EditorComponent
	Messages    chatcomp.MessagesComponent
	Completions dialog.CompletionDialog

	// Providers
	CommandProvider completions.CompletionProvider
	FileProvider    completions.CompletionProvider
	SymbolsProvider completions.CompletionProvider
	AgentsProvider  completions.CompletionProvider

	// State
	ShowCompletionDialog bool
}

func New(app *app.App) *Page {
	p := &Page{
		app: app,
		// Initialize components with verified constructors
		Editor:      chatcomp.NewEditorComponent(app),
		Messages:    chatcomp.NewMessagesComponent(app),
		Completions: dialog.NewCompletionDialogComponent("", nil),
	}

	// Initialize providers
	p.CommandProvider = completions.NewCommandCompletionProvider(app)
	p.FileProvider = completions.NewFileContextGroup(app)
	p.SymbolsProvider = completions.NewSymbolsContextGroup(app)
	p.AgentsProvider = completions.NewAgentsContextGroup(app)

	return p
}

func (p *Page) Init() tea.Cmd {
	return tea.Batch(
		p.Editor.Init(),
		p.Messages.Init(),
		p.Completions.Init(),
	)
}

func (p *Page) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.SetSize(msg.Width, msg.Height)
		// Propagate resize to components
		// Note: We modify msg with -2 height logic similar to tui.go if needed,
		// but tui.go likely already did that before passing to Page?
		// We'll see how tui.go integrates it. For now, assume raw WindowSizeMsg.

		// In tui.go: msg.Height -= 2 // Make space for the status bar
		// We should probably replicate that logic or assume caller handles it.
		// Let's replicate for safety if caller passes raw window size.
		// But Page.SetSize handles internal state.

		// Let's stick to forwarding the msg to components.
		updatedEditor, cmd := p.Editor.Update(msg)
		p.Editor = updatedEditor.(chatcomp.EditorComponent)
		cmds = append(cmds, cmd)

		updatedMessages, cmd := p.Messages.Update(msg)
		p.Messages = updatedMessages.(chatcomp.MessagesComponent)
		cmds = append(cmds, cmd)

		if p.ShowCompletionDialog {
			updatedCompletions, cmd := p.Completions.Update(msg)
			p.Completions = updatedCompletions.(dialog.CompletionDialog)
			cmds = append(cmds, cmd)
		}

		return p, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		keyString := msg.String()

		// Handle completions trigger
		if keyString == "/" &&
			!p.ShowCompletionDialog &&
			p.Editor.Value() == "" &&
			!p.app.IsBashMode {
			p.ShowCompletionDialog = true

			updated, cmd := p.Editor.Update(msg)
			p.Editor = updated.(chatcomp.EditorComponent)
			cmds = append(cmds, cmd)

			p.Completions = dialog.NewCompletionDialogComponent("/", p.CommandProvider)
			updated, cmd = p.Completions.Update(msg)
			p.Completions = updated.(dialog.CompletionDialog)
			cmds = append(cmds, cmd)

			return p, tea.Sequence(cmds...)
		}

		if keyString == "@" &&
			!p.ShowCompletionDialog &&
			!p.app.IsBashMode {
			p.ShowCompletionDialog = true

			updated, cmd := p.Editor.Update(msg)
			p.Editor = updated.(chatcomp.EditorComponent)
			cmds = append(cmds, cmd)

			p.Completions = dialog.NewCompletionDialogComponent("@", p.AgentsProvider, p.FileProvider, p.SymbolsProvider)
			updated, cmd = p.Completions.Update(msg)
			p.Completions = updated.(dialog.CompletionDialog)
			cmds = append(cmds, cmd)

			return p, tea.Sequence(cmds...)
		}

		if p.ShowCompletionDialog {
			switch keyString {
			case "tab", "enter", "esc", "ctrl+c", "up", "down", "ctrl+p", "ctrl+n":
				updated, cmd := p.Completions.Update(msg)
				p.Completions = updated.(dialog.CompletionDialog)
				cmds = append(cmds, cmd)
				return p, tea.Batch(cmds...)
			}

			updated, cmd := p.Editor.Update(msg)
			p.Editor = updated.(chatcomp.EditorComponent)
			cmds = append(cmds, cmd)

			updated, cmd = p.Completions.Update(msg)
			p.Completions = updated.(dialog.CompletionDialog)
			cmds = append(cmds, cmd)

			return p, tea.Batch(cmds...)
		}

		// Fallback to components
		updatedEditor, cmd := p.Editor.Update(msg)
		p.Editor = updatedEditor.(chatcomp.EditorComponent)
		cmds = append(cmds, cmd)

		// Forward certain keys to messages?
		// tui.go didn't explicit forward keys to messages unless scroll/select.
		// Messages.Update usually handles scrolling.
	}

	// Always update messages for events (like streaming)
	// But not KeyPressMsg generally unless it's for scrolling?
	// tui.go calls a.messages.Update(msg) in "default" fallthrough or "tea.MouseWheelMsg".
	// But NOT for every KeyPressMsg unless it fell through.
	// In strict logic, we should be careful.
	// But Editor handles most keys.

	// If msg is NOT KeyPressMsg, forward to Messages (e.g. streaming events)
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		updatedMessages, cmd := p.Messages.Update(msg)
		p.Messages = updatedMessages.(chatcomp.MessagesComponent)
		cmds = append(cmds, cmd)
	}

	// Handle completion close
	if _, ok := msg.(dialog.CompletionDialogCloseMsg); ok {
		p.ShowCompletionDialog = false
	}

	return p, tea.Batch(cmds...)
}

func (p *Page) View() string {
	t := theme.CurrentTheme()

	var mainLayout string
	if p.app.Session.ID == "" {
		mainLayout = p.viewHome()
	} else {
		mainLayout = p.viewChat()
	}

	mainLayout = styles.NewStyle().
		Background(t.Background()).
		Padding(0, 2).
		Render(mainLayout)
	mainLayout = lipgloss.PlaceHorizontal(
		p.width,
		lipgloss.Center,
		mainLayout,
		styles.WhitespaceStyle(t.Background()),
	)

	mainStyle := styles.NewStyle().Background(t.Background())
	mainLayout = mainStyle.Render(mainLayout)

	if theme.CurrentThemeUsesAnsiColors() {
		mainLayout = util.ConvertRGBToAnsi16Colors(mainLayout)
	}

	return mainLayout
}

func (p *Page) Cursor() *tea.Cursor {
	var cursor *tea.Cursor
	// Get base cursor from editor
	cursor = p.Editor.Cursor()
	if cursor == nil {
		return nil
	}

	// Calculate offset based on layout (Home vs Chat)
	var editorX, editorY int
	if p.app.Session.ID == "" {
		_, editorX, editorY = p.calculateHomeLayout()
	} else {
		_, editorX, editorY = p.calculateChatLayout()
	}

	// Apply offset. Note: View() adds Padding(0,2) but tui.go logic returned adjusted offsets.
	// We trust tui.go offset calculation logic adapted below.
	cursor.Position.X += editorX
	cursor.Position.Y += editorY

	return cursor
}

func (p *Page) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	// We don't call SetSize on components because they are not interfaces with that method.
	// They receive size updates via Update(tea.WindowSizeMsg).
	return nil
}

// Internal view helpers

func (p *Page) viewHome() string {
	layout, _, _ := p.calculateHomeLayout()
	return layout
}

func (p *Page) viewChat() string {
	layout, _, _ := p.calculateChatLayout()
	return layout
}

func (p *Page) calculateHomeLayout() (string, int, int) {
	t := theme.CurrentTheme()
	effectiveWidth := p.width - 4
	baseStyle := styles.NewStyle().Foreground(t.Text()).Background(t.Background())
	base := baseStyle.Render
	muted := styles.NewStyle().Foreground(t.TextMuted()).Background(t.Background()).Render

	open := `
                    
█▀▀█ █▀▀█ █▀▀█ █▀▀▄ 
█░░█ █░░█ █▀▀▀ █░░█ 
▀▀▀▀ █▀▀▀ ▀▀▀▀ ▀  ▀ `

	code := `
             ▄
█▀▀▀ █▀▀█ █▀▀█ █▀▀█
█░░░ █░░█ █░░█ █▀▀▀
▀▀▀▀ ▀▀▀▀ ▀▀▀▀ ▀▀▀▀`

	logo := lipgloss.JoinHorizontal(
		lipgloss.Top,
		muted(open),
		base(code),
	)

	versionStyle := styles.NewStyle().
		Foreground(t.TextMuted()).
		Background(t.Background()).
		Width(lipgloss.Width(logo)).
		Align(lipgloss.Right)
	version := versionStyle.Render(p.app.Version)

	logoAndVersion := strings.Join([]string{logo, version}, "\n")
	logoAndVersion = lipgloss.PlaceHorizontal(
		effectiveWidth,
		lipgloss.Center,
		logoAndVersion,
		styles.WhitespaceStyle(t.Background()),
	)

	limit := 5
	if util.IsVSCode() {
		limit = 3
	}

	showVscode := util.IsVSCode()
	commandsView := cmdcomp.New(
		p.app,
		cmdcomp.WithBackground(t.Background()),
		cmdcomp.WithLimit(limit),
		cmdcomp.WithVscode(showVscode),
	)
	cmds := lipgloss.PlaceHorizontal(
		effectiveWidth,
		lipgloss.Center,
		commandsView.View(),
		styles.WhitespaceStyle(t.Background()),
	)

	lines := []string{}
	lines = append(lines, "")
	lines = append(lines, logoAndVersion)
	lines = append(lines, "")
	lines = append(lines, cmds)
	lines = append(lines, "")
	lines = append(lines, "")

	mainHeight := lipgloss.Height(strings.Join(lines, "\n"))

	editorView := p.Editor.View()
	editorWidth := lipgloss.Width(editorView)
	editorView = lipgloss.PlaceHorizontal(
		effectiveWidth,
		lipgloss.Center,
		editorView,
		styles.WhitespaceStyle(t.Background()),
	)
	lines = append(lines, editorView)

	editorLines := p.Editor.Lines()

	mainLayout := lipgloss.Place(
		effectiveWidth,
		p.height,
		lipgloss.Center,
		lipgloss.Center,
		baseStyle.Render(strings.Join(lines, "\n")),
		styles.WhitespaceStyle(t.Background()),
	)

	editorX := max(0, (effectiveWidth-editorWidth)/2)
	editorY := (p.height / 2) + (mainHeight / 2) - 3
	editorYDelta := 3

	if editorLines > 1 {
		editorYDelta = 2
		content := p.Editor.Content()
		editorHeight := lipgloss.Height(content)

		if editorY+editorHeight > p.height {
			difference := (editorY + editorHeight) - p.height
			editorY -= difference
		}
		mainLayout = layout.PlaceOverlay(
			editorX,
			editorY,
			content,
			mainLayout,
		)
	}

	// Completion Overlay
	if p.ShowCompletionDialog {
		p.Completions.SetWidth(editorWidth)
		overlay := p.Completions.View()
		overlayHeight := lipgloss.Height(overlay)

		mainLayout = layout.PlaceOverlay(
			editorX,
			editorY-overlayHeight+2,
			overlay,
			mainLayout,
		)
	}

	// Returns layout, editorX, editorY (adjusted for final output)
	return mainLayout, editorX + 5, editorY + editorYDelta
}

func (p *Page) calculateChatLayout() (string, int, int) {
	effectiveWidth := p.width - 4
	t := theme.CurrentTheme()
	editorView := p.Editor.View()
	lines := p.Editor.Lines()
	messagesView := p.Messages.View()

	editorWidth := lipgloss.Width(editorView)
	editorHeight := max(lines, 5)

	editorView = lipgloss.PlaceHorizontal(
		effectiveWidth,
		lipgloss.Center,
		editorView,
		styles.WhitespaceStyle(t.Background()),
	)

	mainLayout := messagesView + "\n" + editorView

	// Recalculate editorX for centering
	editorX := max(0, (effectiveWidth-editorWidth)/2)
	editorY := p.height - editorHeight

	if lines > 1 {
		content := p.Editor.Content()
		editorHeight := lipgloss.Height(content)
		if editorY+editorHeight > p.height {
			difference := (editorY + editorHeight) - p.height
			editorY -= difference
		}
		mainLayout = layout.PlaceOverlay(
			editorX,
			editorY,
			content,
			mainLayout,
		)
	}

	if p.ShowCompletionDialog {
		p.Completions.SetWidth(editorWidth)
		overlay := p.Completions.View()
		overlayHeight := lipgloss.Height(overlay)

		// In chat, overlay is above editor
		overlayY := (p.height - editorHeight) - overlayHeight + 1

		mainLayout = layout.PlaceOverlay(
			editorX,
			overlayY,
			overlay,
			mainLayout,
		)
	}

	return mainLayout, editorX + 5, editorY + 2
}
