package dialog

import (
	"context"
	"sort"
	"strings"
	"time"

	"slices"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/muesli/reflow/truncate"
	"github.com/sahilm/fuzzy"
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

// SessionDialog interface for the session switching dialog
type SessionDialog interface {
	layout.Modal
}

// sessionItem is a custom list item for sessions that can show delete confirmation
type sessionItem struct {
	session            opencode.Session
	isDeleteConfirming bool
	isCurrentSession   bool
	isBusy             bool
	timeAgo            string
}

func (s sessionItem) Render(
	selected bool,
	width int,
	isFirstInViewport bool,
	baseStyle styles.Style,
) string {
	t := theme.CurrentTheme()

	var text string
	if s.isDeleteConfirming {
		text = "Press again to confirm delete"
	} else {
		prefix := ""
		if s.isCurrentSession {
			prefix = "● "
		} else if s.isBusy {
			prefix = "[⋯] "
		}
		text = prefix + s.session.Title
	}

	// Calculate available width for title (leave space for time)
	timeWidth := len(s.timeAgo)
	availableWidth := width - timeWidth - 4 // 4 for spacing and padding

	truncatedTitle := truncate.StringWithTail(text, uint(availableWidth), "...")

	var itemStyle styles.Style
	if selected {
		if s.isDeleteConfirming {
			// Red background for delete confirmation
			itemStyle = baseStyle.
				Background(t.Error()).
				Foreground(t.BackgroundElement()).
				Width(width).
				PaddingLeft(1)
		} else if s.isCurrentSession {
			// Different style for current session when selected
			itemStyle = baseStyle.
				Background(t.Primary()).
				Foreground(t.BackgroundElement()).
				Width(width).
				PaddingLeft(1).
				Bold(true)
		} else {
			// Normal selection
			itemStyle = baseStyle.
				Background(t.Primary()).
				Foreground(t.BackgroundElement()).
				Width(width).
				PaddingLeft(1)
		}
	} else {
		if s.isDeleteConfirming {
			// Red text for delete confirmation when not selected
			itemStyle = baseStyle.
				Foreground(t.Error()).
				PaddingLeft(1)
		} else if s.isCurrentSession {
			// Highlight current session when not selected
			itemStyle = baseStyle.
				Foreground(t.Primary()).
				PaddingLeft(1).
				Bold(true)
		} else if s.isBusy {
			// Muted style for busy indicator
			itemStyle = baseStyle.
				Foreground(t.Text()).
				PaddingLeft(1)
		} else {
			itemStyle = baseStyle.
				PaddingLeft(1)
		}
	}

	// Build the line with title and time
	titlePart := truncatedTitle
	spacing := width - len(truncatedTitle) - timeWidth - 2
	if spacing < 1 {
		spacing = 1
	}

	timeStyle := baseStyle.Foreground(t.TextMuted())
	if selected {
		timeStyle = baseStyle.Foreground(t.BackgroundElement())
	}

	line := itemStyle.Render(titlePart + strings.Repeat(" ", spacing) + timeStyle.Render(s.timeAgo))
	return line
}

func (s sessionItem) Selectable() bool {
	return true
}

type sessionDialog struct {
	width              int
	height             int
	modal              *modal.Modal
	allSessions        []opencode.Session // All sessions (unfiltered)
	filteredSessions   []opencode.Session // Filtered sessions based on search
	list               list.List[sessionItem]
	app                *app.App
	deleteConfirmation int // -1 means no confirmation, >= 0 means confirming deletion of session at this index
	renameMode         bool
	renameInput        textinput.Model
	renameIndex        int // index of session being renamed
	searchInput        textinput.Model
	searchQuery        string
	sessionStatus      map[string]string // session ID -> status ("busy", "idle")
}

func (s *sessionDialog) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		s.loadSessionStatus(),
	)
}

func (s *sessionDialog) loadSessionStatus() tea.Cmd {
	return func() tea.Msg {
		// Try to get session status from API
		// For now, we'll use a simple heuristic based on session data
		return nil
	}
}

