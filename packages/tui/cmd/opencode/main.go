package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/option"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/tui"
)

func main() {
	// Simple slog setup for stderr
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	// 1. Initialize Client
	server := os.Getenv("OPENCODE_SERVER")
	var opts []option.RequestOption
	if server != "" {
		opts = append(opts, option.WithBaseURL(server))
	}
	client := opencode.NewClient(opts...)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	// 2. Resolve Paths
	pathInfo, err := client.Path.Get(ctx, opencode.PathGetParams{
		Directory: opencode.F(cwd),
	})
	if err != nil {
		fmt.Printf("Error resolving paths: %v\n", err)
		os.Exit(1)
	}

	// 3. Resolve Project
	project, err := client.Project.Current(ctx, opencode.ProjectCurrentParams{
		Directory: opencode.F(cwd),
	})
	if err != nil {
		slog.Error("Failed to get current project", "error", err)
		project = &opencode.Project{
			ID: "unknown",
		}
	}

	// 4. Fetch Agents (Required by app.New to avoid panic)
	agents, err := client.Agent.List(ctx, opencode.AgentListParams{
		Directory: opencode.F(cwd),
	})
	if err != nil {
		// Log but try to proceed with nil? No, app.New panics on empty agents.
		// We'll create a dummy agent if list fails/empty to prevent crash in New.
		slog.Error("Failed to list agents", "error", err)
	}
	// Fallback dummy if empty to prevent startup panic
	if len(agents) == 0 {
		agents = []opencode.Agent{
			{Name: "default", Mode: "chat", Description: "Fallback Agent"},
		}
	}

	// 5. Initialize App State
	// app.New(ctx, version, project, path, agents, client, initModel, initPrompt, initAgent, initSession)
	application, err := app.New(
		ctx,
		"0.0.1",  // version
		project,  // project
		pathInfo, // path
		agents,   // agents
		client,   // httpClient
		nil,      // initialModel
		nil,      // initialPrompt
		nil,      // initialAgent
		nil,      // initialSession
	)
	if err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// 6. Initialize TUI Model
	model := tui.NewModel(application)

	// 7. Run Bubble Tea Program (Standard AltScreen Mode)
	p := tea.NewProgram(
		model,
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
