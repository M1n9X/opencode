package dialog

import (
	"context"
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/list"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

type FileDialog interface {
	layout.Modal
}

type fileDialog struct {
	app          *app.App
	searchDialog *SearchDialog
	modal        *modal.Modal
	width        int
	height       int
}

type fileSelectItem struct {
	path string
}

func (f fileSelectItem) Render(selected bool, width int, baseStyle styles.Style) string {
	t := theme.CurrentTheme()

	// Base style
	itemStyle := baseStyle.
		Background(t.BackgroundPanel()).
		Foreground(t.Text())

	if selected {
		itemStyle = itemStyle.Foreground(t.Primary())
	}

	icon := getFileIcon(f.path)
	text := icon + " " + f.path

	// Layout
	availableWidth := width - 2
	if len(text) > availableWidth {
		text = text[:availableWidth]
	}

	return baseStyle.
		Background(t.BackgroundPanel()).
		PaddingLeft(1).
		Width(width).
		Render(text)
}

func (f fileSelectItem) Selectable() bool {
	return true
}

func getFileIcon(path string) string {
	if strings.HasSuffix(path, "/") {
		return "📁"
	}
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		ext := parts[len(parts)-1]
		switch strings.ToLower(ext) {
		case "go":
			return "🐹"
		case "ts", "tsx", "js", "jsx":
			return "⚡"
		case "md":
			return "📝"
		case "json", "yaml", "yml", "toml":
			return "⚙️"
		case "css", "scss", "html":
			return "🎨"
		case "png", "jpg", "jpeg", "svg", "gif":
			return "🖼️"
		case "sh", "bash", "zsh":
			return "🐚"
		case "dockerfile":
			return "🐳"
		case "gitignore", "dockerignore":
			return "👁️"
		}
	}
	return "📄"
}

type fileSearchResultsMsg []string

func (f *fileDialog) Init() tea.Cmd {
	return tea.Batch(
		f.searchDialog.Init(),
		f.searchFiles(""), // Initial search
	)
}

func (f *fileDialog) searchFiles(query string) tea.Cmd {
	return func() tea.Msg {
		files, err := f.app.Client.Find.Files(
			context.Background(),
			opencode.FindFilesParams{Query: opencode.F(query)},
		)
		if err != nil {
			slog.Error("Failed to search files", "error", err)
			return nil
		}
		if files == nil {
			return fileSearchResultsMsg([]string{})
		}
		return fileSearchResultsMsg(*files)
	}
}

func (f *fileDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
		f.searchDialog.SetWidth(f.calculateWidth())
		f.searchDialog.SetHeight(msg.Height)

	case SearchSelectionMsg:
		if item, ok := msg.Item.(fileSelectItem); ok {
			// Append file path to editor
			return f, tea.Sequence(
				util.CmdHandler(modal.CloseModalMsg{}),
				util.CmdHandler(app.SetEditorContentMsg{Text: item.path, Append: true}),
			)
		}

	case SearchCancelledMsg:
		return f, util.CmdHandler(modal.CloseModalMsg{})

	case SearchQueryChangedMsg:
		return f, f.searchFiles(msg.Query)

	case fileSearchResultsMsg:
		items := make([]list.Item, len(msg))
		for i, path := range msg {
			items[i] = fileSelectItem{path: path}
		}
		f.searchDialog.SetItems(items)
	}

	updatedDialog, cmd := f.searchDialog.Update(msg)
	f.searchDialog = updatedDialog.(*SearchDialog)
	return f, cmd
}

func (f *fileDialog) View() string {
	return f.modal.Render(f.searchDialog.View(), "")
}

func (f *fileDialog) Render(background string) string {
	return f.modal.Render(f.searchDialog.View(), background)
}

func (f *fileDialog) Close() tea.Cmd {
	return nil
}

func (f *fileDialog) SetSize(width, height int) {
	f.width = width
	f.height = height
}

func (f *fileDialog) calculateWidth() int {
	return 60 // Fixed width
}

func NewFileDialog(app *app.App) FileDialog {
	searchDialog := NewSearchDialog("Search files...", 10)

	f := &fileDialog{
		app:          app,
		searchDialog: searchDialog,
	}

	f.modal = modal.New(
		modal.WithTitle("Files"),
		modal.WithMaxWidth(64),
	)
	f.searchDialog.SetWidth(60)

	return f
}
