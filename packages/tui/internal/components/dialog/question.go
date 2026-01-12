package dialog

import (
	"context"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/components/toast"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
)

// QuestionDialog interface for the question prompt dialog
type QuestionDialog interface {
	layout.Modal
}

// QuestionDismissedMsg is sent when the question dialog is dismissed
type QuestionDismissedMsg struct{}

type questionDialog struct {
	app          *app.App
	modal        *modal.Modal
	width        int
	height       int
	request      app.QuestionRequest
	tab          int        // Current question tab (or confirm tab)
	selected     int        // Currently selected option
	answers      [][]string // Answers for each question
	customInputs []string   // Custom "other" inputs for each question
	editing      bool       // Whether we're editing a custom input
	customInput  string     // Current custom input text
}

func (q *questionDialog) Init() tea.Cmd {
	return nil
}

func (q *questionDialog) questions() []app.QuestionInfo {
	return q.request.Questions
}

func (q *questionDialog) single() bool {
	return len(q.questions()) == 1 && !q.questions()[0].Multiple
}

func (q *questionDialog) tabs() int {
	if q.single() {
		return 1
	}
	return len(q.questions()) + 1 // questions + confirm tab
}

func (q *questionDialog) question() *app.QuestionInfo {
	if q.tab < len(q.questions()) {
		return &q.questions()[q.tab]
	}
	return nil
}

func (q *questionDialog) confirm() bool {
	return !q.single() && q.tab == len(q.questions())
}

func (q *questionDialog) options() []app.QuestionOption {
	if question := q.question(); question != nil {
		return question.Options
	}
	return nil
}

func (q *questionDialog) other() bool {
	return q.selected == len(q.options())
}

func (q *questionDialog) multi() bool {
	if question := q.question(); question != nil {
		return question.Multiple
	}
	return false
}

func (q *questionDialog) input() string {
	if q.tab < len(q.customInputs) {
		return q.customInputs[q.tab]
	}
	return ""
}

func (q *questionDialog) customPicked() bool {
	value := q.input()
	if value == "" {
		return false
	}
	if q.tab < len(q.answers) {
		return slices.Contains(q.answers[q.tab], value)
	}
	return false
}

func (q *questionDialog) isPicked(label string) bool {
	return q.tab < len(q.answers) && slices.Contains(q.answers[q.tab], label)
}

func (q *questionDialog) submit() tea.Cmd {
	answers := make([][]string, len(q.questions()))
	for i := range q.questions() {
		if i < len(q.answers) {
			answers[i] = q.answers[i]
		} else {
			answers[i] = []string{}
		}
	}

	return func() tea.Msg {
		ctx := context.Background()
		body := map[string]any{
			"answers": answers,
		}
		err := q.app.Client.Post(ctx, "/question/"+q.request.ID+"/reply", body, nil)
		if err != nil {
			return tea.Sequence(
				toast.NewErrorToast("Failed to submit answer: "+err.Error()),
				func() tea.Msg { return QuestionDismissedMsg{} },
			)()
		}
		return QuestionDismissedMsg{}
	}
}

func (q *questionDialog) reject() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := q.app.Client.Post(ctx, "/question/"+q.request.ID+"/reject", nil, nil)
		if err != nil {
			return tea.Sequence(
				toast.NewErrorToast("Failed to reject question: "+err.Error()),
				func() tea.Msg { return QuestionDismissedMsg{} },
			)()
		}
		return QuestionDismissedMsg{}
	}
}

func (q *questionDialog) pick(answer string, custom bool) {
	for len(q.answers) <= q.tab {
		q.answers = append(q.answers, []string{})
	}
	q.answers[q.tab] = []string{answer}

	if custom {
		for len(q.customInputs) <= q.tab {
			q.customInputs = append(q.customInputs, "")
		}
		q.customInputs[q.tab] = answer
	}
}

