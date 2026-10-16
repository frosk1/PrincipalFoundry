package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// designDocumentMap maps the required top-level design-package documents to the
// repository paths they become after bootstrap.
var designDocumentMap = map[string]string{
	"PROJECT_BRIEF.md":  ".factory/PROJECT_BRIEF.md",
	"PRODUCT_SPEC.md":   "docs/PRODUCT_SPEC.md",
	"ARCHITECTURE.md":   "docs/ARCHITECTURE.md",
	"OPEN_QUESTIONS.md": "docs/OPEN_QUESTIONS.md",
}

// InitConfig describes a repository bootstrap from an approved design package.
type InitConfig struct {
	Name      string
	Architect string
	Executor  string
	ParentDir string
	// DesignDir is the approved design package produced by the discovery phase.
	DesignDir string
	// TaskFile, when set, is installed as the active .factory/CURRENT_TASK.md.
	TaskFile string
}

// CreateFromDesign creates a project repository from an approved discovery
// design package. It validates the package, generates the repository skeleton,
// installs the approved product and architecture documents, records the project
// brief, and optionally activates the first implementation task.
func CreateFromDesign(cfg InitConfig) (string, error) {
	designDir, err := resolveDesignDir(cfg.DesignDir)
	if err != nil {
		return "", err
	}

	designFiles, err := loadDesignPackage(designDir)
	if err != nil {
		return "", err
	}

	taskPath := ""
	if strings.TrimSpace(cfg.TaskFile) != "" {
		if taskPath, err = resolveTaskFile(cfg.TaskFile); err != nil {
			return "", err
		}
	}

	root, files, err := prepare(Config{
		Name:      cfg.Name,
		Architect: cfg.Architect,
		Executor:  cfg.Executor,
		ParentDir: cfg.ParentDir,
	})
	if err != nil {
		return "", err
	}

	// Replace the placeholder skeleton documents with the approved design.
	for dest, content := range designFiles {
		files[dest] = content
	}

	// Bootstrap is complete; advance the phase to SCAFFOLDING only when the
	// first implementation task is activated.
	phase := "REPOSITORY_BOOTSTRAP"
	if taskPath != "" {
		taskContent, err := readTextFile(taskPath)
		if err != nil {
			return "", fmt.Errorf("cannot read task file %s: %w", cfg.TaskFile, err)
		}
		files[".factory/CURRENT_TASK.md"] = readyTaskContent(taskContent)
		phase = "SCAFFOLDING"
	}
	files[".factory/FACTORY_STATE.md"] = factoryStateContent(phase, cfg.Architect, cfg.Executor)

	if err := writeFiles(root, files); err != nil {
		return "", err
	}
	return absolutePath(root), nil
}

// loadDesignPackage validates an approved design package and returns the
// repository-relative destination path for each installed file keyed by content.
func loadDesignPackage(dir string) (map[string]string, error) {
	files := map[string]string{}

	for src, dest := range designDocumentMap {
		content, err := readTextFile(filepath.Join(dir, filepath.FromSlash(src)))
		if err != nil {
			return nil, fmt.Errorf("design package is missing %s: %w", src, err)
		}
		files[dest] = content
	}

	decisionsDir := filepath.Join(dir, "decisions")
	entries, err := os.ReadDir(decisionsDir)
	if err != nil {
		return nil, fmt.Errorf("design package is missing decisions/ directory: %w", err)
	}

	foundIndex := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := readTextFile(filepath.Join(decisionsDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if entry.Name() == "INDEX.md" {
			foundIndex = true
		}
		files["docs/decisions/"+entry.Name()] = content
	}
	if !foundIndex {
		return nil, errors.New("design package is missing decisions/INDEX.md")
	}

	return files, nil
}

func resolveDesignDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", errors.New("design package directory is required (--from)")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("design package not found: %s", dir)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("design package is not a directory: %s", dir)
	}
	return abs, nil
}

func resolveTaskFile(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("task file not found: %s", path)
	}
	if info.IsDir() {
		return "", fmt.Errorf("task file is a directory: %s", path)
	}
	return abs, nil
}

func readTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// readyTaskContent turns a task document into an active CURRENT_TASK.md by
// adding the READY status frontmatter the executor contract expects.
func readyTaskContent(content string) string {
	body := strings.TrimRight(content, "\n")
	return "---\nstatus: READY\n---\n\n" + body + "\n"
}
