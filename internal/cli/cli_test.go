package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frosk1/PrincipalFoundry/docs/prompts"
)

func TestPromptArchitectLoopRendersRepositoryURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{
		"prompt",
		"architect-loop",
		"--repo",
		"https://github.com/frosk1/Principalfoundry",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	got := stdout.String()
	if strings.Contains(got, prompts.RepositoryURLPlaceholder) {
		t.Fatal("rendered prompt still contains repository placeholder")
	}
	if !strings.Contains(got, "GitHub repository: https://github.com/frosk1/Principalfoundry") {
		t.Fatal("rendered prompt missing repository URL")
	}
	if !strings.HasPrefix(got, "# Architect Development Loop Prompt") {
		t.Fatalf("stdout should contain only the prompt, got %q", got[:min(len(got), 80)])
	}
}

func TestPromptArchitectLoopRequiresRepositoryURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"prompt", "architect-loop"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected missing --repo to fail")
	}
	if !strings.Contains(err.Error(), "--repo") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromptArchitectLoopRejectsNonGitHubRepository(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{
		"prompt",
		"architect-loop",
		"--repo",
		"https://gitlab.com/example/project",
	}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected non-GitHub repository to fail")
	}
}

func runGitCLI(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func initStatusRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitCLI(t, dir, "init", "-b", "main")
	runGitCLI(t, dir, "config", "user.email", "test@example.com")
	runGitCLI(t, dir, "config", "user.name", "Test")
	if err := os.MkdirAll(filepath.Join(dir, ".factory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".factory/FACTORY_STATE.md"), []byte("---\nphase: SCAFFOLDING\narchitect: chatgpt\nexecutor: opencode\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".factory/CURRENT_TASK.md"), []byte("---\nstatus: READY\ntask_id: TASK-0001\nrevision: 1\n---\n\n# TASK-0001\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitCLI(t, dir, "add", "-A")
	runGitCLI(t, dir, "commit", "-m", "init")
	return dir
}

func TestExecutorStatusJSONIsParseable(t *testing.T) {
	dir := initStatusRepo(t)

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"executor", "status", "--json", "--dir", dir}, &stdout, &stderr); err != nil {
		t.Fatalf("status failed: %v (%s)", err, stderr.String())
	}

	var parsed map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("status output is not valid JSON: %v\n%s", err, stdout.String())
	}
	if parsed["state"] != "PHASE_BLOCKED" {
		t.Fatalf("state = %v, want PHASE_BLOCKED", parsed["state"])
	}
}

func TestExecutorStatusHumanOutput(t *testing.T) {
	dir := initStatusRepo(t)

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"executor", "status", "--dir", dir}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	out := stdout.String()
	for _, want := range []string{"Factory: SCAFFOLDING", "Task: TASK-0001 r1", "State: PHASE_BLOCKED"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
