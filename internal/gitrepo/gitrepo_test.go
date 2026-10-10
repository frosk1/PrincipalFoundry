package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesMainBranchAndBootstrapCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Init(root); err != nil {
		t.Fatal(err)
	}

	branch := gitOutput(t, root, "symbolic-ref", "--short", "HEAD")
	if branch != "main" {
		t.Fatalf("expected main branch, got %q", branch)
	}

	subject := gitOutput(t, root, "log", "-1", "--pretty=%s")
	if subject != initialCommitMessage {
		t.Fatalf("expected bootstrap commit %q, got %q", initialCommitMessage, subject)
	}

	status := gitOutput(t, root, "status", "--porcelain")
	if status != "" {
		t.Fatalf("expected clean repository, got:\n%s", status)
	}
}

func TestRemoteInstructions(t *testing.T) {
	instructions := RemoteInstructions("/tmp/my project")
	for _, expected := range []string{
		"git remote add origin REMOTE_URL",
		"git push -u origin main",
		"create an empty remote repository",
	} {
		if !strings.Contains(instructions, expected) {
			t.Errorf("remote instructions missing %q", expected)
		}
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %s", strings.Join(args, " "), output)
	}
	return strings.TrimSpace(string(output))
}