func (q *questionDialog) toggle(answer string) {
	for len(q.answers) <= q.tab {
		q.answers = append(q.answers, []string{})
	}

	existing := q.answers[q.tab]
	if slices.Contains(existing, answer) {
		q.answers[q.tab] = slices.DeleteFunc(existing, func(s string) bool {
			return s == answer
		})
	} else {
		q.answers[q.tab] = append(existing, answer)
	}
}

func (q *questionDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		q.width = msg.Width
		q.height = msg.Height

	case tea.KeyPressMsg:
		keyStr := msg.String()

		// When editing custom input
		if q.editing && !q.confirm() {
			if keyStr == "escape" {
				q.editing = false
				return q, nil
			}
			if keyStr == "enter" {
				text := strings.TrimSpace(q.customInput)
				prev := q.input()

				if text == "" {
					// Clear custom input
					if prev != "" {
						for len(q.customInputs) <= q.tab {
							q.customInputs = append(q.customInputs, "")
						}
						q.customInputs[q.tab] = ""
					}
					// Remove from answers
					for len(q.answers) <= q.tab {
						q.answers = append(q.answers, []string{})
					}
					if prev != "" {
						newAnswers := []string{}
						for _, ans := range q.answers[q.tab] {
							if ans != prev {
								newAnswers = append(newAnswers, ans)
							}
						}
						q.answers[q.tab] = newAnswers
					}
					q.editing = false
					return q, nil
				}

				if q.multi() {
					for len(q.customInputs) <= q.tab {
						q.customInputs = append(q.customInputs, "")
					}
					q.customInputs[q.tab] = text

					for len(q.answers) <= q.tab {
						q.answers = append(q.answers, []string{})
					}
					existing := q.answers[q.tab]
					// Remove previous custom value if exists
					if prev != "" {
						existing = slices.DeleteFunc(existing, func(ans string) bool {
							return ans == prev
						})
					}
					// Add new value if not already present
					if !slices.Contains(existing, text) {
						existing = append(existing, text)
					}
					q.answers[q.tab] = existing
					q.editing = false
					return q, nil
				}

				// Single select - pick and move to next
				q.pick(text, true)
				q.editing = false
				if q.single() {
					return q, q.submit()
				}
				q.tab++
				q.selected = 0
				return q, nil
			}

			// Handle text input
			if keyStr == "backspace" {
				if len(q.customInput) > 0 {
					q.customInput = q.customInput[:len(q.customInput)-1]
				}
				return q, nil
			}
			if len(keyStr) == 1 {
				q.customInput += keyStr
				return q, nil
			}
			return q, nil
		}

		// Tab navigation
		if keyStr == "left" || keyStr == "h" {
			q.tab = (q.tab - 1 + q.tabs()) % q.tabs()
			q.selected = 0
			return q, nil
		}
		if keyStr == "right" || keyStr == "l" {
			q.tab = (q.tab + 1) % q.tabs()
			q.selected = 0
			return q, nil
		}

		if q.confirm() {
			if keyStr == "enter" {
				return q, q.submit()
			}
			if keyStr == "escape" || keyStr == "ctrl+c" {
				return q, q.reject()
			}
			return q, nil
		}

		// Option navigation
		opts := q.options()
		total := len(opts) + 1 // options + "Other"

		if keyStr == "up" || keyStr == "k" {
			q.selected = (q.selected - 1 + total) % total
			return q, nil
		}
		if keyStr == "down" || keyStr == "j" {
			q.selected = (q.selected + 1) % total
			return q, nil
		}

		if keyStr == "enter" {
			if q.other() {
				if !q.multi() {
					q.editing = true
					q.customInput = q.input()
					return q, nil
				}
				value := q.input()
				if value != "" && q.customPicked() {
					q.toggle(value)
					return q, nil
				}
				q.editing = true
				q.customInput = q.input()
				return q, nil
			}

			opt := opts[q.selected]
			if q.multi() {
				q.toggle(opt.Label)
				return q, nil
			}

			q.pick(opt.Label, false)
			if q.single() {
				return q, q.submit()
			}
			q.tab++
			q.selected = 0
			return q, nil
		}

		if keyStr == "escape" || keyStr == "ctrl+c" {
			return q, q.reject()
		}
	}

	return q, nil
}

