package dialog

import (
	"context"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/shared"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/diff"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/components/toast"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
	"github.com/sst/opencode/internal/viewport"
)

// PermissionDialog renders a permission request modal with full details.
type PermissionDialog interface {
	layout.Modal
}

type permissionDialog struct {
	app         *app.App
	modal       *modal.Modal
	permission  opencode.Permission
	viewport    viewport.Model
	width       int
	height      int
	lastContent string
}

func NewPermissionDialog(app *app.App, perm opencode.Permission) PermissionDialog {
	vp := viewport.New(viewport.WithHeight(12))
	title := perm.Title
	if title == "" {
		title = "Permission required"
	}
	return &permissionDialog{
		app:        app,
		permission: perm,
		modal: modal.New(
			modal.WithTitle(title),
			modal.WithMaxWidth(96),
		),
		viewport: vp,
	}
}

func (p *permissionDialog) Init() tea.Cmd {
	return p.viewport.Init()
}

func (p *permissionDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		maxWidth := min(96, layout.Current.Container.Width-8)
		p.viewport = viewport.New(
			viewport.WithWidth(maxWidth-4),
			viewport.WithHeight(msg.Height-6),
		)
		p.refreshContent()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return p, p.respond(opencode.SessionPermissionRespondParamsResponseOnce)
		case "a":
			return p, p.respond(opencode.SessionPermissionRespondParamsResponseAlways)
		case "esc":
			return p, p.respond(opencode.SessionPermissionRespondParamsResponseReject)
		}
	}

	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

func (p *permissionDialog) respond(resp opencode.SessionPermissionRespondParamsResponse) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		_, err := p.app.Client.Session.Permissions.Respond(
			ctx,
			p.permission.SessionID,
			p.permission.ID,
			opencode.SessionPermissionRespondParams{
				Response: opencode.F(resp),
			},
		)
		if err != nil {
			return toast.NewErrorToast("Failed to respond to permission: " + err.Error())()
		}
		// Optimistically clear current permission
		p.app.CurrentPermission = opencode.Permission{}
		if len(p.app.Permissions) > 0 {
			p.app.Permissions = p.app.Permissions[1:]
		}
		return modal.CloseModalMsg{}
	}
}

func (p *permissionDialog) Render(background string) string {
	p.refreshContent()
	return p.modal.Render(p.viewport.View(), background)
}

func (p *permissionDialog) Close() tea.Cmd {
	return nil
}

func (p *permissionDialog) refreshContent() {
	content := p.buildContent()
	if content != p.lastContent {
		p.viewport.SetContent(content)
		p.lastContent = content
	}
}

func (p *permissionDialog) buildContent() string {
	t := theme.CurrentTheme()
	base := styles.NewStyle().Background(t.BackgroundPanel()).Foreground(t.Text()).Render
	muted := styles.NewStyle().Background(t.BackgroundPanel()).Foreground(t.TextMuted()).Render

	var sections []string

	// Title/summary
	if p.permission.Title != "" {
		sections = append(sections, base(p.permission.Title))
	}

	// Metadata lines
	meta := p.permission.Metadata
	if filePath, ok := meta["filePath"].(string); ok && filePath != "" {
		sections = append(sections, muted("Path: ")+base(util.RelPath(filePath)))
	}
	if cmd, ok := meta["command"].(string); ok && cmd != "" {
		sections = append(sections, muted("Command: ")+base(cmd))
	}
	if pat := extractPattern(p.permission.Pattern); pat != "" {
		sections = append(sections, muted("Pattern: ")+base(pat))
	}

	// Diff preview
	if diffText, ok := meta["diff"].(string); ok && diffText != "" {
		filename := ""
		if filePath, ok := meta["filePath"].(string); ok {
			filename = filePath
		}
		w := 80
		if p.viewport.Width > 0 {
			w = p.viewport.Width - 2
		}
		formatted, err := diff.FormatUnifiedDiff(filename, diffText, diff.WithWidth(w))
		if err == nil && formatted != "" {
			header := styles.NewStyle().
				Background(t.BackgroundPanel()).
				Foreground(t.Secondary()).
				Bold(true).
				Render("Diff")
			sections = append(sections, header, formatted)
		}
	}

	// Preview content
	if preview, ok := meta["preview"].(string); ok && preview != "" {
		sections = append(sections, muted("Preview:"), base(preview))
	}

	// Helper line
	keyStyle := styles.NewStyle().Background(t.BackgroundPanel()).Foreground(t.Text()).Bold(true).Render
	help := keyStyle("enter") + muted(" accept   ") +
		keyStyle("a") + muted(" always   ") +
		keyStyle("esc") + muted(" reject")
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func extractPattern(p opencode.PermissionPatternUnion) string {
	switch v := p.(type) {
	case nil:
		return ""
	case opencode.PermissionPatternArray:
		if len(v) == 0 {
			return ""
		}
		return strings.Join(v, ", ")
	case shared.UnionString:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
