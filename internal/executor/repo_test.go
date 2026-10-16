package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeRemoteURL(t *testing.T) {
	cases := map[string]string{
		"git@github.com:frosk1/dockspur.git":       "https://github.com/frosk1/dockspur",
		"https://github.com/frosk1/dockspur.git":   "https://github.com/frosk1/dockspur",
		"https://github.com/frosk1/dockspur":       "https://github.com/frosk1/dockspur",
		"ssh://git@github.com/frosk1/dockspur.git": "https://github.com/frosk1/dockspur",
	}
	for input, want := range cases {
		if got := NormalizeRemoteURL(input); got != want {
			t.Errorf("NormalizeRemoteURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsCleanAndBranch(t *testing.T) {
	f := newFixture(t)
	f.write("a.txt", "one\n")
	f.commit("base")

	clean, err := IsClean(f.dir)
	if err != nil {
		t.Fatal(err)
	}
	if !clean {
		t.Fatal("expected a clean worktree")
	}

	branch, err := Branch(f.dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}

	f.write("b.txt", "two\n")
	clean, err = IsClean(f.dir)
	if err != nil {
		t.Fatal(err)
	}
	if clean {
		t.Fatal("expected a dirty worktree after adding an untracked file")
	}
}

func TestFastForwardAppliesRemoteCommits(t *testing.T) {
	f := newFixture(t)
	f.write("a.txt", "one\n")
	f.commit("base")
	f.git("push", "-u", "origin", "main")

	other := cloneRepo(t, f.remote)
	if err := os.WriteFile(filepath.Join(other, "a.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, other, "add", "-A")
	gitIn(t, other, "commit", "-m", "remote change")
	gitIn(t, other, "push", "origin", "main")

	if err := Fetch(f.dir, "origin"); err != nil {
		t.Fatal(err)
	}
	if err := FastForward(f.dir, "origin", "main"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(f.dir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "two" {
		t.Fatalf("expected fast-forwarded content, got %q", data)
	}
}

func TestFastForwardRejectsDivergence(t *testing.T) {
	f := newFixture(t)
	f.write("a.txt", "one\n")
	f.commit("base")
	f.git("push", "-u", "origin", "main")

	other := cloneRepo(t, f.remote)
	if err := os.WriteFile(filepath.Join(other, "a.txt"), []byte("remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, other, "add", "-A")
	gitIn(t, other, "commit", "-m", "remote change")
	gitIn(t, other, "push", "origin", "main")

	f.write("a.txt", "local\n")
	f.commit("local change")

	if err := Fetch(f.dir, "origin"); err != nil {
		t.Fatal(err)
	}
	err := FastForward(f.dir, "origin", "main")
	if err == nil {
		t.Fatal("expected divergent history to be rejected")
	}
	if !strings.Contains(err.Error(), "diverged") {
		t.Fatalf("unexpected error: %v", err)
	}
}
