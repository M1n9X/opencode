package dialog

import (
	"context"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
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

// ProviderDialog interface for provider connection dialog
type ProviderDialog interface {
	layout.Modal
}

// ProviderConnectedMsg is sent when a provider is successfully connected
type ProviderConnectedMsg struct {
	ProviderID string
}

// providerItem represents a provider in the list
type providerItem struct {
	provider    opencode.Provider
	description string
	category    string
}

func (p providerItem) Render(
	selected bool,
	width int,
	isFirstInViewport bool,
	baseStyle styles.Style,
) string {
	t := theme.CurrentTheme()

	nameStyle := baseStyle.Foreground(t.Text())
	descStyle := baseStyle.Foreground(t.TextMuted())

	if selected {
		nameStyle = baseStyle.Background(t.Primary()).Foreground(t.BackgroundElement())
		descStyle = baseStyle.Background(t.Primary()).Foreground(t.BackgroundElement())
	}

	// Render name and description
	nameText := nameStyle.Bold(true).Render(p.provider.Name)
	descText := ""
	if p.description != "" {
		descText = descStyle.Render(" " + p.description)
	}

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

	return itemStyle.Render(nameText + descText)
}

func (p providerItem) Selectable() bool {
	return true
}

type providerDialogMode int

const (
	modeSelectProvider providerDialogMode = iota
	modeEnterAPIKey
)

type providerDialog struct {
	width       int
	height      int
	modal       *modal.Modal
	list        list.List[providerItem]
	app         *app.App
	providers   []opencode.Provider
	mode        providerDialogMode
	selectedID  string
	apiKeyInput textinput.Model
}

func (p *providerDialog) Init() tea.Cmd {
	return nil
}

func (p *providerDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.list.SetMaxWidth(layout.Current.Container.Width - 12)
	case tea.KeyPressMsg:
		if p.mode == modeEnterAPIKey {
			switch msg.String() {
			case "enter":
				apiKey := p.apiKeyInput.Value()
				if strings.TrimSpace(apiKey) != "" {
					return p, p.connectWithAPIKey(p.selectedID, apiKey)
				}
			case "esc":
				// Go back to provider selection
				p.mode = modeSelectProvider
				p.modal.SetTitle("Connect a Provider")
				return p, nil
			default:
				var cmd tea.Cmd
				p.apiKeyInput, cmd = p.apiKeyInput.Update(msg)
				return p, cmd
			}
			return p, nil
		}

		switch msg.String() {
		case "enter":
			if item, idx := p.list.GetSelectedItem(); idx >= 0 {
				// For now, just prompt for API key
				p.selectedID = item.provider.ID
				p.mode = modeEnterAPIKey
				p.modal.SetTitle("Enter API Key - " + item.provider.Name)
				p.setupAPIKeyInput()
				return p, textinput.Blink
			}
		}
	}

	if p.mode == modeSelectProvider {
		var cmd tea.Cmd
		listModel, cmd := p.list.Update(msg)
		p.list = listModel.(list.List[providerItem])
		return p, cmd
	}

	return p, nil
}

func (p *providerDialog) setupAPIKeyInput() {
	t := theme.CurrentTheme()
	bgColor := t.BackgroundPanel()
	textColor := t.Text()
	textMutedColor := t.TextMuted()

	p.apiKeyInput = textinput.New()
	p.apiKeyInput.Placeholder = "Enter your API key..."
	p.apiKeyInput.Focus()
	p.apiKeyInput.EchoMode = textinput.EchoPassword
	p.apiKeyInput.SetWidth(layout.Current.Container.Width - 20)

	p.apiKeyInput.Styles.Blurred.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	p.apiKeyInput.Styles.Blurred.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	p.apiKeyInput.Styles.Focused.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	p.apiKeyInput.Styles.Focused.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	p.apiKeyInput.Styles.Focused.Prompt = styles.NewStyle().
		Background(bgColor).
		Lipgloss()
}

