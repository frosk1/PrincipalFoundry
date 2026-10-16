package executor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/frosk1/PrincipalFoundry/internal/gitrepo"
)

// defaultRemote is the Git remote used for synchronization unless overridden.
const defaultRemote = "origin"

// Root returns the absolute path of the repository containing dir.
func Root(dir string) (string, error) {
	return gitrepo.Output(dir, "rev-parse", "--show-toplevel")
}

// GitDir returns the absolute git directory for the repository containing dir.
func GitDir(dir string) (string, error) {
	return gitrepo.Output(dir, "rev-parse", "--absolute-git-dir")
}

// Branch returns the current branch name for dir.
func Branch(dir string) (string, error) {
	return gitrepo.Output(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// IsClean reports whether the worktree has no uncommitted changes, including
// untracked files.
func IsClean(dir string) (bool, error) {
	out, err := gitrepo.Output(dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// RemoteURL returns the configured URL of a remote, normalized to a canonical
// HTTPS form when it is a GitHub SSH URL.
func RemoteURL(dir, remote string) (string, error) {
	out, err := gitrepo.Output(dir, "remote", "get-url", remote)
	if err != nil {
		return "", err
	}
	return NormalizeRemoteURL(out), nil
}

// NormalizeRemoteURL converts common Git remote URL forms into a canonical
// HTTPS URL, e.g. "git@github.com:owner/repo.git" -> "https://github.com/owner/repo".
func NormalizeRemoteURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, "git@") {
		hostPath := strings.TrimPrefix(url, "git@")
		if host, path, ok := strings.Cut(hostPath, ":"); ok {
			url = "https://" + host + "/" + path
		}
	} else if rest, ok := strings.CutPrefix(url, "ssh://git@"); ok {
		url = "https://" + rest
	}
	return strings.TrimSuffix(url, ".git")
}

// Fetch updates remote-tracking refs for the given remote.
func Fetch(dir, remote string) error {
	return gitrepo.Run(dir, "fetch", "--prune", remote)
}

// FastForward synchronizes the current branch with the remote using fast-forward
// only semantics. It refuses to reconcile divergent or unrelated histories and
// never resets or rebases local work.
func FastForward(dir, remote, branch string) error {
	remoteRef := remote + "/" + branch

	// Confirm the remote-tracking ref exists before comparing.
	if _, err := gitrepo.Output(dir, "rev-parse", "--verify", "--quiet", remoteRef+"^{commit}"); err != nil {
		return fmt.Errorf("remote branch %s not found; cannot synchronize safely", remoteRef)
	}

	out, err := gitrepo.Output(dir, "rev-list", "--left-right", "--count", "HEAD..."+remoteRef)
	if err != nil {
		return err
	}
	ahead, behind, err := parseAheadBehind(out)
	if err != nil {
		return err
	}

	switch {
	case ahead > 0 && behind > 0:
		return fmt.Errorf("local branch and %s have diverged (ahead %d, behind %d); refusing destructive reconciliation", remoteRef, ahead, behind)
	case behind > 0:
		if err := gitrepo.Run(dir, "merge", "--ff-only", remoteRef); err != nil {
			return fmt.Errorf("fast-forward to %s failed: %w", remoteRef, err)
		}
	}
	return nil
}

func parseAheadBehind(out string) (int, int, error) {
	fields := strings.Fields(out)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q", out)
	}
	ahead, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q", out)
	}
	behind, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected rev-list output %q", out)
	}
	return ahead, behind, nil
}

// HeadCommit returns the current HEAD commit hash.
func HeadCommit(dir string) (string, error) {
	return gitrepo.Output(dir, "rev-parse", "--short", "HEAD")
}
