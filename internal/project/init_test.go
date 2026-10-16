package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeDesignFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newDesignPackage writes a minimal, valid design package and returns its path.
func newDesignPackage(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeDesignFile(t, dir, "PROJECT_BRIEF.md", "# Brief\n\nApproved brief.\n")
	writeDesignFile(t, dir, "PRODUCT_SPEC.md", "# Product\n\nApproved product.\n")
	writeDesignFile(t, dir, "ARCHITECTURE.md", "# Architecture\n\nApproved architecture.\n")
	writeDesignFile(t, dir, "OPEN_QUESTIONS.md", "# Open Questions\n\nOQ-01.\n")
	writeDesignFile(t, dir, "decisions/INDEX.md", "# ADR Index\n\n- ADR-0001\n")
	writeDesignFile(t, dir, "decisions/ADR-0001-example.md", "# ADR-0001\n\nDecision.\n")
	return dir
}

func TestCreateFromDesignInstallsApprovedDesign(t *testing.T) {
	design := newDesignPackage(t)
	parent := t.TempDir()

	root, err := CreateFromDesign(InitConfig{
		Name:      "dockspur",
		Architect: "chatgpt",
		Executor:  "opencode",
		ParentDir: parent,
		DesignDir: design,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The repository skeleton is still generated.
	for _, path := range []string{
		"AGENTS.md",
		".opencode/agents/tester.md",
		".opencode/agents/reviewer.md",
		".factory/FACTORY_STATE.md",
		".factory/PROJECT_BRIEF.md",
		".factory/CURRENT_TASK.md",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected skeleton file %s: %v", path, err)
		}
	}

	// Approved documents replace the placeholders.
	if got := readGeneratedFile(t, root, "docs/PRODUCT_SPEC.md"); !strings.Contains(got, "Approved product") {
		t.Errorf("PRODUCT_SPEC not installed from design package:\n%s", got)
	}
	if got := readGeneratedFile(t, root, "docs/ARCHITECTURE.md"); !strings.Contains(got, "Approved architecture") {
		t.Errorf("ARCHITECTURE not installed from design package:\n%s", got)
	}
	if got := readGeneratedFile(t, root, "docs/OPEN_QUESTIONS.md"); !strings.Contains(got, "OQ-01") {
		t.Errorf("OPEN_QUESTIONS not installed from design package:\n%s", got)
	}
	if got := readGeneratedFile(t, root, ".factory/PROJECT_BRIEF.md"); !strings.Contains(got, "Approved brief") {
		t.Errorf("PROJECT_BRIEF not installed from design package:\n%s", got)
	}
	if got := readGeneratedFile(t, root, "docs/decisions/INDEX.md"); !strings.Contains(got, "ADR-0001") {
		t.Errorf("decisions/INDEX.md not installed:\n%s", got)
	}
	if got := readGeneratedFile(t, root, "docs/decisions/ADR-0001-example.md"); !strings.Contains(got, "Decision") {
		t.Errorf("ADR not installed:\n%s", got)
	}

	// Without --task the factory waits for the architect.
	state := readGeneratedFile(t, root, ".factory/FACTORY_STATE.md")
	if !strings.Contains(state, "phase: REPOSITORY_BOOTSTRAP") {
		t.Errorf("expected REPOSITORY_BOOTSTRAP phase, got:\n%s", state)
	}
	current := readGeneratedFile(t, root, ".factory/CURRENT_TASK.md")
	if !strings.Contains(current, "status: IDLE") {
		t.Errorf("expected IDLE current task without --task, got:\n%s", current)
	}
}

func TestCreateFromDesignActivatesFirstTask(t *testing.T) {
	design := newDesignPackage(t)
	task := filepath.Join(design, "implementation", "TASK-0001-contract.md")
	writeDesignFile(t, design, "implementation/TASK-0001-contract.md", "# TASK-0001\n\nDefine the contract.\n")

	root, err := CreateFromDesign(InitConfig{
		Name:      "dockspur",
		Architect: "chatgpt",
		Executor:  "opencode",
		ParentDir: t.TempDir(),
		DesignDir: design,
		TaskFile:  task,
	})
	if err != nil {
		t.Fatal(err)
	}

	current := readGeneratedFile(t, root, ".factory/CURRENT_TASK.md")
	if !strings.HasPrefix(current, "---\nstatus: READY\n---\n") {
		t.Errorf("expected READY frontmatter, got:\n%s", current)
	}
	if !strings.Contains(current, "Define the contract.") {
		t.Errorf("expected task body in CURRENT_TASK, got:\n%s", current)
	}

	state := readGeneratedFile(t, root, ".factory/FACTORY_STATE.md")
	if !strings.Contains(state, "phase: SCAFFOLDING") {
		t.Errorf("expected SCAFFOLDING phase, got:\n%s", state)
	}
}

func TestCreateFromDesignRejectsIncompletePackage(t *testing.T) {
	tests := []struct {
		name   string
		remove string
	}{
		{name: "missing brief", remove: "PROJECT_BRIEF.md"},
		{name: "missing product spec", remove: "PRODUCT_SPEC.md"},
		{name: "missing architecture", remove: "ARCHITECTURE.md"},
		{name: "missing open questions", remove: "OPEN_QUESTIONS.md"},
		{name: "missing decision index", remove: "decisions/INDEX.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			design := newDesignPackage(t)
			if err := os.Remove(filepath.Join(design, filepath.FromSlash(tt.remove))); err != nil {
				t.Fatal(err)
			}
			_, err := CreateFromDesign(InitConfig{
				Name:      "demo",
				ParentDir: t.TempDir(),
				DesignDir: design,
			})
			if err == nil {
				t.Fatalf("expected error when %s is missing", tt.remove)
			}
		})
	}
}

func TestCreateFromDesignRejectsMissingDesignDir(t *testing.T) {
	_, err := CreateFromDesign(InitConfig{
		Name:      "demo",
		ParentDir: t.TempDir(),
		DesignDir: filepath.Join(t.TempDir(), "does-not-exist"),
	})
	if err == nil {
		t.Fatal("expected missing design package to fail")
	}
	if !strings.Contains(err.Error(), "design package not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateFromDesignRejectsExistingDestination(t *testing.T) {
	design := newDesignPackage(t)
	parent := t.TempDir()
	if err := os.MkdirAll(filepath.Join(parent, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := CreateFromDesign(InitConfig{
		Name:      "demo",
		ParentDir: parent,
		DesignDir: design,
	})
	if err == nil {
		t.Fatal("expected existing destination to fail")
	}
}