func (p *providerDialog) connectWithAPIKey(providerID, apiKey string) tea.Cmd {
	return func() tea.Msg {
		// Note: The Go SDK doesn't have an Auth.Set method yet
		// For now, show a message directing users to configure via environment variables
		return tea.Sequence(
			util.CmdHandler(modal.CloseModalMsg{}),
			toast.NewInfoToast("Set API key via environment variable or config file"),
		)()
	}
}

func (p *providerDialog) Render(background string) string {
	t := theme.CurrentTheme()

	if p.mode == modeEnterAPIKey {
		// Show API key input
		var content strings.Builder

		content.WriteString(p.apiKeyInput.View())
		content.WriteString("\n\n")

		mutedStyle := styles.NewStyle().
			Foreground(t.TextMuted()).
			Background(t.BackgroundPanel())
		content.WriteString(mutedStyle.Render("Press Enter to confirm, Esc to go back"))

		return p.modal.Render(content.String(), background)
	}

	// Show provider list
	listView := p.list.View()

	keyStyle := styles.NewStyle().
		Foreground(t.Text()).
		Background(t.BackgroundPanel()).
		Bold(true).
		Render
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted()).Background(t.BackgroundPanel()).Render

	helpText := keyStyle("enter") + mutedStyle(" select   ") + keyStyle("esc") + mutedStyle(" close")

	bgColor := t.BackgroundPanel()
	helpView := styles.NewStyle().
		Background(bgColor).
		Width(layout.Current.Container.Width - 14).
		PaddingLeft(1).
		PaddingTop(1).
		Render(helpText)

	content := strings.Join([]string{listView, helpView}, "\n")

	return p.modal.Render(content, background)
}

func (p *providerDialog) Close() tea.Cmd {
	return nil
}

// Provider priority for sorting (lower = higher priority)
var providerPriority = map[string]int{
	"opencode":       0,
	"anthropic":      1,
	"github-copilot": 2,
	"openai":         3,
	"google":         4,
}

// Provider descriptions
var providerDescriptions = map[string]string{
	"opencode":  "(Recommended)",
	"anthropic": "(Claude Max or API key)",
	"openai":    "(ChatGPT Plus/Pro or API key)",
}

// NewProviderDialog creates a new provider connection dialog
func NewProviderDialog(app *app.App) ProviderDialog {
	// Get providers from app
	ctx := context.Background()
	providersResp, _ := app.Client.App.Providers(ctx, opencode.AppProvidersParams{})

	var providers []opencode.Provider
	if providersResp != nil {
		providers = providersResp.Providers
	}

	// Sort providers by priority
	sort.Slice(providers, func(i, j int) bool {
		pi := providerPriority[providers[i].ID]
		pj := providerPriority[providers[j].ID]
		if pi == 0 && providers[i].ID != "opencode" {
			pi = 99
		}
		if pj == 0 && providers[j].ID != "opencode" {
			pj = 99
		}
		return pi < pj
	})

	// Create list items
	var items []providerItem
	for _, provider := range providers {
		desc := providerDescriptions[provider.ID]
		category := "Other"
		if _, ok := providerPriority[provider.ID]; ok {
			category = "Popular"
		}
		items = append(items, providerItem{
			provider:    provider,
			description: desc,
			category:    category,
		})
	}

	listComponent := list.NewListComponent(
		list.WithItems(items),
		list.WithMaxVisibleHeight[providerItem](10),
		list.WithFallbackMessage[providerItem]("No providers available"),
		list.WithAlphaNumericKeys[providerItem](true),
		list.WithRenderFunc(
			func(item providerItem, selected bool, width int, baseStyle styles.Style) string {
				return item.Render(selected, width, false, baseStyle)
			},
		),
		list.WithSelectableFunc(func(item providerItem) bool {
			return true
		}),
	)
	listComponent.SetMaxWidth(layout.Current.Container.Width - 12)

	return &providerDialog{
		list:      listComponent,
		app:       app,
		providers: providers,
		mode:      modeSelectProvider,
		modal: modal.New(
			modal.WithTitle("Connect a Provider"),
			modal.WithMaxWidth(layout.Current.Container.Width-8),
		),
	}
}
