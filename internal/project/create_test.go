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

	state := readGeneratedFile(t, root, ".factory/FACTORY_STATE.md")
	if !strings.Contains(state, "phase: DISCOVERY") {
		t.Fatalf("expected DISCOVERY phase, got:\n%s", state)
	}
}

func TestCreateGeneratesAdvancedAgentContracts(t *testing.T) {
	parent := t.TempDir()
	root, err := Create(Config{Name: "demo", Architect: "chatgpt", Executor: "opencode", ParentDir: parent})
	if err != nil {
		t.Fatal(err)
	}

	agents := readGeneratedFile(t, root, "AGENTS.md")
	for _, expected := range []string{
		"## Sources of truth",
		"## Authority boundary",
		"## Execution cycle",
		"ARCHITECT_DECISION_REQUIRED",
		"PREFLIGHT",
		"deterministic final verification",
		"Never weaken, delete, skip, or rewrite tests",
		"## Completion conditions",
		"## Git discipline",
	} {
		if !strings.Contains(agents, expected) {
			t.Errorf("AGENTS.md missing contract marker %q", expected)
		}
	}

	tester := readGeneratedFile(t, root, ".opencode/agents/tester.md")
	for _, expected := range []string{
		"### PREFLIGHT",
		"### FINAL",
		"## Final report contract",
		"BLOCKED_ENVIRONMENT",
		"### COMMANDS_AND_RESULTS",
		"### REGRESSION_RISK",
	} {
		if !strings.Contains(tester, expected) {
			t.Errorf("tester.md missing contract marker %q", expected)
		}
	}

	reviewer := readGeneratedFile(t, root, ".opencode/agents/reviewer.md")
	for _, expected := range []string{
		"## Review dimensions",
		"### Architecture compliance",
		"## Verdict contract",
		"FINDINGS",
		"ARCHITECT_DECISION_REQUIRED",
		"### VERIFICATION_ASSESSMENT",
	} {
		if !strings.Contains(reviewer, expected) {
			t.Errorf("reviewer.md missing contract marker %q", expected)
		}
	}
}

func TestCreateRejectsInvalidName(t *testing.T) {
	_, err := Create(Config{Name: "../escape", ParentDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected invalid project name to fail")
	}
}

func readGeneratedFile(t *testing.T, root, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
