package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateGeneratesDiscoveryProject(t *testing.T) {
	parent := t.TempDir()
	root, err := Create(Config{Name: "demo", Architect: "chatgpt", Executor: "opencode", ParentDir: parent})
	if err != nil {
		t.Fatal(err)
	}

	required := []string{
		"AGENTS.md",
		"docs/PRODUCT_SPEC.md",
		"docs/ARCHITECTURE.md",
		".factory/FACTORY_STATE.md",
		".factory/CURRENT_TASK.md",
		".opencode/agents/tester.md",
		".opencode/agents/reviewer.md",
	}
	for _, path := range required {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}

	state, err := os.ReadFile(filepath.Join(root, ".factory/FACTORY_STATE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(state), "phase: DISCOVERY") {
		t.Fatalf("expected DISCOVERY phase, got:\n%s", state)
	}
}

func TestCreateRejectsInvalidName(t *testing.T) {
	_, err := Create(Config{Name: "../escape", ParentDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected invalid project name to fail")
	}
}