func (s *sessionDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.list.SetMaxWidth(layout.Current.Container.Width - 12)
		s.searchInput.SetWidth(layout.Current.Container.Width - 20)
	case tea.KeyPressMsg:
		if s.renameMode {
			switch msg.String() {
			case "enter":
				if _, idx := s.list.GetSelectedItem(); idx >= 0 && idx < len(s.filteredSessions) && idx == s.renameIndex {
					newTitle := s.renameInput.Value()
					if strings.TrimSpace(newTitle) != "" {
						sessionToUpdate := s.filteredSessions[idx]
						return s, tea.Sequence(
							func() tea.Msg {
								ctx := context.Background()
								err := s.app.UpdateSession(ctx, sessionToUpdate.ID, newTitle)
								if err != nil {
									return toast.NewErrorToast("Failed to rename session: " + err.Error())()
								}
								// Update in both lists
								for i := range s.allSessions {
									if s.allSessions[i].ID == sessionToUpdate.ID {
										s.allSessions[i].Title = newTitle
										break
									}
								}
								s.filteredSessions[idx].Title = newTitle
								s.renameMode = false
								s.modal.SetTitle("Switch Session")
								s.updateListItems()
								return toast.NewSuccessToast("Session renamed successfully")()
							},
						)
					}
				}
				s.renameMode = false
				s.modal.SetTitle("Switch Session")
				s.updateListItems()
				return s, nil
			case "esc":
				s.renameMode = false
				s.modal.SetTitle("Switch Session")
				s.updateListItems()
				return s, nil
			default:
				var cmd tea.Cmd
				s.renameInput, cmd = s.renameInput.Update(msg)
				return s, cmd
			}
		} else {
			switch msg.String() {
			case "enter":
				if s.deleteConfirmation >= 0 {
					s.deleteConfirmation = -1
					s.updateListItems()
					return s, nil
				}
				if _, idx := s.list.GetSelectedItem(); idx >= 0 && idx < len(s.filteredSessions) {
					selectedSession := s.filteredSessions[idx]
					return s, tea.Sequence(
						util.CmdHandler(modal.CloseModalMsg{}),
						util.CmdHandler(app.SessionSelectedMsg(&selectedSession)),
					)
				}
			case "n":
				// Only trigger new session if search is empty
				if s.searchQuery == "" {
					return s, tea.Sequence(
						util.CmdHandler(modal.CloseModalMsg{}),
						util.CmdHandler(app.SessionClearedMsg{}),
					)
				}
				// Otherwise, let it fall through to search input
				var cmd tea.Cmd
				s.searchInput, cmd = s.searchInput.Update(msg)
				s.handleSearchChange()
				return s, cmd
			case "r":
				// Only trigger rename if search is empty
				if s.searchQuery == "" {
					if _, idx := s.list.GetSelectedItem(); idx >= 0 && idx < len(s.filteredSessions) {
						s.renameMode = true
						s.renameIndex = idx
						s.setupRenameInput(s.filteredSessions[idx].Title)
						s.modal.SetTitle("Rename Session")
						s.updateListItems()
						return s, textinput.Blink
					}
				}
				// Otherwise, let it fall through to search input
				var cmd tea.Cmd
				s.searchInput, cmd = s.searchInput.Update(msg)
				s.handleSearchChange()
				return s, cmd
			case "ctrl+d":
				if _, idx := s.list.GetSelectedItem(); idx >= 0 && idx < len(s.filteredSessions) {
					if s.deleteConfirmation == idx {
						// Second press - actually delete the session
						sessionToDelete := s.filteredSessions[idx]
						return s, tea.Sequence(
							func() tea.Msg {
								// Remove from both lists
								for i := range s.allSessions {
									if s.allSessions[i].ID == sessionToDelete.ID {
										s.allSessions = slices.Delete(s.allSessions, i, i+1)
										break
									}
								}
								s.filteredSessions = slices.Delete(s.filteredSessions, idx, idx+1)
								s.deleteConfirmation = -1
								s.updateListItems()
								return nil
							},
							s.deleteSession(sessionToDelete.ID),
						)
					} else {
						// First press - enter delete confirmation mode
						s.deleteConfirmation = idx
						s.updateListItems()
						return s, nil
					}
				}
			case "up", "k":
				s.deleteConfirmation = -1
				var cmd tea.Cmd
				listModel, cmd := s.list.Update(msg)
				s.list = listModel.(list.List[sessionItem])
				s.updateListItems()
				return s, cmd
			case "down", "j":
				s.deleteConfirmation = -1
				var cmd tea.Cmd
				listModel, cmd := s.list.Update(msg)
				s.list = listModel.(list.List[sessionItem])
				s.updateListItems()
				return s, cmd
			case "esc":
				if s.deleteConfirmation >= 0 {
					s.deleteConfirmation = -1
					s.updateListItems()
					return s, nil
				}
				if s.searchQuery != "" {
					s.searchInput.SetValue("")
					s.handleSearchChange()
					return s, nil
				}
			case "backspace":
				var cmd tea.Cmd
				s.searchInput, cmd = s.searchInput.Update(msg)
				s.handleSearchChange()
				return s, cmd
			default:
				// Handle text input for search
				var cmd tea.Cmd
				s.searchInput, cmd = s.searchInput.Update(msg)
				s.handleSearchChange()
				return s, cmd
			}
		}
	}

	if !s.renameMode {
		var cmd tea.Cmd
		listModel, cmd := s.list.Update(msg)
		s.list = listModel.(list.List[sessionItem])
		return s, cmd
	}
	return s, nil
}

