package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Outcome is the concise result of one executor cycle.
type Outcome string

const (
	OutcomeNoAction       Outcome = "NO_ACTION"
	OutcomeExecuted       Outcome = "EXECUTED"
	OutcomeBlockedHuman   Outcome = "BLOCKED_HUMAN"
	OutcomeAlreadyRunning Outcome = "ALREADY_RUNNING"
	OutcomeError          Outcome = "ERROR"
)

// Snapshot is the deterministic repository view used for state decisions.
type Snapshot struct {
	RepoURL  string
	Branch   string
	Factory  FactoryState
	Task     Task
	Evidence Evidence
	LockHeld bool
	GitError string
	State    State
	Reason   string
}

// Inspect gathers the factory control state for root and decides the effective
// executor state. It performs no network access and does not mutate anything.
func Inspect(root, remote string, lockHeld bool, gitErr string) (Snapshot, error) {
	snap := Snapshot{LockHeld: lockHeld, GitError: gitErr}

	factory, err := ParseFactoryState(root)
	if err != nil {
		return snap, fmt.Errorf("cannot read %s: %w", factoryStatePath, err)
	}
	snap.Factory = factory

	task, err := ParseTask(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return snap, fmt.Errorf("cannot read %s: %w", currentTaskPath, err)
	}
	snap.Task = task

	evidence, err := ParseEvidence(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return snap, fmt.Errorf("cannot read %s: %w", implementationReportPath, err)
	}
	snap.Evidence = evidence

	if url, err := RemoteURL(root, remote); err == nil {
		snap.RepoURL = url
	}
	if branch, err := Branch(root); err == nil {
		snap.Branch = branch
	}

	facts := Facts{
		LockHeld:        lockHeld,
		GitError:        gitErr,
		Phase:           factory.Phase,
		TaskPresent:     task.Present,
		TaskID:          task.ID,
		TaskStatus:      task.Status,
		Revision:        task.Revision,
		EvidenceMatches: evidence.Matches(task),
		EvidenceStatus:  evidence.Status,
	}
	snap.State, snap.Reason = Decide(facts)
	return snap, nil
}

// TickConfig configures a single executor cycle.
type TickConfig struct {
	Dir      string
	Remote   string
	Executor string
	Attach   string
	Model    string
	Auto     bool
	// LockWait bounds how long a cycle waits for a concurrent lock holder.
	// Zero uses the default (5 seconds).
	LockWait time.Duration
	Runner   Runner
	Stdout   io.Writer
	Stderr   io.Writer
	Verbose  bool
}

// TickResult is the concise result of one executor cycle.
type TickResult struct {
	Outcome      Outcome
	State        State
	TaskID       string
	Revision     int
	ResultCommit string
	Message      string
}

// Tick performs one executor cycle. Repeated invocations are safe and
// idempotent: with no executable task it returns NO_ACTION without invoking the
// executor and without mutating the repository. An ERROR outcome is also
// returned as a non-nil error so callers can surface failure clearly.
func Tick(ctx context.Context, cfg TickConfig) (TickResult, error) {
	stdout := cfg.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	runner := cfg.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	remote := cfg.Remote
	if strings.TrimSpace(remote) == "" {
		remote = defaultRemote
	}

	root, err := Root(cfg.Dir)
	if err != nil {
		return fail(err)
	}
	gitDir, err := GitDir(root)
	if err != nil {
		return fail(err)
	}

	// Wait briefly for a short-lived holder (for example an architect gate
	// checking state) so a timer collision does not starve the executor.
	lockWait := cfg.LockWait
	if lockWait <= 0 {
		lockWait = 5 * time.Second
	}
	lock, err := AcquireWait(gitDir, lockWait)
	if errors.Is(err, ErrAlreadyRunning) {
		return TickResult{Outcome: OutcomeAlreadyRunning, State: StateAlreadyRunning, Message: "an executor cycle is already running"}, nil
	}
	if err != nil {
		return fail(err)
	}
	defer lock.Release()

	// Require a clean worktree before synchronizing or executing.
	if clean, err := IsClean(root); err != nil {
		return fail(err)
	} else if !clean {
		return fail(errors.New("worktree has uncommitted changes; refusing to execute"))
	}

	// Synchronize conservatively: fetch and fast-forward only.
	if err := Fetch(root, remote); err != nil {
		return fail(fmt.Errorf("fetch %s failed: %w", remote, err))
	}
	branch, err := Branch(root)
	if err != nil {
		return fail(err)
	}
	if err := FastForward(root, remote, branch); err != nil {
		return fail(err)
	}
	if clean, err := IsClean(root); err != nil {
		return fail(err)
	} else if !clean {
		return fail(errors.New("worktree changed during synchronization"))
	}

	snap, err := Inspect(root, remote, false, "")
	if err != nil {
		return fail(err)
	}

	switch snap.State {
	case StateReadyToExecute:
		// Execution is permitted; continue below.
	case StateError:
		return fail(errors.New(snap.Reason))
	case StateBlockedHuman:
		return tickResult(OutcomeBlockedHuman, snap), nil
	default:
		return tickResult(OutcomeNoAction, snap), nil
	}

	executor := resolveExecutor(cfg.Executor, snap.Factory)
	command, err := executorCommand(executor, cfg.Attach, cfg.Model, cfg.Auto, executionPrompt(snap.RepoURL, snap.Task))
	if err != nil {
		return fail(err)
	}
	if cfg.Verbose {
		fmt.Fprintf(stdout, "executor command: %s\n", strings.Join(command, " "))
	}

	if err := runner.Run(ctx, root, command, stdout, stderr); err != nil {
		return fail(err)
	}

	// Git and factory evidence remain authoritative; a zero exit code is not proof.
	if clean, err := IsClean(root); err != nil {
		return fail(err)
	} else if !clean {
		return fail(errors.New("executor left uncommitted changes; refusing to report success"))
	}

	post, err := Inspect(root, remote, false, "")
	if err != nil {
		return fail(err)
	}
	if post.Evidence.Matches(post.Task) && post.Evidence.Status == evidenceCompleted {
		result := tickResult(OutcomeExecuted, post)
		result.ResultCommit = post.Evidence.ResultCommit
		result.Message = "completed " + taskRefFromTask(post.Task)
		return result, nil
	}
	return fail(fmt.Errorf("executor did not produce a committed handoff for %s", taskRefFromTask(snap.Task)))
}

func tickResult(outcome Outcome, snap Snapshot) TickResult {
	return TickResult{
		Outcome:  outcome,
		State:    snap.State,
		TaskID:   snap.Task.ID,
		Revision: snap.Task.Revision,
		Message:  snap.Reason,
	}
}

func fail(err error) (TickResult, error) {
	return TickResult{Outcome: OutcomeError, State: StateError, Message: err.Error()}, err
}

func taskRefFromTask(t Task) string {
	id := t.ID
	if id == "" {
		id = "task"
	}
	if t.Revision > 0 {
		return fmt.Sprintf("%s r%d", id, t.Revision)
	}
	return id
}

func resolveExecutor(flag string, factory FactoryState) string {
	if v := strings.ToLower(strings.TrimSpace(flag)); v != "" {
		return v
	}
	if v := strings.ToLower(strings.TrimSpace(factory.Executor)); v != "" {
		return v
	}
	return "opencode"
}
