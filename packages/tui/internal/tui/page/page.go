package page

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

// PageID is a unique identifier for a page.
type PageID string

// Page defines the interface for a UI page.
type Page interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetSize(width, height int) tea.Cmd
}