func (s *sessionDialog) handleSearchChange() {
	newQuery := s.searchInput.Value()
	if newQuery == s.searchQuery {
		return
	}

	s.searchQuery = newQuery
	s.deleteConfirmation = -1

	if newQuery == "" {
		s.filteredSessions = s.allSessions
	} else {
		// Use fuzzy matching
		titles := make([]string, len(s.allSessions))
		for i, sess := range s.allSessions {
			titles[i] = sess.Title
		}

		matches := fuzzy.Find(newQuery, titles)
		s.filteredSessions = make([]opencode.Session, len(matches))
		for i, match := range matches {
			s.filteredSessions[i] = s.allSessions[match.Index]
		}
	}

	s.updateListItems()
	s.list.SetSelectedIndex(0)
}

func (s *sessionDialog) Render(background string) string {
	if s.renameMode {
		// Show rename input instead of list
		t := theme.CurrentTheme()
		renameView := s.renameInput.View()

		mutedStyle := styles.NewStyle().
			Foreground(t.TextMuted()).
			Background(t.BackgroundPanel()).
			Render
		helpText := mutedStyle("Enter to confirm, Esc to cancel")
		helpText = styles.NewStyle().PaddingLeft(1).PaddingTop(1).Render(helpText)

		content := strings.Join([]string{renameView, helpText}, "\n")
		return s.modal.Render(content, background)
	}

	t := theme.CurrentTheme()

	// Search input
	searchView := s.searchInput.View()
	searchView = styles.NewStyle().PaddingLeft(1).PaddingBottom(1).Render(searchView)

	// List view
	listView := s.list.View()

	// Help text
	keyStyle := styles.NewStyle().
		Foreground(t.Text()).
		Background(t.BackgroundPanel()).
		Bold(true).
		Render
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted()).Background(t.BackgroundPanel()).Render

	var leftHelp, rightHelp string
	if s.searchQuery == "" {
		leftHelp = keyStyle("n") + mutedStyle(" new   ") + keyStyle("r") + mutedStyle(" rename")
		rightHelp = keyStyle("ctrl+d") + mutedStyle(" delete")
	} else {
		leftHelp = mutedStyle("Type to search")
		rightHelp = keyStyle("esc") + mutedStyle(" clear")
	}

	bgColor := t.BackgroundPanel()
	helpText := layout.Render(layout.FlexOptions{
		Direction:  layout.Row,
		Justify:    layout.JustifySpaceBetween,
		Width:      layout.Current.Container.Width - 14,
		Background: &bgColor,
	}, layout.FlexItem{View: leftHelp}, layout.FlexItem{View: rightHelp})

	helpText = styles.NewStyle().PaddingLeft(1).PaddingTop(1).Render(helpText)

	content := strings.Join([]string{searchView, listView, helpText}, "\n")

	return s.modal.Render(content, background)
}

func (s *sessionDialog) setupRenameInput(currentTitle string) {
	t := theme.CurrentTheme()
	bgColor := t.BackgroundPanel()
	textColor := t.Text()
	textMutedColor := t.TextMuted()

	s.renameInput = textinput.New()
	s.renameInput.SetValue(currentTitle)
	s.renameInput.Focus()
	s.renameInput.CharLimit = 100
	s.renameInput.SetWidth(layout.Current.Container.Width - 20)

	s.renameInput.Styles.Blurred.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	s.renameInput.Styles.Blurred.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	s.renameInput.Styles.Focused.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	s.renameInput.Styles.Focused.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	s.renameInput.Styles.Focused.Prompt = styles.NewStyle().
		Background(bgColor).
		Lipgloss()
}

