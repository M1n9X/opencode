package dialog

import (
	"context"
	"log/slog"

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

// TagDialog provides quick tag/file autocomplete (parity with OpenTUI DialogTag)
type TagDialog interface {
	layout.Modal
}

type tagDialog struct {
	app          *app.App
	searchDialog *SearchDialog
	modal        *modal.Modal
}

type tagSelectItem struct {
	value string
}

func (t tagSelectItem) Render(selected bool, width int, baseStyle styles.Style) string {
	theme := theme.CurrentTheme()
	style := baseStyle.
		Background(theme.BackgroundPanel()).
		Foreground(theme.Text()).
		PaddingLeft(1).
		Width(width)
	if selected {
		style = style.Foreground(theme.Primary())
	}
	return style.Render(t.value)
}

func (t tagSelectItem) Selectable() bool { return true }

type tagSearchResultsMsg []string

func NewTagDialog(app *app.App) TagDialog {
	searchDialog := NewSearchDialog("Search tags...", 8)
	searchDialog.SetWidth(60)

	d := &tagDialog{
		app:          app,
		searchDialog: searchDialog,
		modal: modal.New(
			modal.WithTitle("Tags"),
			modal.WithMaxWidth(64),
		),
	}
	return d
}

func (t *tagDialog) Init() tea.Cmd {
	return tea.Batch(
		t.searchDialog.Init(),
		t.searchTags(""),
	)
}

func (t *tagDialog) searchTags(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := t.app.Client.Find.Files(
			context.Background(),
			opencode.FindFilesParams{Query: opencode.F(query)},
		)
		if err != nil {
			slog.Error("Failed to search tags", "error", err)
			return nil
		}
		if results == nil {
			return tagSearchResultsMsg{}
		}
		return tagSearchResultsMsg(*results)
	}
}

func (t *tagDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchSelectionMsg:
		if item, ok := msg.Item.(tagSelectItem); ok {
			return t, tea.Sequence(
				util.CmdHandler(modal.CloseModalMsg{}),
				util.CmdHandler(app.SetEditorContentMsg{Text: item.value, Append: true}),
			)
		}
	case SearchCancelledMsg:
		return t, util.CmdHandler(modal.CloseModalMsg{})
	case SearchQueryChangedMsg:
		return t, t.searchTags(msg.Query)
	case tagSearchResultsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = tagSelectItem{value: v}
		}
		t.searchDialog.SetItems(items)
	}

	updatedDialog, cmd := t.searchDialog.Update(msg)
	t.searchDialog = updatedDialog.(*SearchDialog)
	return t, cmd
}

func (t *tagDialog) View() string {
	return t.modal.Render(t.searchDialog.View(), "")
}

func (t *tagDialog) Render(background string) string {
	return t.modal.Render(t.searchDialog.View(), background)
}

func (t *tagDialog) Close() tea.Cmd { return nil }
