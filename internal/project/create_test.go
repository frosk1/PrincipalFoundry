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
		"docs/OPEN_QUESTIONS.md",
		".factory/FACTORY_STATE.md",
		".factory/MILESTONE.md",
		".factory/CURRENT_TASK.md",
		".opencode/agents/tester.md",
		".opencode/agents/reviewer.md",
	}
	for _, path := range required {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}

	milestone := readGeneratedFile(t, root, ".factory/MILESTONE.md")
	if !strings.Contains(milestone, "status: UNSET") {
		t.Fatalf("expected an unset milestone, got:\n%s", milestone)
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
		"docs/OPEN_QUESTIONS.md",
		"REPOSITORY_BOOTSTRAP",
		"## Authority boundary",
		"## Execution cycle",
		"## Finding-driven repair and evidence",
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
		"## Critical invariants and falsifiability",
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
		"## Independent counterexample review",
	} {
		if !strings.Contains(reviewer, expected) {
			t.Errorf("reviewer.md missing contract marker %q", expected)
		}
	}
}

func TestCreateGeneratesCostAwareVerificationGuidance(t *testing.T) {
	cases := []struct {
		executor   string
		testerPath string
		reviewPath string
	}{
		{"opencode", ".opencode/agents/tester.md", ".opencode/agents/reviewer.md"},
		{"claude-code", ".claude/agents/tester.md", ".claude/agents/reviewer.md"},
		{"codex", ".codex/agents/tester.toml", ".codex/agents/reviewer.toml"},
	}
	for _, tc := range cases {
		t.Run(tc.executor, func(t *testing.T) {
			root, err := Create(Config{Name: "demo", Architect: "chatgpt", Executor: tc.executor, ParentDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			agents := readGeneratedFile(t, root, "AGENTS.md")
			for _, marker := range []string{
				"## Efficient verification without weaker gates",
				"an approved repository contract explicitly establishes its equivalence",
				"prove exactly-once assignment",
				"Do not shorten production deadlines",
				"every gate required by .factory/CURRENT_TASK.md passes",
			} {
				if !strings.Contains(agents, marker) {
					t.Errorf("AGENTS.md missing efficient-verification marker %q", marker)
				}
			}
			tester := readGeneratedFile(t, root, tc.testerPath)
			for _, marker := range []string{
				"TEST_EXECUTION_PLAN",
				"Independently run targeted existing tests",
				"Do not automatically duplicate every expensive suite",
				"Never unilaterally shorten test deadlines",
				"who actually ran them",
			} {
				if !strings.Contains(tester, marker) {
					t.Errorf("%s missing efficiency marker %q", tc.testerPath, marker)
				}
			}
			reviewer := readGeneratedFile(t, root, tc.reviewPath)
			if !strings.Contains(reviewer, "equivalence to the required gate") {
				t.Errorf("%s missing optimized-runner equivalence check", tc.reviewPath)
			}
		})
	}
}

func TestCreateGeneratesExecutorEvidenceContract(t *testing.T) {
	root, err := Create(Config{Name: "demo", Architect: "chatgpt", Executor: "opencode", ParentDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	agents := readGeneratedFile(t, root, "AGENTS.md")
	for _, marker := range []string{
		"task_id: TASK-XXXX",
		"result_commit",
		"idempotent",
	} {
		if !strings.Contains(agents, marker) {
			t.Errorf("AGENTS.md missing executor evidence contract marker %q", marker)
		}
	}

	report := readGeneratedFile(t, root, ".factory/IMPLEMENTATION_REPORT.md")
	for _, marker := range []string{"task_id:", "revision:", "status:", "result_commit:"} {
		if !strings.Contains(report, marker) {
			t.Errorf("implementation report placeholder missing frontmatter field %q", marker)
		}
	}
}

func TestCreateGeneratesExecutorSpecificSubagents(t *testing.T) {
	tests := []struct {
		name       string
		executor   string
		expected   []string
		absent     []string
		testerPath string
		marker     string
	}{
		{
			name:       "opencode",
			executor:   "opencode",
			expected:   []string{".opencode/agents/tester.md", ".opencode/agents/reviewer.md"},
			absent:     []string{".claude/agents/tester.md", ".codex/agents/tester.toml"},
			testerPath: ".opencode/agents/tester.md",
			marker:     "mode: subagent",
		},
		{
			name:       "claude-code",
			executor:   "claude-code",
			expected:   []string{".claude/agents/tester.md", ".claude/agents/reviewer.md"},
			absent:     []string{".opencode/agents/tester.md", ".codex/agents/tester.toml"},
			testerPath: ".claude/agents/tester.md",
			marker:     "name: tester",
		},
		{
			name:       "codex",
			executor:   "codex",
			expected:   []string{".codex/agents/tester.toml", ".codex/agents/reviewer.toml"},
			absent:     []string{".opencode/agents/tester.md", ".claude/agents/tester.md"},
			testerPath: ".codex/agents/tester.toml",
			marker:     "developer_instructions = \"\"\"",
		},
		{
			name:     "none",
			executor: "none",
			absent: []string{
				".opencode/agents/tester.md",
				".claude/agents/tester.md",
				".codex/agents/tester.toml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Create(Config{
				Name:      "demo",
				Architect: "chatgpt",
				Executor:  tt.executor,
				ParentDir: t.TempDir(),
			})
			if err != nil {
				t.Fatal(err)
			}

			for _, path := range tt.expected {
				if _, err := os.Stat(filepath.Join(root, path)); err != nil {
					t.Errorf("expected %s for executor %s: %v", path, tt.executor, err)
				}
			}
			for _, path := range tt.absent {
				if _, err := os.Stat(filepath.Join(root, path)); !os.IsNotExist(err) {
					t.Errorf("did not expect %s for executor %s", path, tt.executor)
				}
			}

			if tt.testerPath != "" {
				content := readGeneratedFile(t, root, tt.testerPath)
				if !strings.Contains(content, tt.marker) {
					t.Errorf("%s missing executor-specific marker %q", tt.testerPath, tt.marker)
				}
			}

			state := readGeneratedFile(t, root, ".factory/FACTORY_STATE.md")
			if !strings.Contains(state, "executor: "+tt.executor) {
				t.Errorf("factory state does not record executor %q", tt.executor)
			}
		})
	}
}

func TestCreateDefaultsExecutorToOpenCode(t *testing.T) {
	root, err := Create(Config{Name: "demo", Architect: "chatgpt", ParentDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, ".opencode/agents/tester.md")); err != nil {
		t.Fatalf("expected OpenCode tester for default executor: %v", err)
	}
	state := readGeneratedFile(t, root, ".factory/FACTORY_STATE.md")
	if !strings.Contains(state, "executor: opencode") {
		t.Fatalf("expected normalized default executor in factory state, got:\n%s", state)
	}
}

func TestCreateRejectsUnsupportedExecutor(t *testing.T) {
	_, err := Create(Config{
		Name:      "demo",
		Executor:  "unknown-agent",
		ParentDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected unsupported executor to fail")
	}
	if !strings.Contains(err.Error(), "unsupported executor") {
		t.Fatalf("unexpected error: %v", err)
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
