package dialog

import (
	"context"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/list"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

type McpDialog interface {
	layout.Modal
}

type McpStatus struct {
	Status string `json:"status"`
}

type mcpDialog struct {
	app          *app.App
	searchDialog *SearchDialog
	modal        *modal.Modal
	width        int
	height       int
	status       map[string]McpStatus
	loading      string // name of the MCP currently being toggled
}

type mcpSelectItem struct {
	name        string
	status      string
	enabled     bool
	loading     bool
	displayText string
}

func (m mcpSelectItem) Render(selected bool, width int, baseStyle styles.Style) string {
	t := theme.CurrentTheme()

	// Base style
	itemStyle := baseStyle.
		Background(t.BackgroundPanel()).
		Foreground(t.Text())

	if selected {
		itemStyle = itemStyle.Foreground(t.Primary())
	}

	// Status indicator style
	statusStyle := baseStyle.Background(t.BackgroundPanel())
	var statusIndicator string

	if m.loading {
		statusStyle = statusStyle.Foreground(t.TextMuted())
		statusIndicator = "⋯ Loading"
	} else if m.enabled {
		statusStyle = statusStyle.Foreground(t.Success()).Bold(true)
		statusIndicator = "✓ Enabled"
	} else {
		statusStyle = statusStyle.Foreground(t.TextMuted())
		statusIndicator = "○ Disabled"
	}

	if m.status == "failed" {
		statusStyle = statusStyle.Foreground(t.Error())
		statusIndicator = "✕ Failed"
	}

	// Layout
	availableWidth := width - 2
	maxNameWidth := availableWidth - len(statusIndicator) - 1

	name := m.name
	if len(name) > maxNameWidth {
		name = name[:maxNameWidth]
	}

	namePart := itemStyle.Render(name)
	spacer := strings.Repeat(" ", max(0, availableWidth-len(name)-len(statusIndicator)))
	statusPart := statusStyle.Render(statusIndicator)

	return baseStyle.
		Background(t.BackgroundPanel()).
		PaddingLeft(1).
		Width(width).
		Render(namePart + spacer + statusPart)
}

func (m mcpSelectItem) Selectable() bool {
	return true
}

type mcpStatusLoadedMsg map[string]McpStatus

func (m *mcpDialog) Init() tea.Cmd {
	return tea.Batch(
		m.searchDialog.Init(),
		m.fetchStatus(),
	)
}

func (m *mcpDialog) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		var resp struct {
			Data map[string]McpStatus `json:"data"`
		}
		err := m.app.Client.Get(context.Background(), "/mcp/status", nil, &resp)
		if err != nil {
			return nil // TODO: Handle error
		}
		return mcpStatusLoadedMsg(resp.Data)
	}
}

func (m *mcpDialog) toggleMcp(name string) tea.Cmd {
	m.loading = name
	m.updateItems() // Update UI to show loading state

	return func() tea.Msg {
		ctx := context.Background()
		status, ok := m.status[name]

		var err error
		body := map[string]interface{}{"name": name}

		if ok && status.Status == "connected" {
			// Disable
			err = m.app.Client.Post(ctx, "/mcp/disconnect", body, nil)
		} else {
			// Enable/Retry
			err = m.app.Client.Post(ctx, "/mcp/connect", body, nil)
		}

		if err != nil {
			// TODO: Handle error
		}

		// Refresh status
		var resp struct {
			Data map[string]McpStatus `json:"data"`
		}
		err = m.app.Client.Get(context.Background(), "/mcp/status", nil, &resp)
		if err != nil {
			return nil
		}
		return mcpStatusLoadedMsg(resp.Data)
	}
}

func (m *mcpDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, searchKeys.Remove) {
			// Placeholder
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchDialog.SetWidth(m.calculateWidth())
		m.searchDialog.SetHeight(msg.Height)

	case mcpStatusLoadedMsg:
		m.status = msg
		m.loading = ""
		m.updateItems()
		return m, nil

	case SearchSelectionMsg:
		if item, ok := msg.Item.(mcpSelectItem); ok {
			// Toggle on selection (Enter)
			return m, m.toggleMcp(item.name)
		}

	case SearchCancelledMsg:
		return m, util.CmdHandler(modal.CloseModalMsg{})

	case SearchQueryChangedMsg:
		m.updateItems()
	}

	updatedDialog, cmd := m.searchDialog.Update(msg)
	m.searchDialog = updatedDialog.(*SearchDialog)
	return m, cmd
}

func (m *mcpDialog) View() string {
	return m.modal.Render(m.searchDialog.View(), "")
}

func (m *mcpDialog) Render(background string) string {
	return m.modal.Render(m.searchDialog.View(), background)
}

func (m *mcpDialog) Close() tea.Cmd {
	return nil
}

func (m *mcpDialog) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *mcpDialog) calculateWidth() int {
	return 60 // Fixed width for now
}

func (m *mcpDialog) updateItems() {
	query := m.searchDialog.GetQuery()
	names := make([]string, 0, len(m.status))
	for name := range m.status {
		names = append(names, name)
	}
	sort.Strings(names)

	if query != "" {
		matches := fuzzy.RankFindFold(query, names)
		sort.Sort(matches)
		names = make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Target)
		}
	}

	items := make([]list.Item, 0, len(names))
	for _, name := range names {
		status := m.status[name]
		items = append(items, mcpSelectItem{
			name:    name,
			status:  string(status.Status),
			enabled: status.Status == "connected",
			loading: m.loading == name,
		})
	}
	m.searchDialog.SetItems(items)
}

func NewMcpDialog(app *app.App) McpDialog {
	searchDialog := NewSearchDialog("Search MCPs...", 10)

	m := &mcpDialog{
		app:          app,
		searchDialog: searchDialog,
		status:       make(map[string]McpStatus),
	}

	m.modal = modal.New(
		modal.WithTitle("MCPs"),
		modal.WithMaxWidth(64),
	)
	m.searchDialog.SetWidth(60)

	return m
}