func (q *questionDialog) View() string {
	return q.Render("")
}

func (q *questionDialog) Render(background string) string {
	t := theme.CurrentTheme()

	var content strings.Builder

	// Tab bar (only for multi-question)
	if !q.single() {
		content.WriteString(q.renderTabs())
		content.WriteString("\n\n")
	}

	if !q.confirm() {
		content.WriteString(q.renderQuestion())
	} else {
		content.WriteString(q.renderConfirm())
	}

	content.WriteString("\n")
	content.WriteString(q.renderHelp())

	contentStyle := styles.NewStyle().
		Background(t.BackgroundPanel()).
		Padding(1).
		Width(q.calculateWidth())

	return q.modal.Render(contentStyle.Render(content.String()), background)
}

func (q *questionDialog) renderTabs() string {
	t := theme.CurrentTheme()
	var tabs strings.Builder

	for i, question := range q.questions() {
		isActive := i == q.tab
		isAnswered := i < len(q.answers) && len(q.answers[i]) > 0

		tabStyle := styles.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

		if isActive {
			tabStyle = tabStyle.Background(t.Primary()).Foreground(t.BackgroundElement())
		} else {
			tabStyle = tabStyle.Background(t.BackgroundElement())
			if isAnswered {
				tabStyle = tabStyle.Foreground(t.Text())
			} else {
				tabStyle = tabStyle.Foreground(t.TextMuted())
			}
		}

		tabs.WriteString(tabStyle.Render(question.Header))
		tabs.WriteString(" ")
	}

	// Confirm tab
	confirmActive := q.confirm()
	confirmStyle := styles.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)
	if confirmActive {
		confirmStyle = confirmStyle.Background(t.Primary()).Foreground(t.BackgroundElement())
	} else {
		confirmStyle = confirmStyle.Background(t.BackgroundElement()).Foreground(t.TextMuted())
	}
	tabs.WriteString(confirmStyle.Render("Confirm"))

	return tabs.String()
}

func (q *questionDialog) renderQuestion() string {
	t := theme.CurrentTheme()
	var content strings.Builder

	question := q.question()
	if question == nil {
		return ""
	}

	// Question text
	questionText := question.Question
	if q.multi() {
		questionText += " (select all that apply)"
	}
	content.WriteString(styles.NewStyle().Foreground(t.Text()).Render(questionText))
	content.WriteString("\n\n")

	// Options
	for i, opt := range q.options() {
		active := i == q.selected
		picked := q.isPicked(opt.Label)

		optStyle := styles.NewStyle()
		if active {
			optStyle = optStyle.Background(t.BackgroundElement())
			if picked {
				optStyle = optStyle.Foreground(t.Success())
			} else {
				optStyle = optStyle.Foreground(t.Secondary())
			}
		} else {
			if picked {
				optStyle = optStyle.Foreground(t.Success())
			} else {
				optStyle = optStyle.Foreground(t.Text())
			}
		}

		checkMark := ""
		if picked {
			checkMark = " ✓"
		}

		content.WriteString(optStyle.Render(string(rune('1'+i)) + ". " + opt.Label))
		if checkMark != "" {
			content.WriteString(styles.NewStyle().Foreground(t.Success()).Render(checkMark))
		}
		content.WriteString("\n")

		// Description
		if opt.Description != "" {
			content.WriteString(styles.NewStyle().Foreground(t.TextMuted()).PaddingLeft(3).Render(opt.Description))
			content.WriteString("\n")
		}
	}

	// "Other" option
	otherActive := q.other()
	otherStyle := styles.NewStyle()
	if otherActive {
		otherStyle = otherStyle.Background(t.BackgroundElement())
		if q.customPicked() {
			otherStyle = otherStyle.Foreground(t.Success())
		} else {
			otherStyle = otherStyle.Foreground(t.Secondary())
		}
	} else {
		if q.customPicked() {
			otherStyle = otherStyle.Foreground(t.Success())
		} else {
			otherStyle = otherStyle.Foreground(t.Text())
		}
	}

	checkMark := ""
	if q.customPicked() {
		checkMark = " ✓"
	}

	content.WriteString(otherStyle.Render(string(rune('1'+len(q.options()))) + ". Type your own answer"))
	if checkMark != "" {
		content.WriteString(styles.NewStyle().Foreground(t.Success()).Render(checkMark))
	}
	content.WriteString("\n")

	// Show custom input if editing or has value
	if q.editing {
		content.WriteString(styles.NewStyle().Foreground(t.Text()).PaddingLeft(3).Render(q.customInput + "█"))
		content.WriteString("\n")
	} else if q.input() != "" {
		content.WriteString(styles.NewStyle().Foreground(t.TextMuted()).PaddingLeft(3).Render(q.input()))
		content.WriteString("\n")
	}

	return content.String()
}

