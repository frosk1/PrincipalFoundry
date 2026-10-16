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

// Status inspects the repository containing dir and returns the effective
// executor state. It does not access the network or mutate the repository.
func Status(dir, remote string) (StatusReport, error) {
	if strings.TrimSpace(remote) == "" {
		remote = defaultRemote
	}
	root, err := Root(dir)
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

	snap, err := Inspect(root, remote, Held(gitDir), gitErr)
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
