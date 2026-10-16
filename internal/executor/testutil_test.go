package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitIn runs git in dir and fails the test on error.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

type fixture struct {
	t      *testing.T
	dir    string
	remote string
}

// newFixture creates a work tree with an "origin" bare remote and a clean
// initial commit already pushed.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	gitIn(t, base, "init", "--bare", "-b", "main", remote)

	dir := filepath.Join(base, "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "init", "-b", "main")

	f := &fixture{t: t, dir: dir, remote: remote}
	f.git("config", "user.email", "test@example.com")
	f.git("config", "user.name", "Test")
	f.git("config", "commit.gpgsign", "false")
	f.git("remote", "add", "origin", remote)
	return f
}

func (f *fixture) git(args ...string) string {
	f.t.Helper()
	return gitIn(f.t, f.dir, args...)
}

func (f *fixture) write(rel, content string) {
	f.t.Helper()
	full := filepath.Join(f.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) commit(message string) {
	f.t.Helper()
	f.git("add", "-A")
	f.git("commit", "-m", message)
}

// setupAutonomous writes an autonomous factory state with a READY task and
// pushes it to the remote.
func (f *fixture) setupAutonomous(taskID string, revision int) {
	f.t.Helper()
	f.write(".factory/FACTORY_STATE.md", "---\nphase: AUTONOMOUS\narchitect: chatgpt\nexecutor: opencode\n---\n\n# Factory State\n")
	f.write(".factory/CURRENT_TASK.md", fmt.Sprintf("---\ntask_id: %s\nrevision: %d\nstatus: READY\n---\n\n# %s\n\nDo the bounded thing.\n", taskID, revision, taskID))
	f.write(".factory/IMPLEMENTATION_REPORT.md", "# Implementation Report\n\nNo implementation cycle has run yet.\n")
	f.commit("architect: issue " + taskID)
	f.git("push", "-u", "origin", "main")
}

// setEvidence overwrites the implementation report with matching frontmatter.
func (f *fixture) setEvidence(taskID string, revision int, status string, commit string) {
	f.t.Helper()
	f.write(".factory/IMPLEMENTATION_REPORT.md", fmt.Sprintf("---\ntask_id: %s\nrevision: %d\nstatus: %s\nresult_commit: %s\n---\n\n# Implementation Report\n\nDone.\n", taskID, revision, status, commit))
}

// cloneRepo clones remote into a fresh directory with committer identity set.
func cloneRepo(t *testing.T, remote string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "clone")
	gitIn(t, "", "clone", remote, dir)
	gitIn(t, dir, "config", "user.email", "other@example.com")
	gitIn(t, dir, "config", "user.name", "Other")
	gitIn(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func commits(t *testing.T, dir string) string {
	t.Helper()
	return gitIn(t, dir, "rev-list", "--all", "--count")
}