func (q *questionDialog) renderConfirm() string {
	t := theme.CurrentTheme()
	var content strings.Builder

	content.WriteString(styles.NewStyle().Foreground(t.Text()).Render("Review"))
	content.WriteString("\n\n")

	for i, question := range q.questions() {
		var value string
		if i < len(q.answers) && len(q.answers[i]) > 0 {
			value = strings.Join(q.answers[i], ", ")
		}

		content.WriteString(styles.NewStyle().Foreground(t.TextMuted()).Render(question.Header + ": "))
		if value != "" {
			content.WriteString(styles.NewStyle().Foreground(t.Text()).Render(value))
		} else {
			content.WriteString(styles.NewStyle().Foreground(t.Error()).Render("(not answered)"))
		}
		content.WriteString("\n")
	}

	return content.String()
}

func (q *questionDialog) renderHelp() string {
	t := theme.CurrentTheme()
	keyStyle := styles.NewStyle().Foreground(t.Text()).Bold(true).Render
	mutedStyle := styles.NewStyle().Foreground(t.TextMuted()).Render

	var help strings.Builder

	if !q.single() {
		help.WriteString(keyStyle("⇆"))
		help.WriteString(mutedStyle(" tab  "))
	}

	if !q.confirm() {
		help.WriteString(keyStyle("↑↓"))
		help.WriteString(mutedStyle(" select  "))
	}

	if q.confirm() {
		help.WriteString(keyStyle("enter"))
		help.WriteString(mutedStyle(" submit  "))
	} else if q.multi() {
		help.WriteString(keyStyle("enter"))
		help.WriteString(mutedStyle(" toggle  "))
	} else if q.single() {
		help.WriteString(keyStyle("enter"))
		help.WriteString(mutedStyle(" submit  "))
	} else {
		help.WriteString(keyStyle("enter"))
		help.WriteString(mutedStyle(" confirm  "))
	}

	help.WriteString(keyStyle("esc"))
	help.WriteString(mutedStyle(" dismiss"))

	return help.String()
}

func (q *questionDialog) Close() tea.Cmd {
	return nil
}

func (q *questionDialog) calculateWidth() int {
	return min(60, layout.Current.Container.Width-8)
}

// NewQuestionDialog creates a new question dialog
func NewQuestionDialog(app *app.App, request app.QuestionRequest) QuestionDialog {
	return &questionDialog{
		app:          app,
		request:      request,
		tab:          0,
		selected:     0,
		answers:      make([][]string, len(request.Questions)),
		customInputs: make([]string, len(request.Questions)),
		editing:      false,
		modal: modal.New(
			modal.WithTitle("Question"),
			modal.WithMaxWidth(64),
		),
	}
}
