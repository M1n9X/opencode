package dialog

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
)

var tips = []string{
	"Use <ctrl+x> <space> to toggle full selection in lists.",
	"Use <ctrl+p> / <ctrl+n> to navigate history in chat.",
	"Press <!> to switch to Bash mode for shell commands.",
	"Use <@> in chat to mention agents or files.",
	"Use </> in chat to access commands quickly.",
	"Press <ctrl+z> to suspend the TUI.",
	"Use <ctrl+c> to cancel generation or exit.",
}

type DidYouKnow struct {
	app     *app.App
	visible bool
	tip     string
}

type ToggleTipsMsg struct{}

func NewDidYouKnow(app *app.App) *DidYouKnow {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &DidYouKnow{
		app:     app,
		visible: true,
		tip:     tips[rng.Intn(len(tips))],
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
		// Optional: Keybinding to toggle? Opentui uses specific key.
		if key.Matches(msg, key.NewBinding(key.WithKeys("?"))) { // Example
			// d.visible = !d.visible
		}
	}
	return d, nil
}

func (d *DidYouKnow) View() string {
	if !d.visible {
		return ""
	}
	t := theme.CurrentTheme()

	boxStyle := styles.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border()).
		Padding(0, 1).
		Foreground(t.TextMuted())

	return boxStyle.Render("Did you know? " + d.tip)
}
