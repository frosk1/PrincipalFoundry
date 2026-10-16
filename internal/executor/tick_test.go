package executor

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func noopRunner(calls *int) Runner {
	return RunnerFunc(func(context.Context, string, []string, io.Writer, io.Writer) error {
		(*calls)++
		return nil
	})
}

func TestTickExecutesOnceThenWaits(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	calls := 0
	runner := RunnerFunc(func(context.Context, string, []string, io.Writer, io.Writer) error {
		calls++
		f.setEvidence("TASK-0001", 1, evidenceCompleted, "abc1234")
		f.commit("executor: complete TASK-0001 r1")
		f.git("push", "origin", "main")
		return nil
	})

	first, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: runner, Stdout: io.Discard})
	if err != nil {
		t.Fatalf("first tick failed: %v", err)
	}
	if first.Outcome != OutcomeExecuted {
		t.Fatalf("first outcome = %s (%s)", first.Outcome, first.Message)
	}
	if calls != 1 {
		t.Fatalf("expected exactly one executor invocation, got %d", calls)
	}

	second, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err != nil {
		t.Fatalf("second tick failed: %v", err)
	}
	if second.Outcome != OutcomeNoAction {
		t.Fatalf("second outcome = %s (%s)", second.Outcome, second.Message)
	}
	if second.State != StateWaitingArchitect {
		t.Fatalf("second state = %s", second.State)
	}
	if calls != 1 {
		t.Fatalf("executor was invoked again for the same task revision: %d", calls)
	}
}

func TestTickDoesNotExecuteWhenPhaseNotAutonomous(t *testing.T) {
	f := newFixture(t)
	f.write(".factory/FACTORY_STATE.md", "---\nphase: SCAFFOLDING\narchitect: chatgpt\nexecutor: opencode\n---\n")
	f.write(".factory/CURRENT_TASK.md", "---\ntask_id: TASK-0001\nrevision: 1\nstatus: READY\n---\n\n# TASK-0001\n")
	f.write(".factory/IMPLEMENTATION_REPORT.md", "# Implementation Report\n")
	f.commit("scaffolding")
	f.git("push", "-u", "origin", "main")

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != OutcomeNoAction || result.State != StatePhaseBlocked {
		t.Fatalf("got outcome=%s state=%s", result.Outcome, result.State)
	}
	if calls != 0 {
		t.Fatalf("executor invoked %d times in a blocked phase", calls)
	}
}

func TestTickDoesNotExecutePausedFactory(t *testing.T) {
	f := newFixture(t)
	f.write(".factory/FACTORY_STATE.md", "---\nphase: PAUSED\narchitect: chatgpt\nexecutor: opencode\n---\n")
	f.write(".factory/CURRENT_TASK.md", "---\ntask_id: TASK-0001\nrevision: 1\nstatus: READY\n---\n\n# TASK-0001\n")
	f.write(".factory/IMPLEMENTATION_REPORT.md", "# Implementation Report\n")
	f.commit("paused")
	f.git("push", "-u", "origin", "main")

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != OutcomeNoAction || result.State != StatePaused {
		t.Fatalf("got outcome=%s state=%s", result.Outcome, result.State)
	}
	if calls != 0 {
		t.Fatalf("executor invoked %d times in a paused factory", calls)
	}
}

func TestTickRefusesDirtyWorktree(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)
	if err := os.WriteFile(filepath.Join(f.dir, "uncommitted.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err == nil {
		t.Fatal("expected a dirty worktree to produce an error")
	}
	if result.Outcome != OutcomeError {
		t.Fatalf("outcome = %s", result.Outcome)
	}
	if calls != 0 {
		t.Fatalf("executor invoked %d times with a dirty worktree", calls)
	}
}

func TestTickRefusesDivergentHistory(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	other := cloneRepo(t, f.remote)
	if err := os.WriteFile(filepath.Join(other, "remote.txt"), []byte("remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, other, "add", "-A")
	gitIn(t, other, "commit", "-m", "remote change")
	gitIn(t, other, "push", "origin", "main")

	f.write("local.txt", "local\n")
	f.commit("local change")

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err == nil {
		t.Fatal("expected divergent history to produce an error")
	}
	if result.Outcome != OutcomeError {
		t.Fatalf("outcome = %s", result.Outcome)
	}
	if calls != 0 {
		t.Fatalf("executor invoked %d times on divergent history", calls)
	}
}

func TestTickReportsAlreadyRunning(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	gitDir, err := GitDir(f.dir)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := Acquire(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != OutcomeAlreadyRunning || result.State != StateAlreadyRunning {
		t.Fatalf("got outcome=%s state=%s", result.Outcome, result.State)
	}
	if calls != 0 {
		t.Fatalf("executor invoked %d times while the lock was held", calls)
	}
}

func TestTickReportsExecutorFailure(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	runner := RunnerFunc(func(context.Context, string, []string, io.Writer, io.Writer) error {
		return errors.New("boom")
	})
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: runner, Stdout: io.Discard})
	if err == nil {
		t.Fatal("expected a failing executor to produce an error")
	}
	if result.Outcome != OutcomeError {
		t.Fatalf("outcome = %s", result.Outcome)
	}
}

func TestTickReportsErrorWhenExecutorProducesNoHandoff(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err == nil {
		t.Fatal("expected a no-op executor to be reported as an error")
	}
	if result.Outcome != OutcomeError {
		t.Fatalf("outcome = %s", result.Outcome)
	}
}

func TestTickNoWorkProducesNoChanges(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)
	f.setEvidence("TASK-0001", 1, evidenceCompleted, "abc1234")
	f.commit("executor: complete TASK-0001 r1")
	f.git("push", "origin", "main")

	before := f.git("rev-parse", "HEAD")
	statusBefore := f.git("status", "--porcelain")

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Outcome != OutcomeNoAction || result.State != StateWaitingArchitect {
		t.Fatalf("got outcome=%s state=%s", result.Outcome, result.State)
	}
	if calls != 0 {
		t.Fatalf("executor invoked %d times with no executable work", calls)
	}
	if after := f.git("rev-parse", "HEAD"); after != before {
		t.Fatalf("HEAD changed on a no-op tick: %s -> %s", before, after)
	}
	if afterStatus := f.git("status", "--porcelain"); afterStatus != statusBefore {
		t.Fatalf("worktree changed on a no-op tick: %q -> %q", statusBefore, afterStatus)
	}
}

func TestTickHigherRevisionIsExecutableAgain(t *testing.T) {
	f := newFixture(t)
	f.setupAutonomous("TASK-0001", 1)
	f.setEvidence("TASK-0001", 1, evidenceCompleted, "abc1234")
	f.commit("executor: complete TASK-0001 r1")

	// Architect requests a new revision of the same task.
	f.write(".factory/CURRENT_TASK.md", "---\ntask_id: TASK-0001\nrevision: 2\nstatus: READY\n---\n\n# TASK-0001 r2\n")
	f.commit("architect: request TASK-0001 r2")
	f.git("push", "origin", "main")

	calls := 0
	result, err := Tick(context.Background(), TickConfig{Dir: f.dir, Runner: noopRunner(&calls), Stdout: io.Discard})
	// The no-op runner produces no handoff, so the cycle reports ERROR, but the
	// important assertion is that execution was attempted for the new revision.
	if err == nil {
		t.Fatal("expected an error because the no-op executor produced no handoff")
	}
	if calls != 1 {
		t.Fatalf("expected the new revision to be executable, calls = %d", calls)
	}
	if result.Outcome != OutcomeError {
		t.Fatalf("outcome = %s", result.Outcome)
	}
}