func (s *sessionDialog) setupSearchInput() {
	t := theme.CurrentTheme()
	bgColor := t.BackgroundPanel()
	textColor := t.Text()
	textMutedColor := t.TextMuted()

	s.searchInput = textinput.New()
	s.searchInput.Placeholder = "Search sessions..."
	s.searchInput.Focus()
	s.searchInput.CharLimit = 100
	s.searchInput.SetWidth(layout.Current.Container.Width - 20)

	s.searchInput.Styles.Blurred.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	s.searchInput.Styles.Blurred.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	s.searchInput.Styles.Focused.Placeholder = styles.NewStyle().
		Foreground(textMutedColor).
		Background(bgColor).
		Lipgloss()
	s.searchInput.Styles.Focused.Text = styles.NewStyle().
		Foreground(textColor).
		Background(bgColor).
		Lipgloss()
	s.searchInput.Styles.Focused.Prompt = styles.NewStyle().
		Background(bgColor).
		Lipgloss()
}

func (s *sessionDialog) updateListItems() {
	_, currentIdx := s.list.GetSelectedItem()

	var items []sessionItem
	for i, sess := range s.filteredSessions {
		// Check if session is busy
		isBusy := false
		if status, ok := s.sessionStatus[sess.ID]; ok {
			isBusy = status == "busy"
		}

		// Convert float64 timestamp (milliseconds) to time.Time
		updatedTime := time.UnixMilli(int64(sess.Time.Updated))

		item := sessionItem{
			session:            sess,
			isDeleteConfirming: s.deleteConfirmation == i,
			isCurrentSession:   s.app.Session != nil && s.app.Session.ID == sess.ID,
			isBusy:             isBusy,
			timeAgo:            util.FormatRelativeTime(updatedTime),
		}
		items = append(items, item)
	}
	s.list.SetItems(items)
	if currentIdx >= 0 && currentIdx < len(items) {
		s.list.SetSelectedIndex(currentIdx)
	}
}

func (s *sessionDialog) deleteSession(sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := s.app.DeleteSession(ctx, sessionID); err != nil {
			return toast.NewErrorToast("Failed to delete session: " + err.Error())()
		}
		return nil
	}
}

// ReopenSessionModalMsg is emitted when the session modal should be reopened
type ReopenSessionModalMsg struct{}

func (s *sessionDialog) Close() tea.Cmd {
	if s.renameMode {
		// If in rename mode, exit rename mode and return a command to reopen the modal
		s.renameMode = false
		s.modal.SetTitle("Switch Session")
		s.updateListItems()

		// Return a command that will reopen the session modal
		return func() tea.Msg {
			return ReopenSessionModalMsg{}
		}
	}
	// Normal close behavior
	return nil
}

// NewSessionDialog creates a new session switching dialog
func NewSessionDialog(app *app.App) SessionDialog {
	sessions, _ := app.ListSessions(context.Background())

	// Filter out child sessions and sort by updated time (newest first)
	var filteredSessions []opencode.Session
	for _, sess := range sessions {
		if sess.ParentID != "" {
			continue
		}
		filteredSessions = append(filteredSessions, sess)
	}

	// Sort by updated time (newest first)
	sort.Slice(filteredSessions, func(i, j int) bool {
		return filteredSessions[i].Time.Updated > filteredSessions[j].Time.Updated
	})

	var items []sessionItem
	for _, sess := range filteredSessions {
		// Convert float64 timestamp (milliseconds) to time.Time
		updatedTime := time.UnixMilli(int64(sess.Time.Updated))
		items = append(items, sessionItem{
			session:            sess,
			isDeleteConfirming: false,
			isCurrentSession:   app.Session != nil && app.Session.ID == sess.ID,
			isBusy:             false,
			timeAgo:            util.FormatRelativeTime(updatedTime),
		})
	}

	listComponent := list.NewListComponent(
		list.WithItems(items),
		list.WithMaxVisibleHeight[sessionItem](10),
		list.WithFallbackMessage[sessionItem]("No sessions found"),
		list.WithAlphaNumericKeys[sessionItem](false), // Disable alphanumeric keys for search
		list.WithRenderFunc(
			func(item sessionItem, selected bool, width int, baseStyle styles.Style) string {
				return item.Render(selected, width, false, baseStyle)
			},
		),
		list.WithSelectableFunc(func(item sessionItem) bool {
			return true
		}),
	)
	listComponent.SetMaxWidth(layout.Current.Container.Width - 12)

	dialog := &sessionDialog{
		allSessions:        filteredSessions,
		filteredSessions:   filteredSessions,
		list:               listComponent,
		app:                app,
		deleteConfirmation: -1,
		renameMode:         false,
		renameIndex:        -1,
		sessionStatus:      make(map[string]string),
		modal: modal.New(
			modal.WithTitle("Switch Session"),
			modal.WithMaxWidth(layout.Current.Container.Width-8),
		),
	}

	dialog.setupSearchInput()

	return dialog
}
