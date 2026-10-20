package executor

import "strings"

// StatusReport is a stable, scriptable view of the effective executor state.
type StatusReport struct {
	Repository   string `json:"repository"`
	Branch       string `json:"branch"`
	Phase        string `json:"phase"`
	TaskID       string `json:"task_id"`
	Revision     int    `json:"revision"`
	State        State  `json:"state"`
	Reason       string `json:"reason"`
	LastResult   string `json:"last_result"`
	ResultCommit string `json:"result_commit"`
}

// StatusOptions configures a status inspection.
type StatusOptions struct {
	Dir    string
	Remote string
	// IgnoreLock reports the repository state as if no executor lock were held.
	// Callers that already hold the lock use this to avoid self-reporting
	// ALREADY_RUNNING.
	IgnoreLock bool
}

// Status inspects the repository containing dir and returns the effective
// executor state. It does not access the network or mutate the repository.
func Status(dir, remote string) (StatusReport, error) {
	return StatusWithOptions(StatusOptions{Dir: dir, Remote: remote})
}

// StatusWithOptions inspects the repository and returns the effective executor
// state, honoring the given options. It does not access the network or mutate
// the repository.
func StatusWithOptions(opts StatusOptions) (StatusReport, error) {
	remote := opts.Remote
	if strings.TrimSpace(remote) == "" {
		remote = defaultRemote
	}
	root, err := Root(opts.Dir)
	if err != nil {
		return StatusReport{}, err
	}
	gitDir, err := GitDir(root)
	if err != nil {
		return StatusReport{}, err
	}

	gitErr := ""
	if clean, err := IsClean(root); err != nil {
		gitErr = err.Error()
	} else if !clean {
		gitErr = "worktree has uncommitted changes"
	}

	lockHeld := false
	if !opts.IgnoreLock {
		lockHeld = Held(gitDir)
	}

	snap, err := Inspect(root, remote, lockHeld, gitErr)
	if err != nil {
		return StatusReport{}, err
	}

	lastResult := "NONE"
	if snap.Evidence.Present {
		lastResult = snap.Evidence.Status
	}

	return StatusReport{
		Repository:   snap.RepoURL,
		Branch:       snap.Branch,
		Phase:        snap.Factory.Phase,
		TaskID:       snap.Task.ID,
		Revision:     snap.Task.Revision,
		State:        snap.State,
		Reason:       snap.Reason,
		LastResult:   lastResult,
		ResultCommit: snap.Evidence.ResultCommit,
	}, nil
}
