package completions

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
)

type filesContextGroup struct {
	app      *app.App
	gitFiles []CompletionSuggestion
}

func (cg *filesContextGroup) GetId() string {
	return "files"
}

func (cg *filesContextGroup) GetEmptyMessage() string {
	return "no matching files"
}

func (cg *filesContextGroup) getGitFiles() []CompletionSuggestion {
	items := make([]CompletionSuggestion, 0)

	status, _ := cg.app.Client.File.Status(context.Background(), opencode.FileStatusParams{})
	if status != nil {
		files := *status
		sort.Slice(files, func(i, j int) bool {
			return files[i].Added+files[i].Removed > files[j].Added+files[j].Removed
		})

		for _, file := range files {
			displayFunc := func(s styles.Style) string {
				t := theme.CurrentTheme()
				green := s.Foreground(t.Success()).Render
				red := s.Foreground(t.Error()).Render
				display := file.Path
				if file.Added > 0 {
					display += green(" +" + strconv.Itoa(int(file.Added)))
				}
				if file.Removed > 0 {
					display += red(" -" + strconv.Itoa(int(file.Removed)))
				}
				return display
			}
			item := CompletionSuggestion{
				Display:    displayFunc,
				Value:      file.Path,
				ProviderID: cg.GetId(),
				RawData:    file,
			}
			items = append(items, item)
		}
	}

	return items
}

func (cg *filesContextGroup) GetChildEntries(
	query string,
) ([]CompletionSuggestion, error) {
	items := make([]CompletionSuggestion, 0)

	query = strings.TrimSpace(query)
	if query == "" {
		items = append(items, cg.gitFiles...)
	}

	files, err := cg.app.Client.Find.Files(
		context.Background(),
		opencode.FindFilesParams{Query: opencode.F(query)},
	)
	if err != nil {
		slog.Error("Failed to get completion items", "error", err)
		return items, err
	}
	if files == nil {
		return items, nil
	}

	for _, file := range *files {
		exists := false
		for _, existing := range cg.gitFiles {
			if existing.Value == file {
				if query != "" {
					items = append(items, existing)
				}
				exists = true
			}
		}
		if !exists {
			displayFunc := func(s styles.Style) string {
				// Hack: Strip "p" prefix if it's an artifact
				clean := strings.TrimPrefix(file, "p")
				icon := getFileIcon(clean)
				return s.Render(icon + " " + clean)
			}

			item := CompletionSuggestion{
				Display:    displayFunc,
				Value:      file,
				ProviderID: cg.GetId(),
				RawData:    file,
			}
			items = append(items, item)
		}
	}

	return items, nil
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

func NewFileContextGroup(app *app.App) CompletionProvider {
	cg := &filesContextGroup{
		app: app,
	}
	go func() {
		cg.gitFiles = cg.getGitFiles()
	}()
	return cg
}
