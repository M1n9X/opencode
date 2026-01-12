package dialog

import (
	"fmt"
	"sort"
	"strings"
	"time"

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

type StashDialog interface {
	layout.Modal
}

type stashDialog struct {
	app          *app.App
	searchDialog *SearchDialog
	modal        *modal.Modal
	width        int
	height       int
	toDelete     *int // index of item to confirm delete
}

type stashSelectItem struct {
	index     int
	input     string
	timestamp time.Time
	toDelete  bool
}

func (s stashSelectItem) Render(selected bool, width int, baseStyle styles.Style) string {
	t := theme.CurrentTheme()

	// Base style
	itemStyle := baseStyle.
		Background(t.BackgroundPanel()).
		Foreground(t.Text())

	if selected {
		itemStyle = itemStyle.Foreground(t.Primary())
	}

	if s.toDelete {
		itemStyle = itemStyle.Background(t.Error()).Foreground(t.Text())
	}

	// Calculate time string
	timeStr := util.FormatRelativeTime(s.timestamp)

	lineCount := strings.Count(s.input, "\n") + 1
	infoStr := timeStr
	if lineCount > 1 {
		infoStr += fmt.Sprintf(" (~%d lines)", lineCount)
	}

	timeStyle := baseStyle.
		Background(t.BackgroundPanel()).
		Foreground(t.TextMuted())

	if s.toDelete {
		timeStyle = timeStyle.Background(t.Error()).Foreground(t.Text())
	}

	// Layout
	availableWidth := width - 2
	maxInputWidth := availableWidth - len(infoStr) - 1

	inputPreview := strings.Split(s.input, "\n")[0]
	if len(inputPreview) > maxInputWidth {
		inputPreview = inputPreview[:maxInputWidth]
	}

	if s.toDelete {
		inputPreview = "Press Ctrl+X again to confirm delete"
	}

	inputPart := itemStyle.Render(inputPreview)
	spacer := strings.Repeat(" ", max(0, availableWidth-len(inputPreview)-len(infoStr)))
	timePart := timeStyle.Render(infoStr)

	return baseStyle.
		Background(t.BackgroundPanel()).
		PaddingLeft(1).
		Width(width).
		Render(inputPart + spacer + timePart)
}

func (s stashSelectItem) Selectable() bool {
	return true
}

func (s *stashDialog) Init() tea.Cmd {
	s.updateItems()
	return s.searchDialog.Init()
}

func (s *stashDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// SearchDialog handles Remove (Ctrl+X) via SearchRemoveItemMsg
		// We rely on that.

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.searchDialog.SetWidth(s.calculateWidth())
		s.searchDialog.SetHeight(msg.Height)

	case SearchSelectionMsg:
		if item, ok := msg.Item.(stashSelectItem); ok {
			// Restore stash
			// We remove it from stash upon restore? Opentui does: stash.remove(option.value)
			s.app.State.RemovePromptFromStash(item.index)
			s.app.SaveState()

			return s, tea.Sequence(
				util.CmdHandler(modal.CloseModalMsg{}),
				util.CmdHandler(app.SetEditorContentMsg{Text: item.input}),
			)
		}

	case SearchRemoveItemMsg:
		if item, ok := msg.Item.(stashSelectItem); ok {
			if s.toDelete != nil && *s.toDelete == item.index {
				// Confirmed delete
				s.app.State.RemovePromptFromStash(item.index)
				s.app.SaveState()
				s.toDelete = nil
				s.updateItems()
				return s, nil
			} else {
				// Request confirmation
				idx := item.index
				s.toDelete = &idx
				s.updateItems()
				return s, nil
			}
		}

	case SearchCancelledMsg:
		return s, util.CmdHandler(modal.CloseModalMsg{})

	case SearchQueryChangedMsg:
		s.toDelete = nil // reset delete confirmation on search change
		s.updateItems()
	}

	updatedDialog, cmd := s.searchDialog.Update(msg)
	s.searchDialog = updatedDialog.(*SearchDialog)
	return s, cmd
}

func (s *stashDialog) View() string {
	return s.modal.Render(s.searchDialog.View(), "")
}

func (s *stashDialog) Render(background string) string {
	return s.modal.Render(s.searchDialog.View(), background)
}

func (s *stashDialog) Close() tea.Cmd {
	return nil
}

func (s *stashDialog) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *stashDialog) calculateWidth() int {
	return 60 // Fixed width
}

func (s *stashDialog) updateItems() {
	query := s.searchDialog.GetQuery()
	entries := s.app.State.PromptStash

	// Create selectable items
	// We need to preserve original index for deletion, but display reversed or sorted
	// Actually, `app.State.PromptStash` index is what matters for deletion.

	// Filter first
	type match struct {
		entry app.StashEntry
		index int
	}
	var matches []match

	if query == "" {
		for i, entry := range entries {
			matches = append(matches, match{entry, i})
		}
	} else {
		// Fuzzy search
		// Simple approach: check input
		for i, entry := range entries {
			if fuzzy.MatchFold(query, entry.Input) {
				matches = append(matches, match{entry, i})
			}
		}
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].entry.Timestamp.After(matches[j].entry.Timestamp)
	})

	items := make([]list.Item, 0, len(matches))
	for _, m := range matches {
		toDelete := false
		if s.toDelete != nil && *s.toDelete == m.index {
			toDelete = true
		}

		items = append(items, stashSelectItem{
			index:     m.index,
			input:     m.entry.Input,
			timestamp: m.entry.Timestamp,
			toDelete:  toDelete,
		})
	}

	s.searchDialog.SetItems(items)
}

func NewStashDialog(app *app.App) StashDialog {
	searchDialog := NewSearchDialog("Search stash...", 10)

	s := &stashDialog{
		app:          app,
		searchDialog: searchDialog,
	}

	s.modal = modal.New(
		modal.WithTitle("Stash"),
		modal.WithMaxWidth(64),
	)
	s.searchDialog.SetWidth(60)

	return s
}
